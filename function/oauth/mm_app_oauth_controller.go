package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

func Configure(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	instanceUrl, _ := creq.Values["instance_url"].(string)
	clientId, _ := creq.Values["client_id"].(string)
	clientSecret, _ := creq.Values["client_secret"].(string)

	asUser := appclient.AsActingUser(creq.Context)

	asUser.StoreOAuth2App(apps.OAuth2App{
		RemoteRootURL: instanceUrl,
		ClientID:      clientId,
		ClientSecret:  clientSecret,
	})

	c.JSON(http.StatusOK, apps.NewTextResponse("updated OAuth client credentials"))

}

func HandleConnect(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	c.JSON(http.StatusOK,
		apps.NewTextResponse("[Connect](%s) to Next Cloud.", creq.Context.OAuth2.ConnectURL))

}

func Oauth2Connect(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	url := buildConnectUrl(&creq)

	c.JSON(http.StatusOK, apps.NewDataResponse(url))
}

func Oauth2Complete(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
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
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	jsonResp := Token{}

	json.NewDecoder(resp.Body).Decode(&jsonResp)

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(jsonResp)

	c.JSON(http.StatusOK, apps.NewTextResponse("completed oauth"))
}

func buildConnectUrl(creq *apps.CallRequest) string {
	remoteUrl := creq.Context.ExpandedContext.OAuth2.OAuth2App.RemoteRootURL
	clientId := creq.Context.ExpandedContext.OAuth2.OAuth2App.ClientID
	state := creq.Values["state"].(string)

	url := fmt.Sprintf("%s/index.php/apps/oauth2/authorize?response_type=code&state=%s&client_id=%s", remoteUrl, state, clientId)
	return url
}

func Disconnect(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	asActingUser := appclient.AsActingUser(creq.Context)
	err := asActingUser.StoreOAuth2User(nil)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, apps.CallResponse{
		Text: "Disconnected your NextCloud account",
	})
}
