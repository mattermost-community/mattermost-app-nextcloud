package calendar

import (
	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"strconv"
	"strings"
	"time"
)

type CalendarEventServiceImpl struct {
	creq apps.CallRequest
}

func (c CalendarEventServiceImpl) CreateEventBody(fromDateUTC string, duration string, timezone string) (string, string) {

	from, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", fromDateUTC)
	if err != nil {
		println(err.Error())
	}

	to := prepareEndDate(from, duration)
	description, isPresent := c.creq.Values["description"].(string)
	title := c.creq.Values["title"].(string)

	organizerId := c.creq.Context.ActingUser.Id

	asBot := appclient.AsBot(c.creq.Context)

	organizer, _, _ := asBot.GetUser(organizerId, "")

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
	if isPresent {
		event.SetDescription(description)
	}
	event.SetOrganizer("mailto:"+organizer.Email, ics.WithCN("Owner"))

	if c.creq.Values["attendees"] != nil {
		addAttendeesToEvent(c.creq.Values["attendees"].([]interface{}), asBot, event)
	}

	text := cal.Serialize()
	return id, text

}

func FindAttendeeStatus(client *appclient.Client, event ics.VEvent, userId string) ics.ParticipationStatus {
	user, _, _ := client.GetUser(userId, "")
	for _, a := range event.Attendees() {
		if user.Email == a.Email() {
			return a.ParticipationStatus()
		}
	}
	return ""
}

func addAttendeesToEvent(attendee []interface{}, asBot *appclient.Client, event *ics.VEvent) {

	userIds := make([]string, 0)
	for _, a := range attendee {
		attendeeInfo := a.(map[string]interface{})
		userIds = append(userIds, attendeeInfo["value"].(string))
	}

	users, _, _ := asBot.GetUsersByIds(userIds)

	for _, u := range users {
		event.AddAttendee(u.Email, ics.CalendarUserTypeIndividual, ics.ParticipationStatusNeedsAction, ics.ParticipationRoleReqParticipant, ics.WithRSVP(true))
	}
}

func prepareEndDate(from time.Time, duration string) time.Time {
	if strings.Contains(duration, "All day") {
		date := from.AddDate(0, 0, 1)
		date = date.Add(-time.Minute * time.Duration(from.Minute()))
		date = date.Add(-time.Hour * time.Duration(from.Hour()))
		date = date.Add(-time.Second * time.Duration(from.Second()))
		return date
	}
	durationAndMeasure := strings.Split(duration, " ")
	if durationAndMeasure[1] == "minutes" {
		durationInMinutes, _ := strconv.Atoi(durationAndMeasure[0])
		return from.Add(time.Minute * time.Duration(durationInMinutes))
	}
	if strings.Contains(durationAndMeasure[1], "hour") {
		durationInHours, _ := strconv.ParseFloat(durationAndMeasure[0], 64)
		return from.Add(time.Duration(durationInHours * float64(time.Hour)))
	}

	return time.Time{}
}
