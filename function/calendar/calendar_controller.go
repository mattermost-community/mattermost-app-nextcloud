package calendar

import (
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/jarylc/go-chrono/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
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

	calendarEventService := CalendarEventServiceImpl{creq, asActingUser}
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

	calendarRequestService := CalendarRequestServiceImpl{Url: reqUrl, Token: accessToken}
	calendarService := CalendarServiceImpl{calendarRequestService: calendarRequestService}

	_, err := calendarService.CreateEvent(body)

	if err != nil {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: "Calendar event was not created"})
		return
	}

	DMEventPost(creq, calendarService, calendar, uuid)
	c.JSON(http.StatusOK, apps.NewTextResponse(""))
}

func DMEventPost(creq apps.CallRequest, calendarService CalendarService, calendar string, uuid string) {
	asBot := appclient.AsBot(creq.Context)

	event, eventError := calendarService.GetCalendarEvent()
	if eventError != nil {
		log.Error("Event was not found", calendarService.GetUrl())
		return
	}
	cal, parseError := ics.ParseCalendar(strings.NewReader(event))
	if parseError != nil {
		log.Errorf("Can't parse calendar for event %s", calendarService.GetUrl())
		return
	}
	vEvent := cal.Events()[0]
	calendarTimePostService := CalendarTimePostService{}

	loc := calendarTimePostService.GetMMUserLocation(creq)

	createCalendarEventPostService := CreateCalendarEventPostService{GetMMUser: asBot}
	postDto := CalendarEventPostDTO{vEvent, asBot, calendar, uuid + ".ics", loc, creq}

	post := createCalendarEventPostService.CreateCalendarEventPost(&postDto)
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

	calendarRequestService := CalendarRequestServiceImpl{Url: reqUrl, Token: accessToken}
	calendarService := CalendarServiceImpl{calendarRequestService: calendarRequestService}
	option := creq.State.(map[string]interface{})

	calendarTimePostService := CalendarTimePostService{}

	loc := calendarTimePostService.GetMMUserLocation(creq)

	currentUserTime := time.Now().In(loc)

	dateFormatService := DateFormatLocaleService{}
	parsedLocale := dateFormatService.GetLocaleByTag(creq.Context.ActingUser.Locale)
	calendarPostServiceImpl := CalendarPostServiceImpl{}

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
				SelectStaticOptions: calendarPostServiceImpl.PrepareMeetingDurations(),
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
	calendarId := c.Param("calendarId")
	eventId := c.Param("eventId")
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	user := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	deleteUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s/%s", remoteUrl, user, calendarId, eventId)

	calendarRequestService := CalendarRequestServiceImpl{Url: deleteUrl, Token: token.AccessToken}
	calendarService := CalendarServiceImpl{calendarRequestService: calendarRequestService}
	_, err := calendarService.DeleteUserEvent()

	if err != nil {
		c.JSON(http.StatusOK, apps.NewErrorResponse(errors.New("Event was not deleted")))
	}

	c.JSON(http.StatusOK, apps.NewTextResponse("Event deleted"))
}

func GetUserSelectedEventsDate(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	calendar := creq.Call.State.(map[string]interface{})["value"].(string)
	calendarTimePostService := CalendarTimePostService{}

	loc := calendarTimePostService.GetMMUserLocation(creq)
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
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	calendar := creq.Call.State.(map[string]interface{})["value"].(string)
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s", remoteUrl, userId, calendar)
	asBot := appclient.AsBot(creq.Context)

	calendarTimePostService := CalendarTimePostService{}
	calendarRequestService := CalendarRequestServiceImpl{Url: reqUrl, Token: token.AccessToken}
	calendarService := CalendarServiceImpl{calendarRequestService: calendarRequestService}
	calendarPostServiceImpl := CreateCalendarEventPostService{GetMMUser: asBot}

	location := calendarTimePostService.GetMMUserLocation(creq)

	service := GetEventsService{calendarService, calendarTimePostService, calendarPostServiceImpl, asBot}
	err := service.GetUserEvents(creq, time.Now().In(location), calendar)
	if err != nil {
		c.JSON(http.StatusOK, apps.NewTextResponse("You don`t have events at this date"))
		return
	}
	c.JSON(http.StatusOK, apps.NewDataResponse(nil))
}

func HandleGetEventsTomorrow(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	calendar := creq.Call.State.(map[string]interface{})["value"].(string)
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s", remoteUrl, userId, calendar)
	asBot := appclient.AsBot(creq.Context)

	calendarTimePostService := CalendarTimePostService{}
	calendarRequestService := CalendarRequestServiceImpl{Url: reqUrl, Token: token.AccessToken}
	calendarService := CalendarServiceImpl{calendarRequestService: calendarRequestService}
	calendarPostServiceImpl := CreateCalendarEventPostService{GetMMUser: asBot}

	location := calendarTimePostService.GetMMUserLocation(creq)

	service := GetEventsService{calendarService, calendarTimePostService, calendarPostServiceImpl, asBot}
	err := service.GetUserEvents(creq, time.Now().AddDate(0, 0, 1).In(location), calendar)
	if err != nil {
		c.JSON(http.StatusOK, apps.NewTextResponse("You don`t have events at this date"))
		return
	}
	c.JSON(http.StatusOK, apps.NewDataResponse(nil))
}

func HandleGetEventsAtSelectedDay(c *gin.Context) {
	calendar := c.Param("calendar")
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s", remoteUrl, userId, calendar)
	asBot := appclient.AsBot(creq.Context)

	calendarTimePostService := CalendarTimePostService{}
	calendarRequestService := CalendarRequestServiceImpl{Url: reqUrl, Token: token.AccessToken}
	calendarService := CalendarServiceImpl{calendarRequestService: calendarRequestService}
	calendarPostServiceImpl := CreateCalendarEventPostService{GetMMUser: asBot}

	location := calendarTimePostService.GetMMUserLocation(creq)

	service := GetEventsService{calendarService, calendarTimePostService, calendarPostServiceImpl, asBot}

	fromDateUTC := creq.Values["from-event-date"].(map[string]interface{})["value"].(string)
	from, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", fromDateUTC)
	if err != nil {
		println(err.Error())
	}
	getEventErr := service.GetUserEvents(creq, from.In(location), calendar)
	if getEventErr != nil {
		c.JSON(http.StatusOK, apps.NewTextResponse("You don`t have events at this date"))
		return
	}
	c.JSON(http.StatusOK, apps.NewDataResponse(nil))
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

	calendarRequestService := CalendarRequestServiceImpl{Url: reqUrl, Token: accessToken}
	calendarService := CalendarServiceImpl{calendarRequestService: calendarRequestService}

	eventIcs, _ := calendarService.GetCalendarEvent()

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

	calendarRequestService := CalendarRequestServiceImpl{Url: reqUrl, Token: accessToken}
	calendarService := CalendarServiceImpl{calendarRequestService}

	userCalendars := calendarService.GetUserCalendars()

	if len(userCalendars) == 0 {
		c.JSON(http.StatusOK, apps.NewTextResponse("You don`t have any calendars"))
		return
	}

	asBot := appclient.AsBot(creq.Context)

	calendarPostServiceImpl := CalendarPostServiceImpl{}

	for _, c := range userCalendars {
		post := calendarPostServiceImpl.CreateCalendarPost(c)
		asBot.DMPost(creq.Context.ActingUser.Id, post)
	}
	c.JSON(http.StatusOK, apps.NewTextResponse(""))

}
