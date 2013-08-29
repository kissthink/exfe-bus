package routex

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"logger"
	"model"
	"net/http"
	"sort"
	"strconv"
	"time"
)

func (m *RouteMap) getTutorialData(currentTime time.Time, userId int64, number int) []SimpleLocation {
	data, ok := m.tutorialDatas[userId]
	if !ok {
		return nil
	}
	currentTime = currentTime.UTC()
	now := currentTime.Unix()
	todayTime, _ := time.Parse("2006-01-02 15:04:05", currentTime.Format("2006-01-02 00:00:00"))
	today := todayTime.Unix()
	offset := (now - today) / 10 * 10

	oneDaySeconds := int64(24 * time.Hour / time.Second)
	totalPoint := len(data)
	currentPoint := sort.Search(len(data), func(i int) bool {
		return data[i].Offset >= offset
	})
	if data[currentPoint].Offset != offset && number == 1 {
		return nil
	}

	var ret []SimpleLocation
	for ; number > 0; number-- {
		l := SimpleLocation{
			Timestamp: today + data[currentPoint].Offset,
			GPS:       []float64{data[currentPoint].Latitude, data[currentPoint].Longitude, data[currentPoint].Accuracy},
		}
		l.ToEarth(m.conversion)
		ret = append(ret, l)
		if currentPoint > 0 {
			currentPoint--
		} else {
			currentPoint = totalPoint - 1
			today -= oneDaySeconds
		}
	}
	return ret
}

type BreadcrumbOffset struct {
	Latitude  float64 `json:"earth_to_mars_latitude"`
	Longitude float64 `json:"earth_to_mars_longitude"`
}

