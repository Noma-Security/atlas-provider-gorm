package gormschema_test

import (
	"os"
	"testing"

	"ariga.io/atlas-provider-gorm/gormschema"
	ckmodels "ariga.io/atlas-provider-gorm/internal/testdata/circularfks"
	"ariga.io/atlas-provider-gorm/internal/testdata/customjointable"
	"ariga.io/atlas-provider-gorm/internal/testdata/models"
	"ariga.io/atlas/sdk/recordriver"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

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

func TestPostgreSQLGinIndexWithExtensions(t *testing.T) {
	resetSession()
	l := gormschema.New("postgres")
	sql, err := l.Load(models.NotebookFile{})
	require.NoError(t, err)

	// Log the actual SQL for debugging
	t.Logf("Generated SQL:\n%s", sql)

	// Verify extensions are created
	require.Contains(t, sql, `CREATE EXTENSION IF NOT EXISTS "pg_trgm"`)
	require.Contains(t, sql, `CREATE EXTENSION IF NOT EXISTS "btree_gin"`)

	// Verify the table and index are created
	require.Contains(t, sql, `CREATE TABLE "notebook_files"`)
	require.Contains(t, sql, "USING gin")
	require.Contains(t, sql, "gin_trgm_ops")

	// Compare against expected output
	requireEqualContent(t, sql, "testdata/postgresql_gin_extensions.sql")
	resetSession()
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
