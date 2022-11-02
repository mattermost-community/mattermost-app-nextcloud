package calendar

import (
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/prokhorind/nextcloud/function/user"
)

func HandleWebhookCreateEvent(c *gin.Context) {

	creq := WebhookCalendarRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	calendarWebhookService := CalenderWebhookServiceImpl{creq}
	event, _ := calendarWebhookService.GetCalendarEvent(creq)

	asBot := appclient.AsBot(creq.Context)
	userSettingsService := user.UserSettingsServiceImpl{asBot}

	for _, a := range event.Attendees {
		u, _, _ := asBot.GetUserByEmail(a.Email(), "")
		userSettings := userSettingsService.GetUserSettingsById(u.Id)
		if userSettings.Contains(creq.Values.Data.CalendarData.URI) {
			continue
		}
		post := createPostWithBindings(event, a, "New event", u.Email)
		asBot.DMPost(u.Id, post)
	}
}

func HandleWebhookUpdateEvent(c *gin.Context) {

	creq := WebhookCalendarRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	calendarWebhookService := CalenderWebhookServiceImpl{creq}
	event, _ := calendarWebhookService.GetCalendarEvent(creq)

	asBot := appclient.AsBot(creq.Context)

	userSettingsService := user.UserSettingsServiceImpl{asBot}

	for _, a := range event.Attendees {
		u, _, _ := asBot.GetUserByEmail(a.Email(), "")
		userSettings := userSettingsService.GetUserSettingsById(u.Id)
		if userSettings.Contains(creq.Values.Data.CalendarData.URI) {
			continue
		}
		post := createPostWithBindings(event, a, "Updated event", u.Email)
		asBot.DMPost(u.Id, post)
	}
}

func createPostWithBindings(event *CalendarEventDto, attendee *ics.Attendee, message string, orginizerEmail string) *model.Post {

	post := model.Post{
		Message: message,
	}
	start := event.GetFormattedStartDate("Jan _2 15:04:05")
	end := event.GetFormattedEndDate("Jan _2 15:04:05")
	status := attendee.ParticipationStatus()
	path := fmt.Sprintf("/calendars/%s/events/%s/status", event.CalendarId, event.ID)
	calendarService := CalendarServiceImpl{}
	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       fmt.Sprintf("%s-%s  %s %s", start, end, event.Summary, status),
		Description: event.Description,
		Bindings:    []apps.Binding{},
	}
	commandBinding = calendarService.AddButtonsToEvents(commandBinding, string(status), path)

	if event.OrganizerEmail == orginizerEmail {
		deletePath := fmt.Sprintf("/delete-event/%s/events/%s", event.CalendarId, event.ID)
		createDeleteButton(&commandBinding, "Delete", "Delete", deletePath)
	}
	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)

	return &post
}
