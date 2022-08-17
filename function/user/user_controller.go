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
