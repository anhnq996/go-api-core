package utils

import (
	"encoding/json"
	"fmt"
)

// ToJSON chuyển object sang JSON string
func ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ToJSONPretty chuyển object sang JSON string (pretty format)
func ToJSONPretty(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON parse JSON string sang object
func FromJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// MustToJSON chuyển sang JSON, panic nếu error
func MustToJSON(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

// IsJSON kiểm tra string có phải JSON hợp lệ không
func IsJSON(str string) bool {
	var js interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

// JSONMerge merge 2 JSON objects
func JSONMerge(json1, json2 string) (string, error) {
	var obj1, obj2 map[string]interface{}

	if err := FromJSON(json1, &obj1); err != nil {
		return "", err
	}

	if err := FromJSON(json2, &obj2); err != nil {
		return "", err
	}

	// Merge obj2 vào obj1
	for k, v := range obj2 {
		obj1[k] = v
	}

	return ToJSON(obj1)
}

// JSONExtract extract field từ JSON string
func JSONExtract(jsonStr, field string) (interface{}, error) {
	var obj map[string]interface{}

	if err := FromJSON(jsonStr, &obj); err != nil {
		return nil, err
	}

	value, ok := obj[field]
	if !ok {
		return nil, fmt.Errorf("field %s not found", field)
	}

	return value, nil
}

// CopyStruct copy struct sang struct khác qua JSON
func CopyStruct(src, dst interface{}) error {
	jsonStr, err := ToJSON(src)
	if err != nil {
		return err
	}

	return FromJSON(jsonStr, dst)
}
