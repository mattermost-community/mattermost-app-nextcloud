package file

import (
	"errors"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
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

func TestFileSizeIsValid(t *testing.T) {
	os.Setenv("MAX_FILE_SIZE_MB", "1")
	defer os.Unsetenv("MAX_FILE_SIZE_MB")
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

func TestFileSizeIsNotValid(t *testing.T) {
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
