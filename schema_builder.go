package migration

import (
	"fmt"
	"github.com/gertd/go-pluralize"
	"os"
	"slices"
	"strings"
)

type constraint struct {
	name           string
	expr           string
	operation      string
	primaryColumns []string
	uniqueColumns  []string
	foreignKey     *foreignKey
	index          *index
	check          *check
	// ifExists tolerates the object being absent on a drop. See
	// DropBuilder.IfExists.
	ifExists bool
	// kind is what this constraint *is*, for the drop path.
	//
	// An add knows from the field that is populated — primaryColumns, index,
	// check and so on. A drop names an identifier and nothing else, so every
	// one of those fields is nil and the discriminator has to be recorded
	// separately. Without it DropIndex, DropUniqueKey, DropForeignKey and
	// DropPrimaryKey each matched no branch and emitted nothing at all, which
	// left `ALTER TABLE t ;` — a syntax error on Postgres, and a migration
	// that appeared to drop something.
	kind string
	// named records that a caller set the identifier explicitly, as opposed
	// to taking the derived one. Postgres emits an anonymous UNIQUE (...)
	// otherwise, and promoting every one of them to a named CONSTRAINT would
	// change the output of every migration already written.
	named bool
}

// check is a CHECK constraint: a named boolean expression over the row.
type check struct {
	name       string
	expression string
}

type index struct {
	name    string
	columns []string
	// unique makes this a unique *index*, which is not the same object as a
	// UNIQUE constraint on Postgres even though MySQL implements one with the
	// other. UniqueKey declares the constraint; this declares the index.
	unique bool
	// where is a partial index's predicate. Postgres and SQLite only; see
	// IndexBuilder.Where.
	where string
	// descending names the columns indexed in descending order. A set rather
	// than a parallel slice, because the order that matters is the column
	// order and a second ordered list to keep in step with it is a bug
	// waiting for somebody to insert a column in the middle.
	descending map[string]bool
}

type foreignKey struct {
	name    string
	table   *Table
	columns []string
	// references is the referenced column list. A slice because a composite
	// key is referenced as a whole: FOREIGN KEY (a, b) REFERENCES t (a, b) is
	// one constraint over a pair, and two single-column keys do not mean the
	// same thing — neither of them requires the pair to exist together.
	references []string
	on         string
	onDelete   string
	onUpdate   string
}

// Column type is the column definition
type Column struct {
	table        *Table
	name         string
	dataType     *DataType
	nullable     bool
	defaultValue any
	hasDefault   bool
	unique       bool
	primary      bool
	incrementing bool
	oldName      string
	operation    string
	// foreignKeys  []*foreignKey
}

// Table type is the table definition
type Table struct {
	dialect     string
	name        string
	columns     []*Column
	constraints []*constraint
	operation   string
}

// Schema type is the schema definition
type Schema struct {
	dialect   string
	tableName string
	operation string
	table     *Table
}

func determineDialect() string {
	if dialect := os.Getenv("DB_DRIVER"); dialect != "" {
		return dialect
	}
	// default to sqlite
	return "sqlite"
}

// NewSchema creates a new schema based on the driver provided in the environment variable DB_DRIVER
func NewSchema() *Schema {
	return &Schema{dialect: determineDialect()}
}

// NewSchemaFor creates a schema for an explicitly named dialect.
//
// The DB_DRIVER default suits a migration run, where the environment already
// says which database is being migrated. It does not suit a caller that knows
// its target — a framework package emitting DDL for a connection it holds, or
// a test covering all three dialects in one process — because an environment
// variable is global and a test cannot vary it safely in parallel.
func NewSchemaFor(dialect string) *Schema {
	return &Schema{dialect: dialect}
}

// Create provides callback to create a new table, and returns a schema
func Create(tableName string, tableFunc func(t *Table)) *Schema {
	s := NewSchema()
	s.tableName = tableName
	s.operation = "create"
	t := &Table{name: tableName, dialect: s.dialect}
	s.table = t
	tableFunc(t)
	return s
}

// CreateFor is Create for an explicitly named dialect, for a caller that
// knows its target rather than relying on DB_DRIVER.
func CreateFor(dialect, tableName string, tableFunc func(t *Table)) *Schema {
	s := NewSchemaFor(dialect)
	s.tableName = tableName
	s.operation = "create"
	t := &Table{name: tableName, dialect: s.dialect}
	s.table = t
	tableFunc(t)
	return s
}

// DropFor is Drop for an explicitly named dialect.
func DropFor(dialect, tableName string) *Schema {
	s := NewSchemaFor(dialect)
	s.tableName = tableName
	s.operation = "drop"
	return s
}

// Alter provides callback to alter an existing table, and returns a schema
// AlterFor builds an ALTER for a named dialect, rather than for whatever
// DB_DRIVER says.
//
// CreateFor and DropFor already existed and this did not, which meant an
// alter could not be rendered for a dialect it was not running against — so
// nothing about its portability was testable, and four of its drop helpers
// emitted a syntax error on Postgres for as long as nobody looked.
func AlterFor(dialect, tableName string, tableFunc func(t *Table)) *Schema {
	s := NewSchemaFor(dialect)
	s.operation = "alter"
	s.tableName = tableName
	t := &Table{name: tableName, dialect: s.dialect}
	tableFunc(t)
	s.table = t
	return s
}

func Alter(tableName string, tableFunc func(t *Table)) *Schema {
	s := NewSchema()
	s.tableName = tableName
	s.operation = "alter"
	t := &Table{name: tableName, dialect: s.dialect}
	s.table = t
	tableFunc(t)
	return s
}

// Drop returns a schema to drop a table
func Drop(tableName string) *Schema {
	s := NewSchema()
	s.tableName = tableName
	s.operation = "drop"
	s.table = &Table{name: tableName, dialect: s.dialect}
	return s
}

// HasConstraints returns true if the table has constraints
func (t *Table) HasConstraints() bool {
	return len(t.constraints) > 0
}

// hasPrimaryKeyConstraint reports whether a table-level primary key was
// declared, e.g. through t.PrimaryKey("id", "org_id").
func (t *Table) hasPrimaryKeyConstraint() bool {
	for _, c := range t.constraints {
		if len(c.primaryColumns) > 0 {
			return true
		}
	}
	return false
}

// Increments adds an auto-incrementing column to the table
func (t *Table) Increments(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeIncrements, t.dialect)).Unsigned()
	c.incrementing = true
	return c
}

// BigIncrements adds an auto-incrementing column to the table with big integers
func (t *Table) BigIncrements(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeBigIncrements, t.dialect)).Unsigned()
	c.incrementing = true
	return c
}

// String adds a varchar column to the table
func (t *Table) String(name string, length uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeVarchar, t.dialect).WithLength(length))
	return c
}

// Text adds a text column to the table
func (t *Table) Text(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeText, t.dialect))
	return c
}

// UUID adds a UUID column to the table.
// PostgreSQL uses the native UUID type, MySQL uses CHAR(36), SQLite uses TEXT.
func (t *Table) UUID(name string) *Column {
	dt := NewDataType(name, ColTypeUUID, t.dialect)
	if t.dialect == DriverMySQL {
		dt.length = 36
	}
	return t.AddColumn(name, dt)
}

