package utils

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
	"github.com/syumai/workers/cloudflare/fetch"
)

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func DoReverseProxy(ctx context.Context, remoteUrl string, w http.ResponseWriter, r *http.Request) error {
	req, err := fetch.NewRequest(r.Context(), r.Method, remoteUrl, r.Body)
	if err != nil {
		logger.ErrorContext(ctx, "error creating proxying request", "status", 500, "error", err.Error())
		return err
	}

	req.Header = r.Header.Clone()

	cli := fetch.NewClient()
	resp, err := cli.Do(req, nil)
	if err != nil {
		logger.ErrorContext(ctx, "error reversing proxying request", "status", 500, "error", err.Error())
		return err
	}

	defer resp.Body.Close()
	io.Copy(w, resp.Body)

	w.WriteHeader(resp.StatusCode)
	w.Header().Add("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Add("Content-Length", resp.Header.Get("Content-Length"))
	w.Header().Add("Cache-Control", resp.Header.Get("Cache-Control"))
	w.Header().Add("Content-Encoding", resp.Header.Get("Content-Encoding"))
	w.Header().Add("Content-Security-Policy", resp.Header.Get("Content-Security-Policy"))
	w.Header().Add("Reporting-Endpoints", resp.Header.Get("Reporting-Endpoints"))
	w.Header().Add("Content-Security-Policy-Report-Only", resp.Header.Get("Content-Security-Policy-Report-Only"))

	return nil
}
