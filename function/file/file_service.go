package file

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/prokhorind/nextcloud/function/oauth"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

type FileServiceImpl struct {
	Url   string
	Token string
}

func ValidateFiles(asBot *appclient.Client, files []interface{}) (bool, *string) {
	for _, file := range files {
		f := file.(map[string]interface{})["value"].(string)

		fileInfo, _, err := asBot.GetFileInfo(f)

		if err != nil {
			errorMsg := fmt.Sprintf("Could not get file info for file %s with error %s", f, err.Error())
			log.Error(errorMsg)
			return false, &errorMsg
		}

		maxFileSizeString := os.Getenv("MAX_FILE_SIZE_MB")
		maxFileSize, _ := strconv.Atoi(maxFileSizeString)
		maxFileSizeInBytes := int64(maxFileSize * 1024 * 1024)

		if fileInfo.Size > maxFileSizeInBytes {
			msg := fmt.Sprintf("File above %s MB cannot be uploaded: %s", maxFileSizeString, fileInfo.Name)
			log.Error(msg)
			return false, &msg
		}

	}
	return true, nil
}

func UploadFiles(creq apps.CallRequest, files []interface{}, asBot *appclient.Client, token oauth.Token) []string {
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	folder := creq.Values["Folder"].(map[string]interface{})["value"].(string)

	fileUrl := fmt.Sprintf("%s%s%s%s", remoteUrl, "/remote.php/dav/files/", userId, folder)
	var uploadedFiles []string

	for _, file := range files {
		f := file.(map[string]interface{})["value"].(string)

		fileInfo, _, _ := asBot.GetFileInfo(f)

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
	return uploadedFiles
}

func (s FileServiceImpl) UploadFile(file []byte) (*http.Response, error) {
	req, _ := http.NewRequest("PUT", s.Url, bytes.NewBuffer(file))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, err
}

func AddBot(creq apps.CallRequest) {
	addBotToTeam(creq)
	addBotToChannel(creq)
}

func addBotToTeam(creq apps.CallRequest) {
	teamId := creq.Context.Channel.TeamId
	botId := creq.Context.BotUserID
	asActingUser := appclient.AsActingUser(creq.Context)
	_, _, err := asActingUser.GetTeamMember(teamId, botId, "")

	if err != nil {
		asActingUser.AddTeamMember(teamId, botId)
	}
}

func addBotToChannel(creq apps.CallRequest) {
	channelId := creq.Context.Channel.Id
	botId := creq.Context.BotUserID

	asActingUser := appclient.AsActingUser(creq.Context)

	_, _, err := asActingUser.GetChannelMember(channelId, botId, "")

	if err != nil {
		asActingUser.AddChannelMember(channelId, botId)
	}

}
