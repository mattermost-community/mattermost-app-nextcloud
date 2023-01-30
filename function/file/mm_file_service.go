package file

import (
	"fmt"
	"io"
	"net/http"
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

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)

	return data, err
}
