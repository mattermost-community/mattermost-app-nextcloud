package file

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	"testing"
)

type FileShareServiceTestMock struct {
	Url   string
	Token string
}

func (s FileShareServiceTestMock) GetAllUserShares() (*SharedFilesResponseBody, error) {
	responseBody := SharedFilesResponseBody{}
	model := FileShareModel{Path: "/test-path"}
	oneSlice := []FileShareModel{model}
	responseBody.Data = Data{Element: oneSlice}
	return &responseBody, nil
}

func (s FileShareServiceTestMock) CreateUserShare(filePath string, shareType int32) (*FileShareModel, error) {
	return &FileShareModel{Path: filePath, ShareType: string(shareType)}, nil
}

func TestFindExistedShareByPath(t *testing.T) {
	expectedFilePath := "/test-path"
	service := FileShareServiceTestMock{}
	testedInstance := FileSharesInfo{service}

	model, _ := testedInstance.GetSharesInfo(expectedFilePath, 3)
	actual := model.Path

	if expectedFilePath != actual {
		t.Errorf(" expected %q, actual %q", expectedFilePath, actual)
	}
}

func TestCreateNewShareIfExistedNotFound(t *testing.T) {
	expectedFilePath := "/new-path"
	service := FileShareServiceTestMock{Url: expectedFilePath}
	testedInstance := FileSharesInfo{service}

	model, _ := testedInstance.GetSharesInfo(expectedFilePath, 3)
	actual := model.Path

	if expectedFilePath != actual {
		t.Errorf(" expected %q, actual %q", expectedFilePath, actual)
	}
}

func TestCreateFileSharePostAttachment(t *testing.T) {

	user := &model.User{
		Username: "username",
	}

	channel := &model.Channel{
		Id: "channelId",
	}

	fm := &FileShareModel{
		FileTarget: "/fileTarget.png",
		URL:        "url.com",
		Mimetype:   "png",
	}

	expContext := apps.ExpandedContext{
		Channel: channel,
	}
	creq := apps.CallRequest{
		Context: apps.Context{ExpandedContext: expContext},
	}

	testedInstance := FileSharePostAttachementsImpl{user: user, sm: fm}

	post := testedInstance.CreateFileSharePostWithAttachments(creq)

	expectedHeader := user.Username

	actualAttachment := post.Attachments()[0]

	if expectedHeader != actualAttachment.AuthorName {
		t.Errorf(" expected %q, actual  %q", expectedHeader, actualAttachment.Title)
	}
}
