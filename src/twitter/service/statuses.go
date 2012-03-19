package twitter_service

import (
	"fmt"
	"net/url"
	"log"
	"oauth"
	"encoding/json"
)

type StatusesUpdateArg struct {
	ClientToken  string
	ClientSecret string
	AccessToken  string
	AccessSecret string

	Tweet        string
}

func (t *StatusesUpdateArg) String() string {
	return fmt.Sprintf("{Client:(%s %s) Access:(%s %s) Tweet:%s}", t.ClientToken, t.ClientSecret, t.AccessToken, t.AccessSecret, t.Tweet)
}

type StatusesUpdateReply struct {
	User UserInfo
}

type StatusesUpdate struct {
}

func (t *StatusesUpdate) Do(arg *StatusesUpdateArg, reply *StatusesUpdateReply) error {
	log.Printf("[Info][statuses/update]Call by arg: %s", arg)

	client := oauth.CreateClient(arg.ClientToken, arg.ClientSecret, arg.AccessToken, arg.AccessSecret, "https://api.twitter.com/1/")
	params := make(url.Values)
	params.Add("status", arg.Tweet)

	retReader, err := client.Do("POST", "/statuses/update.json", params)
	if err != nil {
		log.Printf("[Error][statuses/update]Twitter access error: %s", err)
		return err
	}

	decoder := json.NewDecoder(retReader)
	err = decoder.Decode(reply)
	if err != nil {
		// some user will not fill all field, and twitter responses of these fields  are null,
		// which will cause decode error
		// log.Printf("[Error][statuses/update]Parse twitter response error: %s", err)
		// return err
	}

	return nil
}
