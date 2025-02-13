package main

import (
	"net/http"

	short "main/api/v1"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("GET /", short.Handler)
	workers.Serve(nil)
}
