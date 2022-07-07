package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Connect struct {
	Poll struct {
		Token    string `json:"token"`
		Endpoint string `json:"endpoint"`
	} `json:"poll"`
	Login string `json:"login"`
}

type UserTokenV2 struct {
	Server      string `json:"server"`
	LoginName   string `json:"loginName"`
	AppPassword string `json:"appPassword"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type RequestTokenBody struct {
	Code      string `json:"code"`
	GrantType string `json:"grant_type"`
}

type RefreshTokenBody struct {
	RefreshToken string `json:"refresh_token"`
	GrantType    string `json:"grant_type"`
}

func RefreshToken(creq apps.CallRequest) Token {

	clientId := creq.Context.OAuth2.OAuth2App.ClientID
	clientSecret := creq.Context.OAuth2.OAuth2App.ClientSecret
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL

	reqUrl := fmt.Sprintf("%s/index.php/apps/oauth2/api/v1/token", remoteUrl)
	refreshToken := creq.Context.OAuth2.User.(map[string]interface{})["refresh_token"].(string)

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
