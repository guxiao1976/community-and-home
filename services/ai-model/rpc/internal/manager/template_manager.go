package manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"community-and-home/services/ai-model/rpc/model"
)

type TemplateManager struct {
	db            sqlx.SqlConn
	templateModel model.AmPromptTemplateModel
}

func NewTemplateManager(db sqlx.SqlConn, cacheConf cache.CacheConf) *TemplateManager {
	return &TemplateManager{
		db:            db,
		templateModel: model.NewAmPromptTemplateModel(db, cacheConf),
	}
}

func (m *TemplateManager) GetTemplate(ctx context.Context, templateID int64) (*model.AmPromptTemplate, error) {
	template, err := m.templateModel.FindOne(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("find template: %w", err)
	}

	if template.IsActive != 1 {
		return nil, fmt.Errorf("template is not active")
	}

	return template, nil
}

func (m *TemplateManager) GetTemplateByName(ctx context.Context, name string) (*model.AmPromptTemplate, error) {
	query := "SELECT * FROM am_prompt_template WHERE template_name = ? AND is_active = 1 LIMIT 1"

	var template model.AmPromptTemplate
	err := m.db.QueryRowCtx(ctx, &template, query, name)
	if err != nil {
		return nil, fmt.Errorf("find template by name: %w", err)
	}

	return &template, nil
}

func (m *TemplateManager) RenderTemplate(ctx context.Context, templateID int64, variables map[string]string) (string, error) {
	template, err := m.GetTemplate(ctx, templateID)
	if err != nil {
		return "", err
	}

	return m.render(template.Content, variables), nil
}

func (m *TemplateManager) RenderTemplateByName(ctx context.Context, name string, variables map[string]string) (string, error) {
	template, err := m.GetTemplateByName(ctx, name)
	if err != nil {
		return "", err
	}

	return m.render(template.Content, variables), nil
}

func (m *TemplateManager) render(content string, variables map[string]string) string {
	result := content
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func (m *TemplateManager) ListTemplates(ctx context.Context, category string) ([]*model.AmPromptTemplate, error) {
	query := "SELECT * FROM am_prompt_template WHERE is_active = 1"
	args := []interface{}{}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}

	query += " ORDER BY category, template_name"

	var templates []*model.AmPromptTemplate
	err := m.db.QueryRowsCtx(ctx, &templates, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}

	return templates, nil
}

func (m *TemplateManager) CreateTemplate(ctx context.Context, template *model.AmPromptTemplate) error {
	_, err := m.templateModel.Insert(ctx, template)
	if err != nil {
		return fmt.Errorf("insert template: %w", err)
	}

	return nil
}

func (m *TemplateManager) UpdateTemplate(ctx context.Context, template *model.AmPromptTemplate) error {
	err := m.templateModel.Update(ctx, template)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	return nil
}

func (m *TemplateManager) DeleteTemplate(ctx context.Context, templateID int64) error {
	query := "UPDATE am_prompt_template SET is_active = 0 WHERE id = ?"
	_, err := m.db.ExecCtx(ctx, query, templateID)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}

	return nil
}
