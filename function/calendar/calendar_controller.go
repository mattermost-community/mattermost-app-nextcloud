package calendar

import (
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/prokhorind/nextcloud/function/user"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/prokhorind/nextcloud/function/oauth"
)

func HandleCreateEvent(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}

	token := oauthService.RefreshToken()
	accessToken := token.AccessToken

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	calendarEventService := CalendarEventServiceImpl{creq}
	uuid, body := calendarEventService.CreateEventBody()

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	calendar := creq.Values["calendar"].(map[string]interface{})["value"].(string)

	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s/%s.ics", remoteUrl, userId, calendar, uuid)

	calendarService := CalendarServiceImpl{Url: reqUrl, Token: accessToken}
	calendarService.CreateEvent(body)

	c.JSON(http.StatusOK, apps.NewTextResponse("event created:"+reqUrl))
}

func HandleCreateEventForm(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s", remoteUrl, userId)

	accessToken := token.AccessToken

	calendarService := CalendarServiceImpl{Url: reqUrl, Token: accessToken}
	option := creq.State.(map[string]interface{})

	form := &apps.Form{
		Title: "Create Nextcloud calendar event",
		Icon:  "icon.png",
		Fields: []apps.Field{
			{
				Type:       "text",
				Name:       "title",
				Label:      "Title",
				IsRequired: true,
			},
			{
				Type:        "text",
				Name:        "description",
				Label:       "Description",
				TextSubtype: apps.TextFieldSubtypeTextarea,
				IsRequired:  false,
			},
			{
				Type:          "user",
				Name:          "attendees",
				Label:         "Attendees",
				IsRequired:    true,
				SelectIsMulti: true,
			},

			{
				Type:                "static_select",
				Name:                "calendar",
				Label:               "Calendar",
				IsRequired:          true,
				SelectStaticOptions: calendarService.GetUserCalendars(),
				Value:               apps.SelectOption{Label: option["label"].(string), Value: option["value"].(string)},
			},
		},
		Submit: apps.NewCall("/create-calendar-event").WithExpand(apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			Channel:               apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}),
	}

	c.JSON(http.StatusOK, apps.NewFormResponse(*form))
}

func HandleDeleteCalendarEvent(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	calendarService := CalendarServiceImpl{}
	calendarId := c.Param("calendarId")
	eventId := c.Param("eventId")
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	user := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	deleteUrl := fmt.Sprintf("http://localhost:8081/remote.php/dav/calendars/%s/%s/%s.ics", user, calendarId, eventId)

	calendarService.deleteUserEvent(deleteUrl, token.AccessToken)
	c.JSON(http.StatusOK, apps.NewTextResponse("event deleted :"+eventId))
}

func HandleGetCalendarEventsForm(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s", remoteUrl, userId)

	accessToken := token.AccessToken

	calendarService := CalendarServiceImpl{Url: reqUrl, Token: accessToken}

	option := creq.State.(map[string]interface{})
	form := &apps.Form{
		Title: "Nextcloud calendar events",
		Icon:  "icon.png",
		Fields: []apps.Field{

			{
				Type:                "static_select",
				Name:                "calendar",
				Label:               "Calendar",
				IsRequired:          true,
				SelectStaticOptions: calendarService.GetUserCalendars(),
				Value:               apps.SelectOption{Label: option["label"].(string), Value: option["value"].(string)},
			},
		},
		Submit: apps.NewCall("/get-calendar-events").WithExpand(apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			Channel:               apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}),
	}

	c.JSON(http.StatusOK, apps.NewFormResponse(*form))
}

func HandleGetEvents(c *gin.Context) {

	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	calendar := creq.Values["calendar"].(map[string]interface{})["value"].(string)
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s", remoteUrl, userId, calendar)

	now := time.Now()
	from := now.AddDate(0, 0, -1)
	to := now.AddDate(0, 0, 1)
	eventRange := CalendarEventRequestRange{
		From: from,
		To:   to,
	}
	calendarService := CalendarServiceImpl{Url: reqUrl, Token: token.AccessToken}

	events, eventIds := calendarService.GetCalendarEvents(eventRange)
	calEvents := make([]ics.Calendar, len(events))

	for i := 0; i < len(events); i++ {
		cal, _ := ics.ParseCalendar(strings.NewReader(events[i]))
		calEvents[i] = *cal
	}

	asBot := appclient.AsBot(creq.Context)
	status := findAttendeeStatus(asBot, *calEvents[0].Events()[0], creq.Context.ActingUser.Id)
	mmUserId := creq.Context.ActingUser.Id
	organizerEmail := creq.Context.ActingUser.Email
	for i, e := range calEvents {
		post := createCalendarEventPost(e.Events()[0], status, calendar, organizerEmail, eventIds[i])
		asBot.DMPost(mmUserId, post)
	}

	c.JSON(http.StatusOK, apps.NewDataResponse(nil))
}

func findAttendeeStatus(client *appclient.Client, event ics.VEvent, userId string) ics.ParticipationStatus {
	user, _, _ := client.GetUser(userId, "")
	for _, a := range event.Attendees() {
		if user.Email == a.Email() {
			return a.ParticipationStatus()
		}
	}
	return ""
}

