package entity

import (
	"time"

	"cbs_backend/internal/common"

	"github.com/google/uuid"
)

// NotificationTemplate represents tbl_notification_templates table
type NotificationTemplate struct {
	TemplateID        uuid.UUID    `json:"template_id" db:"template_id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TemplateName      string       `json:"template_name" db:"template_name" gorm:"type:varchar(100);unique;not null"`
	NotificationType  string       `json:"notification_type" db:"notification_type" gorm:"type:varchar(50);not null"`
	TitleTemplate     string       `json:"title_template" db:"title_template" gorm:"type:text;not null"`
	MessageTemplate   string       `json:"message_template" db:"message_template" gorm:"type:text;not null"`
	TemplateVariables common.JSONB `json:"template_variables,omitempty" db:"template_variables" gorm:"type:jsonb"`
	IsActive          bool         `json:"is_active" db:"is_active" gorm:"default:true"`
	TemplateCreatedAt time.Time    `json:"template_created_at" db:"template_created_at" gorm:"default:CURRENT_TIMESTAMP"`
	TemplateUpdatedAt time.Time    `json:"template_updated_at" db:"template_updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (NotificationTemplate) TableName() string {
	return "tbl_notification_templates"
}
