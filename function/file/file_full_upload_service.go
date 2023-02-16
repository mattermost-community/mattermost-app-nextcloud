package file

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
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

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("Error during file uploading. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		log.Errorf("request failed with status %s", resp.Status)
		error := fmt.Errorf("request failed with code %d", resp.StatusCode)
		return nil, error
	}

	return resp, err
}
