package kv

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
)

type Diff struct {
	added, removed map[string]string
}

func CalculateDiff(oldMap, newMap map[string]interface{}) Diff {
	return calculateDiffHelper("", oldMap, newMap)
}

func (d1 *Diff) Append(d2 Diff) {
	for sk, sv := range d2.added {
		d1.added[sk] = sv
	}
	for sk, sv := range d2.removed {
		d1.removed[sk] = sv
	}
}

func (d Diff) Print() {
	marked := make(map[string]bool)

	red := func(k, v interface{}) {
		out, err := yaml.Marshal(v)
		if err != nil {
			panic(err)
		}
		color.Red("--- %s: %s", k, strings.TrimSpace(string(out)))
	}
	green := func(k, v interface{}) {
		out, err := yaml.Marshal(v)
		if err != nil {
			panic(err)
		}
		color.Green("+++ %s: %s", k, strings.TrimSpace(string(out)))
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

	// find added or modified fields
	for key, newValue := range newMap {
		keyPath := constructKeyPath(path, key)
		oldValue, ok := oldMap[key]

		if !ok {
			switch actualNewValue := newValue.(type) {
			case string:
				diff.added[keyPath] = actualNewValue
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
					diff.removed[keyPath] = actualOldValue
					diff.added[keyPath] = actualNewValue
				}
			case map[string]interface{}: // old value is string but new value is map
				diff.removed[keyPath] = actualOldValue
				diff.Append(calculateDiffHelper(
					keyPath,
					map[string]interface{}{},
					actualNewValue,
				))
			default:
				panic("trash")
			}
		case map[string]interface{}:
			switch actualNewValue := newValue.(type) {
			case string: // old value is map, but new value is string
				diff.Append(calculateDiffHelper(
					keyPath,
					actualOldValue,
					map[string]interface{}{},
				))
				diff.added[keyPath] = actualNewValue
			case map[string]interface{}: // old value and new value are maps
				diff.Append(calculateDiffHelper(
					keyPath,
					actualOldValue,
					actualNewValue,
				))
			default:
				panic("trash")
			}
		default:
			panic("")
		}
	}

	// find removed fields
	for key, oldValue := range oldMap {
		keyPath := constructKeyPath(path, key)
		if _, ok := newMap[key]; !ok {
			switch actualOldValue := oldValue.(type) {
			case string:
				diff.removed[keyPath] = actualOldValue
			case map[string]interface{}:
				diff.Append(calculateDiffHelper(
					keyPath,
					actualOldValue,
					map[string]interface{}{},
				))
			default:
				panic("")
			}
		}
	}

	return diff
}

func (d Diff) Apply() {
	errCh := make(chan error)
	deleteCh := make(chan interface{}, len(d.removed))
	addCh := make(chan interface{}, len(d.added))

	for k := range d.removed {
		go func(k string) {
			_, err := kv.DeleteTree(k, nil)
			if err != nil {
				errCh <- err
			}
			fmt.Printf("Key %s deleted\n", k)
			deleteCh <- nil
		}(k)
	}
	for i := 0; i < len(d.removed); i++ {
		select {
		case e := <-errCh:
			panic(e)
		case <-deleteCh:
			continue
		}
	}

	for k, v := range d.added {
		go func(k string, v string) {
			_, err := kv.Put(&api.KVPair{Key: k, Value: []byte(v), Flags: 42}, nil)
			if err != nil {
				errCh <- err
			}
			fmt.Printf("Key %s added\n", k)
			addCh <- nil
		}(k, v)
	}
	for i := 0; i < len(d.added); i++ {
		select {
		case e := <-errCh:
			panic(e)
		case <-addCh:
			continue
		}
	}
}
