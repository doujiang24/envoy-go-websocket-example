package main

import (
	"bytes"
	"net/http"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

// Just a demo to request a remote server
func checkData(data []byte) bool {
	req, err := http.NewRequest("POST", "https://httpbin.org/post", bytes.NewBuffer(data))
	if err != nil {
		api.LogDebugf("new request error: %v", err)
		return false
	}

	httpc := &http.Client{}
	resp, err := httpc.Do(req)
	if err != nil {
		api.LogDebugf("query request error: %v", err)
		return false
	}
	resp.Body.Close()

	return resp.StatusCode == 200
}
