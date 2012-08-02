package exfe_service

import (
	"bytes"
	"email/service"
	"exfe/model"
	"fmt"
	"github.com/googollee/godis"
	"gobus"
	"log"
	"net/mail"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"
)

type CrossEmail struct {
	log    *log.Logger
	queue  *gobus.TailDelayQueue
	config *Config
	client *gobus.Client
	tmpl   *template.Template
}

var helper = template.FuncMap{
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},
	"limit": func(s string, max int) string {
		if max > len(s) {
			max = len(s)
		}
		return s[0:max]
	},
	"replace": func(s, old, new string) string {
		return strings.Replace(s, old, new, -1)
	},
}

func NewCrossEmail(config *Config) *CrossEmail {
	provider := "email"
	log := log.New(os.Stderr, fmt.Sprintf("exfe.cross.%s", provider), log.LstdFlags)

	arg := []OneIdentityUpdateArg{}
	redis := godis.New(config.Redis.Netaddr, config.Redis.Db, config.Redis.Password)
	queue, err := gobus.NewTailDelayQueue(getProviderQueueName(provider), config.Cross.Delay[provider], arg, redis)
	if err != nil {
		panic(err)
	}

	t := template.Must(template.New("invitation").Funcs(helper).ParseFiles("./template/default/cross_invitation.html", "./template/default/cross_update.html", "./template/default/cross.ics"))

	return &CrossEmail{
		log:    log,
		queue:  queue,
		config: config,
		client: gobus.CreateClient(config.Redis.Netaddr, config.Redis.Db, config.Redis.Password, provider),
		tmpl:   t,
	}
}

func (e *CrossEmail) Serve() {
	for {
		t, err := e.queue.NextWakeup()
		if err != nil {
			e.log.Printf("next wakeup error: %s", err)
			break
		}
		time.Sleep(t)
		args, err := e.queue.Pop()
		if err != nil {
			e.log.Printf("pop from delay queue failed: %s", err)
			continue
		}
		if args != nil {
			updates := args.([]OneIdentityUpdateArg)

			by_identities := make([]*exfe_model.Identity, 0, 0)
			posts := make([]*exfe_model.Post, 0, 0)
			var old_cross *exfe_model.Cross
			for _, update := range updates {
				by_identities = append(by_identities, &update.By_identity)
				if old_cross == nil && update.Old_cross != nil {
					old_cross = update.Old_cross
				}
				if update.Post != nil {
					posts = append(posts, update.Post)
				}
			}

			arg := &ProviderArg{
				Cross:         &updates[len(updates)-1].Cross,
				Old_cross:     old_cross,
				To_identity:   &updates[0].To_identity,
				By_identities: by_identities,
				Posts:         posts,
				Config:        e.config,
			}

			e.sendMail(arg)
		}
	}
}

func (e *CrossEmail) GetBody(arg *ProviderArg, filename string) (string, string, error) {
	html := bytes.NewBuffer(nil)
	err := e.tmpl.ExecuteTemplate(html, filename, arg)
	if err != nil {
		return "", "", err
	}

	ics := bytes.NewBuffer(nil)
	err = e.tmpl.ExecuteTemplate(ics, "cross.ics", arg)
	if err != nil {
		return "", "", err
	}

	var output []byte
	line_count := 0
	for _, c := range ics.Bytes() {
		line_count++
		output = append(output, c)
		if line_count == 70 {
			output = append(output, 0xd, 0xa, 0x20)
			line_count = 1
			continue
		}
		if c == 0xa {
			output = append(output, 0xd, 0xa)
			line_count = 0
			continue
		}
	}

	return html.String(), string(output), nil
}

func (e *CrossEmail) sendMail(arg *ProviderArg) {
	filename := "cross_invitation.html"
	if arg.Old_cross != nil {
		arg.Diff(e.log)
		filename = "cross_update.html"
	}

	html, ics, err := e.GetBody(arg, filename)
	if err != nil {
		e.log.Printf("template exec error:", err)
		return
	}
	htmls := strings.SplitN(html, "\n\n", 2)

	mail_addr := fmt.Sprintf("x+%d@exfe.com", arg.Cross.Id)
	mailarg := &email_service.MailArg{
		To:      []*mail.Address{&mail.Address{arg.To_identity.External_id, arg.To_identity.Name}},
		From:    &mail.Address{mail_addr, mail_addr},
		Subject: htmls[0],
		Html:    htmls[1],
		FileParts: []email_service.FilePart{
			email_service.FilePart{fmt.Sprintf("x-%d.ics", arg.Cross.Id), []byte(ics)},
		},
	}

	e.client.Send("EmailSend", &mailarg, 5)
}
