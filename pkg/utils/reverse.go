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

	w.Header().Add("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Add("Content-Length", r.Header.Get("Content-Length"))

	logger.InfoContext(ctx, "reverse proxy response", "status", resp.StatusCode, "proxy", true)

	return nil
}
