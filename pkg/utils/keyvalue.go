package utils

import (
	"github.com/syumai/workers/cloudflare/kv"
)

func GetKVUrl(key string) (string, error) {
	counterKV, err := kv.NewNamespace("KV_SHORT_BINDING")
	if err != nil {
		return "", err
	}

	return counterKV.GetString(key, nil)
}
