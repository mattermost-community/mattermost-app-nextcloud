package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type OauthService interface {
	RefreshToken() (*Token, error)
}

type OauthServiceImpl struct {
	Creq apps.CallRequest
}

func (s OauthServiceImpl) RefreshToken() (*Token, error) {

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
	resp, err := client.Do(req)
	defer resp.Body.Close()

	log.Infof("refresh token response status %s", resp.Status)
	if err != nil {
		log.Errorf("Error during refreshing of the token. Error: %s", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Errorf("refresh token response status code %d for user", resp.StatusCode)
		return nil, errors.New("Request for Nextcloud token refresh is failed")
	}

	jsonResp := Token{}
	json.NewDecoder(resp.Body).Decode(&jsonResp)
	log.Infof("refresh token response status code %d for user %s", resp.StatusCode, jsonResp.UserID)

	return &jsonResp, nil

}

func getToken(creq apps.CallRequest) (*Token, error) {
	code, _ := creq.Values["code"].(string)

	clientId := creq.Context.OAuth2.OAuth2App.ClientID
	clientSecret := creq.Context.OAuth2.OAuth2App.ClientSecret
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL

	reqUrl := fmt.Sprintf("%s/index.php/apps/oauth2/api/v1/token", remoteUrl)

	payload := RequestTokenBody{
		Code:      code,
		GrantType: "authorization_code",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", reqUrl, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(clientId, clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error during getting of the token. Error: %s", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Errorf("get token response status code %d for user", resp.StatusCode)
		return nil, errors.New("Request for Nextcloud token is failed")
	}

	defer resp.Body.Close()
	jsonResp := Token{}
	json.NewDecoder(resp.Body).Decode(&jsonResp)
	return &jsonResp, nil
}
