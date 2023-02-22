package calendar

import (
	ics "github.com/arran4/golang-ical"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	"testing"
	"time"
)

type MMClientMock struct {
}

func (s MMClientMock) GetUserByEmail(email, etag string) (*model.User, *model.Response, error) {
	user := model.User{Id: "1", Email: "test@avenga.com", Username: "test"}
	return &user, nil, nil
}

func (s MMClientMock) GetUser(userId, etag string) (*model.User, *model.Response, error) {
	user := model.User{Id: "1", Email: "test@avenga.com", Username: "test"}
	return &user, nil, nil
}

func (s MMClientMock) DMPost(userID string, post *model.Post) (*model.Post, error) {
	return post, nil
}

func (s MMClientMock) GetUsersByIds(userIds []string) ([]*model.User, *model.Response, error) {
	user := model.User{Id: "1", Email: "test1@avenga.com", Username: "test1"}
	user1 := model.User{Id: "2", Email: "test2@avenga.com", Username: "test2"}
	user2 := model.User{Id: "3", Email: "test3@avenga.com", Username: "test3"}

	return []*model.User{&user, &user1, &user2}, nil, nil
}

func TestCreateCalendarPost(t *testing.T) {
	testedInstance := CalendarPostServiceImpl{}
	option := apps.SelectOption{Label: "test", Value: "test"}

	post := testedInstance.CreateCalendarPost(option)
	bindings := post.GetProps()["app_bindings"].([]apps.Binding)[0]

	if len(bindings.Bindings) != 4 {
		t.Error("Wrong number of buttons in create calendar post")
	}

	if bindings.Label != "Calendar "+option.Label {
		t.Error("Wrong label in create calendar post")
	}
}

func TestCreateCalendarEventPost(t *testing.T) {
	testedInstance := CreateCalendarEventPostService{MMClientMock{}}

	postDto := createPostDto("123")

	post := testedInstance.CreateCalendarEventPost(&postDto)
	bindings := post.GetProps()["app_bindings"].([]apps.Binding)

	if len(bindings[0].Bindings) != 2 {
		t.Error("Only delete and view button must be present")
	}

	if len(bindings[0].Bindings[0].Form.Fields) != 5 {
		t.Error("Only five fields should be present")
	}
}

func TestCreateCalendarEventPostWithMeetingButtons(t *testing.T) {
	testedInstance := CreateCalendarEventPostService{MMClientMock{}}

	postDto := createPostDto("https://us04web.zoom.us/j/77756031423?pwd=V8ZGNqGyaFuvPhSlagehAsbkfY3yik.1 https://meet.google.com/ejz-ymdj-edd")

	post := testedInstance.CreateCalendarEventPost(&postDto)
	bindings := post.GetProps()["app_bindings"].([]apps.Binding)

	if len(bindings[0].Bindings) != 4 {
		t.Error("Wrong number of buttons")
	}

	if len(bindings[0].Bindings[0].Form.Fields) != 7 {
		t.Error("Only seven fields should be present")
	}
}

func TestHandleGetEvents(t *testing.T) {
	testedInstance := CreateCalendarEventPostService{MMClientMock{}}

	postDto := createPostDto("https://us04web.zoom.us/j/77756031423?pwd=V8ZGNqGyaFuvPhSlagehAsbkfY3yik.1 https://meet.google.com/ejz-ymdj-edd")

	post := testedInstance.CreateCalendarEventPost(&postDto)
	bindings := post.GetProps()["app_bindings"].([]apps.Binding)

	if len(bindings[0].Bindings) != 4 {
		t.Error("Wrong number of buttons")
	}

	if len(bindings[0].Bindings[0].Form.Fields) != 7 {
		t.Error("Only seven fields should be present")
	}
}

func createPostDto(description string) CalendarEventPostDTO {
	IcalProp := map[string][]string{}
	IcalProp["PARTSTAT"] = []string{"ACCEPTED"}
	property := ics.BaseProperty{IANAToken: "ATTENDEE", Value: "mailto:test@avenga.com", ICalParameters: IcalProp}
	sumProperty := ics.BaseProperty{IANAToken: "DESCRIPTION", Value: description, ICalParameters: map[string][]string{}}
	ianaProperty := ics.IANAProperty{BaseProperty: property}
	descIanaProperty := ics.IANAProperty{BaseProperty: sumProperty}
	base := ics.ComponentBase{Properties: []ics.IANAProperty{ianaProperty, descIanaProperty}}
	event := ics.VEvent{ComponentBase: base}
	asBot := MMClientMock{}
	location, _ := time.LoadLocation("UTC")
	userMap := map[string]interface{}{
		"user_id": "1",
	}
	oAuth2App := apps.OAuth2App{RemoteRootURL: "http://localhost:8081"}
	context := apps.Context{ExpandedContext: apps.ExpandedContext{ActingUser: &model.User{Locale: "en"}, OAuth2: apps.OAuth2Context{User: userMap, OAuth2App: oAuth2App}}}
	return CalendarEventPostDTO{&event, asBot, "test", "test", location, apps.CallRequest{Context: context}}
}
