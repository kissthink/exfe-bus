package routex

import (
	"fmt"
	"logger"
	"model"
	"net/http"
	"strconv"
	"time"
)

const CrossPlaceTag = "xplace"

func (m RouteMap) HandleSearchGeomarks() []Geomark {
	crossIdStr := m.Vars()["cross_id"]
	crossId, err := strconv.ParseInt(crossIdStr, 10, 64)
	if err != nil {
		m.Error(http.StatusBadRequest, err)
		return nil
	}
	ret := make([]Geomark, 0)
	tag := m.Request().URL.Query().Get("tags")
	if tag == "" {
		return ret
	}
	data, err := m.geomarksRepo.Get(crossId)
	if err != nil {
		logger.ERROR("can't get route of cross %d: %s", crossId, err)
		m.Error(http.StatusInternalServerError, err)
		return nil
	}
	if data == nil {
		return ret
	}
	for _, geomark := range data {
		ok := false
		for _, t := range geomark.Tags {
			if t == tag {
				ok = true
			}
			if t == CrossPlaceTag {
				ok = false
				break
			}
		}
		if ok {
			ret = append(ret, geomark)
		}
	}
	return ret
}

func (m RouteMap) HandleGetGeomarks() []Geomark {
	m.Header().Set("Access-Control-Allow-Origin", m.config.AccessDomain)
	m.Header().Set("Access-Control-Allow-Credentials", "true")
	m.Header().Set("Cache-Control", "no-cache")

	token, ok := m.auth(true)
	if !ok {
		m.Error(http.StatusUnauthorized, m.DetailError(-1, "invalid token"))
		return nil
	}
	toMars := false
	if m.Request().URL.Query().Get("coordinate") == "mars" {
		toMars = true
	}
	ret, err := m.getGeomarks(token.Cross, toMars)
	if err != nil {
		logger.ERROR("can't get route of cross %d: %s", token.Cross.ID, err)
		m.Error(http.StatusInternalServerError, err)
		return nil
	}
	return ret
}

func (m RouteMap) getGeomarks(cross model.Cross, toMars bool) ([]Geomark, error) {
	data, err := m.geomarksRepo.Get(int64(cross.ID))
	if err != nil {
		return nil, err
	}

	needCrossPlace := true
	hasDestination := false
	for i, d := range data {
		for _, t := range d.Tags {
			if t == CrossPlaceTag {
				needCrossPlace = false
			}
			if t == "destination" {
				hasDestination = true
			}
		}
		if toMars {
			d.ToMars(m.conversion)
			data[i] = d
		}
	}

	if needCrossPlace {
		var lat, lng float64
		if cross.Place != nil {
			if lng, err = strconv.ParseFloat(cross.Place.Lng, 64); err != nil {
				cross.Place = nil
			} else if lat, err = strconv.ParseFloat(cross.Place.Lat, 64); err != nil {
				cross.Place = nil
			}
		}
		if cross.Place != nil {
			createdAt, err := time.Parse("2006-01-02 15:04:05 -0700", cross.CreatedAt)
			if err != nil {
				createdAt = time.Now()
			}
			updatedAt, err := time.Parse("2006-01-02 15:04:05 -0700", cross.CreatedAt)
			if err != nil {
				updatedAt = time.Now()
			}
			xplace := Geomark{
				Id:          m.xplaceId(int64(cross.ID)),
				Type:        "location",
				CreatedAt:   createdAt.Unix(),
				CreatedBy:   cross.By.Id(),
				UpdatedAt:   updatedAt.Unix(),
				UpdatedBy:   cross.By.Id(),
				Tags:        []string{CrossPlaceTag},
				Icon:        "",
				Title:       cross.Place.Title,
				Description: cross.Place.Description,
				Longitude:   lng,
				Latitude:    lat,
			}
			if !hasDestination {
				xplace.Tags = append(xplace.Tags, "destination")
			}
			data = append(data, xplace)
		}
	}

	return data, nil
}

