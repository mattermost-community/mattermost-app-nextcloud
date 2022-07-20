package calendar

import (
	"errors"
	"fmt"
	"strings"

	ics "github.com/arran4/golang-ical"
)

func GetCalendarEvent(request WebhookCalendarRequest) (*CalendarEventDto, error) {
	calendar := getParsedCalendar(request.Values.Data.ObjectData.Calendardata)
	return getEventUsers(calendar)

}

func getParsedCalendar(calendarData string) ics.Calendar {
	cal, _ := ics.ParseCalendar(strings.NewReader(calendarData))
	return *cal
}

func getEventUsers(calendar ics.Calendar) (*CalendarEventDto, error) {

	for _, e := range calendar.Events() {
		event := CalendarEventDto{}
		attendees := []string{}
		for _, p := range e.Properties {
			switch p.IANAToken {
			case "SUMMARY":
				event.Summary = p.Value
			case "UID":
				event.ID = p.Value
			case "DTSTART":
				event.Start = p.Value
			case "DTEND":
				event.End = p.Value
			case "DESCRIPTION":
				event.Description = p.Value
			case "ATTENDEE":
				email := strings.Split(p.Value, ":")[1]
				attendees = append(attendees, email)
			case "ORGANIZER":
				event.OrganizerEmail = p.Value
			}
			fmt.Println(p.BaseProperty)
		}
		event.AttendeeEmails = attendees
		return &event, nil
	}
	return nil, errors.New("multiple events created")
}
