package common

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB type for handling PostgreSQL JSONB fields
type JSONB map[string]interface{}

// Value returns the JSON-encoded value to store in DB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan reads a JSONB value from the DB and decodes into JSONB map
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for JSONB")
	}

	var m map[string]interface{}
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}

	*j = m
	return nil
}
