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
