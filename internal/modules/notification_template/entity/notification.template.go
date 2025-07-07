// package entity

// import (
// 	"cbs_backend/internal/common"
// 	"time"

// 	"github.com/google/uuid"
// )

// // NotificationTemplate represents tbl_notification_templates table
// type NotificationTemplate struct {
// 	TemplateID        uuid.UUID    `json:"template_id" gorm:"column:template_id;type:uuid;primaryKey;default:uuid_generate_v4()"`
// 	TemplateName      string       `json:"template_name" gorm:"column:template_name;type:varchar(100);unique;not null"`
// 	NotificationType  string       `json:"notification_type" gorm:"column:notification_type;type:varchar(50);not null"`
// 	TitleTemplate     string       `json:"title_template" gorm:"column:title_template;type:text;not null"`
// 	MessageTemplate   string       `json:"message_template" gorm:"column:message_template;type:text;not null"`
// 	TemplateVariables common.JSONB `json:"template_variables,omitempty" gorm:"column:template_variables;type:jsonb"`
// 	// TemplateVariables json.RawMessage `json:"template_variables,omitempty" db:"template_variables"`
// 	IsActive          bool      `json:"is_active" gorm:"column:is_active;default:true"`
// 	TemplateCreatedAt time.Time `json:"template_created_at" gorm:"column:template_created_at;autoCreateTime"`
// 	TemplateUpdatedAt time.Time `json:"template_updated_at" gorm:"column:template_updated_at;autoUpdateTime"`
// }

//	func (NotificationTemplate) TableName() string {
//		return "tbl_notification_templates"
//	}
package entity

import (
	"cbs_backend/internal/common"
	"time"

	"github.com/google/uuid"
)

// NotificationTemplate represents tbl_notification_templates table
type NotificationTemplate struct {
	TemplateID       uuid.UUID `json:"template_id" gorm:"column:template_id;type:uuid;primaryKey;default:uuid_generate_v4()"`
	TemplateName     string    `json:"template_name" gorm:"column:template_name;type:varchar(100);unique;not null"`
	NotificationType string    `json:"notification_type" gorm:"column:notification_type;type:varchar(50);not null"`
	TitleTemplate    string    `json:"title_template" gorm:"column:title_template;type:text;not null"`
	MessageTemplate  string    `json:"message_template" gorm:"column:message_template;type:text;not null"`
	// TemplateVariables []byte    `json:"template_variables,omitempty" gorm:"column:template_variables;type:jsonb"`
	TemplateVariables common.JSONB `json:"template_variables,omitempty" gorm:"column:template_variables;type:jsonb"`
	IsActive          bool         `json:"is_active" gorm:"column:is_active;default:true"`
	TemplateCreatedAt time.Time    `json:"template_created_at" gorm:"column:template_created_at;autoCreateTime"`
	TemplateUpdatedAt time.Time    `json:"template_updated_at" gorm:"column:template_updated_at;autoUpdateTime"`
}

func (NotificationTemplate) TableName() string {
	return "tbl_notification_templates"
}
