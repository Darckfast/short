package utils

import (
	"io"
	"net/http"
	"strings"

	"github.com/syumai/workers/cloudflare/fetch"
)

func DoReverseProxy(remoteUrl string, w http.ResponseWriter, r *http.Request) error {
	remoteUrl = strings.Replace(remoteUrl, "reverse:", "", 1)

	req, err := fetch.NewRequest(r.Context(), r.Method, remoteUrl, r.Body)
	req.Header = r.Header.Clone()

	if err != nil {
		return err
	}

	cli := fetch.NewClient()
	resp, err := cli.Do(req, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	io.Copy(w, resp.Body)
	w.WriteHeader(resp.StatusCode)

	for name, values := range r.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	return nil
}
