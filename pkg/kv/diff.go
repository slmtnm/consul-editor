package kv

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
)

type Diff struct {
	Added, Removed map[string]string
}

func CalculateDiff(oldMap, newMap map[string]interface{}) Diff {
	return calculateDiffHelper("", oldMap, newMap)
}

func (d1 *Diff) Append(d2 Diff) {
	for sk, sv := range d2.Added {
		d1.Added[sk] = sv
	}
	for sk, sv := range d2.Removed {
		d1.Removed[sk] = sv
	}
}

func (d Diff) Print(prefix string) {
	marked := make(map[string]bool)
	if prefix != "" {
		prefix = prefix + "/"
	}

	red := func(k string, v interface{}) {
		out, err := yaml.Marshal(v)
		if err != nil {
			panic(err)
		}
		color.Red("- %s: %s", prefix+k, strings.TrimSpace(string(out)))
	}
	green := func(k string, v interface{}) {
		out, err := yaml.Marshal(v)
		if err != nil {
			panic(err)
		}
		color.Green("+ %s: %s", prefix+k, strings.TrimSpace(string(out)))
	}

	for k, v := range d.Removed {
		if _, ok := marked[k]; !ok {
			red(k, v)
		}
	}
	for k, v := range d.Added {
		if _, ok := marked[k]; !ok {
			green(k, v)
		}
	}
}

func calculateDiffHelper(path string, oldMap, newMap map[string]interface{}) Diff {
	diff := Diff{
		make(map[string]string),
		make(map[string]string),
	}

	constructKeyPath := func(path, key string) string {
		if path == "" {
			return key
		} else {
			return path + "/" + key
		}
	}

	// find Added or modified fields
	for key, newValue := range newMap {
		keyPath := constructKeyPath(path, key)
		oldValue, ok := oldMap[key]

		if !ok {
			switch actualNewValue := newValue.(type) {
			case string:
				diff.Added[keyPath] = actualNewValue
			case map[string]interface{}:
				diff.Append(calculateDiffHelper(
					keyPath,
					map[string]interface{}{},
					actualNewValue,
				))
			default:
				panic("trash")
			}
			continue
		}

		switch actualOldValue := oldValue.(type) {
		case string:
			switch actualNewValue := newValue.(type) {
			case string: // old value and new value are strings
				if actualOldValue != actualNewValue {
					diff.Removed[keyPath] = actualOldValue
					diff.Added[keyPath] = actualNewValue
				}
			case map[string]interface{}: // old value is string but new value is map
				diff.Removed[keyPath] = actualOldValue
				diff.Append(calculateDiffHelper(
					keyPath,
					map[string]interface{}{},
					actualNewValue,
				))
			default:
				panic(fmt.Sprintf("Unsupported type of key %v: %v\n", keyPath, reflect.TypeOf(newValue)))
			}
		case map[string]interface{}:
			switch actualNewValue := newValue.(type) {
			case string: // old value is map, but new value is string
				diff.Append(calculateDiffHelper(
					keyPath,
					actualOldValue,
					map[string]interface{}{},
				))
				diff.Added[keyPath] = actualNewValue
			case map[string]interface{}: // old value and new value are maps
				diff.Append(calculateDiffHelper(
					keyPath,
					actualOldValue,
					actualNewValue,
				))
			default:
				panic(fmt.Sprintf("Unsupported type of key %v: %v\n", keyPath, reflect.TypeOf(newValue)))
			}
		}
	}

	// find Removed fields
	for key, oldValue := range oldMap {
		keyPath := constructKeyPath(path, key)
		if _, ok := newMap[key]; !ok {
			switch actualOldValue := oldValue.(type) {
			case string:
				diff.Removed[keyPath] = actualOldValue
			case map[string]interface{}:
				if len(actualOldValue) == 0 { // empty folder
					diff.Removed[keyPath+"/"] = "(folder)"
				} else {
					diff.Append(calculateDiffHelper(
						keyPath,
						actualOldValue,
						map[string]interface{}{},
					))
				}
			default:
				panic("")
			}
		}
	}

	return diff
}

func (d Diff) Apply(prefix string) {
	if prefix != "" {
		prefix = prefix + "/"
	}

	var wg sync.WaitGroup
	for k := range d.Removed {
		k = prefix + k
		wg.Add(1)
		go func(k string) {
			if _, err := kv.DeleteTree(k, nil); err != nil {
				panic(err)
			}
			fmt.Printf("Key %s deleted\n", color.YellowString(k))
			wg.Done()
		}(k)
	}
	wg.Wait()

	for k, v := range d.Added {
		k = prefix + k
		wg.Add(1)
		go func(k string, v string) {
			if _, err := kv.Put(&api.KVPair{Key: k, Value: []byte(v), Flags: 42}, nil); err != nil {
				panic(err)
			}
			fmt.Printf("Key %s added\n", color.YellowString(k))
			wg.Done()
		}(k, v)
	}
	wg.Wait()
}
