package calendar

import (
	"encoding/xml"
	ics "github.com/arran4/golang-ical"
	"github.com/mattermost/mattermost-plugin-apps/apps"
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
	Response []UserCalendarsResponseItems `xml:"response"`
}

type UserCalendarsResponseItems struct {
	Text     string               `xml:",chardata"`
	Href     string               `xml:"href"`
	Propstat UserCalendarPropstat `xml:"propstat"`
}

type UserCalendarPropstat struct {
	Text   string           `xml:",chardata"`
	Prop   UserCalendarProp `xml:"prop"`
	Status string           `xml:"status"`
}

type UserCalendarProp struct {
	Text        string `xml:",chardata"`
	Displayname string `xml:"displayname"`
	Getctag     string `xml:"getctag"`
}

type UserCalendarEventsResponse struct {
	NextcloudXmlResponseHeaders
	Response []UserCalendarEventsResponseItems `xml:"response"`
}

type UserCalendarEventsResponseItems struct {
	Text     string           `xml:",chardata"`
	Href     string           `xml:"href"`
	Propstat CalendarPropStat `xml:"propstat"`
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

type CalendarEventData struct {
	CalendarStr string
	CalendarId  string
	CalendarIcs ics.Calendar
	Event       ics.VEvent
}

type CalendarEventPostDTO struct {
	event      *ics.VEvent
	bot        GetMMUser
	calendarId string
	eventId    string
	loc        *time.Location
	creq       apps.CallRequest
}
