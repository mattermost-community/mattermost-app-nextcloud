package user

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/prokhorind/nextcloud/function/oauth"
	"net/http"
	"strconv"
)

func HandleUserDoNotDisturbMode(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	accessToken := token.AccessToken

	status := creq.Values["enabled"].(bool)

	oauth.ConfigureWebhooks(creq, accessToken, !status)

	c.JSON(http.StatusOK, apps.NewTextResponse("do not disturb  mode enabled:"+strconv.FormatBool(status)))
}

func HandleCalendarDoNotDisturbMode(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	status := c.Param("status")
	calendarId := c.Param("calendarId")

	asBot := appclient.AsBot(creq.Context)

	userSettingsService := UserSettingsServiceImpl{asBot}
	userId := creq.Context.ActingUser.Id

	us := userSettingsService.GetUserSettingsById(userId)

	if status == "Enable" {
		userSettingsService.SetUserSettingsById(userId, us.RemoveDisabledCalendar(calendarId))

	} else {
		userSettingsService.SetUserSettingsById(userId, us.AddDisabledCalendar(calendarId))
	}

	c.JSON(http.StatusOK, apps.NewTextResponse("Calendar status updated:"+status))

}
