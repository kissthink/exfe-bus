package main

import (
	"encoding/json"
	"github.com/googollee/go-rest"
	"github.com/gorilla/mux"
	"here"
	"model"
	"net"
	"net/http"
	"sync"
	"time"
)

type HereService struct {
	rest.Service `prefix:"/v3/here"`

	Users rest.Processor `path:"/users" method:"POST"`

	here *here.Here
}

func (h HereService) Users_(user here.User) {
	h.here.Add(user)
}

type HereStreaming struct {
	locker sync.Mutex
	ids    map[string][]chan string
}

func (h *HereStreaming) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token invalid", http.StatusBadRequest)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "doesn't support streaming", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c := make(chan string)
	h.locker.Lock()
	h.ids[token] = append(h.ids[token], c)
	h.locker.Unlock()
	defer func() {
		close(c)
		conn.Close()
		h.locker.Lock()
		defer h.locker.Unlock()

		var chs []chan string
		for _, ch := range h.ids[token] {
			if ch == c {
				continue
			}
			chs = append(chs, ch)
		}
		if len(chs) > 0 {
			h.ids[token] = chs
		} else {
			delete(h.ids, token)
		}
	}()

	for {
		select {
		case <-time.After(time.Second):
			conn.SetDeadline(time.Now().Add(time.Second / 10))
			buf := make([]byte, 10)
			_, err := conn.Read(buf)
			if e, ok := err.(net.Error); err == nil || (ok && e.Timeout()) {
				continue
			}
			return
		case data := <-c:
			_, err = bufrw.Write([]byte(data + "\n"))
			if err != nil {
				return
			}
			err = bufrw.Flush()
			if err != nil {
				return
			}
		}
	}
}

func NewHere(config *model.Config) (http.Handler, error) {
	ret := mux.NewRouter()
	service := new(HereService)
	service.here = here.New(config.Here.Threshold, config.Here.SignThreshold, time.Duration(config.Here.TimeoutInSecond)*time.Second)
	handler, err := rest.New(service)
	if err != nil {
		return nil, err
	}
	streaming := &HereStreaming{
		ids: make(map[string][]chan string),
	}
	go func() {
		update := service.here.UpdateChannel()
		for {
			select {
			case id := <-update:
				group := service.here.UserInGroup(id)
				if group == nil {
					group = here.NewGroup()
				}
				buf, _ := json.Marshal(group.Users)
				data := string(buf)
				streaming.locker.Lock()
				for _, s := range streaming.ids[id] {
					s <- data
				}
				streaming.locker.Unlock()
			}
		}
	}()
	ret.Path("/v3/here/streaming").Handler(streaming)
	ret.PathPrefix(handler.Prefix()).Handler(handler)
	return ret, nil
}