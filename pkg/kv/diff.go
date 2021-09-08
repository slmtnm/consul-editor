package kv

import (
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
)

type Map map[string]interface{}

type Diff struct {
	added, removed Map
	ModifiedKeys   []string
}

func CalculateDiff(oldMap Map, newMap Map) Diff {
	return calculateDiffHelper("", oldMap, newMap)
}

func (d1 *Diff) Append(d2 Diff) {
	for sk, sv := range d2.added {
		d1.added[sk] = sv
	}
	for sk, sv := range d2.removed {
		d1.removed[sk] = sv
	}
	d1.ModifiedKeys = append(d1.ModifiedKeys, d2.ModifiedKeys...)
}

func (d Diff) Print() {
	marked := make(map[string]bool)

	red := func(k, v interface{}) {
		out, err := yaml.Marshal(v)
		if err != nil {
			panic(err)
		}
		color.Red("--- %s:\n%s", k, strings.TrimSpace(string(out)))
	}
	green := func(k, v interface{}) {
		out, err := yaml.Marshal(v)
		if err != nil {
			panic(err)
		}
		color.Green("+++ %s:\n%s", k, strings.TrimSpace(string(out)))
	}

	for _, k := range d.ModifiedKeys {
		red(k, d.removed[k])
		green(k, d.added[k])
		marked[k] = true
	}

	for k, v := range d.removed {
		if _, ok := marked[k]; !ok {
			red(k, v)
		}
	}
	for k, v := range d.added {
		if _, ok := marked[k]; !ok {
			green(k, v)
		}
	}
}

func calculateDiffHelper(path string, oldMap Map, newMap Map) Diff {
	diff := Diff{
		make(Map),
		make(Map),
		[]string{},
	}

	constructKeyPath := func(path, key string) string {
		if path == "" {
			return key
		} else {
			return path + "/" + key
		}
	}

	// find added or modified fields
	for key, newValue := range newMap {
		keyPath := constructKeyPath(path, key)

		oldValue, ok := oldMap[key]
		if !ok {
			diff.added[keyPath] = newValue
			continue
		}

		// compare oldValue and newValue
		if oldString, ok := oldValue.(string); ok {
			if newString, ok := newValue.(string); ok { // old value and new value are strings
				if oldString != newString {
					diff.removed[keyPath] = oldString
					diff.added[keyPath] = newString
					diff.ModifiedKeys = append(diff.ModifiedKeys, keyPath)
				}
			} else { // old value is string but new value is map
				diff.removed[keyPath] = oldString
				diff.added[keyPath] = newValue
				diff.ModifiedKeys = append(diff.ModifiedKeys, keyPath)
			}
		} else {
			if newString, ok := newValue.(string); ok { // old value is map, but new value is string
				diff.removed[keyPath] = oldValue
				diff.added[keyPath] = newString
				diff.ModifiedKeys = append(diff.ModifiedKeys, keyPath)
			} else { // old value and new values are both maps, compare them recursive
				diff.Append(calculateDiffHelper(
					keyPath,
					oldValue.(map[string]interface{}),
					newValue.(map[string]interface{}),
				))
			}
		}
	}

	// find removed fields
	for key, oldValue := range oldMap {
		keyPath := constructKeyPath(path, key)
		if _, ok := newMap[key]; !ok {
			diff.removed[keyPath] = oldValue
		}
	}

	return diff
}

func (d Diff) Apply() {
	for k := range d.removed {
		_, err := kv.DeleteTree(k, nil)
		if err != nil {
			panic(err)
		}
	}
	for k, v := range d.added {
		_, err := kv.Put(&api.KVPair{Key: k, Value: []byte(v.(string)), Flags: 42}, nil)
		if err != nil {
			panic(err)
		}
	}
}
