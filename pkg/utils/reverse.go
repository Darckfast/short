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

	w.Header().Add("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Add("Content-Length", r.Header.Get("Content-Length"))
	w.Header().Add("Cache-Control", r.Header.Get("Cache-Control"))

	return nil
}
