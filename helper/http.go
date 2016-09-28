package helper

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
)

//RequestedReturnCodeIsOK makes an HEAD or GET request. If the returncode is 2XX it will return true.
func RequestedReturnCodeIsOK(client http.Client, url, function string) bool {
	var resp *http.Response
	var err error
	switch function {
	case "HEAD":
		resp, err = client.Head(url)
	case "GET":
		resp, err = client.Get(url)
	default:
		err = errors.New("Unknown Function")
	}
	if err == nil && isReturnCodeOK(resp) {
		return true
	}
	return false
}

//SentReturnCodeIsOK makes the given request. If the returncode is 2XX it will return true and the body else the error message.
func SentReturnCodeIsOK(client http.Client, url, function string, data string) (bool, string) {
	var req *http.Request
	var resp *http.Response
	var err error

	req, err = http.NewRequest(function, url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("User-Agent", "Nagflux")
	if err != nil {
		return false, err.Error()
	}

	resp, err = client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()

	if isReturnCodeOK(resp) {
		return true, getBody(resp)
	}
	return false, resp.Status
}

func isReturnCodeOK(resp *http.Response) bool {
	return resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300
}

func getBody(resp *http.Response) string {
	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return string(body)
		}
	}
	return ""
}
