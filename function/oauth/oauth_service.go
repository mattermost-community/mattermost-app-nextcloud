package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type OauthService interface {
	RefreshToken() Token
}

type OauthServiceImpl struct {
	Creq apps.CallRequest
}

func (s OauthServiceImpl) RefreshToken() Token {

	clientId := s.Creq.Context.OAuth2.OAuth2App.ClientID
	clientSecret := s.Creq.Context.OAuth2.OAuth2App.ClientSecret
	remoteUrl := s.Creq.Context.OAuth2.OAuth2App.RemoteRootURL

	reqUrl := fmt.Sprintf("%s/index.php/apps/oauth2/api/v1/token", remoteUrl)
	refreshToken := s.Creq.Context.OAuth2.User.(map[string]interface{})["refresh_token"].(string)

	payload := RefreshTokenBody{
		RefreshToken: refreshToken,
		GrantType:    "refresh_token",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", reqUrl, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(clientId, clientSecret)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	jsonResp := Token{}
	json.NewDecoder(resp.Body).Decode(&jsonResp)

	return jsonResp

}

func ConfigureWebhooks(creq apps.CallRequest, token string, status bool) {
	nextcloudRoot := creq.Context.ExpandedContext.OAuth2.OAuth2App.RemoteRootURL
	mmSiteUrl := creq.Context.MattermostSiteURL
	appId := creq.Context.AppID

	webhookUrl := fmt.Sprintf("%s/plugins/com.mattermost.apps/apps/%s/webhook", mmSiteUrl, appId)
	createEventWebhook := fmt.Sprintf("%s/%s", webhookUrl, "calendar-event-created")
	updateEventWebhook := fmt.Sprintf("%s/%s", webhookUrl, "calendar-event-updated")

	payload := CreateWebhooksBody{Enabled: status, WebhookSecret: "", CalendarEventCreatedURL: createEventWebhook, CalendarEventUpdatedURL: updateEventWebhook}
	body, _ := json.Marshal(payload)
	reqUrl := fmt.Sprintf("%s/index.php/apps/integration_mattermost/webhooks", nextcloudRoot)
	req, _ := http.NewRequest("POST", reqUrl, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}
