package file

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type FileFullUploadService interface {
	UploadFile(file []byte, url string) (*http.Response, error)
}

type FileFullUploadServiceImpl struct {
	Token string
}

func (s *FileFullUploadServiceImpl) UploadFile(file []byte, url string) (*http.Response, error) {
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(file))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		log.Errorf("request failed with status %s", resp.Status)
		error := fmt.Errorf("request failed with code %d", resp.StatusCode)
		return nil, error
	}

	return resp, err
}
