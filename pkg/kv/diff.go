package kv

import (
	"github.com/fatih/color"
)

type Map map[string]interface{}

type Diff struct {
	added, removed Map
}

func CalculateDiff(oldMap Map, newMap Map) Diff {
	return calculateDiffHelper("", oldMap, newMap)
}

func (m Map) String() string {
	return "kek"
}

func (d Diff) Print() {
	for k, v := range d.removed {
		color.Red("%s: %v\n", k, v)
	}
	for k, v := range d.added {
		color.Green("%s: %v\n", k, v)
	}
}

func calculateDiffHelper(path string, oldMap Map, newMap Map) Diff {
	diff := Diff{
		make(Map),
		make(Map),
	}

	// find added or modified fields
	for key, newValue := range newMap {
		var keyPath string
		if path == "" {
			keyPath = key
		} else {
			keyPath = path + "/" + key
		}

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
				}
			} else { // old value is string but new value is map
				diff.removed[keyPath] = oldString
				diff.added[keyPath] = newValue
			}
		} else {
			if newString, ok := newValue.(string); ok { // old value is map, but new value is string
				diff.removed[keyPath] = oldValue
				diff.added[keyPath] = newString
			} else { // old value and new values are both maps, compare them recursive
				subDiff := calculateDiffHelper(
					keyPath,
					oldValue.(map[string]interface{}),
					newValue.(map[string]interface{}),
				)
				for sk, sv := range subDiff.added {
					diff.added[sk] = sv
				}
				for sk, sv := range subDiff.removed {
					diff.removed[sk] = sv
				}
			}
		}
	}

	// find removed fields
	for key, oldValue := range oldMap {
		if _, ok := newMap[key]; !ok {
			diff.removed[key] = oldValue
		}
	}

	return diff
}
