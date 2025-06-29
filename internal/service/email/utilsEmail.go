package email

import "fmt"

func FormatAmount(amount interface{}) string {
	switch v := amount.(type) {
	case int64:
		return fmt.Sprintf("$%.2f", float64(v)/100)
	case float64:
		return fmt.Sprintf("$%.2f", v)
	case string:
		return v
	default:
		return "N/A"
	}
}