func createCalendarEventPost(event *ics.VEvent, status ics.ParticipationStatus, calendarId string, organizerEmail string, eventId string) *model.Post {
	var name, attendees, start, finish, description, organizer string
	attendees = ""
	for _, e := range event.Properties {
		if e.BaseProperty.IANAToken == "DESCRIPTION" {
			description = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "ORGANIZER" {
			organizer = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "ATTENDEE" {
			attendees = attendees + " " + e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "SUMMARY" {
			name = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "DTSTART" {
			start = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "DTEND" {
			finish = e.BaseProperty.Value
		}
	}
	post := model.Post{}
	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       "Event " + name,
		Description: createDescriptionForEvent(description, start, finish, organizer, attendees),
		Bindings:    []apps.Binding{},
	}
	calendarService := CalendarServiceImpl{}
	path := fmt.Sprintf("/calendars/%s/events/%s/status", calendarId, eventId)
	commandBinding = calendarService.AddButtonsToEvents(commandBinding, string(status), path)
	if strings.Contains(organizer, ":") {
		organizer = strings.Split(organizer, ":")[1]
	}
	if organizerEmail == organizer {
		deletePath := fmt.Sprintf("/delete-event/%s/events/%s", calendarId, eventId)
		createDeleteButton(&commandBinding, "Delete", "Delete", deletePath)
	}
	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)

	return &post
}

func createDescriptionForEvent(description string, start string, finish string, organizer string, attendees string) string {
	return fmt.Sprintf("Description %s. Organized by %s. Attendies: %s. Start date: %s, End date: %s", description, organizer, attendees, start, finish)
}

func createDeleteButton(commandBinding *apps.Binding, location apps.Location, label string, deletePath string) {
	expand := apps.Expand{
		OAuth2App:             apps.ExpandAll,
		OAuth2User:            apps.ExpandAll,
		ActingUserAccessToken: apps.ExpandAll,
		ActingUser:            apps.ExpandAll,
	}
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{

		Location: location,
		Label:    label,
		Submit:   apps.NewCall(deletePath).WithExpand(expand),
	})
}

func HandleChangeEventStatus(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	accessToken := token.AccessToken
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	user, _, _ := asActingUser.GetUser(creq.Context.ActingUser.Id, "")

	eventId := c.Param("eventId")
	status := strings.ToUpper(c.Param("status"))
	calendarId := c.Param("calendarId")
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s/%s.ics", remoteUrl, userId, calendarId, eventId)

	calendarService := CalendarServiceImpl{Url: reqUrl, Token: accessToken}

	eventIcs := calendarService.GetCalendarEvent(calendarId, eventId)

	cal, _ := ics.ParseCalendar(strings.NewReader(eventIcs))

	body := calendarService.UpdateAttendeeStatus(cal, user, status)
	calendarService.CreateEvent(body)
	c.JSON(http.StatusOK, apps.NewTextResponse("event status updated:"+status))

}

func HandleGetUserCalendars(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	accessToken := token.AccessToken
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s", remoteUrl, userId)

	calendarService := CalendarServiceImpl{Url: reqUrl, Token: accessToken}

	userCalendars := calendarService.GetUserCalendars()

	asBot := appclient.AsBot(creq.Context)
	userSettingsService := user.UserSettingsServiceImpl{asBot}

	for i, c := range userCalendars {
		us := userSettingsService.GetUserSettingsById(creq.Context.ActingUser.Id)
		post := createCalendarPost(i, c, us.Contains(c.Value))
		asBot.DMPost(creq.Context.ActingUser.Id, post)
	}

	c.JSON(http.StatusOK, apps.NewTextResponse("send calendars to DM:"))

}

func createCalendarPost(i int, option apps.SelectOption, disabled bool) *model.Post {
	post := model.Post{}
	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       "Calendar " + strconv.Itoa(i),
		Description: option.Label,
		Bindings:    []apps.Binding{},
	}

	createCalendarEventsButton(&commandBinding, option, "Calendar", "Create calendar event")
	createGetCalendarEventsButton(&commandBinding, option, "Calendar", "Calendar events")
	if disabled {
		createDoNotDisturbButton(&commandBinding, option, "Enable", "Enable notifications")

	} else {
		createDoNotDisturbButton(&commandBinding, option, "Disable", "Disable notifications")
	}

	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)
	return &post

}

func createDoNotDisturbButton(commandBinding *apps.Binding, option apps.SelectOption, location apps.Location, label string) {
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Submit: apps.NewCall(fmt.Sprintf("/calendars/%s/status/%s", option.Value, location)).WithExpand(apps.Expand{
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			ActingUserAccessToken: apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}),
	})
}

func createGetCalendarEventsButton(commandBinding *apps.Binding, option apps.SelectOption, location apps.Location, label string) {
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Submit: apps.NewCall("/get-calendar-events-form").WithExpand(apps.Expand{
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			ActingUserAccessToken: apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}).WithState(option),
	})
}

func createCalendarEventsButton(commandBinding *apps.Binding, option apps.SelectOption, location apps.Location, label string) {
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Submit: apps.NewCall("/create-calendar-event-form").WithExpand(apps.Expand{
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			ActingUserAccessToken: apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}).WithState(option),
	})
}
