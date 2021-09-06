package pkg

import (
	"strings"

	"github.com/hashicorp/consul/api"
)

type KV struct {
	Key   string
	Value string
}

func KeysToMap(pairs api.KVPairs) map[string]interface{} {
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

func MapToKeys(data map[string]interface{}) ([]*KV, error) {
	return traverse("", data)
}
