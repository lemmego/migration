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
		t.Integer("age").Nullable()

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
		t.Integer("age").Nullable()

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
		t.BigInteger("age").Nullable()

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
		t.BigInteger("age").Nullable()

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
		t.Integer("role_id")
		t.ForeignKey("role_id").
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
		t.Integer("role_id")
		t.ForeignKey("role_id").
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
		t.Integer("role_id")
		t.ForeignKey("role_id").
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
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL,\nINDEX users_email_index (email));"

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
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL,\nINDEX users_email_index (email));"

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
	expected := "CREATE TABLE users (\nemail VARCHAR(100) NOT NULL,\nINDEX users_email_index (email));"

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
		t.Unique("email")

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
		t.Unique("email")

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
		t.Unique("email")

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
	expected := "CREATE TABLE users (\nid INTEGER NOT NULL AUTOINCREMENT,\norg_id INTEGER NOT NULL,\nPRIMARY KEY (id, org_id));"

	schema := Create("users", func(t *Table) {
		t.Increments("id")
		t.Integer("org_id")
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
		t.Integer("age").Nullable()

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
		t.Integer("age").Nullable()

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
		t.Integer("age").Nullable()

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
