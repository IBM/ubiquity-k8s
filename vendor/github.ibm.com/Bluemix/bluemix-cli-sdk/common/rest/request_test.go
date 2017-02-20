package rest

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestQueryParam(t *testing.T) {
	assert := assert.New(t)

	req, err := GetRequest("http://www.example.com?foo=fooVal1").
		Query("foo", "fooVal2").
		Query("bar", "bar Val").
		Build()

	assert.NoError(err)
	assert.Contains(req.URL.String(), "foo=fooVal1")
	assert.Contains(req.URL.String(), "foo=fooVal2")
	assert.Contains(req.URL.String(), "bar=bar+Val")
}

func TestRequestHeader(t *testing.T) {
	assert := assert.New(t)

	req, err := GetRequest("http://www.example.com").
		Set("Accept", "application/json").
		Add("Accept-Encoding", "compress").
		Add("Accept-Encoding", "gzip").
		Build()

	assert.NoError(err)
	assert.Equal("application/json", req.Header.Get("Accept"))
	assert.Equal([]string{"compress", "gzip"}, req.Header["Accept-Encoding"])
}

func TestRequestFormText(t *testing.T) {
	assert := assert.New(t)

	req, err := PostRequest("http://www.example.com").
		Field("foo", "bar").
		Build()

	assert.NoError(err)
	err = req.ParseForm()
	assert.NoError(err)
	assert.Equal(formUrlEncodedContentType, req.Header.Get(contentType))
	assert.Equal("bar", req.FormValue("foo"))
}

func TestRequestFormMultipart(t *testing.T) {
	assert := assert.New(t)

	var prepareFileWithContent = func(text string) (*os.File, error) {
		f, err := ioutil.TempFile("", "BluemixCliRestTest")
		if err != nil {
			return nil, err
		}
		_, err = f.WriteString(text)
		if err != nil {
			return nil, err
		}

		_, err = f.Seek(0, 0)
		if err != nil {
			return nil, err
		}

		return f, err
	}

	f, err := prepareFileWithContent("12345")
	assert.NoError(err)

	req, err := PostRequest("http://www.example.com").
		Field("foo", "bar").
		File("file1", File{Name: f.Name(), Content: f}).
		File("file2", File{Name: "file2.txt", Content: strings.NewReader("abcde"), Type: "text/plain"}).
		Build()

	assert.NoError(err)
	assert.Contains(req.Header.Get(contentType), "multipart/form-data")

	err = req.ParseMultipartForm(int64(5000))
	assert.NoError(err)

	assert.Equal(1, len(req.MultipartForm.Value))
	assert.Equal("bar", req.MultipartForm.Value["foo"][0])

	assert.Equal(2, len(req.MultipartForm.File))

	assert.Equal(1, len(req.MultipartForm.File["file1"]))
	assert.Equal(f.Name(), req.MultipartForm.File["file1"][0].Filename)
	assert.Equal("application/octet-stream", req.MultipartForm.File["file1"][0].Header.Get("Content-Type"))

	assert.Equal(1, len(req.MultipartForm.File["file2"]))
	assert.Equal("file2.txt", req.MultipartForm.File["file2"][0].Filename)
	assert.Equal("text/plain", req.MultipartForm.File["file2"][0].Header.Get("Content-Type"))

	b1 := new(bytes.Buffer)
	f1, _ := req.MultipartForm.File["file1"][0].Open()
	io.Copy(b1, f1)
	assert.Equal("12345", string(b1.Bytes()))

	b2 := new(bytes.Buffer)
	f2, _ := req.MultipartForm.File["file2"][0].Open()
	io.Copy(b2, f2)
	assert.Equal("abcde", string(b2.Bytes()))
}

func TestRequestJSON(t *testing.T) {
	assert := assert.New(t)

	var foo = struct {
		Name string
	}{
		Name: "bar",
	}

	req, err := PostRequest("http://www.example.com").Body(&foo).Build()
	assert.NoError(err)
	body, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)
	assert.Equal("{\"Name\":\"bar\"}", string(body))
}
