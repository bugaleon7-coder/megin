package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap stores arbitrary JSON objects in text/json database columns.
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = make(JSONMap)
		return nil
	}

	switch value := value.(type) {
	case []byte:
		return json.Unmarshal(value, m)
	case string:
		return json.Unmarshal([]byte(value), m)
	default:
		return errors.New("model.JSONMap.Scan: invalid value type")
	}
}
