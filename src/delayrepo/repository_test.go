package delayrepo

import (
	"broker"
	"fmt"
	"github.com/stretchrcom/testify/assert"
	"model"
	"testing"
	"time"
)

type RepoTester struct {
	*Repository
	t *testing.T
}

func (r *RepoTester) Do(key string, data [][]byte) {
	assert.Equal(r.t, key, "123")
	assert.Equal(r.t, fmt.Sprintf("%v", data), "[[97] [98] [99]]")
}

func (r *RepoTester) OnError(err error) {
	r.t.Fatalf("%s", err)
}

func TestRepository(t *testing.T) {
	config := new(model.Config)
	config.Redis.MaxConnections = 1
	config.Redis.Netaddr = "127.0.0.1:6379"
	redis, err := broker.NewRedisPool(config)
	if err != nil {
		t.Fatal(err)
	}

	repo := new(RepoTester)
	timer, err := NewTimer(Always, "delay:test", redis)
	if err != nil {
		t.Fatal(err)
	}
	repo.Repository = New(timer, repo, time.Second)
	repo.t = t
	go repo.Serve()

	ontime := time.Now().Add(time.Second).Unix()

	err = repo.Push(ontime, "123", []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Push(ontime, "123", []byte("b"))
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Push(ontime, "123", []byte("c"))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	repo.Quit()
}