package file

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/pkg/errors"
	"github.com/prokhorind/nextcloud/function/oauth"
	"github.com/prokhorind/nextcloud/function/user"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"strings"
)

func FileUploadForm(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	if len(creq.Context.Post.FileIds) == 0 {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: "Selected post doesn't have any files to be uploaded"})
		return
	}

	oauthService := oauth.OauthServiceImpl{creq}
	token, refreshErr := oauthService.RefreshToken()

	if refreshErr != nil {
		c.JSON(http.StatusOK, apps.NewErrorResponse(refreshErr))
		return
	}

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(*token)

	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	url := fmt.Sprintf("%s%s", remoteUrl, "/remote.php/dav/")

	body := createSearchRequestBody(userId, "")
	fileSearchRequestService := FileSearchServiceRequestServiceImpl{url: url, accessToken: token.AccessToken}
	resp, err := fileSearchRequestService.sendFileSearchRequest(body)
	if err != nil {
		c.JSON(http.StatusOK, apps.NewErrorResponse(errors.New("Request failed during file search")))
	}
	searchService := SearchSelectOptionsImpl{}
	folderSelectOptions, rootSelectOption := searchService.CreateFolderSelectOptions(*resp, userId, "Root", "/", false)

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
	token, refreshErr := oauthService.RefreshToken()

	if refreshErr != nil {
		c.JSON(http.StatusOK, apps.NewErrorResponse(refreshErr))
		return
	}

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(*token)
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
	fileSearchRequestService := FileSearchServiceRequestServiceImpl{url: url, accessToken: accessToken}
	FileSearchResp, err := fileSearchRequestService.sendFileSearchRequest(fileSearchBody)

	if err != nil {
		c.JSON(http.StatusOK, apps.NewErrorResponse(errors.New("Request failed during file search")))
		return
	}

	files := FileSearchResp.FileResponse

	if len(files) == 0 {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: "Files not found"})
		return
	}

	searchService := SearchSelectOptionsImpl{}
	fileSelectOptions := searchService.CreateFileSelectOptions(files)

	folderSearchBody := createSearchRequestBody(userId, "")
	folderSearchResp, folderSearchError := fileSearchRequestService.sendFileSearchRequest(folderSearchBody)

	if folderSearchError != nil {
		c.JSON(http.StatusOK, apps.NewErrorResponse(errors.New("Request failed during folder search")))
		return
	}

	folderSelectOptions, defaultSelectOption := searchService.CreateFolderSelectOptions(*folderSearchResp, userId, "Root", "", true)

	if creq.Values["Folder"] != nil {
		for _, so := range folderSelectOptions {
			if folderName == so.Value {
				defaultSelectOption = so
				break
			}
		}
	}

	if len(fileSelectOptions) == 0 {
		log.Error("Files do not have display name field")
		c.JSON(http.StatusOK, apps.NewErrorResponse(errors.New("Files do not have display name. The outdated version of Nextcloud is used. Contact system administrator")))
		return
	}

	sort.Slice(fileSelectOptions, func(i, j int) bool {
		return fileSelectOptions[i].Label < fileSelectOptions[j].Label
	})

	sort.Slice(folderSelectOptions, func(i, j int) bool {
		return folderSelectOptions[i].Label < folderSelectOptions[j].Label
	})

	preSelectedFilesOptions := make([]apps.SelectOption, 0)
	if creq.Values["Files"] != nil {

		preSelectedFiles := creq.Values["Files"].([]interface{})

		if len(preSelectedFiles) > 0 {

			for _, file := range preSelectedFiles {
				label := file.(map[string]interface{})["label"].(string)
				value := file.(map[string]interface{})["value"].(string)
				option := apps.SelectOption{Label: label, Value: value}
				preSelectedFilesOptions = append(preSelectedFilesOptions, option)
			}
		}
	}

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
				Value:               preSelectedFilesOptions,
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
	token, refreshErr := oauthService.RefreshToken()

	if refreshErr != nil {
		c.JSON(http.StatusOK, apps.NewErrorResponse(refreshErr))
		return
	}
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(*token)
	accessToken := token.AccessToken
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	url := fmt.Sprintf("%s%s", remoteUrl, "/ocs/v2.php/apps/files_sharing/api/v1/shares")

	fileShareService := FileShareServiceImpl{Url: url, Token: accessToken}
	fileSharesInfo := FileSharesInfo{fileShareService}

	files := creq.Values["Files"].([]interface{})
	botService := user.BotServiceImpl{creq}
	botService.AddBot()
	asBot := appclient.AsBot(creq.Context)

	if len(files) == 0 {
		msg := fmt.Sprintf("Please, choose a file to share")
		log.Error(msg)
		c.JSON(http.StatusOK, apps.NewErrorResponse(errors.New(msg)))
		return
	}

	for _, file := range files {
		f := file.(map[string]interface{})["value"].(string)
		sm, err := fileSharesInfo.GetSharesInfo(f, 3)
		if err == nil {
			var userId string
			asBot.KVGet("", fmt.Sprintf("nc-user-%s", sm.UidFileOwner), &userId)
			u, _, _ := asBot.GetUser(userId, "")
			attachmentService := FileSharePostAttachementsImpl{user: u, sm: sm}
			post := attachmentService.CreateFileSharePostWithAttachments(creq)
			asBot.CreatePost(post)

		}
	}
	c.JSON(http.StatusOK, apps.NewTextResponse(""))
}

func FileUpload(c *gin.Context) {
	log.Info("File upload request")
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	oauthService := oauth.OauthServiceImpl{creq}
	token, refreshErr := oauthService.RefreshToken()
	if refreshErr != nil {
		c.JSON(http.StatusOK, apps.NewErrorResponse(refreshErr))
		return
	}
	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(*token)

	files := creq.Values["Files"].([]interface{})

	asBot := appclient.AsBot(creq.Context)
	botService := user.BotServiceImpl{Creq: creq}
	botService.AddBot()
	chunkFileService := FileChunkServiceImpl{Token: token.AccessToken}
	mmFileService := MMFileServiceImpl{creq.Context.BotAccessToken}
	chunkUploadService := ChunkFileUploadServiceImpl{&chunkFileService, mmFileService}
	fileService := FileFullUploadServiceImpl{token.AccessToken}
	fileUploadService := FileUploadServiceImpl{&fileService, &chunkUploadService}

	validFiles, errMsg := fileUploadService.ValidateFiles(asBot, files)
	if !validFiles {
		c.JSON(http.StatusOK, apps.CallResponse{Type: apps.CallResponseTypeError, Text: *errMsg})
		return
	}

	uploadedFiles := fileUploadService.UploadFiles(creq, files, asBot)
	c.JSON(http.StatusOK, apps.NewTextResponse("Uploaded files:  %s", strings.Join(uploadedFiles, ",")))
}
