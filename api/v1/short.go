package short

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"main/pkg/utils"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/fetch"
)

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx, wg := multilogger.SetupContext(&multilogger.SetupOps{
		Request:           r,
		BaselimeApiKey:    cloudflare.Getenv("BASELIME_API_KEY"),
		AxiomApiKey:       cloudflare.Getenv("AXIOM_API_KEY"),
		BetterStackApiKey: cloudflare.Getenv("BETTERSTACK_API_KEY"),
		ServiceName:       cloudflare.Getenv("VERCEL_GIT_REPO_SLUG"),
		RequestGen: func(maxQueue chan int, wg *sync.WaitGroup, method, url, bearer string, body *[]byte) {
			maxQueue <- 1
			wg.Add(1)

			req, _ := fetch.NewRequest(r.Context(), method, url, bytes.NewBuffer(*body))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", bearer)

			client := fetch.NewClient()

			go func() {
				defer wg.Done()

				client.Do(req, nil)

				<-maxQueue
			}()
		},
	})

	defer func() {
		wg.Wait()
		ctx.Done()
	}()

	logger.InfoContext(ctx, "processing request")

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
		logger.InfoContext(ctx, "no short link found", "status", 200)
		return
	}

	if strings.Contains(longUrl, "reverse:") {
		err := utils.DoReverseProxy(ctx, longUrl, w, r)
		if err != nil {
			fmt.Fprintf(w, "<h1>no result found</h1>")
		}

		return
	}

	w.WriteHeader(301)
	w.Header().Set("Cache-Control", "604800")
	w.Header().Set("Location", longUrl)
	w.Write([]byte{}) // wasm require empty body or it error out

	logger.InfoContext(ctx, "request completed", "status", 301)
}
