package migration

import (
	"database/sql"
	"strings"
	"testing"
)

// Every Go example in content/docs/migrations, compiled and run.
//
// The logging docs in this monorepo were fiction for a whole release — every
// sample called a function that did not exist — because nothing ever ran
// them. The migration docs were no better: they recommended t.Integer, which
// is not a method; t.ForeignID(...).Constrained() for a foreign key column,
// which on PostgreSQL declares a BIGSERIAL; TimestampTz with a precision,
// which emits a syntax error; and `return migration.Drop("users")`, which
// does not compile.
//
// A doc example that is a test cannot drift from the API without this file
// going red. Examples are reproduced here as they appear on the page, in the
// same order; where a page shows a fragment, it is wrapped in the smallest
// thing that compiles.

// --- _index.md -------------------------------------------------------------

func mig_20250706000001_create_users_table_up(tx *sql.Tx) error {
	schema := Create("users", func(t *Table) {
		t.BigIncrements("id").Primary()
		t.String("email", 255)
		t.String("handle", 30)
		t.TimestampTz("created_at", 0).Default(Expr("now()"))

		t.Check("users_handle_format", `handle ~ '^[a-z0-9][a-z0-9_]{1,29}$'`)
		t.Index("email").Unique().Name("users_email_key")
	}).Build()

	_, err := tx.Exec(schema)
	return err
}

func mig_20250706000001_create_users_table_down(tx *sql.Tx) error {
	_, err := tx.Exec(Drop("users").Build())
	return err
}

// --- schema-builder.md -----------------------------------------------------

