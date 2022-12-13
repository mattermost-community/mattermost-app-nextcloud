package calendar

import (
	"errors"
	"fmt"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

type CalendarWebhookService interface {
	GetCalendarEvent(request WebhookCalendarRequest) (*CalendarEventDto, error)
}

type CalenderWebhookServiceImpl struct {
	request WebhookCalendarRequest
}

func (c CalenderWebhookServiceImpl) GetCalendarEvent(request WebhookCalendarRequest) (*CalendarEventDto, error) {

	calendar := getParsedCalendar(request.Values.Data.ObjectData.Calendardata)
	event, err := getEventUsers(calendar)
	event.CalendarId = request.Values.Data.CalendarData.URI
	principalUri := strings.Split(request.Values.Data.CalendarData.Principaluri, "/")
	event.EventOwner = principalUri[len(principalUri)-1]
	event.ID = request.Values.Data.ObjectData.URI
	return event, err

}

func getParsedCalendar(calendarData string) ics.Calendar {
	cal, _ := ics.ParseCalendar(strings.NewReader(calendarData))
	return *cal
}

func getEventUsers(calendar ics.Calendar) (*CalendarEventDto, error) {

	for _, e := range calendar.Events() {
		event := CalendarEventDto{}
		for _, p := range e.Properties {
			switch p.IANAToken {
			case "SUMMARY":
				event.Summary = p.Value
			case "DTSTART":
				locale := p.ICalParameters["TZID"][0]
				loc, _ := time.LoadLocation(locale)
				date, _ := time.ParseInLocation(icalTimestampFormatUtcLocal, p.Value, loc)
				event.Start = date.UTC().Format(icalTimestampFormatUtc)
			case "DTEND":
				locale := p.ICalParameters["TZID"][0]
				loc, _ := time.LoadLocation(locale)
				date, _ := time.ParseInLocation(icalTimestampFormatUtcLocal, p.Value, loc)
				event.End = date.UTC().Format(icalTimestampFormatUtc)
			case "DESCRIPTION":
				event.Description = p.Value
			case "ORGANIZER":
				event.OrganizerEmail = p.Value
			case "STATUS":
				event.Status = p.Value
			}
			fmt.Println(p.BaseProperty)
		}
		event.Attendees = e.Attendees()
		return &event, nil
	}
	return nil, errors.New("multiple events created")
}
