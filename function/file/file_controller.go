package file

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/prokhorind/nextcloud/function/oauth"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sort"
	"strconv"
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
	folderSelectOptions := make([]apps.SelectOption, 0)
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
			option := apps.SelectOption{Label: split[1 : len(split)-1], Value: split}
			folderSelectOptions = append(folderSelectOptions, option)
		}
	}

	fileSelectOptions := make([]apps.SelectOption, 0)
	fileInfos, _, _ := asActingUser.GetFileInfosForPost(creq.Context.Post.Id, "")

	for _, fi := range fileInfos {
		option := apps.SelectOption{Label: fi.Name, Value: fi.Id}
		fileSelectOptions = append(fileSelectOptions, option)
	}
	rootSelectOption := apps.SelectOption{Label: "Root", Value: "/"}
	folderSelectOptions = append(folderSelectOptions, rootSelectOption)

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

	fileName := creq.Values["file_name"].(string)
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	url := fmt.Sprintf("%s%s", remoteUrl, "/remote.php/dav/")

	body := createSearchRequestBody(userId, fileName)
	resp := sendFileSearchRequest(url, body, accessToken)

	files := resp.FileResponse
	fileSelectOptions := make([]apps.SelectOption, 0)

	for _, f := range files {
		ref := f.Href
		displayname := f.PropertyStats[0].Property.Displayname
		refs := strings.Split(ref, "/")
		r := strings.NewReplacer("%20", " ")
		sharingPath := r.Replace("/" + strings.Join(refs[5:len(refs)-1], "/") + "/" + displayname)
		option := apps.SelectOption{Label: sharingPath[1:], Value: sharingPath}
		fileSelectOptions = append(fileSelectOptions, option)
	}

	if len(files) == 0 {
		c.JSON(http.StatusOK, apps.NewTextResponse("File %s not found check the file name and try again", fileName))
		return
	}

	sort.Slice(fileSelectOptions, func(i, j int) bool {
		return fileSelectOptions[i].Label < fileSelectOptions[j].Label
	})

	form := &apps.Form{
		Title: "File share ",
		Icon:  "icon.png",
		Fields: []apps.Field{
			{
				Type:                "static_select",
				Name:                "Files",
				Label:               "Files",
				IsRequired:          true,
				SelectIsMulti:       true,
				SelectStaticOptions: fileSelectOptions,
			},
		},
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

	folder := creq.Values["Folder"].(map[string]interface{})["value"].(string)

	files := creq.Values["Files"].([]interface{})

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	fileUrl := fmt.Sprintf("%s%s%s%s", remoteUrl, "/remote.php/dav/files/", userId, folder)

	asBot := appclient.AsBot(creq.Context)
	AddBot(creq)
	var uploadedFiles []string

	for _, file := range files {
		f := file.(map[string]interface{})["value"].(string)

		fileInfo, _, err := asBot.GetFileInfo(f)

		if err != nil {
			log.Errorf("Could not get file info for file %s with error %s", f, err.Error())

			continue
		}

		chunkFileSize, _ := strconv.Atoi(os.Getenv("CHUNK_FILE_SIZE_MB"))

		chunkFileSizeInBytes := int64(chunkFileSize * 1024 * 1024)

		destination := fmt.Sprintf("%s%s", fileUrl, fileInfo.Name)
		if fileInfo.Size <= chunkFileSizeInBytes {
			log.Info("Full file uploading")
			file, _, err := asBot.GetFile(f)
			if err != nil {
				log.Errorf("File was not downloaded from MM %s with error %s", f, err.Error())
				continue
			}
			fileService := FileServiceImpl{Url: destination, Token: token.AccessToken}
			_, uploadError := fileService.UploadFile(file)
			if uploadError != nil {
				log.Errorf("File %s was not auploaded to NC  with error %s", fileInfo.Id, err.Error())
			} else {
				uploadedFiles = append(uploadedFiles, fileInfo.Name)
				log.Infof("file was uploaded %s", fileInfo.Name)

			}
		} else {
			log.Info("Chunk file uploading")
			chunkFolder := fmt.Sprintf("/%s-%s", "temp", uuid.New().String())
			chunkUrl := fmt.Sprintf("%s%s%s%s", remoteUrl, "/remote.php/dav/uploads/", userId, chunkFolder)
			mmfileUrl := fmt.Sprintf("%s/%s/%s", creq.Context.MattermostSiteURL, "api/v4/files", fileInfo.Id)
			fileService := FileChunkServiceImpl{BaseUrl: chunkUrl, Token: token.AccessToken}
			_, err := fileService.createChunkFolder()

			if err != nil {
				log.Errorf("Chunk folder was not created %s", err.Error())
				continue
			}

			allChunksUpload := uploadChunks(chunkFileSizeInBytes, fileInfo, mmfileUrl, creq, fileService)

			if allChunksUpload {
				_, err := fileService.assembleChunk(destination)

				if err != nil {
					log.Errorf("Chunk was not assembled to NC destination %s with error %s", destination, err.Error())
				} else {
					uploadedFiles = append(uploadedFiles, fileInfo.Name)
					log.Infof("file was uploaded %s", destination)
				}
			}
		}

	}
	c.JSON(http.StatusOK, apps.NewTextResponse("Uploaded files:  %s", strings.Join(uploadedFiles, ",")))
}
