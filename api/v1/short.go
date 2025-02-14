package short

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"main/pkg/utils"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
)

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx, wg := multilogger.SetupContext(&multilogger.SetupOps{
		Request:           r,
		BaselimeApiKey:    os.Getenv("BASELIME_API_KEY"),
		AxiomApiKey:       os.Getenv("AXIOM_API_KEY"),
		BetterStackApiKey: os.Getenv("BETTERSTACK_API_KEY"),
		ServiceName:       os.Getenv("VERCEL_GIT_REPO_SLUG"),
	})

	defer func() {
		wg.Wait()
		ctx.Done()
	}()

	logger.InfoContext(ctx, "Processing request")

	urlPath := r.PathValue("id")
	subFolder := r.URL.Query().Get("f")

	if subFolder != "" {
		urlPath = subFolder + "/" + urlPath
	}

	longUrl, err := utils.GetKVUrl(urlPath)
	if err != nil {
		fmt.Fprintf(w, "<h1>no result found</h1>")
		logger.ErrorContext(ctx, "error getting KV value", "status", 200, "error", err.Error())
		return
	}

	if longUrl == "<null>" {
		fmt.Fprintf(w, "<h1>no result found</h1>")
		logger.WarnContext(ctx, "no short link found", "status", 200)
		return
	}

	if strings.Contains(longUrl, "reverse:") {
		err := utils.DoReverseProxy(longUrl, w, r)
		if err != nil {
			fmt.Fprintf(w, "<h1>no result found</h1>")
			logger.ErrorContext(ctx, "error proxing reqeust", "status", 200, "error", err.Error())
		}

		logger.InfoContext(ctx, "request completed", "status", 200)
		return
	}

	w.WriteHeader(301)
	w.Header().Set("Cache-Control", "604800")
	w.Header().Set("Location", longUrl)
	w.Write([]byte{}) // wasm require empty body or it error out

	logger.InfoContext(ctx, "request completed", "status", 301)
}
