package oauth

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

func TestBaseCreate(t *testing.T) {
	base := OAuthBase{
		ClientToken: oauth.Credentials{
			Token:  "123",
			Secret: "abc",
		},
	}

	if base.client != nil {
		t.Error("OAuthBase.HttpClient default value should be nil")
	}

	base.init()
	if base.client == nil {
		t.Error("OAuthBase.HttpClient should not be nil after token()")
	}
}

func TestClientLoad(t *testing.T) {
	f, err := os.Open("oauth_client.json")
	if err != nil {
		fmt.Println("Please run helper to generate oauth_client.json")
		t.Error("Open oauth_client.json error: ", err)
	}
	_, err = LoadClientFromJson(f)
	if err != nil {
		fmt.Println("Please run helper to generate oauth_client.json")
		t.Error(err)
	}
}

func TestClient(t *testing.T) {
	f, _ := os.Open("oauth_client.json")
	client, err := LoadClientFromJson(f)
	if err != nil {
		fmt.Println("Please run helper to generate oauth_client.json")
		t.Error(err)
	}

	params := make(url.Values)
	params.Add("include_entities", "true")
	request, err := client.GetRequest("GET", "statuses/home_timeline.json", params)
	if err != nil {
		t.Error(err)
	}
	ret, err := client.SendRequest(request)
	if err != nil {
		t.Error(err)
	}
	decoder := json.NewDecoder(ret)
	var buf1 []interface{}
	err = decoder.Decode(&buf1)
	if err != nil {
		t.Error(err)
	}

	params = make(url.Values)
	params.Add("status", "test测试")
	ret, err = client.Do("POST", "statuses/update.json", params)
	if err != nil {
		p, _ := ioutil.ReadAll(ret)
		fmt.Println(string(p))
		t.Error(err)
	}
	decoder = json.NewDecoder(ret)
	var buf2 map[string]interface{}
	err = decoder.Decode(&buf2)
	if err != nil {
		t.Error(err)
	}

	ret, err = client.Do("POST", "statuses/destroy/"+buf2["id_str"].(string)+".json", nil)
	if err != nil {
		t.Error(err)
	}
	decoder = json.NewDecoder(ret)
	err = decoder.Decode(&buf2)
	if err != nil {
		t.Error(err)
	}
}