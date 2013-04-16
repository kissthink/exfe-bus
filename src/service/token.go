package main

import (
	"broker"
	"fmt"
	"github.com/googollee/go-logger"
	"github.com/googollee/go-rest"
	"model"
	"net/http"
	"time"
	"token"
)

type Token struct {
	rest.Service `prefix:"/v3/tokens"`

	Create         rest.Processor `method:"POST" path:"/:type"`
	KeyGet         rest.Processor `method:"GET" path:"/key/:key"`
	ResourceGet    rest.Processor `method:"POST" path:"/resources"`
	KeyUpdate      rest.Processor `method:"POST" path:"/key/:key"`
	ResourceUpdate rest.Processor `method:"POST" path:"/resource"`

	log     *logger.SubLogger
	manager *token.Manager
}

func NewToken(config *model.Config, db *broker.DBMultiplexer) (*Token, error) {
	repo, err := NewTokenRepo(config, db)
	if err != nil {
		return nil, err
	}
	token := &Token{
		log:     config.Log.SubPrefix("tokens"),
		manager: token.New(repo),
	}
	return token, nil
}

type CreateArg struct {
	Data               string `json:"data"`
	Resource           string `json:"resource"`
	ExpireAfterSeconds int    `json:"expire_after_seconds"`
}

// 根据resource，data和expire after seconds创建一个token
//
// 例子：
//
//     > curl "http://127.0.0.1:23333/v3/tokens/long" -d '{"data":"abc","resource":"123","expire_after_seconds":300}'
//
// 返回：
//
//     {"key":"0303","data":"abc","touched_at":21341234,"expire_at":66354}
func (s Token) Create_(string, arg CreateArg) (ret model.Token) {
	genType := s.Vars()["type"]
	if genType != "long" || genType != "short" {
		s.Error(http.StatusNotFound, fmt.Errorf("invalid type %s", genType))
		return
	}
	after := time.Duration(arg.ExpireAfterSeconds) * time.Second
	ret, err := s.manager.Create(genType, arg.Resource, arg.Data, after)
	if err != nil {
		s.Error(http.StatusNotFound, err)
		return
	}
	return ret
}

// 根据key获得一个token，如果token不存在，返回错误
//
// 例子：
//
//     > curl "http://127.0.0.1:23333/v3/tokens/key/0303"
//
// 返回：
//
//     [{"key":"0303","data":"abc","touched_at":21341234,"expire_at":66354}]
func (s Token) KeyGet_() []model.Token {
	key := s.Vars()["key"]
	ret, err := s.manager.Get(key, "")
	if err != nil {
		s.Error(http.StatusNotFound, err)
		return nil
	}
	return ret
}

// 根据resource获得一个token，如果token不存在，返回错误
//
// 例子：
//
//     > curl "http://127.0.0.1:23333/v3/tokens/resources" -d '"abc"'
//
// 返回：
//
//     [{"key":"0303","data":"abc","touched_at":21341234,"expire_at":66354}]
func (s Token) ResourceGet_(resource string) []model.Token {
	ret, err := s.manager.Get("", resource)
	if err != nil {
		s.Error(http.StatusNotFound, err)
		return nil
	}
	return ret
}

type UpdateArg struct {
	Data               *string `json:"data"`
	ExpireAfterSeconds *int    `json:"expire_after_seconds"`
	Resource           string  `json:"resource"`
}

// 更新key对应的token的data信息或者expire after seconds
//
// 例子：
//
//     > curl "http://127.0.0.1:23333/v3/tokens/key/0303" -d '{"data":"xyz","expire_after_seconds":13}'
func (s Token) KeyUpdate_(arg UpdateArg) {
	key := s.Vars()["key"]
	if arg.Data != nil {
		err := s.manager.UpdateData(key, *arg.Data)
		if err != nil {
			s.Error(http.StatusBadRequest, err)
			return
		}
	}
	if arg.ExpireAfterSeconds != nil {
		after := time.Duration(*arg.ExpireAfterSeconds) * time.Second
		err := s.manager.Refresh(key, "", after)
		if err != nil {
			s.Error(http.StatusBadRequest, err)
			return
		}
	}
}

// 更新resource对应的token的expire after seconds
//
// 例子：
//
//     > curl "http://127.0.0.1:23333/v3/tokens/resource" -d '{"resource":"abc", "expire_after_seconds":13}'
func (s Token) ResourceUpdate_(arg UpdateArg) {
	if arg.Resource == "" {
		s.Error(http.StatusBadRequest, fmt.Errorf("invalid resource"))
		return
	}
	if arg.ExpireAfterSeconds != nil {
		after := time.Duration(*arg.ExpireAfterSeconds) * time.Second
		err := s.manager.Refresh("", arg.Resource, after)
		if err != nil {
			s.Error(http.StatusBadRequest, err)
			return
		}
	}
}
