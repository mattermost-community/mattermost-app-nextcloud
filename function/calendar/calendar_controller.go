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
	duration := creq.Values["duration"].(map[string]interface{})["value"].(string)

	var timezone string

	if creq.Context.ActingUser.Timezone["useAutomaticTimezone"] == "false" {
		timezone = creq.Context.ActingUser.Timezone["manualTimezone"]
	} else {
		timezone = creq.Context.ActingUser.Timezone["automaticTimezone"]
	}
	uuid, body := calendarEventService.CreateEventBody(fromDateUTC, duration, timezone)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	calendar := creq.Values["calendar"].(map[string]interface{})["value"].(string)

	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s/%s.ics", remoteUrl, userId, calendar, uuid)

	calendarService := CalendarServiceImpl{Url: reqUrl, Token: accessToken}
	calendarService.CreateEvent(body)
	c.JSON(http.StatusOK, apps.NewTextResponse("event created: "+reqUrl))
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

	loc := getMMUserLocation(creq)
	currentUserTime := time.Now().In(loc)

	dateFormatService := DateFormatLocaleService{}
	parsedLocale := dateFormatService.GetLocaleByTag(creq.Context.ActingUser.Locale)

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
				Type:        apps.FieldTypeDynamicSelect,
				Name:        "from-event-date",
				Label:       "From",
				IsRequired:  true,
				Description: "Type \"Today\", \"Tomorrow\" or \"Monday 13:00\" to choose a date",
				Value:       apps.SelectOption{Label: currentUserTime.Format(dateFormatService.GetDateTimeFormatsByLocale(parsedLocale)), Value: currentUserTime.String()},
				SelectDynamicLookup: apps.NewCall("/get-parsed-date").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
					Channel:               apps.ExpandAll,
					ActingUser:            apps.ExpandAll,
				}),
			},
			{
				Type:                apps.FieldTypeStaticSelect,
				Name:                "duration",
				Label:               "Duration",
				IsRequired:          true,
				SelectStaticOptions: prepareMeetingDurations(),
				Value:               apps.SelectOption{Label: "30 minutes", Value: "30 minutes"},
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
				IsRequired:    false,
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

func GetUserSelectedEventsDate(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	calendar := creq.Call.State.(map[string]interface{})["value"].(string)
	loc := getMMUserLocation(creq)
	currentUserTime := time.Now().In(loc)

	dateFormatService := DateFormatLocaleService{}
	parsedLocale := dateFormatService.GetLocaleByTag(creq.Context.ActingUser.Locale)

	form := &apps.Form{
		Title: "Nextcloud calendar events",
		Icon:  "icon.png",
		Fields: []apps.Field{
			{
				Type:        apps.FieldTypeDynamicSelect,
				Name:        "from-event-date",
				Label:       "Date",
				IsRequired:  true,
				Description: "Type \"Today\", \"Tomorrow\" or \"Monday 13:00\" to choose a date",
				Value:       apps.SelectOption{Label: currentUserTime.Format(dateFormatService.GetDateTimeFormatsByLocale(parsedLocale)), Value: currentUserTime.String()},
				SelectDynamicLookup: apps.NewCall("/get-parsed-date").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
					Channel:               apps.ExpandAll,
					ActingUser:            apps.ExpandAll,
				}),
			},
		},
		Submit: apps.NewCall("/get-calendar-events-select-date/" + calendar).WithExpand(apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			Channel:               apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}),
	}
	c.JSON(http.StatusOK, apps.NewFormResponse(*form))
}

func HandleGetEventsToday(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	location := getMMUserLocation(creq)
	calendar := creq.Call.State.(map[string]interface{})["value"].(string)
	HandleGetEvents(c, creq, time.Now().In(location), calendar)
}

func HandleGetEventsTomorrow(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	location := getMMUserLocation(creq)
	calendar := creq.Call.State.(map[string]interface{})["value"].(string)
	HandleGetEvents(c, creq, time.Now().AddDate(0, 0, 1).In(location), calendar)
}

func HandleGetEventsAtSelectedDay(c *gin.Context) {
	calendar := c.Param("calendar")
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	location := getMMUserLocation(creq)
	fromDateUTC := creq.Values["from-event-date"].(map[string]interface{})["value"].(string)
	from, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", fromDateUTC)
	if err != nil {
		println(err.Error())
	}
	HandleGetEvents(c, creq, from.In(location), calendar)
}

func getMMUserLocation(creq apps.CallRequest) *time.Location {
	var timezone string
	var loc *time.Location
	if creq.Context.ActingUser.Timezone["useAutomaticTimezone"] == "false" {
		timezone = creq.Context.ActingUser.Timezone["manualTimezone"]
	} else {
		timezone = creq.Context.ActingUser.Timezone["automaticTimezone"]
	}
	loc, _ = time.LoadLocation(timezone)
	return loc
}