func TestDocsSchemaBuilderExamples(t *testing.T) {
	// "Why declare instead of writing SQL"
	for _, dialect := range []string{DriverPostgres, DriverMySQL, DriverSQLite} {
		sql := CreateFor(dialect, "circle_members", func(t *Table) {
			t.UnsignedBigInt("circle_id")
			t.UnsignedBigInt("user_id")
			t.String("role", 12).Default("member")
			t.TimestampTz("joined_at", 0).Default(Expr("CURRENT_TIMESTAMP"))
			t.PrimaryKey("circle_id", "user_id")
			t.Index("user_id", "joined_at").Name("circle_members_user_idx")
		}).Build()

		// The table on the page, asserted rather than described.
		if dialect == DriverMySQL {
			if !strings.Contains(sql, "BIGINT UNSIGNED") {
				t.Errorf("MySQL: expected BIGINT UNSIGNED in:\n%s", sql)
			}
			if !strings.Contains(sql, "INDEX circle_members_user_idx") ||
				strings.Contains(sql, "CREATE INDEX") {
				t.Errorf("MySQL: expected an inline index in:\n%s", sql)
			}
		} else {
			if !strings.Contains(sql, "CREATE INDEX circle_members_user_idx") {
				t.Errorf("%s: expected a separate CREATE INDEX in:\n%s", dialect, sql)
			}
		}
		if dialect == DriverPostgres && !strings.Contains(sql, "TIMESTAMP WITH TIME ZONE") {
			t.Errorf("postgres: expected a zoned timestamp in:\n%s", sql)
		}
	}

	// "Create, alter, drop"
	create := CreateFor(DriverPostgres, "users", func(t *Table) {
		t.BigIncrements("id").Primary()
		t.String("email", 255)
		t.TimestampTz("created_at", 0).Default(Expr("now()"))
	}).Build()
	if !strings.Contains(create, "CREATE TABLE users") {
		t.Errorf("unexpected create:\n%s", create)
	}

	statements := CreateFor(DriverPostgres, "users", func(t *Table) {
		t.BigIncrements("id").Primary()
		t.Index("id").Name("users_id_idx")
	}).Statements()
	if len(statements) != 2 {
		t.Errorf("expected a table and an index, got %d statements", len(statements))
	}

	if err := CreateFor(DriverPostgres, "users", func(t *Table) {
		t.BigIncrements("id").Primary()
	}).Validate(); err != nil {
		t.Errorf("the Validate example does not validate: %v", err)
	}

	alter := AlterFor(DriverPostgres, "users", func(t *Table) {
		t.Int("view_count").Default(0)
		t.String("nickname", 80).Nullable().Change()
		t.DropColumn("legacy_body")
		t.RenameColumn("body", "body_source")
	}).Build()
	if !strings.Contains(alter, "ALTER TABLE users") {
		t.Errorf("unexpected alter:\n%s", alter)
	}

	if got := DropFor(DriverPostgres, "users").Build(); !strings.Contains(got, "DROP TABLE") {
		t.Errorf("unexpected drop: %q", got)
	}

	// "Columns", and the three traps the page names.
	columns := CreateFor(DriverPostgres, "posts", func(t *Table) {
		t.BigIncrements("id").Primary()
		t.String("email", 255)
		t.Text("bio")
		t.Int("view_count").Default(0)
		t.Boolean("is_published").Default(false)
		t.Json("payload").Nullable()
		t.TimestampTz("created_at", 0)

		t.String("status", 16).Default("draft")
		t.Check("posts_status_valid", "status IN ('draft','published','removed')")

		t.UnsignedBigInt("user_id")
		t.Foreign("user_id").References("id").On("users").OnDelete("CASCADE")
	}).Build()
	for _, want := range []string{
		"payload JSONB",
		"created_at TIMESTAMP WITH TIME ZONE NOT NULL",
		"CONSTRAINT posts_status_valid CHECK",
		"FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
	} {
		if !strings.Contains(normalizeSchema(columns), normalizeSchema(want)) {
			t.Errorf("expected %q in:\n%s", want, columns)
		}
	}

	// The page claims TimestampTz with a precision is broken. Asserted, so
	// the advice cannot quietly stop being true.
	broken := CreateFor(DriverPostgres, "posts", func(t *Table) {
		t.TimestampTz("created_at", 6)
	}).Build()
	if !strings.Contains(broken, "TIMESTAMP WITH TIME ZONE(6)") {
		t.Errorf("TimestampTz(name, 6) no longer emits the invalid form the "+
			"page warns about:\n%s", broken)
	}

	// And that ForeignID declares an auto-increment column.
	fk := CreateFor(DriverPostgres, "posts", func(t *Table) {
		t.BigIncrements("id").Primary()
		t.ForeignID("user_id").Constrained()
	}).Build()
	if !strings.Contains(fk, "user_id BIGSERIAL") {
		t.Errorf("ForeignID no longer emits the BIGSERIAL the page warns "+
			"about:\n%s", fk)
	}

	// And the Enum split the page describes: a native ENUM on MySQL, a TEXT
	// with a CHECK on Postgres, and plain TEXT with no constraint on SQLite.
	//
	// This assertion is here because the first draft of the page said Enum
	// enforced nothing outside MySQL — true once, and fixed since — and this
	// test is what caught it.
	if pg := CreateFor(DriverPostgres, "posts", func(t *Table) {
		t.Enum("status", "draft", "published")
	}).Build(); !strings.Contains(pg, "CHECK") {
		t.Errorf("postgres no longer gets a CHECK from Enum:\n%s", pg)
	}
	if my := CreateFor(DriverMySQL, "posts", func(t *Table) {
		t.Enum("status", "draft", "published")
	}).Build(); !strings.Contains(my, "ENUM(") {
		t.Errorf("mysql no longer gets a native ENUM:\n%s", my)
	}
	if lite := CreateFor(DriverSQLite, "posts", func(t *Table) {
		t.Enum("status", "draft", "published")
	}).Build(); strings.Contains(lite, "CHECK") {
		t.Errorf("sqlite now constrains an Enum; the page says it does not:\n%s", lite)
	}

	// "Modifiers"
	CreateFor(DriverPostgres, "users", func(t *Table) {
		t.String("email", 255).Unique()
		t.String("name", 255).NotNull()
		t.String("phone", 255).Nullable()
		t.BigIncrements("id").Primary()
		t.Int("views").Unsigned()
	}).Build()

	// "Defaults"
	defaults := CreateFor(DriverPostgres, "users", func(t *Table) {
		t.String("status", 255).Default("active")
		t.String("note", 255).Default("it's fine")
		t.Int("age").Default(18)
		t.Boolean("admin").Default(false)
		t.String("nickname", 255).Nullable().Default(nil)
		t.Int("attempts").Default(Expr("(1 + 1)"))
		t.TimestampTz("created_at", 0).Default(Expr("now()"))
	}).Build()
	for _, want := range []string{"DEFAULT 'active'", "DEFAULT 'it''s fine'", "DEFAULT 18"} {
		if !strings.Contains(defaults, want) {
			t.Errorf("expected %q in:\n%s", want, defaults)
		}
	}

	// CurrentTimestamp on a DATETIME(6), which is the case it exists for.
	CreateFor(DriverMySQL, "users", func(t *Table) {
		t.DateTime("created_at", 6).Default(CurrentTimestamp)
	}).Build()

	// "Keys and constraints"
	CreateFor(DriverPostgres, "circle_members", func(t *Table) {
		t.BigIncrements("id").Primary()
		t.UnsignedBigInt("circle_id")
		t.UnsignedBigInt("user_id")
		t.PrimaryKey("circle_id", "user_id")
	}).Build()

	CreateFor(DriverPostgres, "posts", func(t *Table) {
		t.Int("reaction_count").Default(0)
		t.Int("comment_count").Default(0)
		t.String("kind", 16)
		t.Check("posts_kind_valid", "kind IN ('status','text','link')")
		t.Check("posts_counts_sane", "reaction_count >= 0 AND comment_count >= 0")
	}).Build()

	// "Indexes"
	indexes := CreateFor(DriverPostgres, "posts", func(t *Table) {
		t.UnsignedBigInt("user_id")
		t.String("email", 255)
		t.TimestampTz("created_at", 0)
		t.TimestampTz("deleted_at", 0).Nullable()

		t.Index("user_id")
		t.Index("user_id", "created_at").Name("posts_author_idx")
		t.Index("user_id", "created_at").Desc("created_at").Name("posts_recent_idx")
		t.Index("email").Unique().Name("users_email_key")
		t.Index("email").Unique().Where("deleted_at IS NULL").Name("users_live_email_key")
	}).Build()
	for _, want := range []string{
		"CREATE INDEX posts_recent_idx ON posts (user_id, created_at DESC);",
		"CREATE UNIQUE INDEX users_live_email_key ON posts (email) WHERE deleted_at IS NULL;",
	} {
		if !strings.Contains(normalizeSchema(indexes), normalizeSchema(want)) {
			t.Errorf("expected %q in:\n%s", want, indexes)
		}
	}

	// "Unique: index or constraint" — the page says these differ.
	asIndex := CreateFor(DriverPostgres, "users", func(t *Table) {
		t.String("email", 255)
		t.Index("email").Unique().Name("users_email_key")
	}).Build()
	asConstraint := CreateFor(DriverPostgres, "users", func(t *Table) {
		t.String("email", 255)
		t.UniqueKey("email").Name("users_email_key")
	}).Build()
	if asIndex == asConstraint {
		t.Error("the page says a unique index and a unique key differ; they did not")
	}

	// "Foreign keys"
	fks := CreateFor(DriverPostgres, "endorsements", func(t *Table) {
		t.UnsignedBigInt("user_id")
		t.UnsignedBigInt("skill_id")
		t.Foreign("user_id").References("id").On("users").OnDelete("CASCADE")
		t.Foreign("user_id", "skill_id").
			Name("endorsements_claimed_skill").
			References("user_id", "skill_id").
			On("user_skills").
			OnDelete("CASCADE")
	}).Build()
	if !strings.Contains(normalizeSchema(fks),
		normalizeSchema("CONSTRAINT endorsements_claimed_skill FOREIGN KEY (user_id, skill_id)")) {
		t.Errorf("the named composite key did not render:\n%s", fks)
	}
}

