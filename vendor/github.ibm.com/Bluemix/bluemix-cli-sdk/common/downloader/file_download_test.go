package downloader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/common/file_helpers"
)

type DownloadTestSuite struct {
	suite.Suite
	downloader *FileDownloader
}

func TestDownloadTestSuite(t *testing.T) {
	suite.Run(t, new(DownloadTestSuite))
}

func (suite *DownloadTestSuite) SetupTest() {
	tmpDir, err := ioutil.TempDir("", "testfiledownload")
	suite.NoError(err)
	suite.downloader = NewFileDownloader(tmpDir)
}

func (suite *DownloadTestSuite) TearDownTest() {
	if suite.downloader.SaveDirectory() != "" {
		os.RemoveAll(suite.downloader.SaveDirectory())
	}
}

func (suite *DownloadTestSuite) TestDownload_GetFileNameFromHeader() {
	assert := assert.New(suite.T())

	resp := "this is the file content"
	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="test.txt"`)
		fmt.Fprint(w, resp)
	}))

	dest, size, err := suite.downloader.Download(fileServer.URL)

	assert.NoError(err)
	assert.Equal(dest, filepath.Join(suite.downloader.SaveDirectory(), "test.txt"))
	assert.True(file_helpers.FileExists(dest))
	assert.Equal(size, int64(len(resp)))
}

func (suite *DownloadTestSuite) TestDownload_GetFileNameFromURL() {
	assert := assert.New(suite.T())

	fileServer := CreateServerReturnContent("123")
	defer fileServer.Close()

	dest, _, err := suite.downloader.Download(fileServer.URL + "/test.txt")

	assert.NoError(err)
	assert.Equal(dest, filepath.Join(suite.downloader.SaveDirectory(), "test.txt"))
	assert.True(file_helpers.FileExists(dest))
}

func (suite *DownloadTestSuite) TestDownload_SaveAsIndexHTML() {
	assert := assert.New(suite.T())

	fileServer := CreateServerReturnContent("123")
	defer fileServer.Close()

	dest, _, err := suite.downloader.Download(fileServer.URL)

	assert.NoError(err)
	assert.Equal(dest, filepath.Join(suite.downloader.SaveDirectory(), "index.html"))
	assert.True(file_helpers.FileExists(dest))
}

func (suite *DownloadTestSuite) TestDownloadAs() {
	assert := assert.New(suite.T())

	fileServer := CreateServerReturnContent("123")
	defer fileServer.Close()

	dest, _, err := suite.downloader.DownloadAs(fileServer.URL, "out")

	assert.NoError(err)
	assert.Equal(dest, filepath.Join(suite.downloader.SaveDirectory(), "out"))
	assert.True(file_helpers.FileExists(dest))
}

func (suite *DownloadTestSuite) TestDownloadSaveDirNotSet() {
	assert := assert.New(suite.T())

	fileServer := CreateServerReturnContent("123")
	defer fileServer.Close()

	suite.downloader = NewFileDownloader("")
	dest, _, err := suite.downloader.Download(fileServer.URL + "/test.txt")
	defer os.Remove(dest)

	assert.NoError(err)
	assert.True(file_helpers.FileExists(dest))
	assert.Equal(filepath.Dir(dest), ".")
}

func (suite *DownloadTestSuite) TestDownloadRequestHeader() {
	assert := assert.New(suite.T())

	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get("Accept-Language"), "en_US")
	}))
	defer fileServer.Close()

	h := http.Header{}
	h.Add("Accept-Language", "en_US")
	suite.downloader.SetHeader(h)
	suite.downloader.Download(fileServer.URL + "/test.txt")
}

func CreateServerReturnContent(content string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "this is the file content")
	}))
}
