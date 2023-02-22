package file

import (
	"errors"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
	"net/http"
	"os"
	"testing"
)

type Client4Mock struct {
	*appclient.Client
	fileInfo *model.FileInfo
	error    error
}

func (mock Client4Mock) GetFileInfo(fileId string) (*model.FileInfo, *model.Response, error) {

	return mock.fileInfo, nil, mock.error
}

func (mock Client4Mock) GetFile(fileId string) ([]byte, *model.Response, error) {
	return nil, nil, nil
}

type FileFullUploadServiceMock struct {
	used bool
}

func (f *FileFullUploadServiceMock) UploadFile(file []byte, url string) (*http.Response, error) {
	f.used = true
	return nil, nil
}

type FileChunkUploadServiceMock struct {
	used bool
}

func (f FileChunkUploadServiceMock) createChunkFolder(url string) (*http.Response, error) {
	return nil, nil
}

func (f FileChunkUploadServiceMock) uploadFileChunk(file []byte, start string, end string, baseurl string) (*http.Response, error) {
	return nil, nil
}

func (f *FileChunkUploadServiceMock) assembleChunk(dest string, baseurl string) (*http.Response, error) {
	f.used = true
	return nil, nil
}

func (f FileChunkUploadServiceMock) abortChunkUpload(url string) (*http.Response, error) {
	return nil, nil
}

type FileServiceMock struct {
}

func (s FileServiceMock) GetChunkedFile(path string, from string, to string) (chunk []byte, err error) {
	return nil, nil
}

func TestFileSizeIsValid(t *testing.T) {
	os.Setenv("MAX_FILE_SIZE_MB", "1")
	defer os.Unsetenv("MAX_FILE_SIZE_MB")
	os.Setenv("MAX_FILES_SIZE_MB", "2")
	defer os.Unsetenv("MAX_FILES_SIZE_MB")

	model := model.NewInfo("name")
	model.Size = 1024
	mock := Client4Mock{fileInfo: model}
	m := map[string]interface{}{
		"value": "name",
	}
	var arr []interface{}
	arr = append(arr, m)

	testedInstance := FileUploadServiceImpl{}
	isValid, _ := testedInstance.ValidateFiles(mock, arr)

	if !isValid {
		t.Error("Files not valid")
	}
}

func TestFullFileUploadIsUsing(t *testing.T) {
	os.Setenv("MAX_FILE_SIZE_MB", "1")
	defer os.Unsetenv("MAX_FILE_SIZE_MB")
	os.Setenv("CHUNK_FILE_SIZE_MB", "2")
	defer os.Unsetenv("CHUNK_FILE_SIZE_MB")

	os.Setenv("MAX_FILES_SIZE_MB", "3")
	defer os.Unsetenv("MAX_FILES_SIZE_MB")

	m := map[string]interface{}{
		"value": "name",
	}
	var arr []interface{}
	arr = append(arr, m)
	remoteRootUrl := "remoteUrk"
	oauth2APP := apps.OAuth2App{RemoteRootURL: remoteRootUrl}

	user := map[string]interface{}{
		"user_id": "id",
	}

	oauth := apps.OAuth2Context{}
	oauth.OAuth2App = oauth2APP
	oauth.User = user

	context := apps.Context{}

	context.OAuth2 = oauth
	creq := apps.CallRequest{}

	creq.Context = context

	folder := map[string]interface{}{
		"value": "folder_value",
	}

	values := map[string]interface{}{
		"Folder": folder,
	}

	creq.Values = values

	model := model.NewInfo("name")
	model.Size = 1024
	asBotMock := Client4Mock{fileInfo: model}

	fullUpload := FileFullUploadServiceMock{}
	testedInstance := FileUploadServiceImpl{fileFullUploadService: &fullUpload}

	uploaded := testedInstance.UploadFiles(creq, arr, asBotMock)
	if len(uploaded) == 0 || !fullUpload.used {
		t.Error("File should upload by full flow")
	}
}

