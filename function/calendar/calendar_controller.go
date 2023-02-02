package calendar

import (
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/jarylc/go-chrono/v2"
	"github.com/prokhorind/nextcloud/function/user"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
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

	_, err := calendarService.CreateEvent(body)

	if err != nil {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: "Calendar event was not created"})
		return
	}

	DMEventPost(creq, calendarService, calendar, uuid)
	c.JSON(http.StatusOK, apps.NewTextResponse(""))
}

func DMEventPost(creq apps.CallRequest, calendarService CalendarServiceImpl, calendar string, uuid string) {
	asBot := appclient.AsBot(creq.Context)

	event := calendarService.GetCalendarEvent()
	cal, _ := ics.ParseCalendar(strings.NewReader(event))
	vEvent := cal.Events()[0]
	loc := getMMUserLocation(creq)

	postDto := CalendarEventPostDTO{vEvent, asBot, calendar, uuid + ".ics", loc, creq}
	post := createCalendarEventPost(&postDto)
	post.Message = "Event created"
	mmUserId := creq.Context.ActingUser.Id
	asBot.DMPost(mmUserId, post)
}

func RedirectToAMeeting(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	link := fmt.Sprint(creq.State)
	response := apps.CallResponse{Type: apps.CallResponseTypeNavigate, NavigateToURL: link}
	c.JSON(http.StatusOK, response)
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

func DoNothing(c *gin.Context) {
	c.JSON(http.StatusOK, apps.NewTextResponse(""))
	return
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
		postDto := CalendarEventPostDTO{&e, asBot, calendar, eventIds[i], loc, creq}
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

func createCalendarEventPost(postDTO *CalendarEventPostDTO) *model.Post {
	var name, organizer, eventStatus string
	for _, e := range postDTO.event.Properties {
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

	userId := postDTO.creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	remoteUrl := postDTO.creq.Context.OAuth2.OAuth2App.RemoteRootURL
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s/%s", remoteUrl, userId, postDTO.calendarId, postDTO.eventId)

	post := model.Post{}
	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       createNameForEvent(name, postDTO),
		Description: "Going?",
		Bindings:    []apps.Binding{},
	}
	calendarService := CalendarServiceImpl{}

	if eventStatus == "CANCELLED" {
		commandBinding.Label = fmt.Sprintf("Cancelled ~~%s~~", commandBinding.Label)
		commandBinding.Description = fmt.Sprintf("~~%s~~", commandBinding.Description)
		m1 := make(map[string]interface{})
		m1["app_bindings"] = []apps.Binding{commandBinding}

		post.SetProps(m1)

		return &post
	}

	if strings.Contains(organizer, ":") {
		organizer = strings.Split(organizer, ":")[1]
	}
	organizerEmail := postDTO.creq.Context.ActingUser.Email
	status := FindAttendeeStatus(postDTO.bot, *postDTO.event, postDTO.creq.Context.ActingUser.Id)

	if organizerEmail != organizer {
		path := fmt.Sprintf("/users/%s/calendars/%s/events/%s/status", userId, postDTO.calendarId, postDTO.eventId)
		commandBinding = calendarService.AddButtonsToEvents(commandBinding, string(status), path)
	}

	deletePath := fmt.Sprintf("/delete-event/%s/events/%s", postDTO.calendarId, postDTO.eventId)
	сreateDeleteButton(&commandBinding, "Delete", "Delete", deletePath)
	сreateViewButton(&commandBinding, "view-details", organizer, "View Details", postDTO, name, reqUrl)
	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)

	return &post
}

func createMeetingStartButton(commandBinding *apps.Binding, link string, location apps.Location) {
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    fmt.Sprintf("Join %s Meeting", location),
		Submit:   apps.NewCall("/redirect/meeting").WithState(link),
	})
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
			return "@" + mmUser.Username + "-" + email + " "
		}
		return "@" + mmUser.Username + "-" + email + "-" + status + " "
	} else {
		return email + "-" + status + " "
	}
}

func createDateForEventInForm(postDTO *CalendarEventPostDTO) string {
	locale := postDTO.creq.Context.ActingUser.Locale
	dateFormatService := DateFormatLocaleService{}
	parsedLocale := dateFormatService.GetLocaleByTag(locale)
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

	return fmt.Sprintf("%s %s-%s", start.In(postDTO.loc).Format(dayFormat), start.In(postDTO.loc).Format(format), finish.In(postDTO.loc).Format(format))
}

