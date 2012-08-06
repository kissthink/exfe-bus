package exfe_service

import (
	"bytes"
	"encoding/json"
	"exfe/model"
	"log"
	"os"
	"testing"
)

var exfee = exfe_model.Exfee{
	Invitations: []exfe_model.Invitation{
		exfe_model.Invitation{
			Rsvp_status: "ACCEPTED",
			Identity: exfe_model.Identity{
				Id:                1,
				Name:              "Tester1",
				Connected_user_id: 1,
			},
		},
		exfe_model.Invitation{
			Rsvp_status: "ACCEPTED",
			Identity: exfe_model.Identity{
				Id:                2,
				Name:              "Tester2",
				Connected_user_id: 2,
			},
		},
		exfe_model.Invitation{
			Rsvp_status: "ACCEPTED",
			Identity: exfe_model.Identity{
				Id:                3,
				Name:              "Tester3",
				Connected_user_id: 3,
			},
		},
		exfe_model.Invitation{
			Rsvp_status: "ACCEPTED",
			Identity: exfe_model.Identity{
				Id:                4,
				Name:              "Tester4",
				Connected_user_id: 4,
			},
		},
	},
}

func TestProvideArgDiff(t *testing.T) {
	log := log.New(os.Stderr, "test", log.LstdFlags)
	var new_, old exfe_model.Exfee

	buf := bytes.NewBuffer(nil)
	e := json.NewEncoder(buf)
	e.Encode(exfee)

	buf1 := bytes.NewBufferString(buf.String())
	d1 := json.NewDecoder(buf1)
	d1.Decode(&new_)
	buf2 := bytes.NewBufferString(buf.String())
	d2 := json.NewDecoder(buf2)
	d2.Decode(&old)

	new_.Invitations[0].Identity.Connected_user_id = 5
	old.Invitations[1].Identity.Connected_user_id = 6
	old.Invitations[2].Rsvp_status = "DECLINED"
	new_.Invitations[3].Rsvp_status = "DECLINED"

	cross_old := &exfe_model.Cross{
		Exfee: old,
	}
	cross_new := &exfe_model.Cross{
		Exfee: new_,
	}
	arg := &ProviderArg{
		Cross:     cross_new,
		Old_cross: cross_old,
	}

	accepted, declined, newlyInvited, removed := arg.Diff(log)

	if expect, got := 3, len(accepted); expect != got {
		t.Errorf("expect: %d, got: %d", expect, got)
	}
	for _, id := range []uint64{2, 3, 5} {
		if _, ok := accepted[id]; !ok {
			t.Errorf("accepted should have id %d", id)
		}
	}
	if expect, got := 1, len(declined); expect != got {
		t.Errorf("expect: %d, got: %d", expect, got)
	}
	for _, id := range []uint64{4} {
		if _, ok := declined[id]; !ok {
			t.Errorf("accepted should have id %d", id)
		}
	}
	if expect, got := 2, len(newlyInvited); expect != got {
		t.Errorf("expect: %d, got: %d", expect, got)
	}
	for _, id := range []uint64{2, 5} {
		if _, ok := newlyInvited[id]; !ok {
			t.Errorf("accepted should have id %d", id)
		}
	}
	if expect, got := 2, len(removed); expect != got {
		t.Errorf("expect: %d, got: %d", expect, got)
	}
	for _, id := range []uint64{1, 6} {
		if _, ok := removed[id]; !ok {
			t.Errorf("accepted should have id %d", id)
		}
	}
}