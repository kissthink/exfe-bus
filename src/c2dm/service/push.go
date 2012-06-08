package c2dm_service

import (
	"github.com/googollee/go_c2dm"
	"fmt"
	"log"
)

type C2DM struct {
	c2dm *go_c2dm.Client
}

func NewC2DM(email, password, appid string) (*C2DM, error) {
	client, err := go_c2dm.NewClient(email, password, appid)
	if err != nil {
		return nil, err
	}
	return &C2DM{
		c2dm: client,
	}, nil
}

type C2DMSendArg struct {
	DeviceID string
	Message string
	Badge uint
	Sound string
	Cid uint64
	T string
}

func (c *C2DM) C2DMSend(args []C2DMSendArg) error {
	for _, arg := range args {
		load := go_c2dm.NewLoad(arg.DeviceID, arg.Message)
		load.Add("badge", fmt.Sprintf("%d", arg.Badge))
		load.Add("sound", arg.Sound)
		load.Add("cid", fmt.Sprintf("%d", arg.Cid))
		load.Add("t", arg.T)
		load.DelayWhileIdle = true
		load.CollapseKey = 3

		_, err := c.c2dm.Send(load)
		if err != nil {
			log.Printf("Send notification(%s) to device(%s) error: %s", arg.Message, arg.DeviceID, err)
		}
	}
	return nil
}