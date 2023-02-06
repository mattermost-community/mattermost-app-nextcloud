package calendar

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	"testing"
)

func TestCreateEventBody(t *testing.T) {
	attendees := map[string]interface{}{
		"label": "test1",
		"value": "id",
	}
	values := map[string]interface{}{
		"description": "description",
		"title":       "title",
		"attendees":   []interface{}{attendees},
	}
	creq := apps.CallRequest{Values: values, Context: apps.Context{ExpandedContext: apps.ExpandedContext{ActingUser: &model.User{Id: "1"}}}}
	testedInstance := CalendarEventServiceImpl{creq: creq, asBot: MMClientMock{}}

	id, eventBody := testedInstance.CreateEventBody("2023-02-06 01:23:32.76349399 +0200 EET", "30 minutes", "Europe/Kiev")

	if len(id) == 0 || len(eventBody) == 0 {
		t.Error("Error during creation of event body")
	}
}

func TestCreateEventBodyWithDurationInHours(t *testing.T) {
	values := map[string]interface{}{
		"description": "description",
		"title":       "title",
	}
	creq := apps.CallRequest{Values: values, Context: apps.Context{ExpandedContext: apps.ExpandedContext{ActingUser: &model.User{Id: "1"}}}}
	testedInstance := CalendarEventServiceImpl{creq: creq, asBot: MMClientMock{}}

	id, eventBody := testedInstance.CreateEventBody("2023-02-06 01:23:32.76349399 +0200 EET", "2 hours", "Europe/Kiev")

	if len(id) == 0 || len(eventBody) == 0 {
		t.Error("Error during creation of event body")
	}
}
