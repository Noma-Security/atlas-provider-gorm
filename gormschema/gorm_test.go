package gormschema_test

import (
	"os"
	"testing"
	"time"

	"ariga.io/atlas-provider-gorm/gormschema"
	ckmodels "ariga.io/atlas-provider-gorm/internal/testdata/circularfks"
	"ariga.io/atlas-provider-gorm/internal/testdata/customjointable"
	"ariga.io/atlas-provider-gorm/internal/testdata/models"
	"ariga.io/atlas/sdk/recordriver"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gschema "gorm.io/gorm/schema"
)

type testModelNamingStrategyOpClass struct {
	ID        string
	TenantID  string
	UserID    *string
	UpdatedAt *time.Time
	SessionID string
}

func (testModelNamingStrategyOpClass) TableName() string {
	return "test_model_naming_strategy_op_class"
}

func (*testModelNamingStrategyOpClass) Indexes() []gormschema.IndexDefinition[testModelNamingStrategyOpClass] {
	return []gormschema.IndexDefinition[testModelNamingStrategyOpClass]{
		{
			Name: "idx_test_model_naming_strategy_op_class_user_id",
			Columns: []gormschema.Col[testModelNamingStrategyOpClass]{
				{Sel: func(m *testModelNamingStrategyOpClass) any { return &m.TenantID }},
				gormschema.WithOpClass(gormschema.Field(func(m *testModelNamingStrategyOpClass) any { return &m.UserID }), "text_pattern_ops"),
				{Sel: func(m *testModelNamingStrategyOpClass) any { return &m.UpdatedAt }, Sort: "desc"},
				{Sel: func(m *testModelNamingStrategyOpClass) any { return &m.SessionID }, Sort: "desc"},
			},
		},
	}
}

type testModelInvalidIndexType struct {
	ID   string
	Name string
}

func (testModelInvalidIndexType) TableName() string {
	return "test_model_invalid_index_type"
}

func (*testModelInvalidIndexType) Indexes() []gormschema.IndexDefinition[testModelInvalidIndexType] {
	return []gormschema.IndexDefinition[testModelInvalidIndexType]{
		{
			Name: "idx_test_model_invalid_index_type_name",
			Type: "gin()",
			Columns: []gormschema.Col[testModelInvalidIndexType]{
				gormschema.Field(func(m *testModelInvalidIndexType) any { return &m.Name }),
			},
		},
	}
}

