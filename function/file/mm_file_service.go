package file

import (
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strconv"
)

type MMFileService interface {
	GetChunkedFile(path string, from string, to string) (chunk []byte, err error)
}

type MMFileServiceImpl struct {
	token string
}

func (s MMFileServiceImpl) GetChunkedFile(path string, from string, to string) (chunk []byte, err error) {
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	req.Header.Set("Range", fmt.Sprintf("bytes=%s-%s", from, to))

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during getting of chunked files. Error: %s", err)
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)

	return data, err
}
