package models

import (
	"ariga.io/atlas-provider-gorm/gormschema"
)

// NotebookFile is a test model with GIN index and extensions
type NotebookFile struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    string `gorm:"type:text;not null"`
	ScanVersion int    `gorm:"not null"`
	FileName    string `gorm:"type:text;not null"`
}

func (NotebookFile) TableName() string {
	return "notebook_files"
}

func (NotebookFile) Indexes() []gormschema.IndexDefinition[NotebookFile] {
	return []gormschema.IndexDefinition[NotebookFile]{
		{
			Name: "idx_notebook_files_tenant_scan_filename_trgm",
			Columns: []gormschema.Col[NotebookFile]{
				gormschema.Field(func(m *NotebookFile) any { return &m.TenantID }),
				gormschema.Field(func(m *NotebookFile) any { return &m.ScanVersion }),
				gormschema.Class(
					gormschema.Field(func(m *NotebookFile) any { return &m.FileName }),
					"gin_trgm_ops",
				),
			},
			Type:       "gin",
			Extensions: []string{"pg_trgm", "btree_gin"},
		},
	}
}