// "Altering a table", which the page gives as one block and whose spellings
// it tabulates per dialect.
func TestDocsAlterExample(t *testing.T) {
	example := func(t *Table) {
		t.Int("view_count").Default(0)
		t.DropIndex("posts_author_idx")
		t.DropUniqueKey("posts_slug_unique")
		t.DropForeignKey("posts_author_fk")
		t.DropCheck("posts_kind_valid").IfExists()
		t.DropPrimaryKey()
	}

	pg := strings.Join(AlterFor(DriverPostgres, "posts", example).Statements(), " ")
	for _, want := range []string{
		"DROP CONSTRAINT posts_pkey",
		"DROP CONSTRAINT posts_slug_unique",
		"DROP CONSTRAINT posts_author_fk",
		"DROP CONSTRAINT IF EXISTS posts_kind_valid",
		"DROP INDEX posts_author_idx;",
	} {
		if !strings.Contains(normalizeSchema(pg), normalizeSchema(want)) {
			t.Errorf("postgres: expected %q in:\n%s", want, pg)
		}
	}

	// The page says a column change is refused on SQLite.
	func() {
		defer func() {
			if recover() == nil {
				t.Error("SQLite accepted a column change, which the page says it refuses")
			}
		}()
		AlterFor(DriverSQLite, "posts", func(t *Table) {
			t.String("slug", 160).Nullable().Change()
		}).Build()
	}()

	// And that MySQL refuses a tolerant drop.
	if err := AlterFor(DriverMySQL, "posts", func(t *Table) {
		t.DropCheck("posts_kind_valid").IfExists()
	}).Validate(); err == nil {
		t.Error("MySQL accepted IF EXISTS, which the page says it refuses")
	}

	// And that MySQL refuses a partial index.
	if err := CreateFor(DriverMySQL, "posts", func(t *Table) {
		t.String("email", 255)
		t.TimestampTz("deleted_at", 0).Nullable()
		t.Index("email").Where("deleted_at IS NULL").Name("posts_live_idx")
	}).Validate(); err == nil {
		t.Error("MySQL accepted a partial index, which the page says it refuses")
	}
}

// --- column-types.md -------------------------------------------------------

