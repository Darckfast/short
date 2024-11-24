package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
)

const (
	BLOB_API_VERSION = "4"
	DEFAULT_BASE_URL = "https://blob.vercel-storage.com"
)

type ListBlobResult struct {
	Blobs []struct {
		URL string `json:"url"`
	} `json:"blobs"`
}

func FindBlob(fileName string) *ListBlobResult {
	req, _ := http.NewRequest("GET", DEFAULT_BASE_URL, nil)
	req.Header.Set("x-api-version", BLOB_API_VERSION)
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	req.Header.Set("Authorization", "Bearer "+token)
	q := req.URL.Query()

	q.Add("prefix", "short/"+fileName)
	q.Add("limit", "1")
	q.Add("mode", "folded")

	req.URL.RawQuery = q.Encode()

	client := &http.Client{
		Timeout: time.Second * 3,
	}

	res, _ := client.Do(req)
	defer res.Body.Close()
	var result ListBlobResult

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil
	}

	return &result
}

func DownloadBlob(fileUrl string) string {
	req, _ := http.NewRequest("GET", fileUrl, nil)
	req.Header.Set("x-api-version", BLOB_API_VERSION)
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{
		Timeout: time.Second * 3,
	}

	res, _ := client.Do(req)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	return string(body)
}

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx, wg := multilogger.SetupContext(&multilogger.SetupOps{
		Request:        r,
		BaselimeApiKey: os.Getenv("BASELIME_API_KEY"),
		AxiomApiKey:    os.Getenv("AXIOM_API_KEY"),
		ServiceName:    os.Getenv("VERCEL_GIT_REPO_SLUG"),
	})

	defer func() {
		wg.Wait()
		ctx.Done()
	}()

	logger.InfoContext(ctx, "Processing request")

	urlPath := strings.Split(r.URL.Path, "/")
	blobHash := urlPath[len(urlPath)-1]

	blob := FindBlob(blobHash)

	if len(blob.Blobs) == 0 {
		fmt.Fprintf(w, "<h1>no result found</h1>")
		return
	}

	longUrl := DownloadBlob(blob.Blobs[0].URL)

	w.WriteHeader(301)
	w.Header().Set("Cache-Control", "604800")
	w.Header().Set("Location", longUrl)

	logger.InfoContext(ctx, "request completed", "status", 301)
}