// ULID adds a ULID column to the table.
// ULIDs are 26-character Crockford base32 encoded identifiers.
func (t *Table) ULID(name string) *Column {
	dt := NewDataType(name, ColTypeULID, t.dialect)
	if t.dialect == DriverMySQL || t.dialect == DriverPostgres {
		dt.length = 26
	}
	return t.AddColumn(name, dt)
}

// TinyInt adds a tiny integer column to the table
func (t *Table) TinyInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeTinyInt, t.dialect))
	return c
}

// SmallInt adds a small integer column to the table
func (t *Table) SmallInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeSmallInt, t.dialect))
	return c
}

// MediumInt adds a medium integer column to the table
func (t *Table) MediumInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeMediumInt, t.dialect))
	return c
}

// Int adds an integer column to the table
func (t *Table) Int(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeInt, t.dialect))
	return c
}

// BigInt adds a big integer column to the table
func (t *Table) BigInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeBigInt, t.dialect))
	return c
}

// Binary adds a binary column to the table
func (t *Table) Binary(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeBinary, t.dialect))
	return c
}

// Boolean adds a boolean column to the table
func (t *Table) Boolean(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeBool, t.dialect))
	return c
}

// Char adds a char column to the table
func (t *Table) Char(name string, length uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeChar, t.dialect).WithLength(length))
	return c
}

// DateTimeTz adds a date time with timezone column to the table
func (t *Table) DateTimeTz(name string, precision uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeDateTimeTz, t.dialect).WithPrecision(precision))
	return c
}

// DateTime adds a date time column to the table
func (t *Table) DateTime(name string, precision uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeDateTime, t.dialect).WithPrecision(precision))
	return c
}

// Date adds a date column to the table
func (t *Table) Date(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeDate, t.dialect))
	return c
}

// Decimal adds a decimal column to the table
func (t *Table) Decimal(name string, precision uint, scale uint) *Column {
	c := t.AddColumn(
		name,
		NewDataType(name, ColTypeDecimal, t.dialect).
			WithPrecision(precision).
			WithScale(scale),
	)
	return c
}

// Double adds a double column to the table
func (t *Table) Double(name string, precision uint, scale uint) *Column {
	c := t.AddColumn(
		name,
		NewDataType(name, ColTypeDouble, t.dialect).
			WithPrecision(precision).
			WithScale(scale),
	)
	return c
}

// Enum adds an enum column to the table
func (t *Table) Enum(name string, values ...string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeEnum, t.dialect).WithEnumValues(values))
	return c
}

// UnsignedBigTInt adds an unsigned tiny integer column to the table
func (t *Table) UnsignedBigInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeBigInt, t.dialect)).Unsigned()
	return c
}

// UnsignedInt adds an unsigned integer column to the table
func (t *Table) UnsignedInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeInt, t.dialect)).Unsigned()
	return c
}

// UnsignedMediumInt adds an unsigned medium integer column to the table
func (t *Table) UnsignedMediumInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeMediumInt, t.dialect)).Unsigned()
	return c
}

// UnsignedSmallInt adds an unsigned small integer column to the table
func (t *Table) UnsignedSmallInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeSmallInt, t.dialect)).Unsigned()
	return c
}

// UnsignedTinyInt adds an unsigned tiny integer column to the table
func (t *Table) UnsignedTinyInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeTinyInt, t.dialect)).Unsigned()
	return c
}

// Float adds a float column to the table
func (t *Table) Float(name string, precision uint, scale uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeFloat, t.dialect).WithPrecision(precision).WithScale(scale))
	return c
}

// Time adds a time column to the table
func (t *Table) Time(name string, precision uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeTime, t.dialect).WithPrecision(precision))
	return c
}

// Timestamp adds a timestamp column to the table
func (t *Table) Timestamp(name string, precision uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeTimestamp, t.dialect).WithPrecision(precision))
	return c
}

// TimestampTz adds a timestamp with timezone column to the table
func (t *Table) TimestampTz(name string, precision uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeTimestampTz, t.dialect).WithPrecision(precision))
	return c
}

// AddColumn adds a new column to the table
func (t *Table) AddColumn(name string, dataType *DataType) *Column {
	c := &Column{
		name:      name,
		table:     t,
		dataType:  dataType,
		operation: "add",
	}
	t.columns = append(t.columns, c)
	return c
}

// DropColumn drops a column from the table
func (t *Table) DropColumn(name string) *Column {
	c := &Column{
		table:     t,
		name:      name,
		operation: "drop",
	}
	t.columns = append(t.columns, c)
	return c
}

// AlterColumn alters a column in the table
func (t *Table) AlterColumn(name string, dataType string) *Column {
	c := &Column{
		table:     t,
		name:      name,
		operation: "alter",
	}
	t.columns = append(t.columns, c)
	return c
}

// RenameColumn renames a column in the table
func (t *Table) RenameColumn(oldName string, newName string) *Column {
	c := &Column{
		table:     t,
		name:      newName,
		oldName:   oldName,
		operation: "rename",
	}
	t.columns = append(t.columns, c)
	return c
}

// PrimaryKey adds a primary key to the table
func (t *Table) PrimaryKey(columns ...string) {
	// panic if there is already a drop operation for primary key
	for _, c := range t.constraints {
		if c.operation == "drop" && len(c.primaryColumns) > 0 {
			panic("cannot add primary key when there is a drop operation for primary key in the table")
		}
	}

	// panic if there is already a primary key
	for _, c := range t.constraints {
		if c.operation == "add" && len(c.primaryColumns) > 0 {
			panic("multiple primary keys are not allowed in a table")
		}
	}

	c := &constraint{
		name:           t.name + "_pkey",
		operation:      "add",
		primaryColumns: columns,
	}
	t.constraints = append(t.constraints, c)
}

// DropPrimaryKey drops the primary key from the table
func (t *Table) DropPrimaryKey() *DropBuilder {
	// panic if there is already an add operation for primary key
	for _, c := range t.constraints {
		if c.operation == "add" && len(c.primaryColumns) > 0 {
			panic("cannot drop primary key when there is an add operation for primary key in the table")
		}
	}
	c := &constraint{
		name:      t.name + "_pkey",
		operation: "drop",
		kind:      "primary",
	}
	t.constraints = append(t.constraints, c)
	return &DropBuilder{c: c}
}

// Index adds an index to the table
// NamedConstraint is the handle Index and UniqueKey return, so a caller can
// override the generated name.
//
// It exists because a derived name is right for a new table and wrong for an
// existing one. Every index in a migration that has already run has a name in
// somebody's database, and a schema builder that cannot reproduce it forces
// the whole statement to be written by hand — which is how a project ends up
// with eighty raw CREATE INDEX lines and a DSL used only for columns.
//
// Returning a handle rather than taking an optional argument keeps the common
// call unchanged: t.Index("user_id") still reads as a statement, and the
// return value is there for the one caller in twenty that needs it.
type NamedConstraint struct{ c *constraint }

