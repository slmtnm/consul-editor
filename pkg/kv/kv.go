package kv

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/consul/api"
)

type KV struct {
	Key   string
	Value string
}

var kv *api.KV

func init() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	kv = client.KV()
}

func keysToMap(pairs api.KVPairs) map[string]interface{} {
	m := make(map[string]interface{})
	for _, pair := range pairs {
		path := strings.Split(pair.Key, "/")
		parent := m

		for i, segment := range path {
			if i == len(path)-1 {
				if segment != "" {
					parent[segment] = string(pair.Value)
				}
			} else {
				if parent[segment] == nil {
					parent[segment] = make(map[string]interface{})
				}
				switch v := parent[segment].(type) {
				case map[string]interface{}:
					parent = v
				default:
					delete(parent, segment)
				}
			}
		}
	}
	return m
}

func GetKeys(path string) (map[string]interface{}, error) {
	keys, _, err := kv.List(path, nil)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("key '%s' not found", path)
	}

	keysMap := keysToMap(keys)
	return keysMap, nil
}
