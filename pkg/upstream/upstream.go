package upstream

import (
	"fmt"
	"github.com/united-manufacturing-hub/expiremap/pkg/expiremap"
	"github.com/zeebo/xxh3"
	"io"
	"net/http"
	"time"
)

type DataStatus struct {
	data   []byte
	status int
}

var responseMap = expiremap.NewEx[string, DataStatus](time.Minute, time.Second*30)

func callProxy(url string, forceUpstream bool) ([]byte, error, int) {
	rawUrlHash := xxh3.Hash128([]byte(url)).Bytes()
	urlHash := string(rawUrlHash[:])

	if !forceUpstream {
		// Check if we have a cached response
		if k, found := responseMap.Get(urlHash); found {
			kX := *k
			// refresh the cache in the background
			go func() {
				_, _, _ = callProxy(url, true)
			}()
			return kX.data, nil, kX.status
		}
	}

	get, err := http.Get(url)

	if err != nil {
		return nil, err, 0
	}
	// Read the response body
	var body []byte
	body, err = io.ReadAll(get.Body)
	if err != nil {
		return nil, err, 0
	}

	// If the response body is reasonable small, cache it for later
	if len(body) < 1024*1024 {
		responseMap.Set(urlHash, DataStatus{data: body, status: get.StatusCode})
	}

	return body, nil, get.StatusCode
}

func CallUpstreamList(uri string) ([]byte, error, int) {
	//https://proxy.golang.org/:URI/@v/list
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/@v/list", uri), false)

}

func CallUpstreamInfo(uri, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:URI/@v/:VERSION.info
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.info", uri, version), false)
}

func CallUpstreamMod(uri, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:URI/@v/:VERSION.mod
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.mod", uri, version), false)
}

func CallUpstreamZip(uri, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:URI/@v/:VERSION.zip
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.zip", uri, version), false)

}

func CallUpstreamLatest(uri string) ([]byte, error, int) {
	//https://proxy.golang.org/:URI/@latest
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/@latest", uri), false)
}

func CallUpstreamSumDB(trail string) ([]byte, error, int) {
	return callProxy(fmt.Sprintf("https://%s/%s", trail), false)
}