// Name overrides the generated identifier.
//
// Empty is ignored rather than accepted, because an empty index name is not a
// choice anybody makes on purpose and Postgres would reject the statement
// much further from the mistake.
func (n *NamedConstraint) Name(name string) *NamedConstraint {
	if n == nil || n.c == nil || name == "" {
		return n
	}
	n.c.name = name
	if n.c.index != nil {
		n.c.index.name = name
	}
	n.c.named = true
	return n
}

func (t *Table) Index(columns ...string) *IndexBuilder {
	// Always prefixed with the table, because an index name is schema-global
	// in Postgres. Naming from the columns alone would collide: two tables
	// each wanting an index on (user_id, position) would both ask for
	// `user_id_position_index`. The whole batch runs in one transaction, so
	// the second would take the migration down with it.
	indexName := t.name + "_" + strings.Join(columns, "_") + "_index"
	c := &constraint{
		name:      indexName,
		operation: "add",
		index: &index{
			name:       indexName,
			columns:    columns,
			descending: map[string]bool{},
		},
	}
	t.constraints = append(t.constraints, c)
	return &IndexBuilder{c: c}
}

// IndexBuilder is the handle Index returns.
//
// Separate from NamedConstraint, which UniqueKey still returns, because
// Unique and Desc mean something on an index and nothing on a unique
// constraint — and a method that silently does nothing is worse than one that
// is not there to call.
type IndexBuilder struct{ c *constraint }

// Name overrides the generated identifier. See NamedConstraint.Name for why
// this exists at all.
func (b *IndexBuilder) Name(name string) *IndexBuilder {
	if b == nil || b.c == nil || name == "" {
		return b
	}
	b.c.name = name
	if b.c.index != nil {
		b.c.index.name = name
	}
	b.c.named = true
	return b
}

// Unique makes this a unique index.
//
// Not the same thing as UniqueKey, and the difference is not cosmetic: on
// Postgres a UNIQUE constraint and a unique index are different catalog
// objects, so a migration that declared one cannot be reproduced with the
// other — pg_dump renders a constraint as ALTER TABLE ... ADD CONSTRAINT and
// an index as CREATE UNIQUE INDEX. MySQL implements the constraint *with* an
// index, which is why the distinction is easy to miss until a schema has to
// be matched rather than merely created.
//
// Reach for UniqueKey when the uniqueness is part of what the table is, and
// for this when it is an index that happens to be unique — a lookup key, or
// one you will later want a partial predicate on.
func (b *IndexBuilder) Unique() *IndexBuilder {
	if b == nil || b.c == nil || b.c.index == nil {
		return b
	}
	b.c.index.unique = true
	return b
}

// Where makes this a partial index, covering only the rows matching the
// predicate.
//
// Postgres and SQLite both have them; **MySQL has no form of a partial index
// at all**, so Validate refuses one rather than dropping the predicate. That
// is the choice this package has got wrong before — Enum on Postgres emitted
// no constraint and said nothing, which is worse than a malformed statement
// because the schema looks right. An index that silently covers every row
// instead of a tenth of them is the same failure: correct results, a plan
// nobody can account for, and a table that grows an index it was promised
// would stay small.
//
// A predicate is SQL, not a value, and it is interpolated. It belongs in a
// migration beside the schema it constrains, never anywhere near user input.
func (b *IndexBuilder) Where(predicate string) *IndexBuilder {
	if b == nil || b.c == nil || b.c.index == nil || predicate == "" {
		return b
	}
	b.c.index.where = predicate
	return b
}

// Desc indexes the named columns in descending order.
//
// Named columns rather than a direction per position, and only columns
// already in the index: writing the direction into the column list — as
// Index("created_at DESC") — happens to work today because the builder joins
// its arguments verbatim, and stops working the day it quotes an identifier,
// at which point the index is built against a column that does not exist.
//
// A column named here that is not in the index is an error reported by
// Validate rather than ignored, because the ignored version is an index
// silently sorted the wrong way and a query plan nobody can explain.
//
// Postgres, MySQL 8.0 and SQLite 3.8.3 all honour a descending index key.
// Only MySQL before 8.0 parsed and discarded it, which is also the version
// that cannot do a CHECK, so a schema built by this package has already left
// it behind.
func (b *IndexBuilder) Desc(columns ...string) *IndexBuilder {
	if b == nil || b.c == nil || b.c.index == nil {
		return b
	}
	if b.c.index.descending == nil {
		b.c.index.descending = map[string]bool{}
	}
	for _, column := range columns {
		b.c.index.descending[column] = true
	}
	return b
}

// DropIndex drops an index from the table
func (t *Table) DropIndex(indexName string) *DropBuilder {
	c := &constraint{
		name:      indexName,
		operation: "drop",
		kind:      "index",
	}
	t.constraints = append(t.constraints, c)
	return &DropBuilder{c: c}
}

// UniqueKey adds a unique constraint to the table
func (t *Table) UniqueKey(columns ...string) *NamedConstraint {
	// The same rule as Index, and the bug here was the mirror image: the
	// single-column branch overwrote the prefix it had just built, so
	// `email_unique` belonged to whichever table migrated first. A unique
	// constraint is backed by an index in Postgres, so the name is
	// schema-global just the same.
	constraintName := t.name + "_" + strings.Join(columns, "_") + "_unique"
	c := &constraint{
		name:          constraintName,
		operation:     "add",
		uniqueColumns: columns,
	}
	t.constraints = append(t.constraints, c)
	return &NamedConstraint{c: c}
}

// DropUniqueKey drops a unique constraint from the table
func (t *Table) DropUniqueKey(name string) *DropBuilder {
	c := &constraint{
		name:      name,
		operation: "drop",
		kind:      "unique",
	}
	t.constraints = append(t.constraints, c)
	return &DropBuilder{c: c}
}

// Check adds a named CHECK constraint to the table.
//
// The expression is written in SQL and passed through untouched, because a
// check is the one constraint whose whole content is an expression: a regex on
// a handle, a range on a count, a set of permitted states, a relationship
// between two columns. Anything that tried to model it would cover a fraction
// of what people write and send the rest back to raw SQL.
//
// The name is required rather than derived. It is what a later migration
// drops, what Postgres puts in the error when a row is refused, and the thing
// a reader greps for — none of which survives a generated name nobody chose.
//
// Supported everywhere this package targets: Postgres, MySQL 8.0.16 and
// later, and SQLite. Note that MySQL before 8.0.16 parses a CHECK and ignores
// it, which is its own problem and not one this package can detect from here.
func (t *Table) Check(name string, expression string) {
	c := &constraint{
		name:      name,
		operation: "add",
		check:     &check{name: name, expression: expression},
	}
	t.constraints = append(t.constraints, c)
}

// DropCheck removes a CHECK constraint by name.
//
// SQLite cannot: it has no ALTER TABLE DROP CONSTRAINT, and removing one means
// rebuilding the table. That is a migration an application writes for itself,
// so this refuses rather than emitting something SQLite will reject.
func (t *Table) DropCheck(name string) *DropBuilder {
	c := &constraint{
		name:      name,
		operation: "drop",
		kind:      "check",
		check:     &check{name: name},
	}
	t.constraints = append(t.constraints, c)
	return &DropBuilder{c: c}
}

