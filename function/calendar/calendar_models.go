package calendar

import (
	"encoding/xml"
	ics "github.com/arran4/golang-ical"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"time"
)

type NextcloudXmlResponseHeaders struct {
	XMLName xml.Name `xml:"multistatus"`
	Text    string   `xml:",chardata"`
	D       string   `xml:"d,attr"`
	S       string   `xml:"s,attr"`
	Cal     string   `xml:"cal,attr"`
	Cs      string   `xml:"cs,attr"`
	Oc      string   `xml:"oc,attr"`
	Nc      string   `xml:"nc,attr"`
}

type UserCalendarsResponse struct {
	NextcloudXmlResponseHeaders
	Response []struct {
		Text     string `xml:",chardata"`
		Href     string `xml:"href"`
		Propstat struct {
			Text string `xml:",chardata"`
			Prop struct {
				Text        string `xml:",chardata"`
				Displayname string `xml:"displayname"`
				Getctag     string `xml:"getctag"`
			} `xml:"prop"`
			Status string `xml:"status"`
		} `xml:"propstat"`
	} `xml:"response"`
}

type UserCalendarEventsResponse struct {
	NextcloudXmlResponseHeaders
	Response []struct {
		Text     string           `xml:",chardata"`
		Href     string           `xml:"href"`
		Propstat CalendarPropStat `xml:"propstat"`
	} `xml:"response"`
}

type CalendarPropStat struct {
	Text   string       `xml:",chardata"`
	Prop   CalendarProp `xml:"prop"`
	Status string       `xml:"status"`
}

type CalendarProp struct {
	Text         string `xml:",chardata"`
	CalendarData string `xml:"calendar-data"`
}

type Value struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type CalendarEventRequestRange struct {
	From time.Time
	To   time.Time
}

type CalendarEventPostDTO struct {
	event          *ics.VEvent
	status         ics.ParticipationStatus
	bot            appclient.Client
	calendarId     string
	organizerEmail string
	eventId        string
	userId         string
	loc            *time.Location
	remoteUrl      string
	locale         string
}
