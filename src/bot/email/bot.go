package email

import (
	"fmt"
	"formatter"
	"github.com/googollee/goimap"
	"github.com/sloonz/go-iconv"
	"gobot"
	"gobus"
	"model"
	"net/mail"
	"regexp"
	"strings"
)

var replyLine = [...]string{
	"^--$",
	"^--&nbsp;$",
	"-----Original Message-----",
	"________________________________",
	"Sent from my iPhone",
	"Sent from my BlackBerry",
	`^From:.*[mailto:.*]`,
	`^On (.*) wrote:`,
	"发自我的 iPhone",
	`EXFE ·X· <x\+[a-zA-Z0-9]*@exfe.com>$`,
	`^>+`,
}

var removeLine = [...]string{
	`(?iU)<script\b.*>.*</script>`,
	`(?iU)<style\b.*>.*</style>`,
	`(?U)<.*>`,
}

type EmailBot struct {
	config        *model.Config
	crossId       *regexp.Regexp
	retReplacer   *strings.Replacer
	remover       []*regexp.Regexp
	replyLine     []*regexp.Regexp
	localTemplate *formatter.LocalTemplate
	sender        *gobus.Client
}

func NewEmailBot(config *model.Config, localTemplate *formatter.LocalTemplate, sender *gobus.Client) *EmailBot {
	reply := make([]*regexp.Regexp, len(replyLine), len(replyLine))
	for i, l := range replyLine {
		reply[i] = regexp.MustCompile(l)
	}
	remover := make([]*regexp.Regexp, len(removeLine), len(removeLine))
	for i, l := range removeLine {
		remover[i] = regexp.MustCompile(l)
	}
	return &EmailBot{
		config:        config,
		crossId:       regexp.MustCompile(`^.*?\+([0-9]*)@.*$`),
		retReplacer:   strings.NewReplacer("\r\n", "\n", "\r", "\n"),
		remover:       remover,
		replyLine:     reply,
		localTemplate: localTemplate,
		sender:        sender,
	}
}

func (b *EmailBot) GenerateContext(id string) bot.Context {
	return NewEmailContext(id, b)
}

func (b *EmailBot) GetIDFromInput(input interface{}) (id string, content interface{}, e error) {
	msg, ok := input.(*mail.Message)
	if !ok {
		e = fmt.Errorf("input is not a mail.Message")
		return
	}
	from, err := imap.ParseAddress(msg.Header.Get("From"))
	if err != nil {
		e = fmt.Errorf("can't parse From: %s", err)
		return
	}
	to, err := imap.ParseAddress(msg.Header.Get("To"))
	if err != nil {
		e = fmt.Errorf("can't parse To: %s", err)
		return
	}
	date, err := msg.Header.Date()
	if err != nil {
		e = fmt.Errorf("Get message(%v) time error: id, err")
		return
	}
	text, mediatype, charset, err := imap.GetBody(msg, "text/plain")
	if err != nil {
		e = fmt.Errorf("Get message(%v) body failed: %s", id, err)
		return
	}
	text, err = iconv.Conv(text, "UTF-8", charset)
	if err != nil {
		e = fmt.Errorf("Convert message(%v) from %s to utf8 error: %s", id, charset, err)
		return
	}
	if mediatype != "text/plain" {
		text = b.stripGmail(text)
		text = b.stripHtml(text)
	}
	text = b.stripReply(text)

	id = from[0].Address
	content = &Email{
		From:      from[0],
		To:        to,
		Subject:   msg.Header.Get("Subject"),
		CrossID:   b.getCrossId(to),
		Date:      date,
		MessageID: msg.Header.Get("Message-Id"),
		Text:      text,
	}
	return
}

func (b *EmailBot) getCrossId(addrs []*mail.Address) string {
	for _, addr := range addrs {
		ids := b.crossId.FindStringSubmatch(addr.Address)
		if len(ids) > 1 {
			return ids[1]
		}
	}
	return ""
}

func (b *EmailBot) isReplys(line string) bool {
	for _, r := range b.replyLine {
		line = strings.Trim(line, " \t\n\r")
		if r.MatchString(line) {
			return true
		}
	}
	return false
}

func (b *EmailBot) stripReply(content string) string {
	content = b.retReplacer.Replace(content)
	lines := strings.Split(content, "\n")
	ret := make([]string, len(lines), len(lines))

	for i, line := range lines {
		if b.isReplys(line) {
			ret = ret[:i]
			break
		}
		ret[i] = line
	}
	return strings.Trim(strings.Join(ret, "\n"), " \r\n\t")
}

func (b *EmailBot) stripGmail(text string) string {
	pos := strings.Index(text, `<div class="gmail_quote"`)
	if pos >= 0 {
		return strings.Trim(text[:pos], " \t\n\r")
	}
	return text
}

func (b *EmailBot) stripHtml(text string) string {
	for _, r := range b.remover {
		text = r.ReplaceAllString(text, "")
	}
	return strings.Trim(text, " \t\n\r")
}
