package upstream

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func callProxy(url string) ([]byte, error, int) {
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

	return body, nil, get.StatusCode
}

func CallUpstreamList(domain, user, repo string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@v/list
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@v/list", domain, user, repo))
}

func CallUpstreamInfo(domain, user, repo, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@v/:VERSION.info
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@v/%s.info", domain, user, repo, version))
}

func CallUpstreamMod(domain, user, repo, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@v/:VERSION.mod
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@v/%s.mod", domain, user, repo, version))
}
func CallUpstreamModuleSource(domain, user, repo, version string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@v/:VERSION.zip
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@v/%s.zip", domain, user, repo, version))

}

func CallUpstreamLatest(domain, user, repo string) ([]byte, error, int) {
	//https://proxy.golang.org/:DOMAIN/:USER/:REPO/@latest
	return callProxy(fmt.Sprintf("https://proxy.golang.org/%s/%s/%s/@latest", domain, user, repo))
}
