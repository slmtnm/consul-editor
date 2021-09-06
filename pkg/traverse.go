package pkg

import (
	"strconv"
)

func traverse(path string, j interface{}) ([]*KV, error) {
	kvs := []*KV{}

	pathPre := ""
	if path != "" {
		pathPre = path + "/"
	}

	switch v := j.(type) {
	case []interface{}:
		for sk, sv := range v {
			skvs, err := traverse(pathPre+strconv.Itoa(sk), sv)
			if err != nil {
				return nil, err
			}
			kvs = append(kvs, skvs...)
		}
	case map[string]interface{}:
		for sk, sv := range v {
			skvs, err := traverse(pathPre+sk, sv)
			if err != nil {
				return nil, err
			}
			kvs = append(kvs, skvs...)
		}
	case map[interface{}]interface{}:
		for sk, sv := range v {
			skvs, err := traverse(pathPre+sk.(string), sv)
			if err != nil {
				return nil, err
			}
			kvs = append(kvs, skvs...)
		}
	case float64:
		kvs = append(kvs, &KV{Key: path, Value: strconv.FormatFloat(v, 'f', -1, 64)})
	case bool:
		kvs = append(kvs, &KV{Key: path, Value: strconv.FormatBool(v)})
	case nil:
		kvs = append(kvs, &KV{Key: path, Value: ""})
	default:
		kvs = append(kvs, &KV{Key: path, Value: j.(string)})
	}

	return kvs, nil
}
