package kv

type Diff struct {
	added, removed, modified map[string]interface{}
}

func CalculateDiff(path string, oldMap map[string]interface{}, newMap map[string]interface{}) {
	// diff := Diff{}

	// find removed or modified fields
	// for key, newValue := range newMap {
	// 	oldValue, ok := oldMap[key]
	// 	if !ok {
	// 		diff.added[key] = newValue
	// 	}
	// }
}