func createNameForEvent(name string, postDTO *CalendarEventPostDTO) string {
	locale := postDTO.creq.Context.ActingUser.Locale
	dateFormatService := DateFormatLocaleService{}
	parsedLocale := dateFormatService.GetLocaleByTag(locale)
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
	remoteUrl := postDTO.creq.Context.OAuth2.RemoteRootURL
	calendarUrl := fmt.Sprintf("%s%s%s-%s-%s", remoteUrl, "/apps/calendar/timeGridDay/", strconv.Itoa(start.Year()), month, day)
	return fmt.Sprintf("[%s](%s) %s %s-%s", name, calendarUrl, start.In(postDTO.loc).Format(dayFormat), start.In(postDTO.loc).Format(format), finish.In(postDTO.loc).Format(format))
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

func сreateViewButton(commandBinding *apps.Binding, location apps.Location, organizer string, label string, postDTO *CalendarEventPostDTO, name string, reqUrl string) {
	event := postDTO.event
	bot := postDTO.bot
	property := event.GetProperty(ics.ComponentPropertyDescription)
	var description string
	if property == nil {
		description = ""
	} else {
		description = property.Value
	}
	zoomLinks, googleMeetLinks := getZoomAndGoogleMeetLinksFromDescription(description)

	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Form: &apps.Form{
			Title: name,
			Fields: []apps.Field{
				{
					Type:       apps.FieldTypeText,
					Name:       "Date",
					Label:      "Date",
					ReadOnly:   true,
					Value:      createDateForEventInForm(postDTO),
					IsRequired: true,
				},
				{
					Type:        apps.FieldTypeText,
					Name:        "Description",
					Label:       "Description",
					ReadOnly:    true,
					Value:       description,
					TextSubtype: apps.TextFieldSubtypeTextarea,
				},
				{
					Type:                apps.FieldTypeStaticSelect,
					Name:                "Attendees",
					Label:               "Attendees",
					SelectIsMulti:       true,
					Value:               prepareAttendeeStaticSelect(сastUserEmailsToMMUserNicknames(event.Attendees(), *bot)),
					SelectStaticOptions: prepareAttendeeStaticSelect(сastUserEmailsToMMUserNicknames(event.Attendees(), *bot)),
					ReadOnly:            true,
				},
				{
					Type:       apps.FieldTypeText,
					Name:       "Organizer",
					Label:      "Organizer",
					ReadOnly:   true,
					IsRequired: true,
					Value:      сastSingleEmailToMMUserNickname(organizer, "", *bot),
				},
			},
			Submit: apps.NewCall("/do-nothing"),
		},
	})
	i := len(commandBinding.Bindings) - 1
	if len(zoomLinks) != 0 {
		commandBinding.Bindings[i].Form.Fields = append(commandBinding.Bindings[i].Form.Fields, apps.Field{
			Type:        apps.FieldTypeText,
			Name:        "ZoomUrl",
			Label:       "ZoomLink",
			Value:       zoomLinks,
			ReadOnly:    true,
			IsRequired:  true,
			TextSubtype: apps.TextFieldSubtypeURL,
		})
		createMeetingStartButton(commandBinding, strings.Split(zoomLinks, " ")[0], "Zoom")
	}
	if len(googleMeetLinks) != 0 {
		commandBinding.Bindings[i].Form.Fields = append(commandBinding.Bindings[i].Form.Fields, apps.Field{
			Type:        apps.FieldTypeText,
			Name:        "GoogleMeetUrl",
			Label:       "Google-Meet-Link",
			Value:       googleMeetLinks,
			ReadOnly:    true,
			IsRequired:  true,
			TextSubtype: apps.TextFieldSubtypeURL,
		})
		createMeetingStartButton(commandBinding, strings.Split(googleMeetLinks, " ")[0], "Google Meet")
	}
	commandBinding.Bindings[i].Form.Fields = append(commandBinding.Bindings[i].Form.Fields, apps.Field{
		Type:        apps.FieldTypeText,
		Name:        "Event-Import",
		Label:       "Event-Import",
		Description: "Use this link to import event in your calendars",
		Value:       reqUrl,
		ReadOnly:    true,
		IsRequired:  true,
		TextSubtype: apps.TextFieldSubtypeURL,
	})
}

func getZoomAndGoogleMeetLinksFromDescription(description string) (string, string) {
	zoomPattern := regexp.MustCompile(`https:\/\/[\w-]*\.?zoom.us\/(j|my)\/[\d\w?=-]+`)
	googleMeetPattern := regexp.MustCompile(`https?:\/\/(.+?\.)?meet\.google\.com(\/[A-Za-z0-9\-]*)?`)
	zoomLinks := zoomPattern.FindAllString(description, -1)
	googleMeetLinks := googleMeetPattern.FindAllString(description, -1)
	return strings.Join(zoomLinks, " "), strings.Join(googleMeetLinks, " ")
}

func prepareAttendeeStaticSelect(attendees string) []apps.SelectOption {
	options := make([]apps.SelectOption, 0)
	for _, a := range strings.Split(attendees, " ") {
		options = append(options, apps.SelectOption{Label: a, Value: a})
	}
	return options
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

	body, error := calendarService.UpdateAttendeeStatus(cal, user, status)
	if error != nil {
		c.JSON(http.StatusOK, apps.NewTextResponse("Event is no longer valid"))
		return
	}
	_, err := calendarService.CreateEvent(body)

	if err != nil {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: "Event status was not updated"})
		return
	}

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

	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)
	return &post

}

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
