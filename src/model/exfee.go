package model

type Exfee struct {
	ID          uint64       `json:"id"`
	Invitations []Invitation `json:"invitations"`

	Accepted   []Invitation `json:"-"`
	Declined   []Invitation `json:"-"`
	Interested []Invitation `json:"-"`
	Pending    []Invitation `json:"-"`
}

func (e Exfee) Parse() {
	e.Accepted = make([]Invitation, 0)
	e.Declined = make([]Invitation, 0)
	e.Interested = make([]Invitation, 0)
	e.Pending = make([]Invitation, 0)

	for _, i := range e.Invitations {
		switch i.RsvpStatus {
		case RspvAccepted:
			e.Accepted = append(e.Accepted, i)
		case RsvpDecliend:
			e.Declined = append(e.Declined, i)
		case RsvpInterested:
			e.Interested = append(e.Interested, i)
		case RsvpNoresponse:
			e.Pending = append(e.Pending, i)
		}
	}
}