package dropbox

import (
	"broker"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/googollee/go-aws/s3"
	"github.com/mrjones/oauth"
	"io"
	"logger"
	"model"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Dropbox struct {
	oauth  broker.OAuth
	bucket *s3.Bucket
}

func New(config *model.Config) (*Dropbox, error) {
	provider := oauth.ServiceProvider{
		RequestTokenUrl:   "https://api.dropbox.com/1/oauth/request_token",
		AuthorizeTokenUrl: "https://www.dropbox.com/1/oauth/authorize",
		AccessTokenUrl:    "https://api.dropbox.com/1/oauth/access_token",
	}
	aws := s3.New(config.AWS.S3.Domain, config.AWS.S3.Key, config.AWS.S3.Secret)
	aws.SetACL(s3.ACLPublicRead)
	aws.SetLocationConstraint(s3.LC_AP_SINGAPORE)
	bucket, err := aws.GetBucket(fmt.Sprintf("%s-3rdpart-photos", config.AWS.S3.BucketPrefix))
	if err != nil {
		return nil, err
	}
	return &Dropbox{
		oauth:  broker.NewOAuth(config.Thirdpart.Dropbox.Key, config.Thirdpart.Dropbox.Secret, provider),
		bucket: bucket,
	}, nil
}

func (d *Dropbox) Provider() string {
	return "dropbox"
}

func (d *Dropbox) Grab(to model.Recipient, albumID string) ([]model.Photo, error) {
	var data model.OAuthToken
	err := json.Unmarshal([]byte(to.AuthData), &data)
	if err != nil {
		return nil, fmt.Errorf("can't parse %s auth data(%s): %s", to, to.AuthData, err)
	}
	token := oauth.AccessToken{
		Token:  data.Token,
		Secret: data.Secret,
	}
	path := escapePath(albumID)
	path = fmt.Sprintf("https://api.dropbox.com/1/metadata/dropbox%s", path)
	resp, err := broker.HttpResponse(d.oauth.Get(path, nil, &token))
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	decoder := json.NewDecoder(resp)
	var list folderList
	err = decoder.Decode(&list)
	if err != nil {
		return nil, err
	}
	ret := make([]model.Photo, 0)
	for _, c := range list.Contents {
		if !c.ThumbExists {
			logger.DEBUG("%s %s is not picture.", to, c.Path)
			continue
		}
		caption := c.Path[strings.LastIndex(c.Path, "/"):]
		modified, err := time.Parse(time.RFC1123Z, c.Modified)
		if err != nil {
			modified = time.Now()
		}
		photo := model.Photo{
			Caption: caption,
			By: model.Identity{
				ID: to.IdentityID,
			},
			CreatedAt:       modified.Format("2006-01-02 15:04:05"),
			UpdatedAt:       modified.Format("2006-01-02 15:04:05"),
			Provider:        "dropbox",
			ExternalAlbumID: albumID,
			ExternalID:      c.Rev,
		}

		thumb, big, err := d.savePic(c, to, &token)
		if err != nil {
			logger.ERROR("%s %s can't save: %s", to, c.Path, err)
			continue
		}
		photo.Images.Preview.Url = thumb
		photo.Images.Preview.Height = 480
		photo.Images.Preview.Width = 640
		photo.Images.Fullsize.Url = big
		photo.Images.Fullsize.Height = 768
		photo.Images.Fullsize.Width = 1024
		ret = append(ret, photo)
	}
	return ret, nil
}

func (d *Dropbox) Get(to model.Recipient, pictureIDs []string) ([]string, error) {
	var data model.OAuthToken
	err := json.Unmarshal([]byte(to.AuthData), &data)
	if err != nil {
		return nil, fmt.Errorf("can't parse %s auth data(%s): %s", to, to.AuthData, err)
	}
	token := oauth.AccessToken{
		Token:  data.Token,
		Secret: data.Secret,
	}

	var ret []string
	for _, id := range pictureIDs {
		url := fmt.Sprintf("https://api-content.dropbox.com/1/thumbnails/dropbox%s", escapePath(id))
		resp, err := d.oauth.Get(url, map[string]string{"size": "l"}, &token)
		reader, err := broker.HttpResponse(resp, err)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		mime := resp.Header.Get("Content-Type")
		if spliter := strings.Index(mime, ";"); spliter > 0 {
			mime = mime[:spliter]
		}
		buf := bytes.NewBuffer(nil)
		encoder := base64.NewEncoder(base64.StdEncoding, buf)
		_, err = io.Copy(encoder, reader)
		if err != nil {
			return nil, err
		}
		uri := fmt.Sprintf("data:%s;base64,%s", mime, buf.String())
		ret = append(ret, uri)
	}
	return ret, nil
}

func (d *Dropbox) savePic(c content, to model.Recipient, token *oauth.AccessToken) (string, string, error) {
	path := fmt.Sprintf("https://api-content.dropbox.com/1/thumbnails/dropbox%s", escapePath(c.Path))
	thumbPath := fmt.Sprintf("/i%d/dropbox%s", to.IdentityID, getThumbName(c.Path))
	bigPath := fmt.Sprintf("/i%d/dropbox%s", to.IdentityID, c.Path)
	thumb, err := d.saveFile(path, "l", thumbPath, c.MimeType, token)
	if err != nil {
		return "", "", err
	}
	big, err := d.saveFile(path, "xl", bigPath, c.MimeType, token)
	if err != nil {
		return "", "", err
	}

	return thumb, big, nil
}

func (d *Dropbox) saveFile(from, size, to, mime string, token *oauth.AccessToken) (string, error) {
	resp, err := d.oauth.Get(from, map[string]string{"size": size}, token)
	reader, err := broker.HttpResponse(resp, err)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	length, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return "", err
	}

	object, err := d.bucket.CreateObject(to, mime)
	if err != nil {
		return "", err
	}
	object.SetDate(time.Now())
	err = object.SaveReader(reader, int64(length))
	if err != nil {
		return "", err
	}
	return object.URL(), nil
}

func getThumbName(path string) string {
	extIndex := strings.LastIndex(path, ".")
	pathIndex := strings.LastIndex(path, "/")
	thumbName := path
	if extIndex < pathIndex {
		return thumbName + "-thumb"
	}
	return fmt.Sprintf("%s-thumb%s", path[:extIndex], path[extIndex:])
}

func escapePath(path string) string {
	path = url.QueryEscape(path)
	path = strings.Replace(path, "%2F", "/", -1)
	path = strings.Replace(path, "+", "%20", -1)
	return path
}

type content struct {
	Rev         string `json:"rev"`
	ThumbExists bool   `json:"thumb_exists"`
	Bytes       int    `json:"bytes"`
	Modified    string `json:"modified"`
	Path        string `json:"path"`
	IsDir       bool   `json:"is_dir"`
	Root        string `json:"root"`
	MimeType    string `json:"mime_type"`
}

type folderList struct {
	Contents []content `json:"contents"`
}
