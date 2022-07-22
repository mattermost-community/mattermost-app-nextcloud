package file

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/prokhorind/nextcloud/function/oauth"
)

func SearchFolders(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	foldName := creq.Query
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	url := fmt.Sprintf("%s%s", remoteUrl, "/remote.php/dav/")

	body := createSearchRequestBody(userId, foldName)
	resp := sendFileSearchRequest(url, body, token.AccessToken)
	selectOptions := make([]apps.SelectOption, 0)
	for _, f := range resp.FileResponse {
		hasContentType := false

		for _, p := range f.PropertyStats {
			if len(p.Property.Getcontenttype) != 0 {
				hasContentType = true
				break
			}
		}
		if !hasContentType {
			option := apps.SelectOption{Label: f.Href, Value: f.Href}
			selectOptions = append(selectOptions, option)
		}
	}

	c.JSON(200, apps.NewDataResponse(DynamicSelectResponse{Items: selectOptions}))
}

func FileUploadForm(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	url := fmt.Sprintf("%s%s", remoteUrl, "/remote.php/dav/")

	body := createSearchRequestBody(userId, "")
	resp := sendFileSearchRequest(url, body, token.AccessToken)
	selectOptions := make([]apps.SelectOption, 0)
	for _, f := range resp.FileResponse {
		hasContentType := false

		for _, p := range f.PropertyStats {
			if len(p.Property.Getcontenttype) != 0 {
				hasContentType = true
				break
			}
		}
		if !hasContentType {
			split := strings.Split(f.Href, "/remote.php/dav/files/"+userId)[1]
			option := apps.SelectOption{Label: split, Value: split}
			selectOptions = append(selectOptions, option)
		}
	}

	form := &apps.Form{
		Title: "Upload to Nextcloud ",
		Icon:  "icon.png",
		Fields: []apps.Field{

			{
				Type:                "static_select",
				Name:                "Folder",
				Label:               "Folder",
				IsRequired:          true,
				SelectStaticOptions: selectOptions,
			},
		},
		Submit: apps.NewCall("/file-upload").WithState(creq.Context.Post.FileIds).WithExpand(apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			Channel:               apps.ExpandAll,
		}),
	}

	c.JSON(http.StatusOK, apps.NewFormResponse(*form))
}

func FileSearch(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	accessToken := token.AccessToken

	fileName := creq.Values["file_name"].(string)
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	url := fmt.Sprintf("%s%s", remoteUrl, "/remote.php/dav/")

	body := createSearchRequestBody(userId, fileName)
	resp := sendFileSearchRequest(url, body, accessToken)

	files := resp.FileResponse

	for _, f := range files {
		sendFiles(f, &creq)
	}

	c.JSON(http.StatusOK, apps.NewDataResponse(nil))
}

func FileUpload(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	folder := creq.Values["Folder"].(map[string]interface{})["value"].(string)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	filesUrl := fmt.Sprintf("%s%s%s%s", remoteUrl, "/remote.php/dav/files/", userId, folder)

	files := creq.State.([]interface{})

	asBot := appclient.AsBot(creq.Context)
	for _, file := range files {
		f := file.(string)

		fileInfo, _, _ := asBot.GetFileInfo(f)
		fileService := FileServiceImpl{Url: fmt.Sprintf("%s%s", filesUrl, fileInfo.Name), Token: token.AccessToken}

		file, _, _ := asBot.GetFile(f)

		fileService.UploadFile(file)
	}

	c.JSON(http.StatusOK, apps.NewTextResponse("files were uploaded"))
}
