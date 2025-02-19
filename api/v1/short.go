package short

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"main/pkg/utils"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/fetch"
)

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx, wg := multilogger.SetupContext(&multilogger.SetupOps{
		Request:     r,
		AxiomApiKey: cloudflare.Getenv("AXIOM_API_KEY"),
		ServiceName: cloudflare.Getenv("VERCEL_GIT_REPO_SLUG"),
		RequestGen: func(args multilogger.SendLogsArgs) {
			args.MaxQueue <- 1
			args.Wg.Add(1)

			req, _ := fetch.NewRequest(args.Ctx, args.Method, args.Url, bytes.NewBuffer(*args.Body))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", args.Bearer)

			client := fetch.NewClient()

			go func() {
				defer args.Wg.Done()
				client.Do(req, nil)
				<-args.MaxQueue
			}()
		},
	})

	defer func() {
		wg.Wait()
		ctx.Done()
	}()

	urlPath := r.URL.Path
	subFolder := r.URL.Query().Get("f")
	if urlPath == "" || urlPath == "/" {
		urlPath = "index"
	}

	urlPath = strings.Replace(urlPath, "/", "", 1)

	if len(urlPath) > 0 && urlPath[len(urlPath)-1:] == "/" {
		urlPath = urlPath[:len(urlPath)-1]
	}

	urlPathParts := strings.Split(urlPath, "/")
	subPath := ""
	if len(urlPathParts) > 1 {
		subPath = strings.Join(urlPathParts[1:], "/")
		urlPath = urlPathParts[0]
	}

	if subFolder != "" {
		urlPath = subFolder + "/" + urlPath
	}

	longUrl, err := utils.GetKVUrl(urlPath)
	if err != nil {
		fmt.Fprintf(w, "<h1>no result found</h1>")
		logger.ErrorContext(ctx, "error getting KV value", "status", 200, "error", err.Error(), "url", urlPath)
		return
	}

	if longUrl == "<null>" {
		fmt.Fprintf(w, "<h1>no result found</h1>")
		logger.InfoContext(ctx, "no short link found", "status", 200, "url", urlPath)
		return
	}

	flags := strings.Split(longUrl, "|")

	doReverseProxy := false
	cacheControl := ""
	if len(flags) > 0 {
		for _, flag := range flags {
			if flag == "reverse" {
				doReverseProxy = true
			} else if strings.Contains(flag, "cache") {
				cacheControl = strings.Split(flag, "=")[1]
			} else if strings.Contains(flag, "https://") {
				longUrl = flag
			}
		}
	}

	if subPath != "" {
		longUrl += "/" + subPath
	}

	if doReverseProxy {
		err := utils.DoReverseProxy(ctx, longUrl, w, r)
		if cacheControl != "" {
			w.Header().Set("Cache-Control", "public, max-age="+cacheControl)
		}

		logger.InfoContext(ctx, "request completed", "cache", cacheControl, "url", urlPath)
		if err != nil {
			fmt.Fprintf(w, "<h1>no result found</h1>")
		}

		return
	}
	w.WriteHeader(301)
	w.Header().Set("Location", longUrl)
	w.Header().Set("Cache-Control", "public, max-age="+cacheControl)
	w.Write([]byte{}) // wasm require empty body or it error out

	logger.InfoContext(ctx, "request completed", "status", 301, "cache", cacheControl, "url", urlPath)
}
