package calendar

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

func CreateEventBody(creq apps.CallRequest) (string, string) {

	description := creq.Values["description"].(string)
	title := creq.Values["title"].(string)

	organizerId := creq.Context.ActingUser.Id

	attendee := creq.Values["attendees"].(map[string]interface{})["value"].(string)

	userIds := make([]string, 0)
	userIds = append(userIds, attendee)

	asBot := appclient.AsBot(creq.Context)

	organizer, _, _ := asBot.GetUsersByIds([]string{organizerId})
	users, _, e := asBot.GetUsersByIds(userIds)

	if e != nil {
		return "", ""
	}

	newUUid := uuid.New()
	id := newUUid.String()
	cal := ics.NewCalendar()
	event := cal.AddEvent(id)
	event.SetCreatedTime(time.Now())
	event.SetDtStampTime(time.Now())
	event.SetModifiedAt(time.Now())
	event.SetStartAt(time.Now())
	event.SetEndAt(time.Now())
	event.SetSummary(title)
	event.SetLocation("Address")
	event.SetDescription(description)
	event.AddRrule(fmt.Sprintf("FREQ=YEARLY;BYMONTH=%d;BYMONTHDAY=%d", time.Now().Month(), time.Now().Day()))
	event.SetOrganizer(organizer[0].Email, ics.WithCN("Owner"))

	for _, u := range users {
		event.AddAttendee(u.Email, ics.CalendarUserTypeIndividual, ics.ParticipationStatusNeedsAction, ics.ParticipationRoleReqParticipant, ics.WithRSVP(true))
	}
	text := cal.Serialize()

	return id, text

}

func CreateEvent(url string, token string, body string) {

	req, _ := http.NewRequest("PUT", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}

func GetUserCalendars(reqUrl string, accessToken string) []apps.SelectOption {

	calendarsResponse := getUserCalendars(reqUrl, accessToken)

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

func getUserCalendars(url string, token string) UserCalendarsResponse {

	body :=
		`<d:propfind xmlns:d="DAV:" xmlns:cs="http://calendarserver.org/ns/">
	<d:prop>
	   <d:displayname />
	   <cs:getctag />
	</d:prop>
  </d:propfind>`

	req, _ := http.NewRequest("PROPFIND", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	xmlResp := UserCalendarsResponse{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)

	return xmlResp
}