// DropBuilder is the handle every Drop* helper returns.
//
// It exists for IfExists, which a down migration almost always wants and an
// up migration almost never does: reverting a change should not fail because
// the thing it removes was already gone, while an up migration naming a
// constraint that is not there has a typo in it and should say so.
type DropBuilder struct{ c *constraint }

// IfExists makes the drop tolerate the object being absent.
//
// Not available on MySQL for a constraint, an index or a key — it has no
// IF EXISTS for any ALTER TABLE drop action, so Validate refuses it there
// rather than emitting a statement the server will reject. Postgres has it
// for all of them; SQLite has it for DROP INDEX, which is the only one of
// these it can do at all.
func (b *DropBuilder) IfExists() *DropBuilder {
	if b == nil || b.c == nil {
		return b
	}
	b.c.ifExists = true
	return b
}

// ForeignID accepts an id column that references the primary column of another table
func (t *Table) ForeignID(column string) *foreignKey {
	fk := &foreignKey{
		table:      t,
		columns:    []string{column},
		references: []string{"id"},
	}
	fkName := column
	c := &constraint{
		name:       fkName + "_fkey",
		operation:  "add",
		foreignKey: fk,
	}
	// fk.name = c.name
	t.constraints = append(t.constraints, c)
	t.AddColumn(column, NewDataType(column, ColTypeBigIncrements, t.dialect))

	return fk
}

// Foreign adds a foreign key to the table
func (t *Table) Foreign(columns ...string) *foreignKey {
	fk := &foreignKey{
		table:   t,
		columns: columns,
	}
	fkName := strings.Join(columns, "_")
	c := &constraint{
		name:       fkName + "_fkey",
		operation:  "add",
		foreignKey: fk,
	}
	// fk.name = c.name
	t.constraints = append(t.constraints, c)
	return fk
}

// Constrained is shorthand of .References("id").On("pluralized_table_name")
func (f *foreignKey) Constrained() *foreignKey {
	f.references = []string{"id"}
	referencedTable := guessPluralizedTableNameFromColumnName(f.columns[0])
	f.on = referencedTable
	return f
}

// ConstrainedFunc sets the table and name of the foreign key
func (f *foreignKey) ConstrainedFunc(fn func(t *Table) (table, indexName string)) *foreignKey {
	table, indexName := fn(f.table)
	f.name = indexName
	f.references = []string{"id"}
	f.on = table
	return f
}

// References sets the columns that the foreign key references.
//
// Variadic, so References("id") reads exactly as it did, and a composite key
// becomes sayable: Foreign("user_id", "skill_id").References("user_id",
// "skill_id").On("user_skills") is one constraint requiring the pair to exist
// together, which two single-column keys do not express — neither of them
// would stop a row naming a user and a skill that are never paired.
//
// The referenced columns must carry a unique or primary key in that order, on
// every dialect. Validate checks the count against the referencing columns;
// whether the target is actually unique is the database's to refuse, since
// the target table is usually not the one being built here.
func (f *foreignKey) References(columns ...string) *foreignKey {
	if len(columns) == 0 {
		return f
	}
	f.references = columns
	return f
}

// On sets the table that the foreign key references
func (f *foreignKey) On(table string) *foreignKey {
	f.on = table
	return f
}

// Name gives the constraint an identifier.
//
// Opt-in, for the reason NamedConstraint.Name records: every foreign key this
// package has emitted so far is anonymous, so Postgres named it
// <table>_<column>_fkey and that is the name sitting in every database that
// has run those migrations. Emitting the derived name by default would rename
// all of them, which is the one thing a schema builder must not do to a schema
// that already exists.
//
// It is what lets an existing hand-written constraint be restated rather than
// replaced: a composite key added as
// `ALTER TABLE endorsements ADD CONSTRAINT endorsements_claimed_skill ...`
// can now be declared with the table and keep its name.
func (f *foreignKey) Name(name string) *foreignKey {
	if name == "" {
		return f
	}
	f.name = name
	return f
}

// OnDelete adds the ON DELETE clause to the foreign key
func (f *foreignKey) OnDelete(onDelete string) *foreignKey {
	f.onDelete = onDelete
	return f
}

// OnUpdate adds the ON UPDATE clause to the foreign key
func (f *foreignKey) OnUpdate(onUpdate string) *foreignKey {
	f.onUpdate = onUpdate
	return f
}

// DropForeignKey drops a foreign key from the table
func (t *Table) DropForeignKey(name string) *DropBuilder {
	c := &constraint{
		name:      name,
		operation: "drop",
		kind:      "foreign",
	}
	t.constraints = append(t.constraints, c)
	return &DropBuilder{c: c}
}

// Type adds a data type to the column
func (c *Column) Type(dataType *DataType) *Column {
	c.dataType = dataType
	return c
}

// Unsigned adds the unsigned attribute to the column
func (c *Column) Unsigned() *Column {
	c.dataType.unsigned = true
	return c
}

// Nullable adds the nullable attribute to the column
func (c *Column) Nullable() *Column {
	c.nullable = true
	return c
}

// NotNull adds the not null attribute to the column
func (c *Column) NotNull() *Column {
	c.nullable = false
	return c
}

// Default sets the column's default value.
//
// The value is rendered as a SQL literal according to its Go type, so pass the
// value you actually want: Default("active") produces DEFAULT \'active\' and
// Default(18) produces DEFAULT 18. Strings are quoted and their quotes
// escaped, booleans and time.Time are rendered the way the target dialect
// expects, and a nil value produces DEFAULT NULL.
//
// To emit a SQL expression rather than a literal — CURRENT_TIMESTAMP, a
// function call, an arithmetic expression — wrap it in [Expr], which is passed
// through verbatim:
//
//	t.String("status", 255).Default("active")
//	t.DateTime("created_at", 6).Default(Expr("CURRENT_TIMESTAMP"))
//
// A value of a type with no SQL literal form is reported by [Schema.Validate],
// so it surfaces while the statement is built rather than as a driver error.
func (c *Column) Default(defaultValue any) *Column {
	c.defaultValue = defaultValue
	c.hasDefault = true
	return c
}

// Unique adds the unique attribute to the column
func (c *Column) Unique() *Column {
	c.unique = true
	return c
}

// Primary adds the primary attribute to the column
func (c *Column) Primary() *Column {
	c.primary = true
	return c
}

// Change changes the operation of the column to alter
func (c *Column) Change() {
	c.operation = "alter"
}

// Done returns the table that the column belongs to
func (c *Column) Done() *Table {
	return c.table
}

// Build returns the SQL query for the schema
// Json adds a document column.
//
// Postgres gets JSONB, MySQL gets JSON and SQLite gets TEXT — which is not a
// fallback but the way SQLite stores JSON, since its JSON1 functions operate
// on TEXT.
//
// No length argument, because none of the three takes one.
func (t *Table) Json(name string) *Column {
	return t.AddColumn(name, NewDataType(name, ColTypeJson, t.dialect))
}

