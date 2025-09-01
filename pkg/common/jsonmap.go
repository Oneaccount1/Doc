package common

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// 存放 domain层的值对象
// 略显不规范，但代价可接受

type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil || len(m) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (m JSONMap) Scan(src interface{}) error {
	if src == nil {
		m = JSONMap{}
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported scan type: %T", src)
	}
	if len(data) == 0 {
		m = JSONMap{}
		return nil
	}
	var tmp map[string]interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("json unmarshal failed: %w", err)
	}
	m = tmp
	return nil
}

// IsEmpty 检查JSONMap是否为空
func (m JSONMap) IsEmpty() bool {
	return len(m) == 0
}

// GetString 安全地获取字符串值
func (m JSONMap) GetString(key string) (string, bool) {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// ToMap 转换为普通map（用于模板渲染）
func (m JSONMap) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// NewJSONMap 创建新的JSONMap
func NewJSONMap(data map[string]interface{}) JSONMap {
	if data == nil {
		return make(JSONMap)
	}
	return data
}
