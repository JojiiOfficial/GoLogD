package main

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var tr *http.Transport

var HTTPclient *http.Client

func request(url, file string, data []byte, ignoreCert bool) (string, error) {
	if tr == nil {
		tr = &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: ignoreCert},
			MaxIdleConns:        3,
			MaxConnsPerHost:     4,
			MaxIdleConnsPerHost: 3,
		}
	}
	if HTTPclient == nil {
		HTTPclient = &http.Client{Transport: tr}
	}

	var addFile string
	if strings.HasSuffix(url, "/") {
		if strings.HasPrefix(file, "/") {
			file = file[1:]
		}
		addFile = url + file
	} else {
		if strings.HasPrefix(file, "/") {
			file = file[1:]
		}
		addFile = url + "/" + file
	}
	resp, err := HTTPclient.Post(addFile, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		LogError(err.Error())
	}
	resp.Body.Close()
	return string(d), nil
}
