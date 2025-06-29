package entity

import (
	"cbs_backend/internal/common"
	"time"

	"github.com/google/uuid"
)

// SystemSetting represents tbl_system_settings table
type SystemSetting struct {
	SettingID          uuid.UUID    `json:"setting_id" db:"setting_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SettingKey         string       `json:"setting_key" db:"setting_key" gorm:"type:varchar(100);unique;not null"`
	SettingValue       common.JSONB `json:"setting_value" db:"setting_value" gorm:"type:jsonb;not null"`
	SettingDescription *string      `json:"setting_description,omitempty" db:"setting_description" gorm:"type:text"`
	IsPublic           bool         `json:"is_public" db:"is_public" gorm:"default:false"`
	SettingCreatedAt   time.Time    `json:"setting_created_at" db:"setting_created_at" gorm:"default:CURRENT_TIMESTAMP"`
	SettingUpdatedAt   time.Time    `json:"setting_updated_at" db:"setting_updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}
