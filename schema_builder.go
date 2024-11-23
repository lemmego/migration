package migration

import (
	"fmt"
	"github.com/gertd/go-pluralize"
	"os"
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
}

type index struct {
	name    string
	columns []string
}

type foreignKey struct {
	name       string
	table      *Table
	columns    []string
	references string
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

// Alter provides callback to alter an existing table, and returns a schema
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
func (t *Table) DropPrimaryKey() {
	// panic if there is already an add operation for primary key
	for _, c := range t.constraints {
		if c.operation == "add" && len(c.primaryColumns) > 0 {
			panic("cannot drop primary key when there is an add operation for primary key in the table")
		}
	}
	c := &constraint{
		name:      t.name + "_pkey",
		operation: "drop",
	}
	t.constraints = append(t.constraints, c)
}

// Index adds an index to the table
func (t *Table) Index(columns ...string) {
	indexName := ""
	if len(columns) == 1 {
		indexName = t.name + "_" + columns[0] + "_index"
	} else {
		for _, column := range columns {
			indexName += column + "_"
		}
		indexName += "index"
	}
	c := &constraint{
		name:      indexName,
		operation: "add",
		index: &index{
			name:    indexName,
			columns: columns,
		},
	}
	t.constraints = append(t.constraints, c)
}

// DropIndex drops an index from the table
func (t *Table) DropIndex(indexName string) {
	c := &constraint{
		name:      indexName,
		operation: "drop",
	}
	t.constraints = append(t.constraints, c)
}

// UniqueKey adds a unique constraint to the table
func (t *Table) UniqueKey(columns ...string) {
	constraintName := t.name + "_"
	if len(columns) == 1 {
		constraintName = columns[0] + "_unique"
	} else {
		for _, column := range columns {
			constraintName += column + "_"
		}
		constraintName += "unique"
	}
	c := &constraint{
		name:          constraintName,
		operation:     "add",
		uniqueColumns: columns,
	}
	t.constraints = append(t.constraints, c)
}

// DropUniqueKey drops a unique constraint from the table
func (t *Table) DropUniqueKey(name string) {
	c := &constraint{
		name:      name,
		operation: "drop",
	}
	t.constraints = append(t.constraints, c)
}

// ForeignID accepts an id column that references the primary column of another table
func (t *Table) ForeignID(column string) *foreignKey {
	fk := &foreignKey{
		table:      t,
		columns:    []string{column},
		references: "id",
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
	f.references = "id"
	referencedTable := guessPluralizedTableNameFromColumnName(f.columns[0])
	f.on = referencedTable
	return f
}

// ConstrainedFunc sets the table and name of the foreign key
func (f *foreignKey) ConstrainedFunc(fn func(t *Table) (table, indexName string)) *foreignKey {
	table, indexName := fn(f.table)
	f.name = indexName
	f.references = "id"
	f.on = table
	return f
}

// References sets the column that the foreign key references
func (f *foreignKey) References(column string) *foreignKey {
	f.references = column
	return f
}

// On sets the table that the foreign key references
func (f *foreignKey) On(table string) *foreignKey {
	f.on = table
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
func (t *Table) DropForeignKey(name string) {
	c := &constraint{
		name:      name,
		operation: "drop",
	}
	t.constraints = append(t.constraints, c)
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

// Default adds the default value to the column
func (c *Column) Default(defaultValue any) *Column {
	c.defaultValue = defaultValue
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
func (s *Schema) Build() string {
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

func (s *Schema) buildAlter() string {
	switch s.dialect {
	case DriverSQLite:
		return s.buildAlterSQLite()
	case DriverMySQL:
		return s.buildAlterMySQL()
	case DriverPostgres:
		return s.buildAlterPostgreSQL()
	}
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
	for index, column := range s.table.columns {
		if index == len(s.table.columns)-1 && !s.table.HasConstraints() {
			sql += s.buildColumn(column, false)
		} else {
			sql += s.buildColumn(column, true)
		}
	}
	sql += s.buildConstraints()
	sql += ");"
	return sql
}

func (s *Schema) buildCreateMySQL() string {
	sql := "CREATE TABLE " + s.tableName + " ("
	for index, column := range s.table.columns {
		if index == len(s.table.columns)-1 && !s.table.HasConstraints() {
			sql += s.buildColumn(column, false)
		} else {
			sql += s.buildColumn(column, true)
		}
	}
	sql += s.buildConstraints()
	sql += ");"
	return sql
}

func (s *Schema) buildCreatePostgreSQL() string {
	sql := "CREATE TABLE " + s.tableName + " ("
	for index, column := range s.table.columns {
		if index == len(s.table.columns)-1 && !s.table.HasConstraints() {
			sql += s.buildColumn(column, false)
		} else {
			sql += s.buildColumn(column, true)
		}
	}
	sql += s.buildConstraints()
	sql += ");"
	return sql
}

func (s *Schema) buildAlterSQLite() string {
	sql := "ALTER TABLE " + s.tableName + " "
	for index, column := range s.table.columns {
		columnStr := ""
		if index == len(s.table.columns)-1 {
			columnStr = s.buildColumn(column, false)
		} else {
			columnStr = s.buildColumn(column, true)
		}
		switch column.operation {
		case "add":
			sql += "ADD COLUMN " + columnStr
		case "drop":
			sql += "DROP COLUMN " + column.name
		case "alter":
			sql += "ALTER COLUMN " + columnStr
		case "rename":
			sql += "RENAME COLUMN " + column.oldName + " TO " + column.name
		}
	}
	sql += s.buildConstraints()
	sql += ";"
	return sql
}

func (s *Schema) buildAlterMySQL() string {
	sql := "ALTER TABLE " + s.tableName + " "
	for index, column := range s.table.columns {
		columnStr := ""
		if index == len(s.table.columns)-1 {
			columnStr = s.buildColumn(column, false)
		} else {
			columnStr = s.buildColumn(column, true)
		}
		switch column.operation {
		case "add":
			sql += "ADD COLUMN " + columnStr
		case "drop":
			sql += "DROP COLUMN " + column.name
		case "alter":
			sql += "MODIFY COLUMN " + columnStr
		case "rename":
			sql += "RENAME COLUMN " + column.oldName + " TO " + column.name
		}
	}
	sql += s.buildConstraints()
	sql += ";"
	return sql
}

func (s *Schema) buildAlterPostgreSQL() string {
	sql := "ALTER TABLE " + s.tableName + " "
	for index, column := range s.table.columns {
		columnStr := ""
		if index == len(s.table.columns)-1 {
			columnStr = s.buildColumn(column, false)
		} else {
			columnStr = s.buildColumn(column, true)
		}
		switch column.operation {
		case "add":
			sql += "ADD COLUMN " + columnStr
		case "drop":
			sql += "DROP COLUMN " + column.name
		case "alter":
			sql += "ALTER COLUMN " + columnStr
		case "rename":
			sql += "RENAME COLUMN " + column.oldName + " TO " + column.name
		}
	}
	sql += s.buildConstraints()
	sql += ";"
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

func (s *Schema) buildColumn(column *Column, trailingComma bool) string {
	hasCompositePrimaryKey := false

	if column.incrementing && column.table.dialect == DriverSQLite {
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
				"The provided primary columns have been set as unique and the incremental column is set as primary.]", column.name))
		}
		column.primary = true
	}

	sql := "\n" + column.name + " "

	if column.dataType != nil {
		sql += column.dataType.ToString()
	}

	if !column.nullable {
		sql += " NOT NULL"
	}
	if column.defaultValue != nil {
		sql += " DEFAULT " + fmt.Sprintf("%v", column.defaultValue)
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

	// Add trailing comma if trailingComma is true
	if trailingComma {
		return sql + ", "
	}

	return sql
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
				if s.dialect == DriverPostgres {
					prefix = "UNIQUE "
				} else if s.dialect == DriverMySQL {
					prefix = "UNIQUE " + constraint.name + " "
				} else if s.dialect == DriverSQLite {
					prefix = "UNIQUE "
				}
				sql += prefix + "(" + s.buildColumns(constraint.uniqueColumns) + "), "
			}
			if constraint.index != nil {
				sql += "INDEX " + constraint.index.name + " (" + s.buildColumns(constraint.index.columns) + "), "
			}
			if constraint.foreignKey != nil {
				sql += s.buildForeignKey(constraint.foreignKey) + ", "
			}
		case "drop":
			if len(constraint.primaryColumns) > 0 {
				sql += "DROP PRIMARY KEY, "
			}
			if len(constraint.uniqueColumns) > 0 {
				sql += "DROP UNIQUE (" + s.buildColumns(constraint.uniqueColumns) + "), "
			}
			if constraint.index != nil {
				sql += "DROP INDEX " + constraint.index.name + ", "
			}
			if constraint.foreignKey != nil {
				sql += "DROP FOREIGN KEY " + constraint.name + ", "
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

	sql += "FOREIGN KEY (" + s.buildColumns(fk.columns) + ") REFERENCES " + fk.on + "(" + fk.references + ")"
	if fk.onDelete != "" {
		sql += " ON DELETE " + fk.onDelete
	}
	if fk.onUpdate != "" {
		sql += " ON UPDATE " + fk.onUpdate
	}
	return sql
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
