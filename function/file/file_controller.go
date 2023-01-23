package file

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/prokhorind/nextcloud/function/oauth"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"strings"
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

	if len(creq.Context.Post.FileIds) == 0 {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: "Selected post doesn't have any files to be uploaded"})
		return
	}

	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	url := fmt.Sprintf("%s%s", remoteUrl, "/remote.php/dav/")

	body := createSearchRequestBody(userId, "")
	resp := sendFileSearchRequest(url, body, token.AccessToken)
	folderSelectOptions, rootSelectOption := CreateFolderSelectOptions(resp, userId, "Root", "/")

	fileSelectOptions := make([]apps.SelectOption, 0)
	fileInfos, _, _ := asActingUser.GetFileInfosForPost(creq.Context.Post.Id, "")

	for _, fi := range fileInfos {
		option := apps.SelectOption{Label: fi.Name, Value: fi.Id}
		fileSelectOptions = append(fileSelectOptions, option)
	}

	sort.Slice(folderSelectOptions, func(i, j int) bool {
		return folderSelectOptions[i].Label < folderSelectOptions[j].Label
	})

	sort.Slice(fileSelectOptions, func(i, j int) bool {
		return fileSelectOptions[i].Label < fileSelectOptions[j].Label
	})

	form := &apps.Form{
		Title: "Upload to Nextcloud ",
		Icon:  "icon.png",
		Fields: []apps.Field{

			{
				Type:                "static_select",
				Name:                "Folder",
				Label:               "Folder",
				IsRequired:          true,
				SelectStaticOptions: folderSelectOptions,
				Value:               rootSelectOption,
			},

			{
				Type:                "static_select",
				Name:                "Files",
				Label:               "Files",
				IsRequired:          true,
				SelectIsMulti:       true,
				SelectStaticOptions: fileSelectOptions,
				Value:               fileSelectOptions,
			},
		},
		Submit: apps.NewCall("/file-upload").WithExpand(apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			Channel:               apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}),
	}

	c.JSON(http.StatusOK, apps.NewFormResponse(*form))
}

func FileShareForm(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	accessToken := token.AccessToken

	var folderName string
	if creq.Values["Folder"] == nil {
		folderName = ""
	} else {
		folderName = creq.Values["Folder"].(map[string]interface{})["value"].(string)
	}
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	url := fmt.Sprintf("%s%s", remoteUrl, "/remote.php/dav/")

	fileSearchBody := createSearchRequestBody(userId+folderName, "")
	FileSearchResp := sendFileSearchRequest(url, fileSearchBody, accessToken)

	files := FileSearchResp.FileResponse

	if len(files) == 0 {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: "Files not found"})
		return
	}

	fileSelectOptions := CreateFileSelectOptions(files)

	folderSearchBody := createSearchRequestBody(userId, "")
	folderSearchResp := sendFileSearchRequest(url, folderSearchBody, accessToken)
	folderSelectOptions, defaultSelectOption := CreateFolderSelectOptions(folderSearchResp, userId, "Root", "")

	if creq.Values["Folder"] != nil {
		for _, so := range folderSelectOptions {
			if folderName == so.Value {
				defaultSelectOption = so
				break
			}
		}
	}

	if len(fileSelectOptions) == 0 {
		option := apps.SelectOption{Label: "", Value: ""}
		fileSelectOptions = append(fileSelectOptions, option)
	}

	sort.Slice(fileSelectOptions, func(i, j int) bool {
		return fileSelectOptions[i].Label < fileSelectOptions[j].Label
	})

	sort.Slice(folderSelectOptions, func(i, j int) bool {
		return folderSelectOptions[i].Label < folderSelectOptions[j].Label
	})

	form := &apps.Form{
		Title: "File share ",
		Icon:  "icon.png",
		Fields: []apps.Field{
			{
				Type:                "static_select",
				Name:                "Folder",
				Label:               "Folder",
				IsRequired:          true,
				SelectRefresh:       true,
				SelectStaticOptions: folderSelectOptions,
				Value:               defaultSelectOption,
			},
			{
				Type:                "static_select",
				Name:                "Files",
				Label:               "Files",
				IsRequired:          true,
				SelectIsMulti:       true,
				SelectStaticOptions: fileSelectOptions,
			},
		},
		Source: apps.NewCall("/file/search/form").WithExpand(apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			Channel:               apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}),
		Submit: apps.NewCall("/file-share").WithExpand(apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			Channel:               apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}),
	}

	c.JSON(http.StatusOK, apps.NewFormResponse(*form))
}

func FileShare(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)
	accessToken := token.AccessToken
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	url := fmt.Sprintf("%s%s", remoteUrl, "/ocs/v2.php/apps/files_sharing/api/v1/shares")

	fileShareService := FileShareServiceImpl{Url: url, Token: accessToken}

	files := creq.Values["Files"].([]interface{})

	asBot := appclient.AsBot(creq.Context)
	for _, file := range files {
		f := file.(map[string]interface{})["value"].(string)
		sm, err := fileShareService.GetSharesInfo(f, 3)
		if err == nil {
			createFileSharePostWithAttachments(asBot, sm, creq)
		}
	}
	c.JSON(http.StatusOK, apps.NewTextResponse(""))
}

func FileUpload(c *gin.Context) {
	log.Info("File upload request")
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}
	token := oauthService.RefreshToken()
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	files := creq.Values["Files"].([]interface{})

	asBot := appclient.AsBot(creq.Context)
	AddBot(creq)
	validFiles, errMsg := ValidateFiles(asBot, files)
	if !validFiles {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: *errMsg})
		return
	}

	uploadedFiles := UploadFiles(creq, files, asBot, token)
	c.JSON(http.StatusOK, apps.NewTextResponse("Uploaded files:  %s", strings.Join(uploadedFiles, ",")))
}
