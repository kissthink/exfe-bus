package main

import (
	"broker"
	"fmt"
	"github.com/googollee/go-broadcast"
	"github.com/googollee/go-rest"
	"here"
	"math/rand"
	"model"
	"net/http"
	"strings"
	"time"
)

type LiveService struct {
	rest.Service `prefix:"/v3/live"`

	Card      rest.Processor `path:"/cards" method:"POST"`
	Streaming rest.Streaming `path:"/streaming" method:"POST" end:""`

	platform  *broker.Platform
	config    *model.Config
	here      *here.Here
	rand      *rand.Rand
	tokens    map[string]bool
	broadcast map[string]*broadcast.Broadcast
}

func NewLive(config *model.Config) (http.Handler, error) {
	platform, err := broker.NewPlatform(config)
	if err != nil {
		return nil, err
	}
	service := &LiveService{
		config:    config,
		here:      here.New(config.Here.Threshold, config.Here.SignThreshold, time.Duration(config.Here.TimeoutInSecond)*time.Second),
		rand:      rand.New(rand.NewSource(time.Now().Unix())),
		platform:  platform,
		broadcast: make(map[string]*broadcast.Broadcast),
	}

	go service.here.Serve()

	go func() {
		c := service.here.UpdateChannel()
		for {
			group := <-c
			cards := make([]here.Card, 0)
			if group.Name != "" {
				for _, d := range group.Data {
					cards = append(cards, d.Card)
				}
			}
			for token := range group.Data {
				if b, ok := service.broadcast[token]; ok {
					b.Send(cards)
				}
			}
		}
	}()

	return rest.New(service)
}

func (h LiveService) HandleCard(data here.Data) []string {
	h.Header().Set("Access-Control-Allow-Origin", h.config.AccessDomain)
	h.Header().Set("Access-Control-Allow-Credentials", "true")
	h.Header().Set("Cache-Control", "no-cache")

	token := h.Request().URL.Query().Get("token")
	if token == "" {
		token = fmt.Sprintf("%04d", rand.Int31n(10000))
		if h.here.Exist(token) {
			h.Error(http.StatusNotFound, fmt.Errorf("please wait and try again."))
			return nil
		}
		data.Card.Id = fmt.Sprintf("%032d", rand.Int31())
	} else if !h.here.Exist(token) {
		h.Error(http.StatusForbidden, fmt.Errorf("invalid token"))
		return nil
	}
	data.Token = token
	remote := h.Request().RemoteAddr
	remotes := strings.Split(remote, ":")
	data.Traits = append(data.Traits, remotes[0])

	if data.Card.Avatar == "" {
		ids, err := h.platform.GetIdentity(data.Card.Identities)
		if err == nil {
			for _, id := range ids {
				if strings.Index(id.Avatar, "/v2/avatar/default?name=") < 0 {
					data.Card.Avatar = id.Avatar
					break
				}
			}
			h.config.Log.Debug("token %s can't find avatar", data.Token)
		} else {
			h.config.Log.Debug("get avatar failed: %s", err)
		}
	}

	err := h.here.Add(&data)
	h.config.Log.Info("|live|add|t|%s|card|%s|name|%s|long|%s|lang|%s|acc|%s|trait|%s", data.Token, data.Card.Id, data.Card.Name, data.Longitude, data.Latitude, data.Accuracy, data.Traits)

	if err != nil {
		h.Error(http.StatusBadRequest, err)
		return nil
	}

	return []string{token, data.Card.Id}
}

func (h LiveService) HandleStreaming(s rest.Stream) {
	h.Header().Set("Access-Control-Allow-Origin", h.config.AccessDomain)
	h.Header().Set("Access-Control-Allow-Credentials", "true")
	h.Header().Set("Cache-Control", "no-cache")
	token := h.Request().URL.Query().Get("token")
	if !h.here.Exist(token) {
		h.Error(http.StatusForbidden, fmt.Errorf("invalid token"))
		return
	}
	c := make(chan interface{})
	b, ok := h.broadcast[token]
	if !ok {
		b = broadcast.NewBroadcast()
		h.broadcast[token] = b
	}
	b.Register(c)
	defer func() {
		b.Unregister(c)
		close(c)
	}()

	for {
		d := <-c
		cards, ok := d.([]here.Card)
		if !ok {
			continue
		}
		err := s.Write(cards)
		if err != nil || len(cards) == 0 {
			h.config.Log.Info("|live|clear|t|%s|card||name||long||lang||acc||trait|", token)
			return
		}
	}
}