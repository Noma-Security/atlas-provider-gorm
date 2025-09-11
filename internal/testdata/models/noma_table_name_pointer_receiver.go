package models

import (
	"ariga.io/atlas-provider-gorm/gormschema"
)

type TestModelTableNamePointerReceiver struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
	Age  int    `gorm:"column:age"`
}

func (model *TestModelTableNamePointerReceiver) TableName() string {
	return "test_model_table_name_pointer_receiver"
}

func (model *TestModelTableNamePointerReceiver) Indexes() []gormschema.IndexDefinition[TestModelTableNamePointerReceiver] {
	return []gormschema.IndexDefinition[TestModelTableNamePointerReceiver]{
		{
			Name: "idx_test_model_unique",
			Columns: []gormschema.Col[TestModelTableNamePointerReceiver]{
				{Sel: func(m *TestModelTableNamePointerReceiver) any { return &m.Name }},
				{Sel: func(m *TestModelTableNamePointerReceiver) any { return &m.Age }}},
			Unique: true,
		},
	}
}
