package file

import (
	"bytes"
	"fmt"
	"net/http"
)

type FileServiceImpl struct {
	Url   string
	Token string
}

func (s FileServiceImpl) UploadFile(file []byte) {
	req, _ := http.NewRequest("PUT", s.Url, bytes.NewBuffer(file))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}
