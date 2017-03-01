package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
)

type BasicHttpClient struct {
	HTTPClient *http.Client
	username string
	password string
	apiUrl string
	protocol string
}

func NewBasicHttpClient(username, password, apiUrl string) *BasicHttpClient {
	hClient := &BasicHttpClient{
		username: username,
		password: password,
		apiUrl: apiUrl,
		protocol: "https",
		HTTPClient: http.DefaultClient,
	}
	return hClient
}

func (client *BasicHttpClient) DoHttpRequest(path string, masks string, filters string, requestType string, requestBody *bytes.Buffer) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s@%s/%s", client.protocol, client.username, client.password, client.apiUrl, path)
	if requestType == "GET" {
		if filters !="" && masks !="" {
			url += "?objectFilter=" + filters + "&objectMask=filteredMask[" + masks + "]"
		} else if filters !="" {
			url += "?objectFilter=" + filters
		} else if masks !=""{
			url += "?objectMask=" + masks
		}
	}
	return client.makeHttpRequest(url, requestType, requestBody)
}

func (client *BasicHttpClient) makeHttpRequest(url string, requestType string, requestBody *bytes.Buffer) ([]byte, error) {
	
	//Create request Object
	req, err := http.NewRequest(requestType, url, requestBody)
	if err != nil {
		return nil, err
	}
	
	//Log request
	requestInfo, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "\n[http-request]:\n%s\n", string(requestInfo))
	
	//Do http Call
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//Log response
	responseInfo, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "\n[http-response]:\n%s\n", string(responseInfo))
	
	//Read response
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err == nil && resp.StatusCode >= 400 {		
		var decodedResponse map[string]interface{}
		parseErr := json.Unmarshal(responseBody, &decodedResponse)
		if parseErr == nil {
			if errString, ok := decodedResponse["error"]; !ok {
				err = errors.New(fmt.Sprintf("Operation failed with Http Error:%d", resp.StatusCode))
			} else {
				err = errors.New(errString.(string))
			}
		}
	}
	return responseBody, err
}
