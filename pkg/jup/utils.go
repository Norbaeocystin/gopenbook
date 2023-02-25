package jup

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

//Send get request method. Returns bytes and error
func Get(urlstring string) ([]byte, error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", urlstring, nil)
	if err != nil {
		return []byte(""), err
	}
	request.Header.Set("Content-Type", "application/json")
	// Make request
	response, err := client.Do(request)
	if err != nil {
		return []byte(""), err
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return bodyBytes, err
}

//Send get request method. Returns bytes and error
func Post(urlstring string, body []byte) ([]byte, error) {
	client := http.Client{}
	request, err := http.NewRequest("POST", urlstring, bytes.NewReader(body))
	if err != nil {
		return []byte(""), err
	}
	request.Header.Set("Content-Type", "application/json")
	// Make request
	response, err := client.Do(request)
	if err != nil {
		return []byte(""), err
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return bodyBytes, err
}
