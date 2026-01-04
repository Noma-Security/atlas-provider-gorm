package gormschema

import (
	"reflect"
	"strings"
	"testing"

	"ariga.io/atlas/sdk/recordriver"
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

	// FileName should have priority:3 and expression with gin_trgm_ops
	require.Contains(t, tags, "FileName")
	require.Len(t, tags["FileName"], 1)
	require.Contains(t, tags["FileName"][0], "priority:3")
	require.Contains(t, tags["FileName"][0], "expression:file_name gin_trgm_ops")
	require.NotContains(t, tags["FileName"][0], "type:")
}

func TestCollectIndexTagsWithMultipleOpClasses(t *testing.T) {
	model := TestMultiOpClass{}
	baseStruct := reflect.TypeOf(model)

	indexes := model.Indexes()
	defsSlice := reflect.ValueOf(indexes)

	tags, err := collectIndexTagsFromIndexesValue(baseStruct, defsSlice)
	require.NoError(t, err)

	// Field1: first column, has type:gist and expression with text_pattern_ops
	require.Contains(t, tags, "Field1")
	require.Contains(t, tags["Field1"][0], "type:gist")
	require.Contains(t, tags["Field1"][0], "expression:field1 text_pattern_ops")

	// Field2: no operator class
	require.Contains(t, tags, "Field2")
	require.NotContains(t, tags["Field2"][0], "expression:")

	// Field3: has expression with gin_trgm_ops
	require.Contains(t, tags, "Field3")
	require.Contains(t, tags["Field3"][0], "expression:field3 gin_trgm_ops")
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

func TestExtractRequiredExtensions(t *testing.T) {
	// TestNotebookFile doesn't have extensions
	exts := ExtractRequiredExtensions(TestNotebookFile{})
	require.Nil(t, exts)

	// Test with a model that has extensions
	exts = ExtractRequiredExtensions(TestWithExtensions{})
	require.Len(t, exts, 2)
	require.Contains(t, exts, "pg_trgm")
	require.Contains(t, exts, "btree_gin")
}

func TestExtractRequiredExtensionsNilModel(t *testing.T) {
	exts := ExtractRequiredExtensions(nil)
	require.Nil(t, exts)
}

func TestExtractRequiredExtensionsNoIndexesMethod(t *testing.T) {
	// A simple struct with no Indexes() method
	type NoIndexes struct {
		ID   uint
		Name string
	}
	exts := ExtractRequiredExtensions(NoIndexes{})
	require.Nil(t, exts)
}

func TestExtractRequiredExtensionsDeduplication(t *testing.T) {
	exts := ExtractRequiredExtensions(TestMultipleIndexesWithExtensions{})
	// Should deduplicate: both indexes require pg_trgm
	require.Len(t, exts, 2)
	require.Contains(t, exts, "pg_trgm")
	require.Contains(t, exts, "btree_gin")
}

// Test model with multiple indexes that have overlapping extensions
type TestMultipleIndexesWithExtensions struct {
	ID     uint `gorm:"primaryKey"`
	Field1 string
	Field2 string
}

func (TestMultipleIndexesWithExtensions) TableName() string {
	return "multi_idx_ext"
}

func (TestMultipleIndexesWithExtensions) Indexes() []IndexDefinition[TestMultipleIndexesWithExtensions] {
	return []IndexDefinition[TestMultipleIndexesWithExtensions]{
		{
			Name: "idx_1",
			Columns: []Col[TestMultipleIndexesWithExtensions]{
				Class(
					Field(func(m *TestMultipleIndexesWithExtensions) any { return &m.Field1 }),
					"gin_trgm_ops",
				),
			},
			Type:       "gin",
			Extensions: []string{"pg_trgm", "btree_gin"},
		},
		{
			Name: "idx_2",
			Columns: []Col[TestMultipleIndexesWithExtensions]{
				Class(
					Field(func(m *TestMultipleIndexesWithExtensions) any { return &m.Field2 }),
					"gin_trgm_ops",
				),
			},
			Type:       "gin",
			Extensions: []string{"pg_trgm"}, // duplicate, should be deduped
		},
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"FileName", "file_name"},
		{"TenantID", "tenant_id"},
		{"ID", "id"},
		{"Name", "name"},
		{"ScanVersion", "scan_version"},
		{"field1", "field1"},
		{"Field1", "field1"},
		{"HTTPServer", "http_server"},
		{"UserID", "user_id"},
		{"APIKey", "api_key"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Test model with extensions
type TestWithExtensions struct {
	ID       uint `gorm:"primaryKey"`
	TenantID string
	FileName string
}

func (TestWithExtensions) TableName() string {
	return "with_extensions"
}

func (TestWithExtensions) Indexes() []IndexDefinition[TestWithExtensions] {
	return []IndexDefinition[TestWithExtensions]{
		{
			Name: "idx_with_extensions_trgm",
			Columns: []Col[TestWithExtensions]{
				Field(func(m *TestWithExtensions) any { return &m.TenantID }),
				Class(
					Field(func(m *TestWithExtensions) any { return &m.FileName }),
					"gin_trgm_ops",
				),
			},
			Type:       "gin",
			Extensions: []string{"pg_trgm", "btree_gin"},
		},
	}
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

func TestGinIndexWithExtensionsSQLOutput(t *testing.T) {
	resetTestSession()

	l := New("postgres")
	sql, err := l.Load(TestGinModel{})
	require.NoError(t, err)

	// Should contain CREATE EXTENSION statements
	require.Contains(t, sql, `CREATE EXTENSION IF NOT EXISTS "pg_trgm"`)
	require.Contains(t, sql, `CREATE EXTENSION IF NOT EXISTS "btree_gin"`)

	// Should contain CREATE TABLE
	require.Contains(t, sql, `CREATE TABLE "gin_test"`)

	// Should contain CREATE INDEX with USING gin and the expression
	require.Contains(t, sql, "USING gin")
	require.Contains(t, sql, "file_name gin_trgm_ops")

	// Print the SQL for debugging
	t.Logf("Generated SQL:\n%s", sql)
}

// Test model for SQL output test
type TestGinModel struct {
	ID       uint   `gorm:"primaryKey"`
	TenantID string `gorm:"type:text"`
	FileName string `gorm:"type:text"`
}

func (TestGinModel) TableName() string {
	return "gin_test"
}

func (TestGinModel) Indexes() []IndexDefinition[TestGinModel] {
	return []IndexDefinition[TestGinModel]{
		{
			Name: "idx_gin_test_trgm",
			Columns: []Col[TestGinModel]{
				Field(func(m *TestGinModel) any { return &m.TenantID }),
				Class(
					Field(func(m *TestGinModel) any { return &m.FileName }),
					"gin_trgm_ops",
				),
			},
			Type:       "gin",
			Extensions: []string{"pg_trgm", "btree_gin"},
		},
	}
}

func resetTestSession() {
	sess, ok := recordriver.Session("gorm")
	if ok {
		sess.Statements = nil
	}
}
