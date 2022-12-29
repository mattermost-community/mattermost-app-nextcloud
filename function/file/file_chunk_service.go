package file

import (
	"bytes"
	"fmt"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type FileChunkServiceImpl struct {
	BaseUrl string
	Token   string
}

func (f FileChunkServiceImpl) createChunkFolder() (*http.Response, error) {
	req, _ := http.NewRequest("MKCOL", f.BaseUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, err
}

func (f FileChunkServiceImpl) uploadFileChunk(file []byte, start string, end string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s-%s", f.BaseUrl, start, end)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(file))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, err
}

func (f FileChunkServiceImpl) assembleChunk(dest string) (*http.Response, error) {
	url := fmt.Sprintf("%s/.file", f.BaseUrl)
	req, _ := http.NewRequest("MOVE", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))
	req.Header.Set("Destination", dest)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, err
}

func (f FileChunkServiceImpl) abortChunkUpload() (*http.Response, error) {
	req, _ := http.NewRequest("DELETE", f.BaseUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, err
}

func uploadChunks(chunkFileSizeInBytes int64, fileInfo *model.FileInfo, mmfileUrl string, creq apps.CallRequest, fileService FileChunkServiceImpl) bool {
	var low int64
	var high int64
	for low = 0; low < fileInfo.Size; low += chunkFileSizeInBytes + 1 {
		high = chunkFileSizeInBytes + low
		chunkUploaded := uploadChunk(mmfileUrl, creq, low, high, fileService)
		if !chunkUploaded {
			return false
		}
	}
	return true
}

func uploadChunk(mmfileUrl string, creq apps.CallRequest, low int64, high int64, fileService FileChunkServiceImpl) bool {
	chunk, err := GetChunkedFile(mmfileUrl, creq.Context.BotAccessToken, fmt.Sprint(low), fmt.Sprint(high))

	if err != nil {
		log.Errorf("Chunk was not downloaded from MM %s", err.Error())
		fileService.abortChunkUpload()
		return false
	}

	_, uploadError := fileService.uploadFileChunk(chunk, fmt.Sprintf("%016d", low), fmt.Sprintf("%016d", high))
	if uploadError != nil {
		fileService.abortChunkUpload()
		log.Errorf("Chunk was not uploaded to NC %s", uploadError.Error())
		return false
	}
	return true
}
