package file

import (
	"bytes"
	"fmt"
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
