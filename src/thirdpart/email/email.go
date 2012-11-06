package email

import (
	"model"
	"thirdpart"
)

type Email struct {
	helper thirdpart.Helper
}

func New(helper thirdpart.Helper) *Email {
	return &Email{
		helper: helper,
	}
}

func (e *Email) Provider() string {
	return "email"
}

func (e *Email) Send(to *model.Recipient, privateMessage string, publicMessage string, info *model.InfoData) (string, error) {
	return e.helper.SendEmail(to.ExternalID, privateMessage)
}
