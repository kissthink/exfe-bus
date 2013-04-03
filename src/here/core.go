package here

import (
	"math"
	"strconv"
	"time"
)

type Identity struct {
	ExternalID       string `json:"external_id"`
	ExternalUsername string `json:"external_username"`
	Provider         string `json:"provider"`
}

type Card struct {
	Id         string     `json:"id"`
	Name       string     `json:"name"`
	Avatar     string     `json:"avatar"`
	Bio        string     `json:"bio"`
	IsMe       bool       `json:"is_me"`
	Identities []Identity `json:"identities"`
	Timestamp  int64      `json:"timestamp"`
}

type Data struct {
	Token     string   `json:"token"`
	Latitude  string   `json:"latitude"`
	Longitude string   `json:"longitude"`
	Accuracy  string   `json:"accuracy"`
	Traits    []string `json:"traits"`
	Card      Card     `json:"card"`

	latitude  float64 `json:"-"`
	longitude float64 `json:"-"`
	accuracy  float64 `json:"-"`

	UpdatedAt time.Time `json:"-"`
}

type Group struct {
	Name            string
	CenterLatitude  float64
	CenterLongitude float64
	Traits          map[string]int
	Data            map[string]*Data
}

func NewGroup() *Group {
	return &Group{
		Traits: make(map[string]int),
		Data:   make(map[string]*Data),
	}
}

func (g *Group) Add(data *Data) {
	data.UpdatedAt = time.Now()
	data.Card.Timestamp = time.Now().Unix()
	g.Data[data.Token] = data
	g.calcuate()
}

func (g *Group) Remove(data *Data) {
	delete(g.Data, data.Token)
	g.calcuate()
}

func (g *Group) Clear(limit time.Duration) []string {
	var remove []string
	for k, u := range g.Data {
		if u.UpdatedAt.Add(limit).Before(time.Now()) {
			remove = append(remove, k)
		}
	}
	for _, k := range remove {
		delete(g.Data, k)
	}
	g.calcuate()
	return remove
}

func (g *Group) Distant(u *Data) float64 {
	a := math.Cos(g.CenterLatitude) * math.Cos(u.latitude) * math.Cos(g.CenterLongitude-u.longitude)
	b := math.Sin(g.CenterLatitude) * math.Sin(u.latitude)
	return math.Acos(a + b)
}

func (g *Group) HasTraits(traits []string) bool {
	for _, t := range traits {
		if _, ok := g.Traits[t]; ok {
			return true
		}
	}
	return false
}

func (g *Group) calcuate() {
	g.CenterLatitude, g.CenterLongitude, g.Traits = 0, 0, make(map[string]int)
	n := 0
	var userId string
	for k, u := range g.Data {
		if len(u.Traits) == 0 {
			a := u.accuracy
			coeff := float64(n) * a
			g.CenterLatitude = (coeff*g.CenterLatitude + u.latitude) / (coeff + 1)
			g.CenterLongitude = (coeff*g.CenterLongitude + u.longitude) / (coeff + 1)
			n += 1
		}
		for _, t := range u.Traits {
			g.Traits[t] += 1
		}
		userId = k
	}
	if u, ok := g.Data[userId]; ok && (g.CenterLatitude == 0 || g.CenterLongitude == 0) {
		g.CenterLatitude, g.CenterLongitude = u.latitude, u.longitude
	}
}

type Cluster struct {
	Groups    map[string]*Group
	UserGroup map[string]string

	distantThreshold float64
	signThreshold    float64
	timeout          time.Duration
}

func NewCluster(threshold, signThreshold float64, timeout time.Duration) *Cluster {
	return &Cluster{
		Groups:           make(map[string]*Group),
		UserGroup:        make(map[string]string),
		distantThreshold: threshold,
		signThreshold:    signThreshold,
		timeout:          timeout,
	}
}

func (c *Cluster) AddUser(data *Data) error {
	var err error
	if data.Latitude != "" && data.Longitude != "" && data.Accuracy != "" {
		data.latitude, err = strconv.ParseFloat(data.Latitude, 64)
		if err != nil {
			return err
		}
		data.longitude, err = strconv.ParseFloat(data.Longitude, 64)
		if err != nil {
			return err
		}
		data.accuracy, err = strconv.ParseFloat(data.Accuracy, 64)
		if err != nil {
			return err
		}
	} else {
		data.Latitude = ""
		data.Longitude = ""
		data.Accuracy = ""
	}
	groupKey := ""
	var distant float64 = -1
	for k, group := range c.Groups {
		d := group.Distant(data)
		if len(data.Traits) > 0 && d < c.signThreshold && group.HasTraits(data.Traits) {
			groupKey, distant = k, 0
		}
		if distant < 0 || d < distant {
			groupKey, distant = k, d
		}
	}
	var group *Group
	if groupKey != "" && distant < c.distantThreshold {
		group = c.Groups[groupKey]
	} else {
		group = NewGroup()
		group.Name = data.Token
		groupKey = data.Token
	}
	group.Add(data)
	c.Groups[groupKey] = group
	c.UserGroup[data.Token] = groupKey
	return nil
}

func (c *Cluster) Clear() []string {
	var remove []string
	var ret []string
	for k, group := range c.Groups {
		ret = append(ret, group.Clear(c.timeout)...)
		if len(group.Data) == 0 {
			remove = append(remove, k)
		} else {
			c.Groups[k] = group
		}
	}
	for _, r := range remove {
		delete(c.Groups, r)
	}
	return ret
}
