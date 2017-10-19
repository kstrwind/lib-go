package larix

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//HTTP client
type HttpClient struct {
	Ip      string // ip: 127.0.0.1
	Port    int
	Headers map[string]string
	//The timeout includes connection time, any
	// redirects, and reading the response body
	Timeout_ms int64
	Host       string
}

//simple check if status OK
func isHttpOk(status int) bool {
	if status >= 400 {
		return false
	}
	return true
}

//todo:
//add Headers and so on
func (hc *HttpClient) AddHeader(field string, value string) {
}

func (hc *HttpClient) Request(method string, uri string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d%s", hc.Ip, hc.Port, uri)

	client := &http.Client{
		Timeout: time.Duration(hc.Timeout_ms) * time.Millisecond,
	}

	r_method := strings.ToUpper(method)
	req, err := http.NewRequest(r_method, url, body)
	if err != nil {
		return []byte{}, err
	}

	if hc.Host != "" {
		req.Host = hc.Host
	}

	for k, v := range hc.Headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log_info := map[string]interface{}{
			"message": "http request failed",
			"method":  method,
			"url":     url,
			"err_msg": err.Error(),
		}
		LogWarn(log_info)
		return []byte{}, err
	}
	defer resp.Body.Close()
	if !isHttpOk(resp.StatusCode) {
		log_info := map[string]interface{}{
			"message":   "http request status wrong",
			"method":    method,
			"http_code": resp.StatusCode,
			"headers":   fmt.Sprintf("%v", resp.Header),
			"url":       url,
		}
		LogWarn(log_info)
		return []byte{}, errors.New(fmt.Sprintf("http status is %d", resp.StatusCode))
	}

	//decode json
	res_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log_info := map[string]interface{}{
			"message":   "http response body read failed",
			"method":    method,
			"http_code": resp.StatusCode,
			"headers":   fmt.Sprintf("%v", resp.Header),
			"error":     err.Error(),
			"url":       url,
		}
		LogWarn(log_info)
		return []byte{}, err
	}

	return res_body, err
}
