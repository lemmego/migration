package migration

import (
	"os"
	"strings"
	"testing"
)

func TestSQLiteIncrements(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nid INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT);"

	schema := Create("users", func(t *Table) {
		t.Increments("id").Primary()
	}).Build()

	// Normalize both the expected and generated schema strings

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected: \n%s, \nGot: \n%s", expected, schema)
	}
}

func TestMySQLIncrements(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nid INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT);"

	schema := Create("users", func(t *Table) {
		t.Increments("id").Primary()
	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestPostgresIncrements(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\nid SERIAL NOT NULL PRIMARY KEY CHECK (id > 0));"

	schema := Create("users", func(t *Table) {
		t.Increments("id").Primary()
	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteBigIncrements(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nid INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT);"

	schema := Create("users", func(t *Table) {
		t.BigIncrements("id").Primary()
	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLBigIncrements(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nid BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT);"

	schema := Create("users", func(t *Table) {
		t.BigIncrements("id").Primary()
	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestPostgresBigIncrements(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\nid BIGSERIAL NOT NULL PRIMARY KEY CHECK (id > 0));"

	schema := Create("users", func(t *Table) {
		t.BigIncrements("id").Primary()
	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteBool(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nactive BOOLEAN);"

	schema := Create("users", func(t *Table) {
		t.Boolean("active").Nullable()
	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLBool(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nactive BOOLEAN);"

	schema := Create("users", func(t *Table) {
		t.Boolean("active").Nullable()
	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestPostgresBool(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\nactive BOOLEAN);"

	schema := Create("users", func(t *Table) {
		t.Boolean("active").Nullable()
	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteSmallInt(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nage SMALLINT);"

	schema := Create("users", func(t *Table) {
		t.SmallInt("age").Nullable()
	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLSmallInt(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nage SMALLINT);"

	schema := Create("users", func(t *Table) {
		t.SmallInt("age").Nullable()
	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteMediumInt(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nage MEDIUMINT);"

	schema := Create("users", func(t *Table) {
		t.MediumInt("age").Nullable()
	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLMediumInt(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nage MEDIUMINT);"

	schema := Create("users", func(t *Table) {
		t.MediumInt("age").Nullable()
	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteInt(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nage INTEGER);"

	schema := Create("users", func(t *Table) {
		t.Int("age").Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLInt(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nage INT);"

	schema := Create("users", func(t *Table) {
		t.Int("age").Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteBigInt(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nage BIGINT);"

	schema := Create("users", func(t *Table) {
		t.BigInt("age").Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLBigInt(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nage BIGINT);"

	schema := Create("users", func(t *Table) {
		t.BigInt("age").Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteFloat(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\namount FLOAT(8, 2));"

	schema := Create("users", func(t *Table) {
		t.Float("amount", 8, 2).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLFloat(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\namount FLOAT(8, 2));"

	schema := Create("users", func(t *Table) {
		t.Float("amount", 8, 2).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteDouble(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\namount DOUBLE(10,2));"

	schema := Create("users", func(t *Table) {
		t.Double("amount", 10, 2).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLDouble(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\namount DOUBLE(10,2));"

	schema := Create("users", func(t *Table) {
		t.Double("amount", 10, 2).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteDecimal(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\namount DECIMAL(10,2));"

	schema := Create("users", func(t *Table) {
		t.Decimal("amount", 10, 2).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestPostgresDecimal(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\namount DECIMAL(10,2));"

	schema := Create("users", func(t *Table) {
		t.Decimal("amount", 10, 2).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLDecimal(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\namount DECIMAL(10,2));"

	schema := Create("users", func(t *Table) {
		t.Decimal("amount", 10, 2).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteChar(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nname CHAR(100));"

	schema := Create("users", func(t *Table) {
		t.Char("name", 100).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLChar(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nname CHAR(100));"

	schema := Create("users", func(t *Table) {
		t.Char("name", 100).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestPostgresChar(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\nname CHAR(100));"

	schema := Create("users", func(t *Table) {
		t.Char("name", 100).Nullable()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteForeignKey(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nid INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,\nrole_id INTEGER NOT NULL,\nFOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE ON UPDATE CASCADE);"

	schema := Create("users", func(t *Table) {
		t.Increments("id").Primary()
		t.Int("role_id")
		t.Foreign("role_id").
			References("id").
			On("roles").
			OnDelete("CASCADE").
			OnUpdate("CASCADE")

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", normalizedExpected, normalizedSchema)
	}
}

func TestMySQLForeignKey(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nid INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,\nrole_id INT NOT NULL,\nFOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE ON UPDATE CASCADE);"

	schema := Create("users", func(t *Table) {
		t.Increments("id").Primary()
		t.Int("role_id")
		t.Foreign("role_id").
			References("id").
			On("roles").
			OnDelete("CASCADE").
			OnUpdate("CASCADE")

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestPostgresForeignKey(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\nid SERIAL NOT NULL PRIMARY KEY CHECK (id > 0),\nrole_id INTEGER NOT NULL,\nFOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE ON UPDATE CASCADE);"

	schema := Create("users", func(t *Table) {
		t.Increments("id").Primary()
		t.Int("role_id")
		t.Foreign("role_id").
			References("id").
			On("roles").
			OnDelete("CASCADE").
			OnUpdate("CASCADE")

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteConstrained(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\norg_id INTEGER NOT NULL,\nFOREIGN KEY (org_id) REFERENCES orgs(id));"

	schema := Create("users", func(t *Table) {
		t.ForeignID("org_id").Constrained()
	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteConstrainedFunc(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\norg_id INTEGER NOT NULL,\nCONSTRAINT f_orgs_id FOREIGN KEY (org_id) REFERENCES orgs(id));"

	schema := Create("users", func(t *Table) {
		t.ForeignID("org_id").ConstrainedFunc(func(t *Table) (table string, indexName string) {
			return "orgs", "f_orgs_id"
		})
	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteUnique(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL UNIQUE);"

	schema := Create("users", func(t *Table) {
		t.String("email", 100).Unique()

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestMySQLUnique(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL UNIQUE);"

	schema := Create("users", func(t *Table) {
		t.String("email", 100).Unique()

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestPostgresUnique(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL UNIQUE);"

	schema := Create("users", func(t *Table) {
		t.String("email", 100).Unique()

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteIndex(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	// SQLite has no inline INDEX clause; the index is its own statement.
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL);\nCREATE INDEX users_email_index ON users (email);"

	schema := Create("users", func(t *Table) {
		t.String("email", 100)
		t.Index("email")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestMySQLIndex(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	// MySQL does accept an inline INDEX clause, so it keeps one statement.
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL, INDEX users_email_index (email));"

	schema := Create("users", func(t *Table) {
		t.String("email", 100)
		t.Index("email")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestPostgresIndex(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	// SQLite has no inline INDEX clause; the index is its own statement.
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL);\nCREATE INDEX users_email_index ON users (email);"

	schema := Create("users", func(t *Table) {
		t.String("email", 100)
		t.Index("email")

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteUniqueIndex(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL,\nUNIQUE(email));"

	schema := Create("users", func(t *Table) {
		t.String("email", 100)
		t.UniqueKey("email")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestMySQLUniqueIndex(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL,\nUNIQUE email_unique (email));"

	schema := Create("users", func(t *Table) {
		t.String("email", 100)
		t.UniqueKey("email")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestPostgresUniqueIndex(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL,\n UNIQUE (email));"

	schema := Create("users", func(t *Table) {
		t.String("email", 100)
		t.UniqueKey("email")

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLitePrimaryConstraint(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\nid INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,\norg_id INTEGER NOT NULL,\nUNIQUE (id, org_id));"

	schema := Create("users", func(t *Table) {
		t.Increments("id")
		t.Int("org_id")
		t.PrimaryKey("id", "org_id")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteTimestamps(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "CREATE TABLE users (\ncreated_at TIMESTAMP NOT NULL,\nupdated_at TIMESTAMP NOT NULL);"

	schema := Create("users", func(t *Table) {
		t.Timestamp("created_at", 0)
		t.Timestamp("updated_at", 0)

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestMySQLTimestamps(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\ncreated_at TIMESTAMP NOT NULL,\nupdated_at TIMESTAMP NOT NULL);"

	schema := Create("users", func(t *Table) {
		t.Timestamp("created_at", 0)
		t.Timestamp("updated_at", 0)

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestPostgresTimestamps(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "CREATE TABLE users (\ncreated_at TIMESTAMP NOT NULL,\nupdated_at TIMESTAMP NOT NULL);"

	schema := Create("users", func(t *Table) {
		t.Timestamp("created_at", 0)
		t.Timestamp("updated_at", 0)

	}).Build()

	// Normalize both the expected and generated schema strings
	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteRenameColumn(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "ALTER TABLE users RENAME COLUMN username TO name;"

	schema := Alter("users", func(t *Table) {
		t.RenameColumn("username", "name")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestMySQLRenameColumn(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "ALTER TABLE users RENAME COLUMN username TO name;"

	schema := Alter("users", func(t *Table) {
		t.RenameColumn("username", "name")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestPostgresRenameColumn(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "ALTER TABLE users RENAME COLUMN username TO name;"

	schema := Alter("users", func(t *Table) {
		t.RenameColumn("username", "name")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteAlterColumn(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "ALTER TABLE users ALTER COLUMN name VARCHAR(100) NOT NULL;"

	schema := Alter("users", func(t *Table) {
		t.String("name", 100).Change()

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLAlterColumn(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "ALTER TABLE users MODIFY COLUMN name VARCHAR(100) NOT NULL;"

	schema := Alter("users", func(t *Table) {
		t.String("name", 100).Change()

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestPostgresAlterColumn(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "ALTER TABLE users ALTER COLUMN name VARCHAR(100) NOT NULL;"

	schema := Alter("users", func(t *Table) {
		t.String("name", 100).Change()

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteAddColumnToExistingTable(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "ALTER TABLE users ADD COLUMN age INTEGER;"

	schema := Alter("users", func(t *Table) {
		t.Int("age").Nullable()

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestMySQLAddColumnToExistingTable(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "ALTER TABLE users ADD COLUMN age INT;"

	schema := Alter("users", func(t *Table) {
		t.Int("age").Nullable()

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestPostgresAddColumnToExistingTable(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	expected := "ALTER TABLE users ADD COLUMN age INTEGER;"

	schema := Alter("users", func(t *Table) {
		t.Int("age").Nullable()

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)
	if normalizedSchema != normalizedExpected {
		t.Errorf("Expected schema to be %s, got %s", expected, schema)
	}
}

func TestSQLiteDropColumn(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "ALTER TABLE users DROP COLUMN username;"

	schema := Alter("users", func(t *Table) {
		t.DropColumn("username")

	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestMySQLDropColumn(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "ALTER TABLE users DROP COLUMN username;"

	schema := Alter("users", func(t *Table) {
		t.DropColumn("username")
	}).Build()

	normalizedExpected := normalizeSchema(expected)
	normalizedSchema := normalizeSchema(schema)

	if normalizedSchema != normalizedExpected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestSQLiteDropTable(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	expected := "DROP TABLE users;"
	schema := Drop("users").Build()

	if schema != expected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

func TestMySQLDropTable(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "DROP TABLE users;"
	schema := Drop("users").Build()

	if schema != expected {
		t.Errorf("\nExpected:\n %s \nGot:\n %s", expected, schema)
	}
}

// Normalize schema string by removing extra spaces, tabs, and newlines
func normalizeSchema(schema string) string {
	schema = strings.ReplaceAll(schema, "\n", "")
	schema = strings.ReplaceAll(schema, "\t", "")
	schema = strings.ReplaceAll(schema, " ", "")
	return schema
}

// MySQL refuses "AUTO_INCREMENT" on a column that is not a key:
// Error 1075, "there can be only one auto column and it must be defined as a
// key". Every other Increments test marks the column .Primary() explicitly, so
// the unmarked form — which is the natural thing to write, since it implies a
// primary key elsewhere — went uncovered.
func TestMySQLIncrementsWithoutExplicitPrimary(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nid INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY);"

	schema := Create("users", func(t *Table) {
		t.Increments("id")
	}).Build()

	if normalizeSchema(schema) != normalizeSchema(expected) {
		t.Errorf("\nExpected: \n%s, \nGot: \n%s", expected, schema)
	}
}

func TestMySQLBigIncrementsWithoutExplicitPrimary(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")
	expected := "CREATE TABLE users (\nid BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY);"

	schema := Create("users", func(t *Table) {
		t.BigIncrements("id")
	}).Build()

	if normalizeSchema(schema) != normalizeSchema(expected) {
		t.Errorf("\nExpected: \n%s, \nGot: \n%s", expected, schema)
	}
}

// An explicitly marked column must not gain a second PRIMARY KEY.
func TestMySQLIncrementsPrimaryIsNotDuplicated(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")

	schema := Create("users", func(t *Table) {
		t.Increments("id").Primary()
	}).Build()

	if count := strings.Count(schema, "PRIMARY KEY"); count != 1 {
		t.Errorf("expected exactly one PRIMARY KEY, got %d:\n%s", count, schema)
	}
}

// When the table declares its own primary key, the inline one must be left off
// so the two do not collide.
func TestMySQLIncrementsDefersToTablePrimaryKey(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")

	schema := Create("memberships", func(t *Table) {
		t.Increments("id")
		t.Int("org_id")
		t.PrimaryKey("id", "org_id")
	}).Build()

	if strings.Contains(schema, "AUTO_INCREMENT PRIMARY KEY") {
		t.Errorf("an inline PRIMARY KEY must not be added when the table declares one:\n%s", schema)
	}
	if !strings.Contains(schema, "PRIMARY KEY (id, org_id)") {
		t.Errorf("expected the declared composite primary key:\n%s", schema)
	}
}

// Postgres uses SERIAL, which carries no such requirement, and must be
// unaffected by the MySQL fix.
func TestPostgresIncrementsWithoutExplicitPrimaryIsUnchanged(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")

	schema := Create("users", func(t *Table) {
		t.Increments("id")
	}).Build()

	if strings.Contains(schema, "PRIMARY KEY") {
		t.Errorf("postgres must not gain an inline PRIMARY KEY:\n%s", schema)
	}
}

// MySQL refuses an index on a TEXT or BLOB column without a key length
// (Error 1170). Catching it while the statement is built turns a confusing
// mid-migration driver error into a message naming the column and the fix.
func TestMySQLRejectsIndexedTextColumns(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")

	cases := map[string]func(t *Table){
		"column unique":  func(t *Table) { t.Text("email").Unique() },
		"column primary": func(t *Table) { t.Text("email").Primary() },
		"table unique":   func(t *Table) { t.Text("email"); t.UniqueKey("email") },
		"table primary":  func(t *Table) { t.Text("email"); t.PrimaryKey("email") },
		"table index":    func(t *Table) { t.Text("email"); t.Index("email") },
		// Types without a dedicated builder are reachable through AddColumn.
		"blob unique": func(t *Table) {
			t.AddColumn("payload", NewDataType("payload", ColTypeBlob, DriverMySQL)).Unique()
		},
		"long text": func(t *Table) {
			t.AddColumn("body", NewDataType("body", ColTypeLongText, DriverMySQL)).Unique()
		},
	}

	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			schema := Create("users", build)

			err := schema.Validate()
			if err == nil {
				t.Fatalf("expected %s to be rejected", name)
			}
			if !strings.Contains(err.Error(), "String(") {
				t.Fatalf("the error should name the fix: %v", err)
			}

			// Build panics rather than emitting DDL the server will refuse.
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected Build to panic on an invalid schema")
					}
				}()
				_ = schema.Build()
			}()
		})
	}
}

// A bounded column is indexable, and an unindexed text column is fine.
func TestMySQLAcceptsIndexableColumns(t *testing.T) {
	os.Setenv("DB_DRIVER", "mysql")

	schema := Create("users", func(t *Table) {
		t.BigIncrements("id").Primary()
		t.String("email", 255).Unique()
		t.Text("bio") // not indexed, so perfectly fine
	})
	if err := schema.Validate(); err != nil {
		t.Fatalf("a bounded unique column must be accepted: %v", err)
	}
	if built := schema.Build(); !strings.Contains(built, "VARCHAR(255)") {
		t.Fatalf("unexpected DDL: %s", built)
	}
}

// Postgres and SQLite index text columns without complaint, so the check must
// not fire for them.
func TestOnlyMySQLRejectsIndexedText(t *testing.T) {
	for _, driver := range []string{"postgres", "sqlite"} {
		t.Run(driver, func(t *testing.T) {
			os.Setenv("DB_DRIVER", driver)
			schema := Create("users", func(t *Table) { t.Text("email").Unique() })
			if err := schema.Validate(); err != nil {
				t.Fatalf("%s indexes text fine; it must not be rejected: %v", driver, err)
			}
			if built := schema.Build(); built == "" {
				t.Fatal("expected DDL to be produced")
			}
		})
	}
}