func HandleGetEvents(c *gin.Context, creq apps.CallRequest, date time.Time, calendar string) {
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s", remoteUrl, userId, calendar)

	loc := getMMUserLocation(creq)

	asBot := appclient.AsBot(creq.Context)
	mmUserId := creq.Context.ActingUser.Id
	organizerEmail := creq.Context.ActingUser.Email

	from, to := prepareTimeRangeForGetEventsRequest(date)
	eventRange := CalendarEventRequestRange{
		From: from,
		To:   to,
	}
	calendarService := CalendarServiceImpl{Url: reqUrl, Token: token.AccessToken}

	events, eventIds := calendarService.GetCalendarEvents(eventRange)
	calendarEvents := make([]ics.VEvent, 0)

	for i := 0; i < len(events); i++ {
		cal, _ := ics.ParseCalendar(strings.NewReader(events[i]))
		event := *cal.Events()[0]
		if len(event.Properties) != 0 {
			calendarEvents = append(calendarEvents, event)
		}
	}

	dailyCalendarEvents := make([]ics.VEvent, 0)

	for _, e := range calendarEvents {
		at, _ := e.GetStartAt()
		endAt, _ := e.GetEndAt()
		localStartTime := at.In(loc)
		localEndTime := endAt.In(loc)
		if localStartTime.Day() == date.Day() || localEndTime.Day() == date.Day() {
			dailyCalendarEvents = append(dailyCalendarEvents, e)
		}
	}

	if len(dailyCalendarEvents) == 0 {
		c.JSON(http.StatusOK, apps.NewTextResponse("You don`t have events at this date"))
		return
	}

	for i, e := range dailyCalendarEvents {
		status := findAttendeeStatus(asBot, e, creq.Context.ActingUser.Id)
		postDto := CalendarEventPostDTO{&e, status, *asBot, calendar, organizerEmail, eventIds[i], userId, loc, creq.Context.OAuth2.RemoteRootURL, creq.Context.ActingUser.Locale}
		post := createCalendarEventPost(&postDto)
		asBot.DMPost(mmUserId, post)
	}

	c.JSON(http.StatusOK, apps.NewDataResponse(nil))
}

