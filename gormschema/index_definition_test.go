package gormschema

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test model for index definition tests
type TestNotebookFile struct {
	ID          uint `gorm:"primaryKey"`
	TenantID    string
	ScanVersion int
	FileName    string
}

func (TestNotebookFile) TableName() string {
	return "notebook_files"
}

func (TestNotebookFile) Indexes() []IndexDefinition[TestNotebookFile] {
	return []IndexDefinition[TestNotebookFile]{
		{
			Name: "idx_notebook_files_tenant_scan_filename_trgm",
			Columns: []Col[TestNotebookFile]{
				Field(func(m *TestNotebookFile) any { return &m.TenantID }),
				Field(func(m *TestNotebookFile) any { return &m.ScanVersion }),
				Class(
					Field(func(m *TestNotebookFile) any { return &m.FileName }),
					"gin_trgm_ops",
				),
			},
			Type: "gin",
		},
	}
}

// Test model with simple index type (no operator class)
type TestSimpleGinIndex struct {
	ID   uint `gorm:"primaryKey"`
	Data string
}

func (TestSimpleGinIndex) TableName() string {
	return "simple_gin"
}

func (TestSimpleGinIndex) Indexes() []IndexDefinition[TestSimpleGinIndex] {
	return []IndexDefinition[TestSimpleGinIndex]{
		{
			Name: "idx_simple_gin_data",
			Columns: []Col[TestSimpleGinIndex]{
				Field(func(m *TestSimpleGinIndex) any { return &m.Data }),
			},
			Type: "gin",
		},
	}
}

// Test model with multiple operator classes
type TestMultiOpClass struct {
	ID     uint `gorm:"primaryKey"`
	Field1 string
	Field2 string
	Field3 string
}

func (TestMultiOpClass) TableName() string {
	return "multi_opclass"
}

func (TestMultiOpClass) Indexes() []IndexDefinition[TestMultiOpClass] {
	return []IndexDefinition[TestMultiOpClass]{
		{
			Name: "idx_multi_opclass",
			Columns: []Col[TestMultiOpClass]{
				Class(
					Field(func(m *TestMultiOpClass) any { return &m.Field1 }),
					"text_pattern_ops",
				),
				Field(func(m *TestMultiOpClass) any { return &m.Field2 }), // no op class
				Class(
					Field(func(m *TestMultiOpClass) any { return &m.Field3 }),
					"gin_trgm_ops",
				),
			},
			Type: "gist",
		},
	}
}

func TestCollectIndexTagsWithType(t *testing.T) {
	model := TestSimpleGinIndex{}
	baseStruct := reflect.TypeOf(model)

	// Get the Indexes() return value
	indexes := model.Indexes()
	defsSlice := reflect.ValueOf(indexes)

	tags, err := collectIndexTagsFromIndexesValue(baseStruct, defsSlice)
	require.NoError(t, err)

	// Check that Data field has the correct index tag with type:gin
	require.Contains(t, tags, "Data")
	require.Len(t, tags["Data"], 1)
	require.Contains(t, tags["Data"][0], "index:idx_simple_gin_data")
	require.Contains(t, tags["Data"][0], "type:gin")
}

func TestCollectIndexTagsWithOpClass(t *testing.T) {
	model := TestNotebookFile{}
	baseStruct := reflect.TypeOf(model)

	indexes := model.Indexes()
	defsSlice := reflect.ValueOf(indexes)

	tags, err := collectIndexTagsFromIndexesValue(baseStruct, defsSlice)
	require.NoError(t, err)

	// TenantID should have index tag with type:gin (first column gets type)
	require.Contains(t, tags, "TenantID")
	require.Len(t, tags["TenantID"], 1)
	require.Contains(t, tags["TenantID"][0], "index:idx_notebook_files_tenant_scan_filename_trgm")
	require.Contains(t, tags["TenantID"][0], "priority:1")
	require.Contains(t, tags["TenantID"][0], "type:gin")
	require.NotContains(t, tags["TenantID"][0], "class:")

	// ScanVersion should have priority:2, no type (not first), no class
	require.Contains(t, tags, "ScanVersion")
	require.Len(t, tags["ScanVersion"], 1)
	require.Contains(t, tags["ScanVersion"][0], "priority:2")
	require.NotContains(t, tags["ScanVersion"][0], "type:")
	require.NotContains(t, tags["ScanVersion"][0], "class:")

	// FileName should have priority:3 and class:gin_trgm_ops
	require.Contains(t, tags, "FileName")
	require.Len(t, tags["FileName"], 1)
	require.Contains(t, tags["FileName"][0], "priority:3")
	require.Contains(t, tags["FileName"][0], "class:gin_trgm_ops")
	require.NotContains(t, tags["FileName"][0], "type:")
}

