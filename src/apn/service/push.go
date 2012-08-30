package apn_service

import (
	"github.com/virushuo/Go-Apns"
	"log"
)

type Apn struct {
	cert   string
	key    string
	server string
	apn    *goapns.Apn
	id     uint32
}

func NewApn(cert, key, server, rootca string) (*Apn, error) {
	apn, err := goapns.Connect(cert, key, server)
	if err != nil {
		return nil, err
	}
	ret := &Apn{
		cert:   cert,
		key:    key,
		server: server,
		apn:    apn,
		id:     0,
	}
	go errorListen(ret)
	return ret, nil
}

func errorListen(apn *Apn) {
	for {
		apnerr := <-apn.apn.Errorchan
		log.Printf("Apn error: cmd %d, status %d, id %d", apnerr.Command, apnerr.Status, apnerr.Identifier)
		panic("apn error")
	}
}

type ExfePush struct {
	Cid uint64 `json:"cid"`
	T   string `json:"t"`
}

type ApnSendArg struct {
	DeviceToken string
	Alert       string
	Badge       uint
	Sound       string
	Cid         uint64
	T           string
}

func (a *Apn) ApnSend(args []ApnSendArg) error {
	context := goapns.Context{}
	for _, arg := range args {
		context.Aps.Alert = arg.Alert
		context.Aps.Badge = arg.Badge
		context.Aps.Sound = arg.Sound
		context.Arg = ExfePush{
			Cid: arg.Cid,
			T:   arg.T,
		}
		notification := goapns.Notification{
			DeviceToken: arg.DeviceToken,
			Identifier:  a.id,
			Context:     &context,
		}
		a.id++
		err := a.apn.SendNotification(&notification)
		if err != nil {
			log.Printf("Send notification(%s) to device(%s) error: %s", arg.Alert, arg.DeviceToken, err)
			return err
		}
	}
	return nil
}
