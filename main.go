package jsontools

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type stackElement struct {
	currentMap  map[string]interface{}
	currentPath string
}

func retrieveJSONTagsFromStruct[T any](fields []string) ([]string, error) {
	var zero T
	v := reflect.ValueOf(zero)

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type T must be a struct")
	}

	res := make([]string, 0)
	for _, fieldName := range fields {
		v := reflect.ValueOf(zero)
		t := v.Type()
		var keysToFilt []string
		fieldNames := strings.Split(fieldName, ".")
		something := true
		for len(fieldNames) > 0 && something {
			something = false
			for i := 0; i < v.NumField(); i++ {
				field := t.Field(i)
				if field.Name != fieldNames[0] {
					continue
				}

				something = true
				fieldNames = fieldNames[1:]

				jsonTag := field.Tag.Get("json")
				if jsonTag == "" || jsonTag == "-" {
					continue
				}

				tagParts := strings.Split(jsonTag, ",")
				keysToFilt = append(keysToFilt, tagParts[0])
				if len(fieldNames) > 0 && t.Kind() == reflect.Struct {
					v = v.Field(i)
					t = field.Type
					break
				}

				if len(fieldNames) == 0 {
					break
				}
			}
		}

		res = append(res, strings.Join(keysToFilt, "."))
	}

	return res, nil
}

func iterateNestedMap(root map[string]interface{}, onValueDetected func(path, key string, value interface{}) bool) bool {
	stack := []stackElement{{root, ""}}

	for len(stack) > 0 {
		// Pop the last element from the stack (LIFO)
		element := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Sort keys for consistent order (optional)
		var keys []string
		for key := range element.currentMap {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := element.currentMap[key]
			newPath := key
			if element.currentPath != "" {
				newPath = element.currentPath + "." + key
			}

			// Check if the value is a nested map
			if nestedMap, ok := value.(map[string]interface{}); ok {
				// Push nested map onto the stack
				stack = append(stack, stackElement{nestedMap, newPath})
			} else {
				// Process leaf node (e.g., print path and value)
				if onValueDetected != nil {
					if onValueDetected(element.currentPath, key, value) {
						return true
					}
				}
			}
		}
	}

	return false
}

func ParseFilters[T any](filters map[string]interface{}) (map[string]interface{}, error) {
	type filter struct {
		path     string
		jsonPath string
		value    interface{}
	}
	// Parse filters to []filter
	var flatFilters []filter
	iterateNestedMap(filters, func(path, key string, value interface{}) bool {
		pathKey := fmt.Sprintf("%s.%s", path, key)
		if path == "" {
			pathKey = key
		}
		tag, err := retrieveJSONTagsFromStruct[T]([]string{pathKey})
		if err != nil || len(tag) == 0 {
			return false
		}
		flatFilters = append(flatFilters, filter{
			path:     pathKey,
			jsonPath: tag[0],
			value:    value,
		})
		return false
	})

	// Parse JSON tags to map of path => value
	parsedFilters := make(map[string]interface{})
	for _, v := range flatFilters {
		parsedFilters[v.jsonPath] = v.value
	}

	return parsedFilters, nil
}

func ParseToStruct(data interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal data: %v", err)
	}

	return nil
}

func FilterBy[T any](data map[string]interface{}, filts map[string]interface{}) (map[string]T, error) {
	filters, err := ParseFilters[T](filts)
	if err != nil {
		return nil, err
	}

	filterData := make(map[string]interface{})
	for k, v := range data {
		entrie, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		match := true
		for kk, vv := range filters {
			subMatch := iterateNestedMap(entrie, func(path, key string, value interface{}) bool {
				pathKey := fmt.Sprintf("%s.%s", path, key)
				if path == "" {
					pathKey = key
				}
				if kk == pathKey {
					switch val := vv.(type) {
					case []any:
						for _, v := range val {
							if reflect.DeepEqual(v, value) {
								return true
							}
						}
					default:
						return reflect.DeepEqual(vv, value)
					}
				}
				return false
			})

			if !subMatch {
				match = false
				break
			}
		}

		if match {
			filterData[k] = v
		}
	}

	res := make(map[string]T)
	for k, v := range filterData {
		var dummy T
		if err := ParseToStruct(v, &dummy); err != nil {
			return nil, err
		}
		res[k] = dummy
	}

	return res, nil
}