func TestCollectIndexTagsWithMultipleOpClasses(t *testing.T) {
	model := TestMultiOpClass{}
	baseStruct := reflect.TypeOf(model)

	indexes := model.Indexes()
	defsSlice := reflect.ValueOf(indexes)

	tags, err := collectIndexTagsFromIndexesValue(baseStruct, defsSlice)
	require.NoError(t, err)

	// Field1: first column, has type:gist and class:text_pattern_ops
	require.Contains(t, tags, "Field1")
	require.Contains(t, tags["Field1"][0], "type:gist")
	require.Contains(t, tags["Field1"][0], "class:text_pattern_ops")

	// Field2: no class
	require.Contains(t, tags, "Field2")
	require.NotContains(t, tags["Field2"][0], "class:")

	// Field3: has class:gin_trgm_ops
	require.Contains(t, tags, "Field3")
	require.Contains(t, tags["Field3"][0], "class:gin_trgm_ops")
}

// Test model without Type or OpClass (backward compatibility)
type TestBasicIndex struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Email     string
	DeletedAt *string
}

func (TestBasicIndex) TableName() string {
	return "basic_index"
}

func (TestBasicIndex) Indexes() []IndexDefinition[TestBasicIndex] {
	return []IndexDefinition[TestBasicIndex]{
		{
			Name: "idx_basic_name_email",
			Columns: []Col[TestBasicIndex]{
				Field(func(m *TestBasicIndex) any { return &m.Name }),
				Field(func(m *TestBasicIndex) any { return &m.Email }),
			},
		},
		{
			Name:   "idx_basic_unique_email",
			Unique: true,
			Columns: []Col[TestBasicIndex]{
				Field(func(m *TestBasicIndex) any { return &m.Email }),
			},
		},
		{
			Name:  "idx_basic_partial",
			Where: "deleted_at IS NULL",
			Columns: []Col[TestBasicIndex]{
				Desc(Field(func(m *TestBasicIndex) any { return &m.Name })),
			},
		},
	}
}

func TestCollectIndexTagsWithoutTypeOrClass(t *testing.T) {
	model := TestBasicIndex{}
	baseStruct := reflect.TypeOf(model)

	indexes := model.Indexes()
	defsSlice := reflect.ValueOf(indexes)

	tags, err := collectIndexTagsFromIndexesValue(baseStruct, defsSlice)
	require.NoError(t, err)

	// Basic composite index - no type, no class
	require.Contains(t, tags, "Name")
	nameTag := findTagContaining(tags["Name"], "idx_basic_name_email")
	require.NotEmpty(t, nameTag)
	require.Contains(t, nameTag, "priority:1")
	require.NotContains(t, nameTag, "type:")
	require.NotContains(t, nameTag, "class:")

	require.Contains(t, tags, "Email")
	emailTag := findTagContaining(tags["Email"], "idx_basic_name_email")
	require.NotEmpty(t, emailTag)
	require.Contains(t, emailTag, "priority:2")

	// Unique index
	uniqueEmailTag := findTagContaining(tags["Email"], "idx_basic_unique_email")
	require.NotEmpty(t, uniqueEmailTag)
	require.Contains(t, uniqueEmailTag, "unique")

	// Partial index with sort
	partialTag := findTagContaining(tags["Name"], "idx_basic_partial")
	require.NotEmpty(t, partialTag)
	require.Contains(t, partialTag, "where:deleted_at IS NULL")
	require.Contains(t, partialTag, "sort:desc")
}

func findTagContaining(tags []string, substr string) string {
	for _, tag := range tags {
		if strings.Contains(tag, substr) {
			return tag
		}
	}
	return ""
}

func TestClassHelper(t *testing.T) {
	col := Field(func(m *TestNotebookFile) any { return &m.FileName })
	require.Equal(t, "", col.OpClass)

	col = Class(col, "gin_trgm_ops")
	require.Equal(t, "gin_trgm_ops", col.OpClass)
}

func TestColChaining(t *testing.T) {
	// Test that Class can be chained with other modifiers
	col := Class(
		Desc(Field(func(m *TestNotebookFile) any { return &m.FileName })),
		"gin_trgm_ops",
	)
	require.Equal(t, "desc", col.Sort)
	require.Equal(t, "gin_trgm_ops", col.OpClass)
}
