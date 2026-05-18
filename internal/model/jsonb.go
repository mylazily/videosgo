package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// JSONB 自定义 JSONB 类型，实现 driver.Valuer 和 sql.Scanner 接口
type JSONB json.RawMessage

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("JSONB: cannot scan %T into JSONB", value)
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return fmt.Errorf("JSONB: unmarshal error: %w", err)
	}

	*j = JSONB(result)
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*j = nil
		return nil
	}
	var result json.RawMessage
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("JSONB: unmarshal error: %w", err)
	}
	*j = JSONB(result)
	return nil
}

// String 返回 JSON 字符串
func (j JSONB) String() string {
	return string(j)
}

// ToMap 将 JSONB 解析为 map[string]interface{}
func (j JSONB) ToMap() (map[string]interface{}, error) {
	if len(j) == 0 {
		return nil, nil
	}
	var result map[string]interface{}
	err := json.Unmarshal(j, &result)
	return result, err
}

// ToSlice 将 JSONB 解析为 []interface{}
func (j JSONB) ToSlice() ([]interface{}, error) {
	if len(j) == 0 {
		return nil, nil
	}
	var result []interface{}
	err := json.Unmarshal(j, &result)
	return result, err
}

// ErrNotFound 未找到错误
var ErrNotFound = errors.New("record not found")