func TestDocsColumnTypeTable(t *testing.T) {
	// Every row of every mapping table on the page.
	for _, row := range []struct{ method, postgres, mysql, sqlite string }{
		{"Increments", "SERIAL", "INT", "INTEGER"},
		{"BigIncrements", "BIGSERIAL", "BIGINT", "INTEGER"},
		{"TinyInt", "SMALLINT", "TINYINT", "TINYINT"},
		{"SmallInt", "SMALLINT", "SMALLINT", "SMALLINT"},
		{"MediumInt", "INTEGER", "MEDIUMINT", "MEDIUMINT"},
		{"Int", "INTEGER", "INT", "INTEGER"},
		{"BigInt", "BIGINT", "BIGINT", "BIGINT"},
		{"Date", "DATE", "DATE", "DATE"},
		{"Text", "TEXT", "TEXT", "TEXT"},
		{"Json", "JSONB", "JSON", "TEXT"},
		{"UUID", "UUID", "CHAR", "TEXT"},
		{"Boolean", "BOOLEAN", "BOOLEAN", "BOOLEAN"},
		{"Binary", "BYTEA", "BINARY", "BINARY"},
	} {
		t.Run(row.method, func(t *testing.T) {
			for dialect, want := range map[string]string{
				DriverPostgres: row.postgres,
				DriverMySQL:    row.mysql,
				DriverSQLite:   row.sqlite,
			} {
				sql := CreateFor(dialect, "probe", func(tb *Table) {
					declare(tb, row.method, "col")
				}).Build()
				if !strings.Contains(normalizeSchema(sql), normalizeSchema("col "+want)) {
					t.Errorf("%s: %s should give %q, got:\n%s",
						dialect, row.method, want, sql)
				}
			}
		})
	}

	// The soft-delete trio the page shows.
	CreateFor(DriverPostgres, "posts", func(t *Table) {
		t.TimestampTz("created_at", 0).Default(Expr("now()"))
		t.TimestampTz("updated_at", 0).Default(Expr("now()"))
		t.TimestampTz("deleted_at", 0).Nullable()
	}).Build()

	// MySQL refuses a JSON column in a key, with advice of its own.
	err := CreateFor(DriverMySQL, "posts", func(t *Table) {
		t.Json("payload")
		t.Index("payload").Name("posts_payload_idx")
	}).Validate()
	if err == nil || !strings.Contains(err.Error(), "cannot use in an index at all") {
		t.Errorf("MySQL did not refuse a JSON index as the page says: %v", err)
	}
}

func declare(t *Table, method, name string) {
	switch method {
	case "Increments":
		t.Increments(name)
	case "BigIncrements":
		t.BigIncrements(name)
	case "TinyInt":
		t.TinyInt(name)
	case "SmallInt":
		t.SmallInt(name)
	case "MediumInt":
		t.MediumInt(name)
	case "Int":
		t.Int(name)
	case "BigInt":
		t.BigInt(name)
	case "Date":
		t.Date(name)
	case "Text":
		t.Text(name)
	case "Json":
		t.Json(name)
	case "UUID":
		t.UUID(name)
	case "Boolean":
		t.Boolean(name)
	case "Binary":
		t.Binary(name)
	}
}

// --- writing-migrations.md -------------------------------------------------

func mig_docs_up(tx *sql.Tx) error {
	schema := Create("posts", func(t *Table) {
		t.BigIncrements("id").Primary()
		t.UnsignedBigInt("user_id")
		t.Text("body").Default("")
		t.TimestampTz("deleted_at", 0).Nullable()

		t.Foreign("user_id").References("id").On("users").OnDelete("CASCADE")
		t.Check("posts_body_present", "length(body) > 0")
		t.Index("user_id", "deleted_at").Name("posts_author_idx")
	}).Build()

	if _, err := tx.Exec(schema); err != nil {
		return err
	}

	for _, statement := range []string{
		`ALTER TABLE posts ADD COLUMN search_vector tsvector
		   GENERATED ALWAYS AS (to_tsvector('english', coalesce(body, ''))) STORED`,
		`CREATE INDEX posts_search_idx ON posts USING GIN (search_vector)`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func mig_docs_down(tx *sql.Tx) error {
	for _, table := range []string{"post_comments", "posts"} {
		if _, err := tx.Exec(Drop(table).Build()); err != nil {
			return err
		}
	}
	return nil
}

// The IfExists example from the same page.
func TestDocsDownExample(t *testing.T) {
	got := AlterFor(DriverPostgres, "users", func(t *Table) {
		t.DropCheck("users_handle_not_reserved").IfExists()
	}).Build()
	if !strings.Contains(got, "DROP CONSTRAINT IF EXISTS users_handle_not_reserved") {
		t.Errorf("unexpected: %s", got)
	}

	// The migration functions above exist to be compiled; referencing them
	// keeps the compiler from calling them dead.
	_ = mig_20250706000001_create_users_table_up
	_ = mig_20250706000001_create_users_table_down
	_ = mig_docs_up
	_ = mig_docs_down
}
