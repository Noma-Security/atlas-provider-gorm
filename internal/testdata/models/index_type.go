package models

import (
	"time"

	"ariga.io/atlas-provider-gorm/gormschema"
)

type TestModelIndexType struct {
	ID      string `gorm:"column:id"`
	Name    string `gorm:"column:name"`
	Profile string `gorm:"column:profile"`
}

func (model TestModelIndexType) TableName() string {
	return "test_model_index_type"
}

func (model *TestModelIndexType) Indexes() []gormschema.IndexDefinition[TestModelIndexType] {
	return []gormschema.IndexDefinition[TestModelIndexType]{
		{
			Name: "idx_test_model_index_type_name_gin",
			Type: "gin",
			Columns: []gormschema.Col[TestModelIndexType]{
				{Sel: func(m *TestModelIndexType) any { return &m.Name }},
			},
		},
		{
			Name: "idx_test_model_index_type_name_profile_gin",
			Type: "gin",
			Columns: []gormschema.Col[TestModelIndexType]{
				{Sel: func(m *TestModelIndexType) any { return &m.Name }},
				{Sel: func(m *TestModelIndexType) any { return &m.Profile }},
			},
		},
	}
}

type TestModelIndexOpClass struct {
	ID        string     `gorm:"column:id"`
	TenantID  string     `gorm:"column:tenant_id"`
	UserID    *string    `gorm:"column:user_id"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
	SessionID string     `gorm:"column:session_id"`
}

func (model TestModelIndexOpClass) TableName() string {
	return "test_model_index_op_class"
}

func (model *TestModelIndexOpClass) Indexes() []gormschema.IndexDefinition[TestModelIndexOpClass] {
	return []gormschema.IndexDefinition[TestModelIndexOpClass]{
		{
			Name: "idx_test_model_index_op_class_user_id",
			Columns: []gormschema.Col[TestModelIndexOpClass]{
				{Sel: func(m *TestModelIndexOpClass) any { return &m.TenantID }},
				gormschema.WithOpClass(gormschema.Field(func(m *TestModelIndexOpClass) any { return &m.UserID }), "text_pattern_ops"),
				{Sel: func(m *TestModelIndexOpClass) any { return &m.UpdatedAt }, Sort: "desc"},
				{Sel: func(m *TestModelIndexOpClass) any { return &m.SessionID }, Sort: "desc"},
			},
		},
	}
}