// buildDropAction is one ALTER TABLE drop action, spelled the way this dialect
// spells it.
//
// The three disagree more here than anywhere else in the builder, which is
// why this is a switch rather than a format string:
//
//   - a primary key is DROP PRIMARY KEY on MySQL and DROP CONSTRAINT
//     <table>_pkey on Postgres, which needs a name MySQL does not use;
//   - a unique constraint is backed by an index on MySQL, so it goes with
//     DROP INDEX, while Postgres drops the constraint;
//   - a foreign key is DROP FOREIGN KEY on MySQL and DROP CONSTRAINT on
//     Postgres;
//   - an index is not an ALTER TABLE action at all outside MySQL — it is a
//     statement of its own, emitted by droppedIndexes.
//
// SQLite can do none of them except dropping an index, and refuses rather
// than emitting something it will reject.
func (s *Schema) buildDropAction(c *constraint) string {
	ifExists := ""
	if c.ifExists {
		ifExists = "IF EXISTS "
	}

	switch c.kind {
	case "index":
		if s.dialect != DriverMySQL {
			// Emitted as its own DROP INDEX statement instead.
			return ""
		}
		return "DROP INDEX " + c.name
	case "primary":
		s.requireAlterableConstraints("dropping")
		if s.dialect == DriverMySQL {
			return "DROP PRIMARY KEY"
		}
		return "DROP CONSTRAINT " + ifExists + c.name
	case "unique":
		s.requireAlterableConstraints("dropping")
		if s.dialect == DriverMySQL {
			return "DROP INDEX " + c.name
		}
		return "DROP CONSTRAINT " + ifExists + c.name
	case "foreign":
		s.requireAlterableConstraints("dropping")
		if s.dialect == DriverMySQL {
			return "DROP FOREIGN KEY " + c.name
		}
		return "DROP CONSTRAINT " + ifExists + c.name
	case "check":
		s.requireAlterableConstraints("dropping")
		return "DROP CONSTRAINT " + ifExists + c.name
	}
	return ""
}

// droppedIndexes emits a DROP INDEX for every dialect that does not take one
// as an ALTER TABLE action — which is every dialect but MySQL.
//
// Postgres and SQLite drop an index by name alone, without naming the table,
// because an index name is schema-global in both.
func (s *Schema) droppedIndexes() []string {
	if s.table == nil || s.dialect == DriverMySQL {
		return nil
	}
	var statements []string
	for _, c := range s.table.constraints {
		if c.operation != "drop" || c.kind != "index" {
			continue
		}
		ifExists := ""
		if c.ifExists {
			ifExists = "IF EXISTS "
		}
		statements = append(statements, "DROP INDEX "+ifExists+c.name+";")
	}
	return statements
}

// unindexableMySQLTypes cannot carry an index without a key length. MySQL
// rejects them outright: "BLOB/TEXT column used in key specification without a
// key length" (Error 1170). Postgres and SQLite index them happily, so this is
// a MySQL-only restriction.
var unindexableMySQLTypes = map[string]bool{
	// JSON is in here for a different reason than the rest: MySQL refuses it
	// in a key outright rather than for want of a length ("JSON column
	// 'data' cannot be used in key specification"), and no length would help.
	// Indexing inside a JSON document there means a generated column over
	// the extracted value, which is not something this builder can guess.
	ColTypeJson:       true,
	ColTypeText:       true,
	ColTypeTinyText:   true,
	ColTypeMediumText: true,
	ColTypeLongText:   true,
	ColTypeBlob:       true,
	ColTypeTinyBlob:   true,
	ColTypeMediumBlob: true,
	ColTypeLongBlob:   true,
}

// Validate reports schema problems the target database would reject, so they
// surface while building the statement rather than as a driver error part-way
// through a migration.
func (s *Schema) Validate() error {
	if s.table == nil {
		return nil
	}

	// A default with no SQL literal form is dropped by buildColumn, which
	// would silently produce a table without the default the migration asked
	// for. Report it instead, on every dialect.
	for _, column := range s.table.columns {
		if !column.hasDefault {
			continue
		}
		if _, err := defaultLiteral(s.table.dialect, column.defaultValue); err != nil {
			return fmt.Errorf("migration: %s.%s: %w", s.tableName, column.name, err)
		}
	}

	// Both of the checks below are dialect-independent: they catch a
	// declaration that cannot mean what it says on any database.
	for _, c := range s.table.constraints {
		// MySQL has no IF EXISTS on any drop of a constraint, a key or an
		// index — not on the ALTER TABLE action and not on DROP INDEX. The
		// check covers indexes too: exempting them would silently discard the
		// tolerance the caller asked for, and the only sign would be a
		// rollback that fails on a second attempt.
		if c.ifExists && s.dialect == DriverMySQL {
			return fmt.Errorf(
				"migration: %s: MySQL has no IF EXISTS when dropping %s — neither "+
					"on an ALTER TABLE drop action nor on DROP INDEX; drop it "+
					"unconditionally, or branch on the dialect",
				s.tableName, c.name)
		}
		if c.index != nil && c.index.where != "" && s.dialect == DriverMySQL {
			return fmt.Errorf(
				"migration: %s: index %s has a WHERE predicate and MySQL has no "+
					"partial index; declare it without Where, or branch on the "+
					"dialect and write this one as raw SQL",
				s.tableName, c.index.name)
		}
		if c.index != nil {
			for column := range c.index.descending {
				if !slices.Contains(c.index.columns, column) {
					return fmt.Errorf(
						"migration: %s: index %s orders %q descending but does not index it; "+
							"name it in Index(...) as well, or drop it from Desc(...)",
						s.tableName, c.index.name, column)
				}
			}
		}
		if c.foreignKey != nil && len(c.foreignKey.references) > 0 &&
			len(c.foreignKey.references) != len(c.foreignKey.columns) {
			return fmt.Errorf(
				"migration: %s: foreign key on (%s) references (%s) — a composite key "+
					"must reference exactly as many columns as it carries",
				s.tableName, strings.Join(c.foreignKey.columns, ", "),
				strings.Join(c.foreignKey.references, ", "))
		}
	}

	if s.dialect != DriverMySQL {
		return nil
	}

	// Keyed to the type, not to a bool: the error's advice depends on which
	// type it is, and "text or blob" was being passed as a literal — so a
	// JSON column was told to declare itself as String(name, 255).
	unindexable := map[string]string{}
	for _, column := range s.table.columns {
		if column.dataType == nil || !unindexableMySQLTypes[column.dataType.genericName] {
			continue
		}
		unindexable[column.name] = column.dataType.genericName

		if column.unique {
			return unindexableColumnError(s.tableName, column.name, column.dataType.genericName, "UNIQUE")
		}
		if column.primary {
			return unindexableColumnError(s.tableName, column.name, column.dataType.genericName, "PRIMARY KEY")
		}
	}
	if len(unindexable) == 0 {
		return nil
	}

	// The same restriction applies to table-level keys and indexes.
	for _, c := range s.table.constraints {
		for _, group := range [][]string{c.primaryColumns, c.uniqueColumns} {
			for _, name := range group {
				if kind, ok := unindexable[name]; ok {
					return unindexableColumnError(s.tableName, name, kind, "a key")
				}
			}
		}
		if c.index != nil {
			for _, name := range c.index.columns {
				if kind, ok := unindexable[name]; ok {
					return unindexableColumnError(s.tableName, name, kind, "an index")
				}
			}
		}
	}
	return nil
}