func TestSQLiteConfig(t *testing.T) {
	resetSession()
	l := gormschema.New("sqlite")
	sql, err := l.Load(
		models.WorkingAgedUsers{},
		models.Pet{},
		models.UserPetHistory{},
		ckmodels.Event{},
		ckmodels.Location{},
		models.TopPetOwner{},
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlite_default.sql")
	resetSession()
	l = gormschema.New("sqlite", gormschema.WithConfig(&gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}))
	sql, err = l.Load(models.UserPetHistory{}, models.Pet{}, models.User{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlite_no_fk.sql")
	resetSession()
}

func TestAutoMigrateModelTableName(t *testing.T) {
	resetSession()

	modelTableNameMapping := map[any]string{
		models.TestModelValueReceiver{}:            models.TestModelValueReceiver{}.TableName(),
		models.TestModelTableNamePointerReceiver{}: (&models.TestModelTableNamePointerReceiver{}).TableName(),
	}
	l := gormschema.New("postgres")
	for model, tableName := range modelTableNameMapping {
		t.Run(tableName, func(t *testing.T) {
			sql, err := l.Load(
				model,
			)
			require.NoError(t, err)
			require.Contains(t, sql, tableName)
		})
	}
}

func TestAutoMigrateModelIndexType(t *testing.T) {
	resetSession()

	l := gormschema.New("postgres")
	sql, err := l.Load(models.TestModelIndexType{})
	require.NoError(t, err)
	require.Contains(t, sql, `CREATE INDEX IF NOT EXISTS "idx_test_model_index_type_name_gin" ON "test_model_index_type" USING gin("name");`)
	require.Contains(t, sql, `CREATE INDEX IF NOT EXISTS "idx_test_model_index_type_name_profile_gin" ON "test_model_index_type" USING gin("name","profile");`)
}

func TestAutoMigrateModelIndexTypeEmptyUsesDatabaseDefault(t *testing.T) {
	resetSession()

	l := gormschema.New("postgres")
	sql, err := l.Load(models.TestModelValueReceiver{})
	require.NoError(t, err)
	require.Contains(t, sql, `CREATE UNIQUE INDEX IF NOT EXISTS "idx_test_model_unique" ON "test_model_value_receiver" ("name","age");`)
	require.NotContains(t, sql, `USING btree`)
}

func TestAutoMigrateModelIndexOpClass(t *testing.T) {
	resetSession()

	l := gormschema.New("postgres")
	sql, err := l.Load(models.TestModelIndexOpClass{})
	require.NoError(t, err)
	require.Contains(t, sql, `CREATE INDEX IF NOT EXISTS "idx_test_model_index_op_class_user_id" ON "test_model_index_op_class" ("tenant_id","user_id" text_pattern_ops,"updated_at" desc,"session_id" desc);`)
}

func TestAutoMigrateModelIndexOpClassHonorsNamingStrategy(t *testing.T) {
	resetSession()

	l := gormschema.New("postgres", gormschema.WithConfig(&gorm.Config{
		NamingStrategy: gschema.NamingStrategy{NoLowerCase: true},
	}))
	sql, err := l.Load(testModelNamingStrategyOpClass{})
	require.NoError(t, err)
	require.Contains(t, sql, `CREATE INDEX IF NOT EXISTS "idx_test_model_naming_strategy_op_class_user_id" ON "test_model_naming_strategy_op_class" ("TenantID","UserID" text_pattern_ops,"UpdatedAt" desc,"SessionID" desc);`)
}

func TestAutoMigrateModelInvalidIndexType(t *testing.T) {
	resetSession()

	_, err := gormschema.New("postgres").Load(testModelInvalidIndexType{})
	require.Error(t, err)
	require.ErrorContains(t, err, `invalid Type "gin()"`)
}

func TestPostgreSQLConfig(t *testing.T) {
	resetSession()
	l := gormschema.New("postgres")
	sql, err := l.Load(
		models.WorkingAgedUsers{},
		ckmodels.Location{},
		ckmodels.Event{},
		models.UserPetHistory{},
		models.User{},
		models.Pet{},
		models.TopPetOwner{},
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/postgresql_default.sql")
	resetSession()
	l = gormschema.New("postgres", gormschema.WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/postgresql_no_fk.sql")
}

func TestMySQLConfig(t *testing.T) {
	resetSession()
	l := gormschema.New("mysql")
	sql, err := l.Load(
		models.WorkingAgedUsers{},
		ckmodels.Location{},
		ckmodels.Event{},
		models.UserPetHistory{},
		models.User{},
		models.Pet{},
		models.TopPetOwner{},
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_default.sql")
	resetSession()
	l = gormschema.New("mysql", gormschema.WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_no_fk.sql")
	resetSession()
	l = gormschema.New("mysql",
		gormschema.WithModelPosition(map[any]string{
			&customjointable.Person{}:              "/internal/testdata/customjointable/models.go:11",
			&customjointable.Address{}:             "/internal/testdata/customjointable/models.go:17",
			&customjointable.PersonAddress{}:       "/internal/testdata/customjointable/models.go:22",
			&customjointable.TopCrowdedAddresses{}: "/internal/testdata/customjointable/models.go:29",
		}),
		gormschema.WithJoinTable(&customjointable.Person{}, "Addresses", &customjointable.PersonAddress{}),
	)
	sql, err = l.Load(customjointable.Address{}, customjointable.Person{}, customjointable.TopCrowdedAddresses{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_custom_join_table.sql")
	resetSession()
	l = gormschema.New("mysql", gormschema.WithModelPosition(map[any]string{
		&customjointable.Person{}:              "/internal/testdata/customjointable/models.go:11",
		&customjointable.Address{}:             "/internal/testdata/customjointable/models.go:17",
		&customjointable.PersonAddress{}:       "/internal/testdata/customjointable/models.go:22",
		&customjointable.TopCrowdedAddresses{}: "/internal/testdata/customjointable/models.go:29",
	}))
	sql, err = l.Load(customjointable.PersonAddress{}, customjointable.Address{}, customjointable.Person{}, customjointable.TopCrowdedAddresses{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_custom_join_table.sql")
	resetSession()
	l = gormschema.New("mysql", gormschema.WithModelPosition(map[any]string{
		&customjointable.Person{}:              "/internal/testdata/customjointable/models.go:11",
		&customjointable.Address{}:             "/internal/testdata/customjointable/models.go:17",
		&customjointable.PersonAddress{}:       "/internal/testdata/customjointable/models.go:22",
		&customjointable.TopCrowdedAddresses{}: "/internal/testdata/customjointable/models.go:29",
	}))
	sql, err = l.Load(customjointable.Address{}, customjointable.PersonAddress{}, customjointable.Person{}, customjointable.TopCrowdedAddresses{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_custom_join_table.sql") // position of tables should not matter
}

func TestSQLServerConfig(t *testing.T) {
	resetSession()
	l := gormschema.New("sqlserver", gormschema.WithStmtDelimiter("\nGO"))
	sql, err := l.Load(
		models.WorkingAgedUsers{},
		ckmodels.Location{},
		ckmodels.Event{},
		models.UserPetHistory{},
		models.User{},
		models.Pet{},
		models.TopPetOwner{},
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlserver_default.sql")
	resetSession()
	l = gormschema.New("sqlserver",
		gormschema.WithStmtDelimiter("\nGO"),
		gormschema.WithConfig(
			&gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: true,
			}))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlserver_no_fk.sql")
}

func resetSession() {
	sess, ok := recordriver.Session("gorm")
	if ok {
		sess.Statements = nil
	}
}

func requireEqualContent(t *testing.T, actual, fileName string) {
	buf, err := os.ReadFile(fileName)
	require.NoError(t, err)
	require.Equal(t, string(buf), actual)
}
