package gobus

import (
	"fmt"
	"github.com/googollee/go-logger"
	"testing"
)

type gobusTest struct {
}

type AddArgs struct {
	A int
	B int
}

func (t *gobusTest) Add(meta *HTTPMeta, args AddArgs, reply *int) error {
	*reply = args.A + args.B
	return nil
}

const gobusUrl = "http://127.0.0.1:12345"

func TestGobus(t *testing.T) {
	l, err := logger.New(logger.Stderr, "test gobus", logger.LstdFlags)
	if err != nil {
		panic(err)
	}
	s, err := NewServer(gobusUrl, l)
	if err != nil {
		t.Fatalf("create gobus server fail: %s", err)
	}
	test := new(gobusTest)
	s.Register(test)
	go s.ListenAndServe()

	c, err := NewClient(fmt.Sprintf("%s/%s", gobusUrl, "gobusTest"))
	if err != nil {
		t.Fatalf("create gobus client fail: %s", err)
	}
	var reply int
	err = c.Do("Add", &AddArgs{2, 4}, &reply)
	if err != nil {
		t.Fatalf("call Add error: %s", err)
	}
	if expect, got := 6, reply; got != expect {
		t.Error("expect: %d, got: %d", expect, got)
	}
}
