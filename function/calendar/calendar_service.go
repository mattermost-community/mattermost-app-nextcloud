package calendar

import (
	"encoding/xml"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	"io"
	"net/http"
	"strings"
)

const (
	icalTimestampFormatUtcLocal = "20060102T150405"
	icalTimestampFormatUtc      = "20060102T150405Z"
)

type CalendarService interface {
	CreateEvent(body string)
	GetUserCalendars() []apps.SelectOption
	GetCalendarEvents(event CalendarEventRequestRange) []string
	AddButtonsToEvents(commandBinding apps.Binding, status string, path string) apps.Binding
}

type CalendarServiceImpl struct {
	Url   string
	Token string
}

func (c CalendarServiceImpl) CreateEvent(body string) {

	req, _ := http.NewRequest("PUT", c.Url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}

func (c CalendarServiceImpl) GetUserCalendars() []apps.SelectOption {

	calendarsResponse := c.getUserCalendars()

	selectOptions := make([]apps.SelectOption, 0)

	for _, r := range calendarsResponse.Response {

		calendarName := r.Propstat.Prop.Displayname

		if len(calendarName) > 0 {
			splitUrl := strings.Split(r.Href, "/")
			val := splitUrl[len(splitUrl)-2]
			selectOption := apps.SelectOption{
				Label: calendarName,
				Value: val,
			}
			selectOptions = append(selectOptions, selectOption)
		}
	}
	return selectOptions
}

func (c CalendarServiceImpl) getUserCalendars() UserCalendarsResponse {

	body :=
		`<d:propfind xmlns:d="DAV:" xmlns:cs="http://calendarserver.org/ns/">
	<d:prop>
	   <d:displayname />
	   <cs:getctag />
	</d:prop>
  </d:propfind>`

	req, _ := http.NewRequest("PROPFIND", c.Url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	xmlResp := UserCalendarsResponse{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)

	return xmlResp
}

func (c CalendarServiceImpl) GetCalendarEvents(event CalendarEventRequestRange) []string {

	resp := c.getCalendarEvents(event)
	events := make([]string, 0)

	for _, r := range resp.Response {
		events = append(events, r.Propstat.Prop.CalendarData)
	}
	return events
}

func (c CalendarServiceImpl) getCalendarEvents(event CalendarEventRequestRange) UserCalendarEventsResponse {

	from := event.From.UTC().Format(icalTimestampFormatUtc)
	to := event.To.UTC().Format(icalTimestampFormatUtc)

	body := fmt.Sprintf(`<c:calendar-query xmlns:c="urn:ietf:params:xml:ns:caldav"
    xmlns:cs="http://calendarserver.org/ns/"
    xmlns:ca="http://apple.com/ns/ical/" 
    xmlns:d="DAV:">                                                            
    <d:prop>                
        <c:calendar-data />
    </d:prop>  
        <c:filter>
        <c:comp-filter name="VCALENDAR">
            <c:comp-filter name="VEVENT">
                <c:time-range start="%s" end="%s"/>
            </c:comp-filter>
        </c:comp-filter>
    </c:filter>
</c:calendar-query> `, from, to)

	req, _ := http.NewRequest("REPORT", c.Url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Depth", "1")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	xmlResp := UserCalendarEventsResponse{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)

	return xmlResp

}

func (c CalendarServiceImpl) GetCalendarEvent(calendarId string, eventId string) string {
	req, _ := http.NewRequest("GET", c.Url, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	event, _ := io.ReadAll(resp.Body)

	return string(event)
}

func (c CalendarServiceImpl) UpdateAttendeeStatus(cal *ics.Calendar, user *model.User, status string) string {
	for _, e := range cal.Events() {
		for _, a := range e.Attendees() {
			if user.Email == a.Email() {
				a.ICalParameters["PARTSTAT"] = []string{status}
				break
			}
		}
	}
	return cal.Serialize()
}

func (c CalendarServiceImpl) AddButtonsToEvents(commandBinding apps.Binding, status string, path string) apps.Binding {
	if status != "ACCEPTED" {
		commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
			Location: "Accept",
			Label:    "Accept",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "accepted")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
				ActingUser:            apps.ExpandAll,
			}),
		})
	}

	if status != "DECLINED" {
		commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
			Location: "Decline",
			Label:    "Decline",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "declined")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
				ActingUser:            apps.ExpandAll,
			}),
		})
	}

	if status != "TENTATIVE" {
		commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
			Location: "Tentative",
			Label:    "Tentative",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "tentative")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
				ActingUser:            apps.ExpandAll,
			}),
		})
	}
	return commandBinding
}