func unindexableColumnError(table, column, columnType, usage string) error {
	// The advice has to match the reason. A TEXT column is refused for want
	// of a key length, which a bounded String fixes; a JSON column is refused
	// outright, and suggesting String(name, 255) there would be telling
	// somebody to store a document in a 255-character field.
	if columnType == ColTypeJson {
		return fmt.Errorf(
			"migration: %s.%s is json, which MySQL cannot use in %s at all; "+
				"index a generated column over the extracted value instead",
			table, column, usage)
	}
	return fmt.Errorf(
		"migration: %s.%s is %s, which MySQL cannot use in %s without a key length; "+
			"declare it with String(%q, 255) instead",
		table, column, columnType, usage, column)
}

func (s *Schema) Build() string {
	// Build has no error return and every migration calls it directly, so an
	// invalid schema panics rather than handing the driver DDL it will refuse.
	// Callers that want to check without panicking can use Validate.
	if err := s.Validate(); err != nil {
		panic(err)
	}

	return strings.Join(s.Statements(), "\n")
}

// Statements returns the DDL as individual statements.
//
// Build joins them with a newline, which works for a migration running
// through a driver that accepts several statements at once. MySQL's does not
// unless the DSN opts in with multiStatements, so a caller executing DDL
// itself needs them separately — and splitting Build's output on semicolons
// afterwards is guesswork the builder does not have to make anyone do.
func (s *Schema) Statements() []string {
	if err := s.Validate(); err != nil {
		panic(err)
	}

	statements := []string{}
	if s.operation == "alter" {
		// An alter can be several statements, because SQLite takes one action
		// per ALTER TABLE. buildOperation joins them for Build; this keeps
		// them separate, which is what Statements promises.
		statements = append(statements, s.alterStatements()...)
	} else if operation := s.buildOperation(); operation != "" {
		statements = append(statements, operation)
	}
	statements = append(statements, s.droppedIndexes()...)
	statements = append(statements, s.trailingIndexes()...)
	return statements
}

// trailingIndexStatements renders the indexes that cannot be declared inside
// CREATE TABLE. Only MySQL accepts an inline INDEX clause; everywhere else an
// index is its own statement.
// closeDefinition trims the separator left dangling when the constraint list
// contributes nothing.
func closeDefinition(sql string) string {
	return strings.TrimRight(sql, ", \n\t")
}

func (s *Schema) trailingIndexStatements() string {
	return strings.Join(s.trailingIndexes(), "\n")
}

func (s *Schema) trailingIndexes() []string {
	if s.table == nil || s.operation == "drop" {
		return nil
	}
	// MySQL takes an index inline, but only inside CREATE TABLE — there is no
	// inline form in an ALTER, so one added there needs its own statement on
	// every dialect.
	if s.dialect == DriverMySQL && s.operation != "alter" {
		return nil
	}

	var statements []string
	for _, constraint := range s.table.constraints {
		if constraint.index == nil || constraint.operation == "drop" {
			continue
		}
		keyword := "CREATE INDEX "
		if constraint.index.unique {
			keyword = "CREATE UNIQUE INDEX "
		}
		statement := keyword + constraint.index.name +
			" ON " + s.tableName + " (" + s.buildIndexColumns(constraint.index) + ")"
		if constraint.index.where != "" {
			statement += " WHERE " + constraint.index.where
		}
		statements = append(statements, statement+";")
	}
	return statements
}

func (s *Schema) buildOperation() string {
	switch s.operation {
	case "create":
		return s.buildCreate()
	case "alter":
		return s.buildAlter()
	case "drop":
		return s.buildDrop()
	}
	return ""
}

func (s *Schema) buildCreate() string {
	switch s.dialect {
	case DriverSQLite:
		return s.buildCreateSQLite()
	case DriverMySQL:
		return s.buildCreateMySQL()
	case DriverPostgres:
		return s.buildCreatePostgreSQL()
	}
	return ""
}

// buildAlter assembles one ALTER TABLE from its actions.
//
// Actions are collected and joined rather than concatenated, which is a fix
// as much as a tidy-up: the three dialect-specific versions this replaces
// each appended straight onto one string, so two column operations in a
// single Alter came out as `ADD COLUMN a INT ADD COLUMN b INT` with no comma
// between them — valid-looking and refused by every server.
//
// An ALTER with no actions emits nothing at all, rather than
// `ALTER TABLE t ;` — a syntax error on Postgres, and what a migration that
// only drops an index would otherwise produce.
func (s *Schema) buildAlter() string {
	return strings.Join(s.alterStatements(), "\n")
}

// alterActions is every ALTER TABLE action this schema asks for, in the order
// it was declared.
func (s *Schema) alterActions() []string {
	actions := []string{}

	for _, column := range s.table.columns {
		switch column.operation {
		case "add":
			actions = append(actions, "ADD COLUMN "+s.columnDefinition(column))
		case "drop":
			actions = append(actions, "DROP COLUMN "+column.name)
		case "alter":
			actions = append(actions, s.changeColumn(column)...)
		case "rename":
			actions = append(actions, "RENAME COLUMN "+column.oldName+" TO "+column.name)
		}
	}

	for _, constraint := range s.table.constraints {
		switch constraint.operation {
		case "add":
			if action := s.buildAddAction(constraint); action != "" {
				actions = append(actions, action)
			}
		case "drop":
			if action := s.buildDropAction(constraint); action != "" {
				actions = append(actions, action)
			}
		}
	}

	return actions
}

// alterStatements is the ALTER TABLE work as individual statements.
//
// Postgres and MySQL take a comma-separated list of actions in one statement.
// **SQLite takes exactly one**: its ALTER TABLE grammar is RENAME TO, RENAME
// COLUMN, ADD COLUMN or DROP COLUMN, singular, so
// `ALTER TABLE t ADD COLUMN a INT, ADD COLUMN b INT` is a syntax error there.
// Found by executing the output rather than by reading it, which is the whole
// reason alter_execution_test.go exists.
func (s *Schema) alterStatements() []string {
	if s.table == nil || s.operation != "alter" {
		return nil
	}
	return s.alterStatementsFrom(s.alterActions())
}

func (s *Schema) alterStatementsFrom(actions []string) []string {
	if len(actions) == 0 {
		return nil
	}
	if s.dialect == DriverSQLite {
		statements := make([]string, 0, len(actions))
		for _, action := range actions {
			statements = append(statements, "ALTER TABLE "+s.tableName+" "+action+";")
		}
		return statements
	}
	return []string{"ALTER TABLE " + s.tableName + " " + strings.Join(actions, ", ") + ";"}
}

// columnDefinition is a column's full declaration, without the separator the
// CREATE path needs.
func (s *Schema) columnDefinition(column *Column) string {
	return strings.TrimSuffix(strings.TrimSpace(s.buildColumn(column)), ",")
}

