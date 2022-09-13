package calendar

import (
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"net/http"
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

func HandleGetEventsForm(c *gin.Context) {
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

	form := &apps.Form{
		Title: "Create Nextcloud calendar event",
		Icon:  "icon.png",
		Fields: []apps.Field{

			{
				Type:                "static_select",
				Name:                "calendar",
				Label:               "Calendar",
				IsRequired:          true,
				SelectStaticOptions: calendarService.GetUserCalendars(),
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

	events := calendarService.GetCalendarEvents(eventRange)

	asBot := appclient.AsBot(creq.Context)
	for _, e := range events {

		post := model.Post{
			Message:   e,
			ChannelId: creq.Context.Channel.Id,
		}
		asBot.CreatePost(&post)

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
