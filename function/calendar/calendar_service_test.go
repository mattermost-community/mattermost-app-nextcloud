package calendar

import (
	ics "github.com/arran4/golang-ical"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
	"net/http"
	"testing"
	"time"
)

type CalendarEventServiceImplMock struct {
	icsResponse string
	error       error
}

func (c CalendarEventServiceImplMock) getUrl() string {
	return ""
}

func (c CalendarEventServiceImplMock) getUserCalendars() (UserCalendarsResponse, error) {
	prop := UserCalendarProp{"test", "test", "1"}
	propstat := UserCalendarPropstat{Text: "test", Prop: prop, Status: "test"}
	items := UserCalendarsResponseItems{"test", "/remote.php/dav/calendars/admin/custom/431b2eba-713f-427f-a058-65bd595db528.ics", propstat}
	response := UserCalendarsResponse{Response: []UserCalendarsResponseItems{items}}
	return response, c.error
}

func (c CalendarEventServiceImplMock) deleteUserEvent() (*http.Response, error) {
	return nil, nil
}

func (c CalendarEventServiceImplMock) getCalendarEvents(event CalendarEventRequestRange) (UserCalendarEventsResponse, error) {
	prop := CalendarProp{CalendarData: "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//arran4//Golang ICS Library\nTZID:Europe/Kiev\nBEGIN:VEVENT\nUID:431b2eba-713f-427f-a058-65bd595db528\nCREATED:20230205T180704Z\nDTSTAMP:20230205T180704Z\nLAST-MODIFIED:20230205T180704Z\nDTSTART;TZID=Europe/Kiev:20230205T200658\nDTEND;TZID=Europe/Kiev:20230205T203658\nSUMMARY:123\nLOCATION:Address\nDESCRIPTION:123\nORGANIZER;CN=Owner:mailto:ostapss222@gmail.com\nATTENDEE;CUTYPE=INDIVIDUAL;PARTSTAT=NEEDS-ACTION;ROLE=REQ-PARTICIPANT;RSVP=\n true;SCHEDULE-STATUS=5.0:mailto:ostap.melnychuk@avenga.com\nEND:VEVENT\nEND:VCALENDAR\n"}
	propStat := CalendarPropStat{Prop: prop}
	item := UserCalendarEventsResponseItems{"", "/remote.php/dav/calendars/admin/custom/431b2eba-713f-427f-a058-65bd595db528.ics", propStat}
	response := UserCalendarEventsResponse{Response: []UserCalendarEventsResponseItems{item}}
	return response, c.error
}

func (c CalendarEventServiceImplMock) createEvent(body string) (*http.Response, error) {
	return nil, nil
}
func (c CalendarEventServiceImplMock) getCalendarEvent() (string, error) {
	return c.icsResponse, c.error
}

func TestGetCalendarEvents(t *testing.T) {
	testedInstance := CalendarServiceImpl{calendarRequestService: CalendarEventServiceImplMock{}}
	from := time.Now().AddDate(0, 0, -1)
	now := time.Now()
	eventsData := testedInstance.GetCalendarEvents(CalendarEventRequestRange{From: from, To: now})
	if len(eventsData) != 1 {
		t.Error("Wrong number of events and eventIds returned")
	}
}

func TestGetCalendarEventsError(t *testing.T) {
	testedInstance := CalendarServiceImpl{calendarRequestService: CalendarEventServiceImplMock{error: errors.New("test")}}
	from := time.Now().AddDate(0, 0, -1)
	now := time.Now()
	eventsData := testedInstance.GetCalendarEvents(CalendarEventRequestRange{From: from, To: now})
	if len(eventsData) != 0 {
		t.Error("Wrong number of events and eventIds returned")
	}
}

func TestGetUserCalendars(t *testing.T) {
	testedInstance := CalendarServiceImpl{calendarRequestService: CalendarEventServiceImplMock{}}
	calendars := testedInstance.GetUserCalendars()
	if len(calendars) != 1 {
		t.Error("Wrong number of calendars returned")
	}
}

func TestGetUserCalendarsError(t *testing.T) {
	testedInstance := CalendarServiceImpl{calendarRequestService: CalendarEventServiceImplMock{error: errors.New("test")}}
	calendars := testedInstance.GetUserCalendars()
	if len(calendars) != 0 {
		t.Error("Wrong number of calendars returned")
	}
}

func TestUserStatusUpdate(t *testing.T) {
	testedInstance := CalendarServiceImpl{}

	property := ics.BaseProperty{IANAToken: "ATTENDEE", Value: "mailto:ostap.melnychuk@avenga.com", ICalParameters: map[string][]string{}}
	ianaProperty := ics.IANAProperty{BaseProperty: property}
	base := ics.ComponentBase{Properties: []ics.IANAProperty{ianaProperty}}
	event := ics.VEvent{ComponentBase: base}
	cal := ics.Calendar{Components: []ics.Component{&event}}

	user := model.User{Email: "ostap.melnychuk@avenga.com"}

	calStr, err := testedInstance.UpdateAttendeeStatus(&cal, &user, "DECLINED")
	status := cal.Components[0].UnknownPropertiesIANAProperties()[0].ICalParameters["PARTSTAT"][0]
	if len(calStr) == 0 || err != nil || status != "DECLINED" {
		t.Error("User status was not updated")
	}
}

func TestUserStatusUpdateFail(t *testing.T) {
	testedInstance := CalendarServiceImpl{}
	user := model.User{Email: "ostap.melnychuk@avenga.com"}

	_, err := testedInstance.UpdateAttendeeStatus(nil, &user, "DECLINED")
	if err == nil {
		t.Error("TestUpdateUserStatusUpdateFail should have returned error")
	}
}

func TestAddButtonsToEvent(t *testing.T) {
	testedInstance := CalendarServiceImpl{}
	commandBinding := apps.Binding{Bindings: make([]apps.Binding, 0)}

	commandBinding = testedInstance.AddButtonsToEvents(commandBinding, "NEEDS-ACTION", "test")

	if len(commandBinding.Bindings[0].Bindings) != 3 {
		t.Error("Not all buttons were added to an event")
	}

	if commandBinding.Bindings[0].Label != "Going?" {
		t.Error("Wrong label in event buttons")
	}
}

func TestAddButtonsToEventWithChosenStatus(t *testing.T) {
	testedInstance := CalendarServiceImpl{}
	commandBinding := apps.Binding{Bindings: make([]apps.Binding, 0)}

	commandBinding = testedInstance.AddButtonsToEvents(commandBinding, "ACCEPTED", "test")

	if len(commandBinding.Bindings[0].Bindings) != 2 {
		t.Error("Not all buttons were added to an event")
	}

	if commandBinding.Bindings[0].Label != "ACCEPTED" {
		t.Error("Wrong label in event buttons")
	}
}
