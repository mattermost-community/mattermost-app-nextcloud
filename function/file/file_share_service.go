package file

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

type FileSharesInfo struct {
	shareService FileShareService
}

func (s FileSharesInfo) GetSharesInfo(filePath string, shareType int32) (*FileShareModel, error) {
	shares, err := s.shareService.GetAllUserShares()

	if err != nil {
		return nil, err
	}

	for _, el := range shares.Data.Element {
		if el.Path == filePath {
			return &el, nil
		}
	}

	return s.shareService.CreateUserShare(filePath, shareType)
}

type FileShareService interface {
	GetAllUserShares() (*SharedFilesResponseBody, error)
	CreateUserShare(filePath string, shareType int32) (*FileShareModel, error)
}

type FileShareServiceImpl struct {
	Url   string
	Token string
}

func (s FileShareServiceImpl) GetAllUserShares() (*SharedFilesResponseBody, error) {

	req, _ := http.NewRequest("GET", s.Url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	req.Header.Set("OCS-APIRequest", "true")

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during getting of user shares. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("request failed with status %s", resp.Status)
		error := fmt.Errorf("request failed with code %d", resp.StatusCode)
		return nil, error
	}

	xmlResp := SharedFilesResponseBody{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)

	return &xmlResp, err
}

func (s FileShareServiceImpl) CreateUserShare(filePath string, shareType int32) (*FileShareModel, error) {
	payload := FileShareRequestBody{filePath, shareType}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", s.Url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	req.Header.Set("OCS-APIRequest", "true")

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during creating of user share. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("request failed with status %s", resp.Status)
		error := fmt.Errorf("request failed with code %d", resp.StatusCode)
		return nil, error
	}

	xmlResp := SharedFileResponseBody{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)

	return &xmlResp.Data, err
}

type FileSharePostAttachements interface {
	CreateFileSharePostWithAttachments(creq apps.CallRequest) *model.Post
}

type FileSharePostAttachementsImpl struct {
	user *model.User
	sm   *FileShareModel
}

func (f FileSharePostAttachementsImpl) CreateFileSharePostWithAttachments(creq apps.CallRequest) *model.Post {

	post := model.Post{}
	post.ChannelId = creq.Context.Channel.Id
	attachments := f.createAttachments()
	post.AddProp("attachments", attachments)
	return &post
}

func (f FileSharePostAttachementsImpl) createAttachments() []*model.SlackAttachment {
	attachment := model.SlackAttachment{}
	attachment.AuthorName = f.user.Username
	attachment.Title = f.sm.FileTarget[1:]
	attachment.TitleLink = f.sm.URL
	attachment.Footer = f.sm.Mimetype

	attachments := make([]*model.SlackAttachment, 0)

	attachments = append(attachments, &attachment)
	return attachments
}
