package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type FileDownloader struct {
	saveDir string
	h       http.Header
	client  *http.Client
}

func NewFileDownloader(saveDir string) *FileDownloader {
	return &FileDownloader{
		saveDir: saveDir,
		client:  http.DefaultClient,
	}
}

func (d *FileDownloader) DownloadAs(url string, outputFileName string) (string, int64, error) {
	if outputFileName == "" {
		return "", 0, fmt.Errorf("download: output file name is empty")
	}
	return d.download(url, outputFileName)
}

func (d *FileDownloader) Download(url string) (string, int64, error) {
	return d.download(url, "")
}

func (d *FileDownloader) download(url string, outputFileName string) (string, int64, error) {
	req, err := d.makeRequest(url)
	if err != nil {
		return "", 0, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", 0, fmt.Errorf("Error downloading the file. Remote return code %d", resp.StatusCode)
	}

	if outputFileName == "" {
		outputFileName = d.determinOutputFileName(resp, url)
	}

	dest := filepath.Join(d.saveDir, outputFileName)

	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return dest, 0, err
	}
	defer f.Close()

	size, err := io.Copy(f, resp.Body)
	if err != nil {
		return dest, size, err
	}

	return dest, size, nil
}

func (d *FileDownloader) SaveDirectory() string {
	return d.saveDir
}

func (d *FileDownloader) determinOutputFileName(resp *http.Response, url string) string {
	fname := getFileNameFromHeader(resp.Header.Get("Content-Disposition"))

	if fname == "" {
		fname = getFileNameFromUrl(url)
	}

	if fname == "" {
		fname = "index.html"
	}

	saveDir := d.saveDir
	if saveDir == "" {
		saveDir = "."
	}

	return fname
}

func getFileNameFromUrl(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return ""
	}

	p := u.Path
	p = path.Clean(p)
	if p == "." {
		return ""
	}

	fields := strings.Split(p, "/")
	if len(fields) == 0 {
		return ""
	}

	return fields[len(fields)-1]
}

func getFileNameFromHeader(header string) string {
	if header == "" {
		return ""
	}

	for _, field := range strings.Split(header, ";") {
		field = strings.TrimSpace(field)

		if strings.HasPrefix(field, "filename=") {
			name := strings.TrimLeft(field, "filename=")
			return strings.Trim(name, `"`)
		}
	}

	return ""
}

func (d *FileDownloader) SetHeader(h http.Header) {
	d.h = h
}

func (d *FileDownloader) makeRequest(downloadUrl string) (*http.Request, error) {
	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header = d.h
	return req, nil
}