func prepareTimeRangeForGetEventsRequest(chosenDate time.Time) (time.Time, time.Time) {
	date := chosenDate.Add(-time.Minute * time.Duration(chosenDate.Minute()))
	date = date.Add(-time.Hour * time.Duration(chosenDate.Hour()))
	date = date.Add(-time.Second * time.Duration(chosenDate.Second()))
	return date.AddDate(0, 0, -1), date.AddDate(0, 0, 1)
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

func createCalendarEventPost(postDTO *CalendarEventPostDTO) *model.Post {
	var name, description, organizer, eventStatus string
	for _, e := range postDTO.event.Properties {
		if e.BaseProperty.IANAToken == "DESCRIPTION" {
			description = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "ORGANIZER" {
			organizer = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "SUMMARY" {
			name = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "STATUS" {
			eventStatus = e.BaseProperty.Value
		}
	}

	post := model.Post{}
	commandBinding := apps.Binding{
		Location: "embedded",
		AppID:    "nextcloud",
		Label:    createNameForEvent(name, postDTO),
		Description: сreateDescriptionForEvent(description, сastSingleEmailToMMUserNickname(organizer, "", postDTO.bot),
			сastUserEmailsToMMUserNicknames(postDTO.event.Attendees(), postDTO.bot)),
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

	if postDTO.organizerEmail != organizer {
		path := fmt.Sprintf("/users/%s/calendars/%s/events/%s/status", postDTO.userId, postDTO.calendarId, postDTO.eventId)
		commandBinding = calendarService.AddButtonsToEvents(commandBinding, string(postDTO.status), path)
	}

	if postDTO.organizerEmail == organizer {
		deletePath := fmt.Sprintf("/delete-event/%s/events/%s", postDTO.calendarId, postDTO.eventId)
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
	if len(attendeesNicknames) != 0 {
		attendeesNicknames = attendeesNicknames[:len(attendeesNicknames)-1]
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

func createNameForEvent(name string, postDTO *CalendarEventPostDTO) string {
	dateFormatService := DateFormatLocaleService{}
	parsedLocale := dateFormatService.GetLocaleByTag(postDTO.locale)
	start, _ := postDTO.event.GetStartAt()
	finish, _ := postDTO.event.GetEndAt()

	format := dateFormatService.GetTimeFormatsByLocale(parsedLocale)
	dayFormat := dateFormatService.GetFullFormatsByLocale(parsedLocale)
	day := strconv.Itoa(start.Day())
	month := strconv.Itoa(int(start.Month()))
	if len(day) < 2 {
		day = "0" + day
	}
	if len(month) < 2 {
		month = "0" + month
	}
	calendarUrl := fmt.Sprintf("%s%s%s-%s-%s", postDTO.remoteUrl, "/apps/calendar/timeGridDay/", strconv.Itoa(start.Year()), month, day)
	return fmt.Sprintf("[%s](%s) %s %s-%s", name, calendarUrl, start.In(postDTO.loc).Format(dayFormat), start.In(postDTO.loc).Format(format), finish.In(postDTO.loc).Format(format))
}

func сreateDescriptionForEvent(description string, organizer string, attendees string) string {
	finalDesc := ""

	if len(description) != 0 {
		finalDesc += fmt.Sprintf("Description %s. ", description)
	}
	finalDesc += fmt.Sprintf("Organized by %s. ", organizer[:len(organizer)-1])
	if len(attendees) != 0 {
		finalDesc += fmt.Sprintf("Attendees: %s. ", attendees)
	}

	return finalDesc
}

func prepareMeetingDurations() []apps.SelectOption {
	var durations []apps.SelectOption
	durations = append(durations, apps.SelectOption{
		Label: "15 minutes",
		Value: "15 minutes",
	})
	durations = append(durations, apps.SelectOption{
		Label: "30 minutes",
		Value: "30 minutes",
	})
	durations = append(durations, apps.SelectOption{
		Label: "45 minutes",
		Value: "45 minutes",
	})
	durations = append(durations, apps.SelectOption{
		Label: "1 hour",
		Value: "1 hour",
	})
	durations = append(durations, apps.SelectOption{
		Label: "1.5 hours",
		Value: "1.5 hours",
	})
	durations = append(durations, apps.SelectOption{
		Label: "2 hours",
		Value: "2 hours",
	})
	durations = append(durations, apps.SelectOption{
		Label: "All day",
		Value: "All day",
	})
	return durations
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
		dateFormatService := DateFormatLocaleService{}
		parsedLocale := dateFormatService.GetLocaleByTag(creq.Context.ActingUser.Locale)
		so = apps.SelectOption{Label: t.Format(dateFormatService.GetDateTimeFormatsByLocale(parsedLocale)), Value: t.String()}
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

	if len(userCalendars) == 0 {
		c.JSON(http.StatusOK, apps.NewTextResponse("You don`t have any calendars"))
		return
	}

	asBot := appclient.AsBot(creq.Context)
	userSettingsService := user.UserSettingsServiceImpl{asBot}

	for _, c := range userCalendars {
		us := userSettingsService.GetUserSettingsById(creq.Context.ActingUser.Id)
		post := createCalendarPost(c, us.Contains(c.Value))
		asBot.DMPost(creq.Context.ActingUser.Id, post)
	}
	c.JSON(http.StatusOK, apps.NewTextResponse(""))

}

func createCalendarPost(option apps.SelectOption, disabled bool) *model.Post {
	post := model.Post{}
	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       "Calendar " + option.Label,
		Description: "Calendar actions",
		Bindings:    []apps.Binding{},
	}

	createGetCalendarEventsButton(&commandBinding, option, "Calendar", "Today", "today")
	createGetCalendarEventsButton(&commandBinding, option, "Calendar", "Tomorrow", "tomorrow")
	createGetCalendarEventsButton(&commandBinding, option, "Calendar", "Select date", "select-date-form")
	createCalendarEventsButton(&commandBinding, option, "Calendar", "Create event")

	//if disabled {
	//	createDoNotDisturbButton(&commandBinding, option, "Enable", "Enable notifications")
	//
	//} else {
	//	createDoNotDisturbButton(&commandBinding, option, "Disable", "Disable notifications")
	//}

	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)
	return &post

}

//func createDoNotDisturbButton(commandBinding *apps.Binding, option apps.SelectOption, location apps.Location, label string) {
//	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
//		Location: location,
//		Label:    label,
//		Submit: apps.NewCall(fmt.Sprintf("/calendars/%s/status/%s", option.Value, location)).WithExpand(apps.Expand{
//			OAuth2App:             apps.ExpandAll,
//			OAuth2User:            apps.ExpandAll,
//			ActingUserAccessToken: apps.ExpandAll,
//			ActingUser:            apps.ExpandAll,
//		}),
//	})
//}

func createGetCalendarEventsButton(commandBinding *apps.Binding, option apps.SelectOption, location apps.Location, label string, day string) {
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Submit: apps.NewCall("/get-calendar-events-" + day).WithExpand(apps.Expand{
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
