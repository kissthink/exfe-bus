package notifier

import (
	"broker"
	"fmt"
	"formatter"
	"model"
)

type Conversation struct {
	localTemplate *formatter.LocalTemplate
	config        *model.Config
	platform      *broker.Platform
}

func NewConversation(localTemplate *formatter.LocalTemplate, config *model.Config, platform *broker.Platform) *Conversation {
	return &Conversation{
		localTemplate: localTemplate,
		config:        config,
		platform:      platform,
	}
}

func (c *Conversation) Update(updates model.ConversationUpdates) error {
	if len(updates) == 0 {
		return fmt.Errorf("len(updates) == 0")
	}

	to := updates[0].To
	if to.Provider == "twitter" {
		c.config.Log.Debug("not send to twitter: %s", to)
		return nil
	}
	selfUpdates := true
	for _, update := range updates {
		if !to.SameUser(&update.Post.By) {
			selfUpdates = false
		}
	}
	if selfUpdates {
		c.config.Log.Debug("not send with all self updates: %s", to)
		return nil
	}

	text, err := c.getConversationContent(updates)
	if err != nil {
		return err
	}

	_, err = c.platform.Send(to, text)

	if err != nil {
		return fmt.Errorf("send error: %s", err)
	}
	return nil
}

func (c *Conversation) getConversationContent(updates []model.ConversationUpdate) (string, error) {
	arg, err := ArgFromUpdates(updates, c.config)
	if err != nil {
		return "", err
	}

	content, err := GetContent(c.localTemplate, "conversation", arg.To, arg)
	if err != nil {
		return "", err
	}

	return content, nil
}

type UpdateArg struct {
	model.ThirdpartTo
	Cross model.Cross
	Posts []*model.Post
}

func (a UpdateArg) Link() string {
	return fmt.Sprintf("%s/#!token=%s", a.Config.SiteUrl, a.To.Token)
}

func ArgFromUpdates(updates []model.ConversationUpdate, config *model.Config) (*UpdateArg, error) {
	if updates == nil && len(updates) == 0 {
		return nil, fmt.Errorf("no update info")
	}

	to := updates[0].To
	cross := updates[0].Cross
	posts := make([]*model.Post, len(updates))

	for i, update := range updates {
		if !to.Equal(&update.To) {
			return nil, fmt.Errorf("updates not send to same recipient: %s, %s", to, update.To)
		}
		if !cross.Equal(&update.Cross) {
			return nil, fmt.Errorf("updates not send to same exfee: %d, %d", cross.ID, update.Cross.ID)
		}
		posts[i] = &updates[i].Post
	}

	ret := &UpdateArg{
		Cross: cross,
		Posts: posts,
	}
	ret.To = to
	err := ret.Parse(config)
	if err != nil {
		return nil, nil
	}

	return ret, nil
}

func (a UpdateArg) Timezone() string {
	if a.To.Timezone != "" {
		return a.To.Timezone
	}
	return a.Cross.Time.BeginAt.Timezone
}
