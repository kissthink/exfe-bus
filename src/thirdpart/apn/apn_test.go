package apn

import (
	"fmt"
	"github.com/virushuo/Go-Apns"
	"model"
	"testing"
	"thirdpart"
)

type FakeBroker struct {
	errChan       chan apns.NotificationError
	notifications []*apns.Notification
}

func (b *FakeBroker) Reset() {
	b.notifications = make([]*apns.Notification, 0)
	b.errChan = make(chan apns.NotificationError)
}

func (b *FakeBroker) GetErrorChan() <-chan apns.NotificationError {
	return b.errChan
}

func (b *FakeBroker) Send(n *apns.Notification) error {
	b.notifications = append(b.notifications, n)
	return nil
}

func errHandler(err apns.NotificationError) {
	fmt.Println(err)
}

var to = &model.Recipient{
	ExternalID:       "54321",
	ExternalUsername: "to",
	AuthData:         "",
	Provider:         "iOS",
	IdentityID:       321,
	UserID:           2,
}

var data = &thirdpart.InfoData{
	CrossID: 12345,
	Type:    thirdpart.CrossUpdate,
}

func TestSend(t *testing.T) {
	broker := new(FakeBroker)
	apn, err := New(broker, errHandler)
	if err != nil {
		t.Fatalf("can't create apn: %s", err)
	}

	{
		broker.Reset()
		apn.Send(to, `\(AAAAAAAA name1\), \(AAAAAAAA name2\) and \(AAAAAAAA name3\) are accepted on \(“some cross”\), \(IIIII name1\), \(IIIII name2\) and \(IIIII name3\) interested, \(UUUU name1\), \(UUUU name2\) and \(UUUU name3\) are unavailable, \(PPPPPPP name1\), \(PPPPPPP name2\) and \(PPPPPPP name3\) are pending. \(3 of 10 accepted\). https://exfe.com/#!token=932ce5324321433253`, "public message", data)
		results := []string{
			`AAAAAAAA name1, AAAAAAAA name2 and AAAAAAAA name3 are accepted on “some cross”, IIIII name1, IIIII name2 and IIIII name3…(1/3)`,
			`interested, UUUU name1, UUUU name2 and UUUU name3 are unavailable, PPPPPPP name1, PPPPPPP name2 and PPPPPPP name3 are pending.…(2/3)`,
			`3 of 10 accepted. (3/3)`,
		}
		for i, r := range results {
			if got, expect := broker.notifications[i].Payload.Aps.Alert, r; got != expect {
				t.Errorf("%d got: %s, expect %s", i, got, expect)
			}
		}
	}
}