func TestChunkFileUploadIsUsing(t *testing.T) {
	os.Setenv("MAX_FILE_SIZE_MB", "1")
	defer os.Unsetenv("MAX_FILE_SIZE_MB")
	os.Setenv("CHUNK_FILE_SIZE_MB", "1")
	os.Setenv("MAX_FILES_SIZE_MB", "3")
	defer os.Unsetenv("MAX_FILES_SIZE_MB")

	m := map[string]interface{}{
		"value": "name",
	}
	var arr []interface{}
	arr = append(arr, m)
	remoteRootUrl := "remoteUrk"
	oauth2APP := apps.OAuth2App{RemoteRootURL: remoteRootUrl}

	user := map[string]interface{}{
		"user_id": "id",
	}

	oauth := apps.OAuth2Context{}
	oauth.OAuth2App = oauth2APP
	oauth.User = user

	context := apps.Context{}

	context.OAuth2 = oauth
	creq := apps.CallRequest{}

	creq.Context = context

	folder := map[string]interface{}{
		"value": "folder_value",
	}

	values := map[string]interface{}{
		"Folder": folder,
	}

	creq.Values = values

	model := model.NewInfo("name")
	model.Size = 1024 * 1025
	asBotMock := Client4Mock{fileInfo: model}

	fileServiceMock := FileServiceMock{}
	fileChunkUploadServiceMock := FileChunkUploadServiceMock{}
	chunkFileUploadService := ChunkFileUploadServiceImpl{fileChunkService: &fileChunkUploadServiceMock, MMFileService: fileServiceMock}
	testedInstance := FileUploadServiceImpl{fileChunkUploadService: &chunkFileUploadService}

	uploaded := testedInstance.UploadFiles(creq, arr, asBotMock)
	if len(uploaded) == 0 || !fileChunkUploadServiceMock.used {
		t.Error("File should upload")
	}
}

func TestFileSizeIsNotValid(t *testing.T) {
	os.Setenv("MAX_FILES_SIZE_MB", "2")
	defer os.Unsetenv("MAX_FILES_SIZE_MB")
	os.Setenv("MAX_FILE_SIZE_MB", "1")
	defer os.Unsetenv("MAX_FILE_SIZE_MB")

	fileInfo := model.NewInfo("name")
	fileInfo.Size = 1024 * 1025
	mock := Client4Mock{fileInfo: fileInfo}
	m := map[string]interface{}{
		"value": "name",
	}
	var arr []interface{}
	arr = append(arr, m)

	testedInstance := FileUploadServiceImpl{}
	isValid, _ := testedInstance.ValidateFiles(mock, arr)

	if isValid {
		t.Error("Validation should fail")
	}
}

func TestFilesSizeIsNotValid(t *testing.T) {
	os.Setenv("MAX_FILES_SIZE_MB", "2")
	defer os.Unsetenv("MAX_FILES_SIZE_MB")
	os.Setenv("MAX_FILE_SIZE_MB", "1")
	defer os.Unsetenv("MAX_FILE_SIZE_MB")

	fileInfo := model.NewInfo("name")
	fileInfo.Size = 1024 * 1024
	mock := Client4Mock{fileInfo: fileInfo}
	m := map[string]interface{}{
		"value": "name",
	}

	var arr []interface{}
	arr = append(arr, m)
	arr = append(arr, m)
	arr = append(arr, m)

	testedInstance := FileUploadServiceImpl{}
	isValid, _ := testedInstance.ValidateFiles(mock, arr)

	if isValid {
		t.Error("Validation should fail")
	}
}

func TestFileNotFound(t *testing.T) {
	os.Setenv("MAX_FILE_SIZE_MB", "1")
	defer os.Unsetenv("MAX_FILE_SIZE_MB")
	expected := "Could not get file info for file name with error error"

	mock := Client4Mock{error: errors.New("error")}
	m := map[string]interface{}{
		"value": "name",
	}
	var arr []interface{}
	arr = append(arr, m)

	testedInstance := FileUploadServiceImpl{}
	_, errMsg := testedInstance.ValidateFiles(mock, arr)

	if expected != *errMsg {
		t.Error("Validation should fail with Could not get file error")
	}
}
