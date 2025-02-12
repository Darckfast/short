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

	urlPath := strings.Split(r.URL.Path, "/")
	subFolder := r.URL.Query().Get("f")

	blobHash := urlPath[len(urlPath)-1]

	if subFolder != "" {
		blobHash = subFolder + "/" + blobHash
	}

	longUrl, err := utils.GetKVUrl(blobHash)
	if err != nil {
		fmt.Fprintf(w, "<h1>no result found</h1>")
		logger.WarnContext(ctx, "no short link found", "status", 200, "error", err.Error())
		return
	}

	w.WriteHeader(301)
	w.Header().Set("Cache-Control", "604800")
	w.Header().Set("Location", longUrl)

	logger.InfoContext(ctx, "request completed", "status", 301)
}
