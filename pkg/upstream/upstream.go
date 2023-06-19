package upstream

import (
	"fmt"
	"github.com/united-manufacturing-hub/expiremap/pkg/expiremap"
	"github.com/zeebo/xxh3"
	"go.uber.org/zap"
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

	zap.S().Debugf("Calling proxy: %s", url)
	get, err := http.Get(url)
	zap.S().Debugf("Response status: %s", get.Status)
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

func CallUpstreamList(domain, user, repo string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@v/list
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@v/list", domain, user, repo), false)
}

func CallUpstreamInfo(domain, user, repo, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@v/:VERSION.info
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@v/%s.info", domain, user, repo, version), false)
}

func CallUpstreamMod(domain, user, repo, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@v/:VERSION.mod
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@v/%s.mod", domain, user, repo, version), false)
}
func CallUpstreamModuleSource(domain, user, repo, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@v/:VERSION.zip
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@v/%s.zip", domain, user, repo, version), false)

}

func CallUpstreamLatest(domain, user, repo string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@latest
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@latest", domain, user, repo), false)
}

func CallUpstreamSumDB(domain, trail string) ([]byte, error, int) {
	return callProxy(fmt.Sprintf("https://%s/%s", domain, trail), false)
}
