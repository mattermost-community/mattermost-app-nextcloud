package file

import (
	"errors"
	"github.com/mattermost/mattermost-server/v6/model"
	"net/http"
	"testing"
)

type MMFileServiceMock struct {
	getChunkedFileError error
}

func (s MMFileServiceMock) GetChunkedFile(path string, from string, to string) (chunk []byte, err error) {

	return nil, s.getChunkedFileError
}

type FileChunkServiceMock struct {
	createChunkFolderError error
	uploadFileChunkError   error
	assembleFileChunkError error
	abortChunkUploadError  error
}

func (f FileChunkServiceMock) createChunkFolder(url string) (*http.Response, error) {

	return nil, f.createChunkFolderError
}

func (f FileChunkServiceMock) uploadFileChunk(file []byte, start string, end string, baseurl string) (*http.Response, error) {
	return nil, f.uploadFileChunkError
}

func (f FileChunkServiceMock) assembleChunk(dest string, baseurl string) (*http.Response, error) {

	return nil, f.assembleFileChunkError
}

func (f FileChunkServiceMock) abortChunkUpload(url string) (*http.Response, error) {

	return nil, f.abortChunkUploadError
}

func TestChunksUpload(t *testing.T) {
	fileService := MMFileServiceMock{}
	fileChunkService := FileChunkServiceMock{}
	testedInstance := ChunkFileUploadServiceImpl{MMFileService: fileService, fileChunkService: fileChunkService}

	var fileSize int64 = 10
	var chunkFileSize int64 = 1
	fileInfo := &model.FileInfo{Size: fileSize}
	url := "test_url"
	mmFileUrl := "mm_file_url"

	uploaded := testedInstance.uploadChunks(chunkFileSize, fileInfo, url, mmFileUrl)

	if !uploaded {
		t.Error("Chunks should upload")
	}
}

func TestMMFileNotFoundDuringChunksUpload(t *testing.T) {
	fileService := MMFileServiceMock{getChunkedFileError: errors.New("error")}
	fileChunkService := FileChunkServiceMock{}
	testedInstance := ChunkFileUploadServiceImpl{MMFileService: fileService, fileChunkService: fileChunkService}

	var fileSize int64 = 10
	var chunkFileSize int64 = 1
	fileInfo := &model.FileInfo{Size: fileSize}
	url := "test_url"
	mmFileUrl := "mm_file_url"

	uploaded := testedInstance.uploadChunks(chunkFileSize, fileInfo, url, mmFileUrl)

	if uploaded {
		t.Error("File couldn't upload cause of get chunked file error")
	}
}

func TestFileChunkNotUpload(t *testing.T) {
	fileService := MMFileServiceMock{getChunkedFileError: errors.New("error")}
	fileChunkService := FileChunkServiceMock{}
	testedInstance := ChunkFileUploadServiceImpl{MMFileService: fileService, fileChunkService: fileChunkService}

	var fileSize int64 = 10
	var chunkFileSize int64 = 1
	fileInfo := &model.FileInfo{Size: fileSize}
	url := "test_url"
	mmFileUrl := "mm_file_url"

	uploaded := testedInstance.uploadChunks(chunkFileSize, fileInfo, url, mmFileUrl)

	if uploaded {
		t.Error("File couldn't upload cause of get chunked file error")
	}
}

func TestUploadFailedDuringChunkUpload(t *testing.T) {
	fileService := MMFileServiceMock{}
	fileChunkService := FileChunkServiceMock{uploadFileChunkError: errors.New("error")}
	testedInstance := ChunkFileUploadServiceImpl{MMFileService: fileService, fileChunkService: fileChunkService}

	var fileSize int64 = 10
	var chunkFileSize int64 = 1
	fileInfo := &model.FileInfo{Size: fileSize}
	url := "test_url"
	mmFileUrl := "mm_file_url"

	uploaded := testedInstance.uploadChunks(chunkFileSize, fileInfo, url, mmFileUrl)

	if uploaded {
		t.Error("File couldn't upload cause of  chunk upload failed")
	}
}