// changeColumn is what "this column should now look like this" becomes.
//
// One action on MySQL and SQLite, which both restate the whole column.
// Several on Postgres, which has no such verb: it changes a type, a
// nullability and a default as three separate actions. `ALTER COLUMN x
// VARCHAR(255)` is a syntax error there. Verified against a server rather
// than read off the grammar.
func (s *Schema) changeColumn(column *Column) []string {
	if s.dialect == DriverMySQL {
		return []string{"MODIFY COLUMN " + s.columnDefinition(column)}
	}
	if s.dialect == DriverSQLite {
		// SQLite has no column alteration at all beyond a rename: changing a
		// type or a nullability means rebuilding the table. Refusing is the
		// same choice DropCheck already makes.
		panic("migration: SQLite cannot alter a column's type or nullability, so " +
			"changing " + s.tableName + "." + column.name + " is not expressible; " +
			"rebuild the table in the migration itself")
	}

	actions := []string{}
	if column.dataType != nil {
		actions = append(actions,
			"ALTER COLUMN "+column.name+" TYPE "+column.dataType.ToString())
	}
	if column.nullable {
		actions = append(actions, "ALTER COLUMN "+column.name+" DROP NOT NULL")
	} else {
		actions = append(actions, "ALTER COLUMN "+column.name+" SET NOT NULL")
	}
	if column.hasDefault {
		if literal, err := defaultLiteralWithPrecision(
			s.dialect, column.defaultValue, column.dataType.Precision(),
		); err == nil {
			actions = append(actions,
				"ALTER COLUMN "+column.name+" SET DEFAULT "+literal)
		}
	}
	return actions
}

// buildAddAction is one ALTER TABLE add action.
//
// The CREATE path renders these as part of a table definition; in an ALTER
// each one is an action and needs its verb, which is why they are not the
// same code.
func (s *Schema) buildAddAction(c *constraint) string {
	switch {
	case len(c.primaryColumns) > 0:
		return "ADD PRIMARY KEY (" + s.buildColumns(c.primaryColumns) + ")"
	case len(c.uniqueColumns) > 0:
		s.requireAlterableConstraints("adding")
		if c.named {
			return "ADD CONSTRAINT " + c.name + " UNIQUE (" + s.buildColumns(c.uniqueColumns) + ")"
		}
		return "ADD UNIQUE (" + s.buildColumns(c.uniqueColumns) + ")"
	case c.check != nil:
		s.requireAlterableConstraints("adding")
		return "ADD CONSTRAINT " + c.check.name + " CHECK (" + c.check.expression + ")"
	case c.foreignKey != nil:
		s.requireAlterableConstraints("adding")
		return "ADD " + strings.TrimSpace(s.buildForeignKey(c.foreignKey))
	}
	// An index added in an ALTER comes out as its own CREATE INDEX, from
	// trailingIndexes — including on MySQL, where the inline form has no
	// ALTER equivalent.
	return ""
}

func (s *Schema) buildDrop() string {
	switch s.dialect {
	case DriverSQLite:
		return s.buildDropSQLite()
	case DriverMySQL:
		return s.buildDropMySQL()
	case DriverPostgres:
		return s.buildDropPostgreSQL()
	}
	return ""
}

func (s *Schema) buildCreateSQLite() string {
	sql := "CREATE TABLE " + s.tableName + " ("

	// SQLite requires incrementing column to be a primary key.
	// If the user doesn't mark the incremental column as a
	// primary column, we will do it for them.
	//
	// If there is an incremental column, however the primary
	// key is a multi-column composite key (e.g. t.PrimaryKey("id", "org_id")),
	// then we will set the incremental key as a primary column,
	// and convert the composite primary key into a composite unique key.
	// This is to keep the schema compatible across different dbs.
	// Misleading key type? - Yes. Gets the job done? - Also yes.

	hasCompositePrimaryKey := false
	for _, column := range s.table.columns {
		if column.incrementing {
			for _, c := range column.table.constraints {
				if len(c.primaryColumns) == 1 {
					c.primaryColumns = []string{}
					break
				}

				if len(c.primaryColumns) > 1 {
					for _, primaryColumn := range c.primaryColumns {
						c.uniqueColumns = append(c.uniqueColumns, primaryColumn)
					}
					c.primaryColumns = []string{}
					hasCompositePrimaryKey = true
					break
				}
			}

			if hasCompositePrimaryKey {
				fmt.Println(fmt.Sprintf("[Warning: %s column is marked as incremental, however a composite primary key is provided.\n"+
					"The provided primary columns have been set as a unique key and the incremental column has been set as the primary key.]", column.name))
			}
			column.primary = true
		}

		sql += s.buildColumn(column)
	}

	// The separator after the last column is not trimmed here, and must not
	// be. Trimming it needs a guard listing every constraint kind that might
	// follow, and such a list is wrong the next time a kind is added — miss
	// one and the column runs straight into CONSTRAINT, which SQLite refuses
	// at "NOT NULLCONSTRAINT".
	//
	// So append whatever the constraints render to and let closeDefinition
	// take the trailing separator off the end. It
	// already did exactly that for every other dialect.
	sql += s.buildConstraints()
	sql = closeDefinition(sql)
	sql += "\n);"
	return sql
}

func (s *Schema) HasUniqueConstraints() bool {
	return slices.ContainsFunc(s.table.constraints, func(c *constraint) bool {
		return len(c.uniqueColumns) > 0
	})
}

func (s *Schema) HasForeignKeys() bool {
	return slices.ContainsFunc(s.table.constraints, func(c *constraint) bool {
		return c.foreignKey != nil
	})
}

func (s *Schema) HasNonPrimaryConstraints() bool {
	return slices.ContainsFunc(s.table.constraints, func(c *constraint) bool {
		return len(c.uniqueColumns) > 0 || c.foreignKey != nil
	})
}

func (s *Schema) HasPrimaryKeyOnTable() bool {
	for _, constraint := range s.table.constraints {
		if len(constraint.primaryColumns) > 0 {
			return true
		}
	}

	return false
}

func (s *Schema) HasCompositePrimaryKey() bool {
	for _, constraint := range s.table.constraints {
		if len(constraint.primaryColumns) > 1 {
			return true
		}
	}

	return false
}

func (s *Schema) buildCreateMySQL() string {
	sql := "CREATE TABLE " + s.tableName + " ("
	for index, column := range s.table.columns {
		if index == len(s.table.columns)-1 && !s.table.HasConstraints() {
			sql += strings.TrimSuffix(s.buildColumn(column), ", ")
		} else {
			sql += s.buildColumn(column)
		}
	}
	sql += s.buildConstraints()
	sql = closeDefinition(sql)
	sql += ");"
	return sql
}

func (s *Schema) buildCreatePostgreSQL() string {
	sql := "CREATE TABLE " + s.tableName + " ("
	for index, column := range s.table.columns {
		if index == len(s.table.columns)-1 && !s.table.HasConstraints() {
			sql += strings.TrimSuffix(s.buildColumn(column), ", ")
		} else {
			sql += s.buildColumn(column)
		}
	}
	sql += s.buildConstraints()
	sql = closeDefinition(sql)
	sql += ");"
	return sql
}

func (s *Schema) buildDropSQLite() string {
	return "DROP TABLE " + s.tableName + ";"
}

