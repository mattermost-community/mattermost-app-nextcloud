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
	"strings"
)

func HandleWebhookCreateEvent(c *gin.Context) {

	creq := WebhookCalendarRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	calendarWebhookService := CalenderWebhookServiceImpl{creq}
	event, _ := calendarWebhookService.GetCalendarEvent(creq)

	principalUri := strings.Split(creq.Values.Data.CalendarData.Principaluri, "/")
	principal := principalUri[len(principalUri)-1]
	asBot := appclient.AsBot(creq.Context)

	var userId string
	asBot.KVGet("", fmt.Sprintf("nc-user-%s", principal), &userId)

	userSettingsService := user.UserSettingsServiceImpl{asBot}

	u, _, err := asBot.GetUser(userId, "")
	if err != nil {
		return
	}
	userSettings := userSettingsService.GetUserSettingsById(u.Id)

	if userSettings.Contains(creq.Values.Data.CalendarData.URI) {
		return
	}

	for _, a := range event.Attendees {
		if a.Email() == u.Email {
			post := createPostWithBindings(event, a, "New event", u.Email)
			asBot.DMPost(u.Id, post)
			break
		}
	}
}

func HandleWebhookUpdateEvent(c *gin.Context) {

	creq := WebhookCalendarRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	calendarWebhookService := CalenderWebhookServiceImpl{creq}
	event, _ := calendarWebhookService.GetCalendarEvent(creq)
	asBot := appclient.AsBot(creq.Context)
	principalUri := strings.Split(creq.Values.Data.CalendarData.Principaluri, "/")
	principal := principalUri[len(principalUri)-1]

	var userId string
	asBot.KVGet("", fmt.Sprintf("nc-user-%s", principal), &userId)
	userSettingsService := user.UserSettingsServiceImpl{asBot}

	u, _, err := asBot.GetUser(userId, "")
	if err != nil {
		return
	}

	userSettings := userSettingsService.GetUserSettingsById(u.Id)

	if userSettings.Contains(creq.Values.Data.CalendarData.URI) {
		return
	}

	for _, a := range event.Attendees {
		if a.Email() == u.Email {
			post := createPostWithBindings(event, a, "Updated event", u.Email)
			asBot.DMPost(u.Id, post)
			break
		}
	}
}

func createPostWithBindings(event *CalendarEventDto, attendee *ics.Attendee, message string, orginizerEmail string) *model.Post {

	post := model.Post{
		Message: message,
	}
	start := event.GetFormattedStartDate("Jan _2 15:04:05")
	end := event.GetFormattedEndDate("Jan _2 15:04:05")
	status := attendee.ParticipationStatus()
	eventOwner := event.EventOwner
	path := fmt.Sprintf("/users/%s/calendars/%s/events/%s/status", eventOwner, event.CalendarId, event.ID)
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
