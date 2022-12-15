package calendar

import (
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/jarylc/go-chrono/v2"
	"github.com/prokhorind/nextcloud/function/user"
	log "github.com/sirupsen/logrus"
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
	fromDateUTC := creq.Values["from-event-date"].(map[string]interface{})["value"].(string)
	toDateUTC := creq.Values["to-event-date"].(map[string]interface{})["value"].(string)

	var timezone string

	if creq.Context.ActingUser.Timezone["useAutomaticTimezone"] == "false" {
		timezone = creq.Context.ActingUser.Timezone["manualTimezone"]
	} else {
		timezone = creq.Context.ActingUser.Timezone["automaticTimezone"]
	}
	uuid, body := calendarEventService.CreateEventBody(fromDateUTC, toDateUTC, timezone)

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
				Type:       apps.FieldTypeText,
				Name:       "title",
				Label:      "Title",
				IsRequired: true,
			},
			{
				Type:       apps.FieldTypeDynamicSelect,
				Name:       "from-event-date",
				Label:      "From",
				IsRequired: true,
				SelectDynamicLookup: apps.NewCall("/get-parsed-date").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
					Channel:               apps.ExpandAll,
					ActingUser:            apps.ExpandAll,
				}),
			},
			{
				Type:       apps.FieldTypeDynamicSelect,
				Name:       "to-event-date",
				Label:      "To",
				IsRequired: true,
				SelectDynamicLookup: apps.NewCall("/get-parsed-date").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
					Channel:               apps.ExpandAll,
					ActingUser:            apps.ExpandAll,
				}),
			},
			{
				Type:        apps.FieldTypeText,
				Name:        "description",
				Label:       "Description",
				TextSubtype: apps.TextFieldSubtypeTextarea,
				IsRequired:  false,
			},
			{
				Type:          apps.FieldTypeUser,
				Name:          "attendees",
				Label:         "Attendees",
				IsRequired:    true,
				SelectIsMulti: true,
			},

			{
				Type:                apps.FieldTypeStaticSelect,
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
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	user := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	deleteUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s/%s", remoteUrl, user, calendarId, eventId)

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
				Type:       apps.FieldTypeDynamicSelect,
				Name:       "from-event-date",
				Label:      "From",
				IsRequired: true,
				SelectDynamicLookup: apps.NewCall("/get-parsed-date").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
					Channel:               apps.ExpandAll,
					ActingUser:            apps.ExpandAll,
				}),
			},
			{
				Type:       apps.FieldTypeDynamicSelect,
				Name:       "to-event-date",
				Label:      "To",
				IsRequired: true,
				SelectDynamicLookup: apps.NewCall("/get-parsed-date").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
					Channel:               apps.ExpandAll,
					ActingUser:            apps.ExpandAll,
				}),
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

	fromDateUTC := creq.Values["from-event-date"].(map[string]interface{})["value"].(string)
	toDateUTC := creq.Values["to-event-date"].(map[string]interface{})["value"].(string)

	from, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", fromDateUTC)
	to, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", toDateUTC)

	eventRange := CalendarEventRequestRange{
		From: from,
		To:   to,
	}
	calendarService := CalendarServiceImpl{Url: reqUrl, Token: token.AccessToken}

	events, eventIds := calendarService.GetCalendarEvents(eventRange)
	if len(events) == 0 {
		c.JSON(http.StatusOK, apps.NewTextResponse("You do not have events in this calendar"))
		return
	}
	calEvents := make([]ics.Calendar, len(events))

	for i := 0; i < len(events); i++ {
		cal, _ := ics.ParseCalendar(strings.NewReader(events[i]))
		calEvents[i] = *cal
	}

	asBot := appclient.AsBot(creq.Context)
	mmUserId := creq.Context.ActingUser.Id
	organizerEmail := creq.Context.ActingUser.Email
	for i, e := range calEvents {
		status := findAttendeeStatus(asBot, *e.Events()[0], creq.Context.ActingUser.Id)
		post := createCalendarEventPost(e.Events()[0], status, *asBot, calendar, organizerEmail, eventIds[i], userId)
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

func createCalendarEventPost(event *ics.VEvent, status ics.ParticipationStatus, bot appclient.Client, calendarId string, organizerEmail string, eventId string, userId string) *model.Post {
	var name, start, finish, description, organizer, eventStatus string
	for _, e := range event.Properties {
		if e.BaseProperty.IANAToken == "DESCRIPTION" {
			description = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "ORGANIZER" {
			organizer = e.BaseProperty.Value
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
		if e.BaseProperty.IANAToken == "STATUS" {
			eventStatus = e.BaseProperty.Value
		}
	}
	post := model.Post{}
	commandBinding := apps.Binding{
		Location: "embedded",
		AppID:    "nextcloud",
		Label:    "Event " + name,
		Description: сreateDescriptionForEvent(description, сastDateToSpecificFormat(start, "Jan _2 15:04:05"),
			сastDateToSpecificFormat(finish, "Jan _2 15:04:05"), сastSingleEmailToMMUserNickname(organizer, "", bot),
			сastUserEmailsToMMUserNicknames(event.Attendees(), bot)),
		Bindings: []apps.Binding{},
	}
	calendarService := CalendarServiceImpl{}

	if eventStatus == "CANCELLED" {
		commandBinding.Label = fmt.Sprintf("~~%s~~", commandBinding.Label)
		commandBinding.Description = fmt.Sprintf("~~%s~~", commandBinding.Description)
		m1 := make(map[string]interface{})
		m1["app_bindings"] = []apps.Binding{commandBinding}

		post.SetProps(m1)

		return &post
	}

	if strings.Contains(organizer, ":") {
		organizer = strings.Split(organizer, ":")[1]
	}

	if organizerEmail != organizer {
		path := fmt.Sprintf("/users/%s/calendars/%s/events/%s/status", userId, calendarId, eventId)
		commandBinding = calendarService.AddButtonsToEvents(commandBinding, string(status), path)
	}

	if organizerEmail == organizer {
		deletePath := fmt.Sprintf("/delete-event/%s/events/%s", calendarId, eventId)
		сreateDeleteButton(&commandBinding, "Delete", "Delete", deletePath)
	}
	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)

	return &post
}

func сastUserEmailsToMMUserNicknames(attendees []*ics.Attendee, bot appclient.Client) string {
	var attendeesNicknames string
	for _, attendee := range attendees {
		attendeesNicknames += сastSingleEmailToMMUserNickname(attendee.Email(), attendee.ICalParameters["PARTSTAT"][0], bot)
	}
	return attendeesNicknames
}

func сastSingleEmailToMMUserNickname(email string, status string, bot appclient.Client) string {
	if strings.Contains(email, ":") {
		email = strings.Split(email, ":")[1]
	}
	mmUser, _, err := bot.GetUserByEmail(email, "")
	if err == nil {
		if status == "" {
			return "@" + mmUser.Username + " "
		}
		return "@" + mmUser.Username + " - " + status + " "
	} else {
		return email + " - " + " "
	}
}

func сastDateToSpecificFormat(dateStr string, outputFormat string) string {
	date, error := time.Parse(icalTimestampFormatUtc, dateStr)

	if error != nil {
		date, _ := time.Parse(icalTimestampFormatUtcLocal, dateStr)
		return date.Format(outputFormat)
	}

	return date.Format(outputFormat)
}

func сreateDescriptionForEvent(description string, start string, finish string, organizer string, attendees string) string {
	return fmt.Sprintf("Description %s. Organized by %sAttendies: %s. Start date: %s, End date: %s", description, organizer, attendees, start, finish)
}

func сreateDeleteButton(commandBinding *apps.Binding, location apps.Location, label string, deletePath string) {
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
	userId := c.Param("userId")

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s/%s", remoteUrl, userId, calendarId, eventId)

	calendarService := CalendarServiceImpl{Url: reqUrl, Token: accessToken}

	eventIcs := calendarService.GetCalendarEvent()

	cal, _ := ics.ParseCalendar(strings.NewReader(eventIcs))

	body := calendarService.UpdateAttendeeStatus(cal, user, status)
	calendarService.CreateEvent(body)
	c.JSON(http.StatusOK, apps.NewTextResponse("event status updated:"+status))

}

func HandleGetParsedCalendarDate(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	ch, err := chrono.New()
	if err != nil {
		log.Error(err)
	}

	now := time.Now()

	t, err := ch.ParseDate(creq.Query, now)
	var so apps.SelectOption
	if err != nil || t == nil {
		so = apps.SelectOption{Label: "", Value: ""}
	} else {
		so = apps.SelectOption{Label: t.Format(time.ANSIC), Value: t.String()}
	}
	var soOptions []apps.SelectOption
	soOptions = append(soOptions, so)
	c.JSON(http.StatusOK, apps.NewLookupResponse(soOptions))
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
