package email

import (
	"bytes"
	"fmt"
	"html/template"

	"cbs_backend/internal/modules/notification_template/entity"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TemplateManager struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewTemplateManager(db *gorm.DB, logger *zap.Logger) *TemplateManager {
	return &TemplateManager{
		db:     db,
		logger: logger,
	}
}

func (tm *TemplateManager) GetTemplate(templateName string) (*entity.NotificationTemplate, error) {
	var notifTemplate entity.NotificationTemplate
	err := tm.db.Where("template_name = ? AND is_active = ?", templateName, true).
		First(&notifTemplate).Error
	if err != nil {
		return nil, fmt.Errorf("template not found: %s, error: %v", templateName, err)
	}
	return &notifTemplate, nil
}

// Render nội dung từ template + data
func (tm *TemplateManager) RenderTemplate(notifTemplate *entity.NotificationTemplate, data map[string]interface{}) (string, string, error) {
	// Title
	titleTemplate, err := template.New("title").Parse(notifTemplate.TitleTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse title template: %v", err)
	}

	var titleBuf bytes.Buffer
	if err := titleTemplate.Execute(&titleBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute title template: %v, data: %+v", err, data)
	}

	// Body
	bodyTemplate, err := template.New("body").Parse(notifTemplate.MessageTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse message template: %v", err)
	}

	var bodyBuf bytes.Buffer
	if err := bodyTemplate.Execute(&bodyBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute message template: %v, data: %+v", err, data)
	}

	return titleBuf.String(), bodyBuf.String(), nil
}
