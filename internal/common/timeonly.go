package common

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type TimeOnly time.Time

const layout = "15:04:05" // Chỉ giờ, không ngày

func (t TimeOnly) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format(layout))), nil
}

func (t *TimeOnly) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	parsed, err := time.Parse(layout, str)
	if err != nil {
		return err
	}
	*t = TimeOnly(parsed)
	return nil
}

func (t TimeOnly) Value() (driver.Value, error) {
	return time.Time(t).Format(layout), nil
}

func (t *TimeOnly) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*t = TimeOnly(v)
		return nil
	case string:
		parsed, err := time.Parse(layout, v)
		if err != nil {
			return err
		}
		*t = TimeOnly(parsed)
		return nil
	default:
		return fmt.Errorf("unsupported scan type for TimeOnly: %T", value)
	}
}
