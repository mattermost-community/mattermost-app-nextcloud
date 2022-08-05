package calendar

import (
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
)

func HandleWebhookCreateEvent(c *gin.Context) {

	creq := WebhookCalendarRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	calendarWebhookService := CalenderWebhookServiceImpl{creq}
	event, _ := calendarWebhookService.GetCalendarEvent(creq)

	asBot := appclient.AsBot(creq.Context)

	for _, a := range event.Attendees {
		u, _, _ := asBot.GetUserByEmail(a.Email(), "")
		post := createPostWithBindings(event, a, "New event")
		asBot.DMPost(u.Id, post)
	}
}

func HandleWebhookUpdateEvent(c *gin.Context) {

	creq := WebhookCalendarRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	calendarWebhookService := CalenderWebhookServiceImpl{creq}
	event, _ := calendarWebhookService.GetCalendarEvent(creq)

	asBot := appclient.AsBot(creq.Context)

	for _, a := range event.Attendees {
		u, _, _ := asBot.GetUserByEmail(a.Email(), "")
		post := createPostWithBindings(event, a, "Updated event")
		asBot.DMPost(u.Id, post)
	}
}

func createPostWithBindings(event *CalendarEventDto, attendee *ics.Attendee, message string) *model.Post {

	post := model.Post{
		Message: message,
	}
	start := event.GetFormattedStartDate("Jan _2 15:04:05")
	end := event.GetFormattedEndDate("Jan _2 15:04:05")
	status := attendee.ParticipationStatus()
	path := fmt.Sprintf("/calendars/%s/events/%s/status", event.CalendarId, event.ID)

	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       fmt.Sprintf("%s-%s  %s %s", start, end, event.Summary, status),
		Description: event.Description,
		Bindings:    []apps.Binding{},
	}

	if status != "ACCEPTED" {
		commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
			Location: "Accept",
			Label:    "Accept",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "accepted")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
			}),
		})
	}

	if status != "DECLINED" {
		commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
			Location: "Decline",
			Label:    "Decline",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "declined")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
			}),
		})
	}

	if status != "TENTATIVE" {
		commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
			Location: "Tentative",
			Label:    "Tentative",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "tentative")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
			}),
		})
	}

	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)

	return &post
}
