package file

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/prokhorind/nextcloud/function/oauth"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type GetFileInfo interface {
	GetFileInfo(fileId string) (*model.FileInfo, *model.Response, error)
}

type FilesUploadService interface {
	ValidateFiles(asBot *appclient.Client, files []interface{}) (bool, *string)
	UploadFiles(creq apps.CallRequest, files []interface{}, asBot *appclient.Client, token oauth.Token) []string
}

type FileUploadServiceImpl struct {
	fileFullUploadService  FileFullUploadService
	fileChunkUploadService ChunkFileUploadServiceImpl
}

func (fileUpload FileUploadServiceImpl) ValidateFiles(asBot GetFileInfo, files []interface{}) (bool, *string) {
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

func (fileUpload FileUploadServiceImpl) UploadFiles(creq apps.CallRequest, files []interface{}, asBot *appclient.Client) []string {
	chunkFileSize, _ := strconv.Atoi(os.Getenv("CHUNK_FILE_SIZE_MB"))
	chunkFileSizeInBytes := int64(chunkFileSize * 1024 * 1024)
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	folder := creq.Values["Folder"].(map[string]interface{})["value"].(string)

	fileUrl := fmt.Sprintf("%s%s%s%s", remoteUrl, "/remote.php/dav/files/", userId, folder)
	var uploadedFiles []string

	for _, file := range files {
		f := file.(map[string]interface{})["value"].(string)
		fileInfo, _, _ := asBot.GetFileInfo(f)
		destination := fmt.Sprintf("%s%s", fileUrl, fileInfo.Name)

		if fileInfo.Size <= chunkFileSizeInBytes {
			uploadedFiles = fileUpload.fullFileUpload(asBot, f, destination, fileInfo, uploadedFiles)
		} else {
			uploadedFiles = fileUpload.chunkFileUpload(creq, fileInfo, chunkFileSizeInBytes, destination, uploadedFiles)
		}

	}
	return uploadedFiles
}

func (fileUpload FileUploadServiceImpl) chunkFileUpload(creq apps.CallRequest, fileInfo *model.FileInfo, chunkFileSizeInBytes int64, destination string, uploadedFiles []string) []string {
	log.Info("Chunk file uploading")
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	chunkFolder := fmt.Sprintf("/%s-%s", "temp", uuid.New().String())
	chunkUrl := fmt.Sprintf("%s%s%s%s", remoteUrl, "/remote.php/dav/uploads/", userId, chunkFolder)
	mmfileUrl := fmt.Sprintf("%s/%s/%s", creq.Context.MattermostSiteURL, "api/v4/files", fileInfo.Id)
	_, err := fileUpload.fileChunkUploadService.fileChunkService.createChunkFolder(chunkUrl)

	if err != nil {
		log.Errorf("Chunk folder was not created %s", err.Error())
		return uploadedFiles
	}
	//chunkUploadService := ChunkFileUploadServiceImpl{ fileUpload.chunkService, fileUpload.mmFileService}
	allChunksUpload := fileUpload.fileChunkUploadService.uploadChunks(chunkFileSizeInBytes, fileInfo, chunkUrl, mmfileUrl)

	if allChunksUpload {
		_, err := fileUpload.fileChunkUploadService.fileChunkService.assembleChunk(destination, chunkUrl)

		if err != nil {
			log.Errorf("Chunk was not assembled to NC destination %s with error %s", destination, err.Error())
			return uploadedFiles
		}
		uploadedFiles = append(uploadedFiles, fileInfo.Name)
		log.Infof("file was uploaded %s", destination)
	}
	return uploadedFiles
}

func (fileUpload FileUploadServiceImpl) fullFileUpload(asBot *appclient.Client, f string, destination string, fileInfo *model.FileInfo, uploadedFiles []string) []string {
	log.Info("Full file uploading")
	file, _, err := asBot.GetFile(f)
	if err != nil {
		log.Errorf("File was not downloaded from MM %s with error %s", f, err.Error())
		return uploadedFiles
	}
	_, uploadError := fileUpload.fileFullUploadService.UploadFile(file, destination)
	if uploadError != nil {
		log.Errorf("File %s was not auploaded to NC  with error %s", fileInfo.Id, err.Error())
		return uploadedFiles
	}
	uploadedFiles = append(uploadedFiles, fileInfo.Name)
	log.Infof("file was uploaded %s", fileInfo.Name)
	return uploadedFiles
}
