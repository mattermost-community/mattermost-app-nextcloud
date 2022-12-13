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
	"time"
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
	var timezone string
	var loc *time.Location

	if u.Timezone["useAutomaticTimezone"] == "false" {
		timezone = u.Timezone["manualTimezone"]
		loc, _ = time.LoadLocation(timezone)
	} else {
		timezone = u.Timezone["automaticTimezone"]
		loc, _ = time.LoadLocation(timezone)
	}

	from, _ := time.Parse(icalTimestampFormatUtc, event.Start)
	to, _ := time.Parse(icalTimestampFormatUtc, event.End)
	localizedFrom := from.In(loc).Format(icalTimestampFormatUtcLocal)
	localizedTo := to.In(loc).Format(icalTimestampFormatUtcLocal)
	event.Start = localizedFrom
	event.End = localizedTo

	userSettings := userSettingsService.GetUserSettingsById(u.Id)

	if userSettings.Contains(creq.Values.Data.CalendarData.URI) {
		return
	}

	for _, a := range event.Attendees {
		if a.Email() == u.Email {
			post := createPostWithBindings(event, a, *asBot, "New event", u.Email, timezone)
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

	if event.Status == "CANCELLED" {
		return
	}

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

	var timezone string
	var loc *time.Location

	if u.Timezone["useAutomaticTimezone"] == "false" {
		timezone = u.Timezone["manualTimezone"]
		loc, _ = time.LoadLocation(timezone)
	} else {
		timezone = u.Timezone["automaticTimezone"]
		loc, _ = time.LoadLocation(timezone)
	}

	from, _ := time.Parse(icalTimestampFormatUtc, event.Start)
	to, _ := time.Parse(icalTimestampFormatUtc, event.End)
	localizedFrom := from.In(loc).Format(icalTimestampFormatUtcLocal)
	localizedTo := to.In(loc).Format(icalTimestampFormatUtcLocal)
	event.Start = localizedFrom
	event.End = localizedTo

	for _, a := range event.Attendees {
		if a.Email() == u.Email {
			post := createPostWithBindings(event, a, *asBot, "Updated event", u.Email, timezone)
			asBot.DMPost(u.Id, post)
			break
		}
	}
}

func createPostWithBindings(event *CalendarEventDto, attendee *ics.Attendee, bot appclient.Client, message string, orginizerEmail string, timezone string) *model.Post {

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
		Location: "embedded",
		AppID:    "nextcloud",
		Label:    fmt.Sprintf("%s %s-%s  %s %s", timezone, start, end, event.Summary, status),
		Description: fmt.Sprintf("Organizer %s Description: %s Attendies %s",
			castEmailToMMUsername(event.OrganizerEmail, bot), event.Description, prepareAllAttendeesUsernames(event.Attendees, bot)),
		Bindings: []apps.Binding{},
	}
	commandBinding = calendarService.AddButtonsToEvents(commandBinding, string(status), path)

	if event.OrganizerEmail == orginizerEmail {
		deletePath := fmt.Sprintf("/delete-event/%s/events/%s", event.CalendarId, event.ID)
		сreateDeleteButton(&commandBinding, "Delete", "Delete", deletePath)
	}
	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)

	return &post
}

func castEmailToMMUsername(email string, bot appclient.Client) string {
	if strings.Contains(email, ":") {
		email = strings.Split(email, ":")[1]
	}
	mmUser, _, err := bot.GetUserByEmail(email, "")
	if err != nil {
		return email
	}
	return "@" + mmUser.Username
}

func prepareAllAttendeesUsernames(attendees []*ics.Attendee, bot appclient.Client) string {
	var attendeesUsernames string
	for _, attendee := range attendees {
		attendeesUsernames += castEmailToMMUsername(attendee.Email(), bot) + " - " + attendee.ICalParameters["PARTSTAT"][0] + " "
	}
	return attendeesUsernames
}