func (s *Schema) buildDropMySQL() string {
	return "DROP TABLE " + s.tableName + ";"
}

func (s *Schema) buildDropPostgreSQL() string {
	return "DROP TABLE " + s.tableName + ";"
}

func (s *Schema) buildColumn(column *Column) string {
	sql := "\n" + column.name + " "

	if column.dataType != nil {
		sql += column.dataType.ToString()
	}

	if !column.nullable {
		sql += " NOT NULL"
	}
	if column.hasDefault {
		// An unrenderable default is rejected by Validate, which Build runs
		// before reaching here; the error is ignored rather than duplicated.
		if literal, err := defaultLiteralWithPrecision(
			column.table.dialect, column.defaultValue, column.dataType.Precision(),
		); err == nil {
			sql += " DEFAULT " + literal
		}
	}
	if column.unique {
		sql += " UNIQUE"
	}
	if column.primary {
		sql += " PRIMARY KEY"
	}
	if column.table.dialect == DriverSQLite && column.incrementing {
		sql += " AUTOINCREMENT"
	}
	if column.table.dialect == DriverMySQL && column.incrementing {
		sql += " AUTO_INCREMENT"
		// MySQL rejects an auto-increment column that is not a key:
		// "there can be only one auto column and it must be defined as a key".
		// The SQLite path already promotes the incrementing column to the
		// primary key when the caller did not; do the same here rather than
		// emitting DDL the server will refuse.
		if !column.primary && !column.table.hasPrimaryKeyConstraint() {
			sql += " PRIMARY KEY"
		}
	}

	// This column level foreign key is not being executed at all.
	// if len(column.foreignKeys) > 0 {
	// 	for _, fk := range column.foreignKeys {
	// 		sql += ", " + s.buildForeignKey(fk)
	// 	}
	// }

	if column.dataType != nil {
		sql += column.dataType.suffix
	}

	return sql + ", "
}

// requireAlterableConstraints refuses the combination SQLite cannot express.
//
// SQLite has no ALTER TABLE ADD/DROP CONSTRAINT: changing one means creating a
// new table, copying the rows and swapping the names, which is a migration an
// application writes deliberately rather than something a builder should
// improvise under it. Refusing loudly beats emitting a statement the database
// rejects at run time, and beats silently dropping the constraint — which is
// the failure this whole change exists to remove.
func (s *Schema) requireAlterableConstraints(what string) {
	if s.dialect == DriverSQLite {
		panic("migration: SQLite cannot alter a table's constraints, so " + what +
			" a CHECK on " + s.tableName + " is not expressible; declare it in " +
			"Create, or rebuild the table in the migration itself")
	}
}

func (s *Schema) buildConstraints() string {
	sql := ""
	for _, constraint := range s.table.constraints {
		switch constraint.operation {
		case "add":
			if len(constraint.primaryColumns) > 0 {
				sql += "PRIMARY KEY (" + s.buildColumns(constraint.primaryColumns) + "), "
			}
			if len(constraint.uniqueColumns) > 0 {
				prefix := ""
				// An explicit name is emitted as a real CONSTRAINT on every
				// dialect, because the caller asked for an identifier and a
				// unique constraint that reports itself as
				// users_email_key1 in an error is the reason they asked.
				//
				// Without one Postgres and SQLite stay anonymous and let the
				// server derive it — not an oversight but the behaviour every
				// migration already written depends on, so promoting them
				// would rewrite schemas that exist.
				switch {
				case constraint.named:
					prefix = "CONSTRAINT " + constraint.name + " UNIQUE "
				case s.dialect == DriverMySQL:
					prefix = "UNIQUE " + constraint.name + " "
				default:
					prefix = "UNIQUE "
				}
				sql += prefix + "(" + s.buildColumns(constraint.uniqueColumns) + "), "
			}
			if constraint.index != nil && s.dialect == DriverMySQL {
				// Inline INDEX is a MySQL extension. The other dialects get a
				// separate CREATE INDEX appended after the table instead.
				keyword := "INDEX "
				if constraint.index.unique {
					keyword = "UNIQUE INDEX "
				}
				sql += keyword + constraint.index.name + " (" + s.buildIndexColumns(constraint.index) + "), "
			}
			if constraint.check != nil {
				// Inside CREATE TABLE a constraint is part of the definition;
				// in an ALTER it is an action, so it needs the verb. Without
				// this the alter path emitted `ALTER TABLE t CONSTRAINT ...`,
				// which no dialect accepts — and adding a check to a table
				// that already exists is at least as common as declaring one
				// with it.
				verb := ""
				if s.operation == "alter" {
					s.requireAlterableConstraints("adding")
					verb = "ADD "
				}
				sql += verb + "CONSTRAINT " + constraint.check.name +
					" CHECK (" + constraint.check.expression + "), "
			}
			if constraint.foreignKey != nil {
				sql += s.buildForeignKey(constraint.foreignKey) + ", "
			}
		case "drop":
			if action := s.buildDropAction(constraint); action != "" {
				sql += action + ", "
			}
		}
	}
	// Remove trailing comma if there is any
	if len(sql) > 0 {
		sql = sql[:len(sql)-2]
	}
	return sql
}

func (s *Schema) buildForeignKey(fk *foreignKey) string {
	sql := ""
	if fk.name != "" {
		sql += "\nCONSTRAINT " + fk.name + " "
	} else {
		sql += "\n"
	}

	sql += "FOREIGN KEY (" + s.buildColumns(fk.columns) + ") REFERENCES " + fk.on + "(" + s.buildColumns(fk.references) + ")"
	if fk.onDelete != "" {
		sql += " ON DELETE " + fk.onDelete
	}
	if fk.onUpdate != "" {
		sql += " ON UPDATE " + fk.onUpdate
	}
	return sql
}

// buildIndexColumns renders an index's key, with a direction where one was
// asked for.
//
// Separate from buildColumns because a direction belongs to an index key and
// to nothing else: a primary key, a unique constraint and a foreign key are
// column lists, and appending ASC/DESC to any of them is a syntax error on at
// least one dialect.
func (s *Schema) buildIndexColumns(idx *index) string {
	if idx == nil {
		return ""
	}
	if len(idx.descending) == 0 {
		return s.buildColumns(idx.columns)
	}
	rendered := make([]string, 0, len(idx.columns))
	for _, column := range idx.columns {
		if idx.descending[column] {
			rendered = append(rendered, column+" DESC")
			continue
		}
		rendered = append(rendered, column)
	}
	return strings.Join(rendered, ", ")
}

func (s *Schema) buildColumns(columns []string) string {
	sql := ""
	for _, column := range columns {
		sql += column + ", "
	}
	return sql[:len(sql)-2]
}

// String returns the SQL query for the schema
func (s *Schema) String() string {
	return s.Build()
}

func guessPluralizedTableNameFromColumnName(columnName string) string {
	pluralize := pluralize.NewClient()
	if strings.HasSuffix(columnName, "id") {
		nameParts := strings.Split(columnName, "_")
		if len(nameParts) > 1 {
			return pluralize.Plural(nameParts[len(nameParts)-2])
		}
		return pluralize.Plural(nameParts[0])
	}
	return pluralize.Plural(columnName)
}
