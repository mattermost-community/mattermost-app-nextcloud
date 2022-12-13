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

func (c CalendarEventServiceImpl) CreateEventBody(fromDateUTC string, toDateUTC string, timezone string) (string, string) {

	from, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", fromDateUTC)
	if err != nil {
		println(err.Error())
	}
	to, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", toDateUTC)

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
	cal.SetTzid(timezone)
	event := cal.AddEvent(id)
	event.SetCreatedTime(time.Now().UTC())
	event.SetDtStampTime(time.Now().UTC())
	event.SetModifiedAt(time.Now().UTC())
	event.SetProperty(ics.ComponentPropertyDtStart, from.Format(icalTimestampFormatUtcLocal), &ics.KeyValues{Key: "TZID", Value: []string{timezone}})
	event.SetProperty(ics.ComponentPropertyDtEnd, to.Format(icalTimestampFormatUtcLocal), &ics.KeyValues{Key: "TZID", Value: []string{timezone}})
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