func (m RouteMap) HandleSetGeomark(mark Geomark) {
	m.Header().Set("Access-Control-Allow-Origin", m.config.AccessDomain)
	m.Header().Set("Access-Control-Allow-Credentials", "true")
	m.Header().Set("Cache-Control", "no-cache")

	token, ok := m.auth(true)
	if !ok || token.Readonly {
		m.Error(http.StatusUnauthorized, m.DetailError(-1, "invalid token"))
		return
	}

	mark.Type = m.Vars()["mark_type"]
	mark.Id = fmt.Sprintf("%s.%s", m.Vars()["mark_id"], m.Vars()["suffix"])
	suffix := m.Vars()["suffix"]
	mark.UpdatedAt, mark.Action = time.Now().Unix(), ""
	if m.Request().URL.Query().Get("coordinate") == "mars" {
		mark.ToEarth(m.conversion)
	}

	for i := len(mark.Tags) - 1; i >= 0; i-- {
		if mark.Tags[i] == CrossPlaceTag {
			go func() {
				if err := m.syncCrossPlace(&mark, token.Cross, mark.UpdatedBy); err != nil {
					logger.ERROR("can't set cross %d place: %s", token.Cross.ID, err)
				}
				m.castLocker.RLock()
				broadcast := m.crossCast[int64(token.Cross.ID)]
				m.castLocker.RUnlock()

				if suffix != CrossPlaceTag {
					if err := m.geomarksRepo.Delete(int64(token.Cross.ID), mark.Type, mark.Id); err != nil {
						logger.ERROR("can't delete cross %d geomark %s %s: %s", token.Cross.ID, mark.Type, mark.Id, err)
					}
					m.routexRepo.Update(token.UserId, int64(token.Cross.ID))
					mark.Action = "delete"
					if broadcast != nil {
						broadcast.Send(mark)
					}
					time.Sleep(time.Second / 10)
				}

				if broadcast != nil {
					mark.Id, mark.Action = m.xplaceId(int64(token.Cross.ID)), ""
					broadcast.Send(mark)
				}
				return
			}()
			return
		}
	}

	if suffix != "location" && suffix != "route" {
		m.Error(http.StatusBadRequest, fmt.Errorf("invalid suffix: %s", suffix))
		return
	}

	if err := m.geomarksRepo.Set(int64(token.Cross.ID), mark); err != nil {
		m.Error(http.StatusInternalServerError, err)
		return
	}
	m.routexRepo.Update(token.UserId, int64(token.Cross.ID))

	go func() {
		mark.Action = "update"
		m.castLocker.RLock()
		broadcast := m.crossCast[int64(token.Cross.ID)]
		m.castLocker.RUnlock()
		if broadcast != nil {
			broadcast.Send(mark)
		}
	}()

	return
}

func (m RouteMap) HandleDeleteGeomark() {
	m.Header().Set("Access-Control-Allow-Origin", m.config.AccessDomain)
	m.Header().Set("Access-Control-Allow-Credentials", "true")
	m.Header().Set("Cache-Control", "no-cache")

	token, ok := m.auth(true)
	if !ok || token.Readonly {
		m.Error(http.StatusUnauthorized, m.DetailError(-1, "invalid token"))
		return
	}

	var mark Geomark
	mark.Type = m.Vars()["mark_type"]
	mark.Id = fmt.Sprintf("%s.%s", m.Vars()["mark_id"], m.Vars()["suffix"])
	suffix := m.Vars()["suffix"]
	if suffix == "location" || suffix == "route" {
		if err := m.geomarksRepo.Delete(int64(token.Cross.ID), mark.Type, mark.Id); err != nil {
			m.Error(http.StatusInternalServerError, err)
			return
		}
	}
	m.routexRepo.Update(token.UserId, int64(token.Cross.ID))

	go func() {
		if suffix == CrossPlaceTag {
			by := ""
			for _, i := range token.Cross.Exfee.Invitations {
				if i.Identity.UserID == token.UserId {
					by = i.Identity.Id()
				}
			}
			if err := m.syncCrossPlace(nil, token.Cross, by); err != nil {
				logger.ERROR("remove cross %d place error: %s", token.Cross.ID, err)
			}
		}
		mark.Action = "delete"
		m.castLocker.RLock()
		broadcast := m.crossCast[int64(token.Cross.ID)]
		m.castLocker.RUnlock()
		if broadcast != nil {
			broadcast.Send(mark)
		}
	}()

	return
}

func (m RouteMap) xplaceId(crossId int64) string {
	return fmt.Sprintf("%d."+CrossPlaceTag, crossId)
}

func (m RouteMap) syncCrossPlace(geomark *Geomark, cross model.Cross, by string) error {
	updatedBy := model.FromIdentityId(by)
	place := model.Place{}

	if cross.Place != nil {
		place.ID = cross.Place.ID
	}
	if geomark != nil {
		place.Title = geomark.Title
		place.Description = geomark.Description
		place.Lng = fmt.Sprintf("%.7f", geomark.Longitude)
		place.Lat = fmt.Sprintf("%.7f", geomark.Latitude)
		place.Provider = "routex"
		place.ExternalID = fmt.Sprintf("%d", cross.ID)
	}
	updateCross := map[string]interface{}{"place": place}
	return m.platform.BotCrossUpdate("cross_id", fmt.Sprintf("%d", cross.ID), updateCross, updatedBy)
}