func (o BreadcrumbOffset) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"earth_to_mars_latitude":%.4f,"earth_to_mars_longitude":%.4f}`, o.Latitude, o.Longitude)), nil
}

func (m RouteMap) HandleUpdateBreadcrums(breadcrumbs []SimpleLocation) BreadcrumbOffset {
	m.Header().Set("Access-Control-Allow-Origin", m.config.AccessDomain)
	m.Header().Set("Access-Control-Allow-Credentials", "true")
	m.Header().Set("Cache-Control", "no-cache")

	var token Token
	var ret BreadcrumbOffset
	token, ok := m.auth(false)
	if !ok || token.Readonly {
		m.Error(http.StatusUnauthorized, m.DetailError(-1, "invalid token"))
		return ret
	}
	m.Vars()["user_id"] = fmt.Sprintf("%d", token.UserId)

	return m.HandleUpdateBreadcrumsInner(breadcrumbs)
}

func (m RouteMap) HandleUpdateBreadcrumsInner(breadcrumbs []SimpleLocation) BreadcrumbOffset {
	var ret BreadcrumbOffset

	userIdStr, breadcrumb := m.Vars()["user_id"], breadcrumbs[0]
	mars, earth := breadcrumb, breadcrumb
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		m.Error(http.StatusBadRequest, err)
		return ret
	}
	if len(breadcrumb.GPS) < 3 {
		m.Error(http.StatusBadRequest, fmt.Errorf("invalid breadcrumb: %+v", breadcrumb))
		return ret
	}
	if m.Request().URL.Query().Get("coordinate") == "mars" {
		breadcrumb.ToEarth(m.conversion)
		earth = breadcrumb
	} else {
		mars.ToMars(m.conversion)
	}
	lat, lng, acc := breadcrumb.GPS[0], breadcrumb.GPS[1], breadcrumb.GPS[2]
	if acc <= 0 {
		m.Error(http.StatusBadRequest, fmt.Errorf("invalid accuracy: %f", acc))
		return ret
	}
	if acc > 70 {
		logger.INFO("routex", "user", userId, "breadcrumb", fmt.Sprintf("%.7f", lat), fmt.Sprintf("%.7f", lng), acc, "accuracy too large, ignore")
		ret = BreadcrumbOffset{
			Latitude:  mars.GPS[0] - earth.GPS[0],
			Longitude: mars.GPS[1] - earth.GPS[1],
		}
		return ret
	}

	breadcrumb.Timestamp = time.Now().Unix()
	last, err := m.breadcrumbCache.Load(userId)
	distance := float64(100)
	if err == nil && len(last.GPS) >= 3 {
		lastLat, lastLng := last.GPS[0], last.GPS[1]
		distance = Distance(lat, lng, lastLat, lastLng)
	}
	var crossIds []int64
	action := ""
	if distance > 30 {
		action = "save_to_history"
		logger.INFO("routex", "user", userId, "breadcrumb", fmt.Sprintf("%.7f", lat), fmt.Sprintf("%.7f", lng), acc)
		if crossIds, err = m.breadcrumbCache.SaveCross(userId, breadcrumb); err != nil {
			logger.ERROR("can't save cache %d: %s with %+v", userId, err, breadcrumb)
			m.Error(http.StatusInternalServerError, err)
			return ret
		}
		if err := m.breadcrumbCache.Save(userId, breadcrumb); err != nil {
			logger.ERROR("can't save cache %d: %s with %+v", userId, err, breadcrumb)
		}
		if err := m.breadcrumbsRepo.Save(userId, breadcrumb); err != nil {
			logger.ERROR("can't save user %d breadcrumb: %s with %+v", userId, err, breadcrumb)
		}
	} else {
		logger.INFO("routex", "user", userId, "breadcrumb", fmt.Sprintf("%.7f", lat), fmt.Sprintf("%.7f", lng), acc, "distance", fmt.Sprintf("%.2f", distance), "nosave")
		if crossIds, err = m.breadcrumbCache.SaveCross(userId, breadcrumb); err != nil {
			logger.ERROR("can't save cache %d: %s with %+v", userId, err, breadcrumb)
			m.Error(http.StatusInternalServerError, err)
			return ret
		}
		if err := m.breadcrumbsRepo.Update(userId, breadcrumb); err != nil {
			logger.ERROR("can't update user %d breadcrumb: %s with %+v", userId, err, breadcrumb)
		}
	}

	ret = BreadcrumbOffset{
		Latitude:  mars.GPS[0] - earth.GPS[0],
		Longitude: mars.GPS[1] - earth.GPS[1],
	}

	go func() {
		route := Geomark{
			Id:        m.breadcrumbsId(userId),
			Action:    action,
			Type:      "route",
			Tags:      []string{"breadcrumbs"},
			Positions: []SimpleLocation{breadcrumb},
		}
		for _, cross := range crossIds {
			m.castLocker.RLock()
			b, ok := m.crossCast[cross]
			m.castLocker.RUnlock()
			if !ok {
				continue
			}
			b.Send(route)
		}
	}()

	return ret
}

func (m RouteMap) HandleGetBreadcrums() []Geomark {
	m.Header().Set("Access-Control-Allow-Origin", m.config.AccessDomain)
	m.Header().Set("Access-Control-Allow-Credentials", "true")
	m.Header().Set("Cache-Control", "no-cache")

	toMars := m.Request().URL.Query().Get("coordinate") == "mars"
	token, ok := m.auth(true)
	if !ok {
		m.Error(http.StatusUnauthorized, m.DetailError(-1, "invalid token"))
		return nil
	}
	return m.getBreadcrumbs(token.Cross, toMars)
}

func (m RouteMap) getBreadcrumbs(cross model.Cross, toMars bool) []Geomark {
	ret := make([]Geomark, 0)
	for _, invitation := range cross.Exfee.Invitations {
		userId := invitation.Identity.UserID
		route := Geomark{
			Id:   m.breadcrumbsId(userId),
			Type: "route",
		}

		if route.Positions = m.getTutorialData(time.Now().UTC(), userId, 100); route.Positions == nil {
			var err error
			if route.Positions, err = m.breadcrumbsRepo.Load(userId, int64(cross.ID), time.Now().Unix()); err != nil {
				logger.ERROR("can't get user %d breadcrumbs of cross %d: %s", userId, cross.ID, err)
				continue
			}
		}
		if len(route.Positions) == 0 {
			continue
		}
		if toMars {
			route.ToMars(m.conversion)
		}
		ret = append(ret, route)
	}
	return ret
}

func (m RouteMap) HandleGetUserBreadcrums() Geomark {
	m.Header().Set("Access-Control-Allow-Origin", m.config.AccessDomain)
	m.Header().Set("Access-Control-Allow-Credentials", "true")
	m.Header().Set("Cache-Control", "no-cache")

	toMars, userIdStr := m.Request().URL.Query().Get("coordinate") == "mars", m.Vars()["user_id"]
	token, ok := m.auth(true)
	var ret Geomark
	if !ok {
		m.Error(http.StatusUnauthorized, m.DetailError(-1, "invalid token"))
		return ret
	}
	var userId int64
	for _, invitation := range token.Cross.Exfee.Invitations {
		if fmt.Sprintf("%d", invitation.Identity.UserID) == userIdStr {
			userId = invitation.Identity.UserID
			break
		}
	}
	if userId == 0 {
		m.Error(http.StatusUnauthorized, m.DetailError(-1, "invalid token"))
		return ret
	}
	after := time.Now().UTC()
	if afterTimstamp := m.Request().URL.Query().Get("after_timestamp"); afterTimstamp != "" {
		timestamp, err := strconv.ParseInt(afterTimstamp, 10, 64)
		if err != nil {
			m.Error(http.StatusBadRequest, err)
			return ret
		}
		after = time.Unix(timestamp, 0)
	}
	if ret.Positions = m.getTutorialData(after, userId, 100); ret.Positions == nil {
		var err error
		if ret.Positions, err = m.breadcrumbsRepo.Load(userId, int64(token.Cross.ID), after.Unix()); err != nil {
			if err == redis.ErrNil {
				m.Error(http.StatusNotFound, err)
			} else {
				logger.ERROR("can't get user %d breadcrumbs of cross %d: %s", userId, token.Cross.ID, err)
				m.Error(http.StatusInternalServerError, err)
			}
			return ret
		}
	}
	ret.Id, ret.Type = m.breadcrumbsId(userId), "route"
	if toMars {
		ret.ToMars(m.conversion)
	}
	return ret
}

func (m RouteMap) HandleGetUserBreadcrumsInner() Geomark {
	toMars, userIdStr := m.Request().URL.Query().Get("coordinate") == "mars", m.Vars()["user_id"]
	var ret Geomark
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		m.Error(http.StatusBadRequest, err)
		return ret
	}

	l, err := m.breadcrumbCache.Load(userId)
	if err != nil {
		if err == redis.ErrNil {
			m.Error(http.StatusNotFound, fmt.Errorf("can't find any breadcrumbs"))
		} else {
			logger.ERROR("can't get user %d breadcrumbs: %s", userId, err)
			m.Error(http.StatusInternalServerError, err)
		}
		return ret
	}
	ret.Id, ret.Type = m.breadcrumbsId(userId), "route"
	ret.Positions = []SimpleLocation{l}
	if toMars {
		ret.ToMars(m.conversion)
	}
	return ret
}

func (m RouteMap) breadcrumbsId(userId int64) string {
	return fmt.Sprintf("breadcrumbs.%d", userId)
}
