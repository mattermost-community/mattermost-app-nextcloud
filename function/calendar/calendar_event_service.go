package calendar

import (
	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"time"
)

type CalendarEventServiceImpl struct {
	creq apps.CallRequest
}

func (c CalendarEventServiceImpl) CreateEventBody() (string, string) {

	description := c.creq.Values["description"].(string)
	title := c.creq.Values["title"].(string)

	organizerId := c.creq.Context.ActingUser.Id

	attendee := c.creq.Values["attendees"].(map[string]interface{})["value"].(string)

	userIds := make([]string, 0)
	userIds = append(userIds, attendee)

	asBot := appclient.AsBot(c.creq.Context)

	organizer, _, _ := asBot.GetUser(organizerId, "")
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
	event.SetOrganizer("mailto:"+organizer.Email, ics.WithCN("Owner"))
	for _, u := range users {
		event.AddAttendee(u.Email, ics.CalendarUserTypeIndividual, ics.ParticipationStatusNeedsAction, ics.ParticipationRoleReqParticipant, ics.WithRSVP(true))
	}
	text := cal.Serialize()
	return id, text

}
