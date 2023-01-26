package file

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	"net/http"
)

type FileSharesInfo struct {
	shareService FileShareService
}

func (s FileSharesInfo) GetSharesInfo(filePath string, shareType int32) (*FileShareModel, error) {
	shares, err := s.shareService.getAllUserShares()

	if err != nil {
		return nil, err
	}

	for _, el := range shares.Data.Element {
		if el.Path == filePath {
			return &el, nil
		}
	}

	return s.shareService.createUserShare(filePath, shareType)
}

type FileShareService interface {
	getAllUserShares() (*SharedFilesResponseBody, error)
	createUserShare(filePath string, shareType int32) (*FileShareModel, error)
}

type FileShareServiceImpl struct {
	Url   string
	Token string
}

func (s FileShareServiceImpl) getAllUserShares() (*SharedFilesResponseBody, error) {

	req, _ := http.NewRequest("GET", s.Url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	req.Header.Set("OCS-APIRequest", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	xmlResp := SharedFilesResponseBody{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)

	return &xmlResp, err
}

func (s FileShareServiceImpl) createUserShare(filePath string, shareType int32) (*FileShareModel, error) {
	payload := FileShareRequestBody{filePath, shareType}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", s.Url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	req.Header.Set("OCS-APIRequest", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	xmlResp := SharedFileResponseBody{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)

	return &xmlResp.Data, err
}

func createFileSharePostWithAttachments(user *model.User, sm *FileShareModel, creq apps.CallRequest) *model.Post {

	post := model.Post{}
	post.ChannelId = creq.Context.Channel.Id
	attachments := createAttachments(user, sm)
	post.AddProp("attachments", attachments)
	return &post
}

func createAttachments(user *model.User, sm *FileShareModel) []model.SlackAttachment {
	attachment := model.SlackAttachment{}
	attachment.AuthorName = user.Username
	attachment.Title = sm.FileTarget[1:]
	attachment.TitleLink = sm.URL
	attachment.Footer = sm.Mimetype

	attachments := make([]model.SlackAttachment, 0)

	attachments = append(attachments, attachment)
	return attachments
}
