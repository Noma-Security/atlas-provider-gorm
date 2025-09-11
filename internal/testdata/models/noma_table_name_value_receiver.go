package models

import (
	"ariga.io/atlas-provider-gorm/gormschema"
)

type TestModelValueReceiver struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
	Age  int    `gorm:"column:age"`
}

func (model TestModelValueReceiver) TableName() string {
	return "test_model_value_receiver"
}

func (model *TestModelValueReceiver) Indexes() []gormschema.IndexDefinition[TestModelValueReceiver] {
	return []gormschema.IndexDefinition[TestModelValueReceiver]{
		{
			Name: "idx_test_model_unique",
			Columns: []gormschema.Col[TestModelValueReceiver]{
				{Sel: func(m *TestModelValueReceiver) any { return &m.Name }},
				{Sel: func(m *TestModelValueReceiver) any { return &m.Age }}},
			Unique: true,
		},
	}
}
