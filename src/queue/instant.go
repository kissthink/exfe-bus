package main

import (
	"fmt"
	"gobus"
	"model"
)

type Instant struct {
	services map[string]*gobus.Client
}

func NewInstant(services map[string]*gobus.Client) *Instant {
	return &Instant{
		services: services,
	}
}

// 即时发送队列
//
// 例子：
//
// > curl 'http://127.0.0.1:23334/Instant?method=Push' -d '{"service":"Conversation",
//       "method":"Update",
//       "merge_key":"email_cross123",
//       "tos":[{"identity_id":12,"user_id":3,"name":"email1 name","auth_data":"","timezone":"+0800","token":"recipient_email1_token","language":"en_US","provider":"email","external_id":"sender1@gmail.com","external_username":"sender1@gmail.com"},{"identity_id":12,"user_id":3,"name":"email1 name","auth_data":"","timezone":"+0800","token":"recipient_email1_token","language":"en_US","provider":"email","external_id":"sender2@hotmail.com","external_username":"sender2@hotmail.com"}],
//       "data":{"to":{"identity_id":12,"user_id":3,"name":"email1 name","auth_data":"","timezone":"+0800","token":"recipient_email1_token","language":"en_US","provider":"email","external_id":"sender1@gmail.com","external_username":"sender1@gmail.com"},"cross":{"id":123,"by_identity":{"id":11,"name":"email1 name","nickname":"email1 nick","bio":"email1 bio","timezone":"+0800","connected_user_id":1,"avatar_filename":"http://path/to/email1.avatar","provider":"email","external_id":"sender2@hotmail.com","external_username":"sender2@hotmail.com"},"title":"Test Cross","description":"test cross description","time":{"begin_at":{"date_word":"","date":"","time_word":"","time":"","timezone":""},"origin":"","output_format":0},"place":{"id":0,"title":"","description":"","lng":"","lat":"","provider":"","external_id":""},"exfee":{"id":0,"name":"","invitations":null}},"post":{"id":1,"by_identity":{"id":11,"name":"email1 name","nickname":"email1 nick","bio":"email1 bio","timezone":"+0800","connected_user_id":1,"avatar_filename":"http://path/to/email1.avatar","provider":"email","external_id":"sender2@hotmail.com","external_username":"sender2@hotmail.com"},"content":"email1 post sth","via":"abc","created_at":"2012-10-24 16:31:00"}}}'
func (i *Instant) Push(meta *gobus.HTTPMeta, arg model.QueuePush, count *int) error {
	client, ok := i.services[arg.Service]
	if !ok {
		return fmt.Errorf("can't find service %s", arg.Service)
	}

	*count = len(arg.Tos)
	if *count == 0 {
		*count = 1
	}

	go func() {
		if len(arg.Tos) == 0 {
			var r int
			err := client.Do(arg.Method, []interface{}{arg.Data}, &r)
			if err != nil {
				meta.Log.Err("send to %s, method %s fail: %s", arg.Service, arg.Method, err)
				return
			}
		}

		data, ok := arg.Data.(map[string]interface{})
		for _, to := range arg.Tos {
			if ok {
				data["to"] = to
			}
			var r int
			err := client.Do(arg.Method, []interface{}{data}, &r)
			if err != nil {
				meta.Log.Err("send to %s, method %s fail: %s", arg.Service, arg.Method, err)
			}
		}
	}()

	return nil
}
