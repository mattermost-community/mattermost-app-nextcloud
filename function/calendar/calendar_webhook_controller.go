package calendar

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
)

func HandleWebhookCreateEvent(c *gin.Context) {

	creq := WebhookCalendarRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	stm, _ := json.Marshal(creq)
	fmt.Println(string(stm))
	calendarWebhookService := CalenderWebhookServiceImpl{creq}
	event, _ := calendarWebhookService.GetCalendarEvent(creq)

	asBot := appclient.AsBot(creq.Context)

	for _, a := range event.AttendeeEmails {
		u, _, _ := asBot.GetUserByEmail(a, "")
		post := createPostWithBindings(event)
		asBot.DMPost(u.Id, post)
	}

}

func createPostWithBindings(event *CalendarEventDto) *model.Post {

	post := model.Post{
		Message: string(event.Summary) + " " + event.GetFormattedStartDate("Jan _2 15:04:05") + " " + event.GetFormattedEndDate("Jan _2 15:04:05"),
	}
	start := event.GetFormattedStartDate("Jan _2 15:04:05")
	end := event.GetFormattedEndDate("Jan _2 15:04:05")

	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       fmt.Sprintf("%s-%s  %s", start, end, event.Summary),
		Description: event.Description,
		Bindings:    []apps.Binding{},
	}

	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: "Accept",
		Label:    "Accept",
		Submit: apps.NewCall("/event-accept").WithExpand(apps.Expand{
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			ActingUserAccessToken: apps.ExpandAll,
		}),
	})

	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: "Decline",
		Label:    "Decline",
		Submit: apps.NewCall("/event-decline").WithExpand(apps.Expand{
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			ActingUserAccessToken: apps.ExpandAll,
		}),
	})

	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)

	return &post
}
