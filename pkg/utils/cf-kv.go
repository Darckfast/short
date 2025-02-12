package utils

import (
	"github.com/syumai/workers/cloudflare/kv"
)

func GetKVUrl(key string) (string, error) {
	counterKV, _ := kv.NewNamespace("short")

	return counterKV.GetString(key, nil)
}
