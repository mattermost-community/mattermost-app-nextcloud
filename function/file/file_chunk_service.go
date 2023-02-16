package file

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mattermost/mattermost-server/v6/model"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

type FileChunkService interface {
	createChunkFolder(url string) (*http.Response, error)
	uploadFileChunk(file []byte, start string, end string, url string) (*http.Response, error)
	assembleChunk(dest string, url string) (*http.Response, error)
	abortChunkUpload(url string) (*http.Response, error)
}

type FileChunkServiceImpl struct {
	Token string
}

func (f *FileChunkServiceImpl) createChunkFolder(url string) (*http.Response, error) {
	req, _ := http.NewRequest("MKCOL", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during refreshing of the token. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		log.Errorf("request failed with status %s", resp.Status)
		error := fmt.Errorf("request failed with code %d", resp.StatusCode)
		return nil, error
	}

	return resp, err
}

func (f *FileChunkServiceImpl) uploadFileChunk(file []byte, start string, end string, baseurl string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s-%s", baseurl, start, end)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(file))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during uploading of file chunks. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		log.Errorf("request failed with status %s", resp.Status)
		error := fmt.Errorf("request failed with code %d", resp.StatusCode)
		return nil, error
	}
	return resp, err
}

func (f *FileChunkServiceImpl) assembleChunk(dest string, baseurl string) (*http.Response, error) {
	url := fmt.Sprintf("%s/.file", baseurl)
	req, _ := http.NewRequest("MOVE", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))
	req.Header.Set("Destination", dest)

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during assembling chunks. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		log.Errorf("request failed with status %s", resp.Status)
		error := fmt.Errorf("request failed with code %d", resp.StatusCode)
		return nil, error
	}

	return resp, err
}

func (f *FileChunkServiceImpl) abortChunkUpload(url string) (*http.Response, error) {
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during aborting of chunk uploading. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		log.Errorf("request failed with status %s", resp.Status)
		error := fmt.Errorf("request failed with code %d", resp.StatusCode)
		return nil, error
	}

	return resp, err
}

type ChunkFileUploadService interface {
	uploadChunks(chunkFileSizeInBytes int64, fileInfo *model.FileInfo, url string, mmfileUrl string) bool
	getFileChunkService() FileChunkService
	getMMFileService() MMFileService
}

type ChunkFileUploadServiceImpl struct {
	fileChunkService FileChunkService
	MMFileService    MMFileService
}

func (s ChunkFileUploadServiceImpl) getFileChunkService() FileChunkService {
	return s.fileChunkService
}

func (s ChunkFileUploadServiceImpl) getMMFileService() MMFileService {
	return s.MMFileService
}

func (s ChunkFileUploadServiceImpl) uploadChunks(chunkFileSizeInBytes int64, fileInfo *model.FileInfo, url string, mmfileUrl string) bool {
	var low int64
	var high int64
	for low = 0; low < fileInfo.Size; low += chunkFileSizeInBytes + 1 {
		log.Debugf("Uploaded %d bytes from %d for file %s", low, chunkFileSizeInBytes, fileInfo.Name)
		high = chunkFileSizeInBytes + low
		chunkUploaded := s.uploadChunk(low, high, url, mmfileUrl)
		if !chunkUploaded {
			log.Errorf("Uploaded %d bytes from %d for file %s failed", low, chunkFileSizeInBytes, fileInfo.Name)
			return false
		}
	}
	log.Debug("Finished uploading")
	return true
}

func (s ChunkFileUploadServiceImpl) uploadChunk(low int64, high int64, url string, mmfileUrl string) bool {
	chunk, err := s.MMFileService.GetChunkedFile(mmfileUrl, fmt.Sprint(low), fmt.Sprint(high))

	if err != nil {
		log.Errorf("Chunk was not downloaded from MM %s", err.Error())
		s.fileChunkService.abortChunkUpload(url)
		return false
	}

	_, uploadError := s.fileChunkService.uploadFileChunk(chunk, fmt.Sprintf("%016d", low), fmt.Sprintf("%016d", high), url)
	if uploadError != nil {
		s.fileChunkService.abortChunkUpload(url)
		log.Errorf("Chunk was not uploaded to NC %s", uploadError.Error())
		return false
	}
	return true
}
