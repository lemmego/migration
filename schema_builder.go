package migration

import (
	"fmt"
	"os"
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
	table      *Table
	columns    []string
	references string
	on         string
	onDelete   string
	onUpdate   string
}

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
	foreignKeys  []*foreignKey
}

type Table struct {
	dialect     string
	name        string
	columns     []*Column
	constraints []*constraint
	operation   string
}

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

func NewSchema() *Schema {
	return &Schema{dialect: determineDialect()}
}

func (s *Schema) Create(tableName string, tableFunc func(t *Table) error) *Schema {
	s.tableName = tableName
	s.operation = "create"
	t := &Table{name: tableName, dialect: s.dialect}
	s.table = t
	if err := tableFunc(t); err != nil {
		panic(err)
	}
	return s
}

func (s *Schema) Table(tableName string, tableFunc func(t *Table) error) *Schema {
	s.tableName = tableName
	s.operation = "alter"
	t := &Table{name: tableName, dialect: s.dialect}
	s.table = t
	if err := tableFunc(t); err != nil {
		panic(err)
	}
	return s
}

func (s *Schema) Drop(tableName string) *Schema {
	s.tableName = tableName
	s.operation = "drop"
	s.table = &Table{name: tableName, dialect: s.dialect}
	return s
}

func (t *Table) HasConstraints() bool {
	return len(t.constraints) > 0
}

func (t *Table) Increments(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeIncrements, t.dialect))
	c.incrementing = true
	return c
}

func (t *Table) BigIncrements(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeBigIncrements, t.dialect))
	c.incrementing = true
	return c
}

func (t *Table) Char(name string, length uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeChar, t.dialect).WithLength(length))
	return c
}

func (t *Table) String(name string, length uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeVarchar, t.dialect).WithLength(length))
	return c
}

func (t *Table) Text(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeText, t.dialect))
	return c
}

func (t *Table) TinyInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeTinyInt, t.dialect))
	return c
}

func (t *Table) Boolean(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeBool, t.dialect))
	return c
}

func (t *Table) SmallInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeSmallInt, t.dialect))
	return c
}

func (t *Table) MediumInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeMediumInt, t.dialect))
	return c
}

func (t *Table) Int(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeInt, t.dialect))
	return c
}

func (t *Table) BigInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeBigInt, t.dialect))
	return c
}

func (t *Table) UnsignedBigInt(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeUnsignedBigInt, t.dialect))
	return c
}

func (t *Table) Float(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeFloat, t.dialect))
	return c
}

func (t *Table) Double(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeDouble, t.dialect))
	return c
}

func (t *Table) Decimal(name string, precision uint, scale uint) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeDecimal, t.dialect).WithPrecision(precision).WithScale(scale))
	return c
}

func (t *Table) Date(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeDate, t.dialect))
	return c
}

func (t *Table) DateTime(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeDateTime, t.dialect))
	return c
}

func (t *Table) Time(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeTime, t.dialect))
	return c
}

func (t *Table) Timestamp(name string) *Column {
	c := t.AddColumn(name, NewDataType(name, ColTypeTimestamp, t.dialect))
	return c
}

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

func (t *Table) DropColumn(name string) *Column {
	c := &Column{
		table:     t,
		name:      name,
		operation: "drop",
	}
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) AlterColumn(name string, dataType string) *Column {
	c := &Column{
		table:     t,
		name:      name,
		operation: "alter",
	}
	t.columns = append(t.columns, c)
	return c
}

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

func (t *Table) DropIndex(indexName string) {
	c := &constraint{
		name:      indexName,
		operation: "drop",
	}
	t.constraints = append(t.constraints, c)
}

func (t *Table) Unique(columns ...string) {
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

func (t *Table) DropUnique(name string) {
	c := &constraint{
		name:      name,
		operation: "drop",
	}
	t.constraints = append(t.constraints, c)
}

func (t *Table) ForeignKey(columns ...string) *foreignKey {
	fk := &foreignKey{
		table:   t,
		columns: columns,
	}
	fkName := ""
	for _, column := range columns {
		fkName += column + "_"
	}
	c := &constraint{
		name:       fkName + "_fkey",
		operation:  "add",
		foreignKey: fk,
	}
	t.constraints = append(t.constraints, c)
	return fk
}

func (f *foreignKey) References(table string) *foreignKey {
	f.references = table
	return f
}

func (f *foreignKey) On(on string) *foreignKey {
	f.on = on
	return f
}

func (f *foreignKey) OnDelete(onDelete string) *foreignKey {
	f.onDelete = onDelete
	return f
}

func (f *foreignKey) OnUpdate(onUpdate string) *foreignKey {
	f.onUpdate = onUpdate
	return f
}

func (t *Table) DropForeignKey(name string) {
	c := &constraint{
		name:      name,
		operation: "drop",
	}
	t.constraints = append(t.constraints, c)
}

func (c *Column) Type(dataType *DataType) *Column {
	c.dataType = dataType
	return c
}

func (c *Column) Nullable() *Column {
	c.nullable = true
	return c
}

func (c *Column) NotNull() *Column {
	c.nullable = false
	return c
}

func (c *Column) Default(defaultValue any) *Column {
	c.defaultValue = defaultValue
	return c
}

func (c *Column) Unique() *Column {
	c.unique = true
	return c
}

func (c *Column) Primary() *Column {
	c.primary = true
	return c
}

func (c *Column) ForeignKey(columns []string, references string, onDelete string, onUpdate string) *Column {
	fk := &foreignKey{
		columns:    columns,
		references: references,
		onDelete:   onDelete,
		onUpdate:   onUpdate,
	}
	c.foreignKeys = append(c.foreignKeys, fk)
	return c
}

func (c *Column) Change() {
	c.operation = "alter"
}

func (c *Column) Done() *Table {
	return c.table
}

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
	case DialectSQLite:
		return s.buildCreateSQLite()
	case DialectMySQL:
		return s.buildCreateMySQL()
	case DialectPostgres:
		return s.buildCreatePostgreSQL()
	}
	return ""
}

func (s *Schema) buildAlter() string {
	switch s.dialect {
	case DialectSQLite:
		return s.buildAlterSQLite()
	case DialectMySQL:
		return s.buildAlterMySQL()
	case DialectPostgres:
		return s.buildAlterPostgreSQL()
	}
	return ""
}

func (s *Schema) buildDrop() string {
	switch s.dialect {
	case DialectSQLite:
		return s.buildDropSQLite()
	case DialectMySQL:
		return s.buildDropMySQL()
	case DialectPostgres:
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
	if column.table.dialect == DialectSQLite && column.incrementing {
		sql += " AUTOINCREMENT"
	}
	if column.table.dialect == DialectMySQL && column.incrementing {
		sql += " AUTO_INCREMENT"
	}
	if len(column.foreignKeys) > 0 {
		for _, fk := range column.foreignKeys {
			sql += ", " + s.buildForeignKey(fk)
		}
	}

	if column.dataType != nil && column.table.dialect == DialectPostgres && (column.dataType.genericName == ColTypeIncrements || column.dataType.genericName == ColTypeBigIncrements) {
		sql += " CHECK (" + column.name + " > 0)"
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
				if s.dialect == DialectPostgres {
					prefix = "UNIQUE "
				} else if s.dialect == DialectMySQL {
					prefix = "UNIQUE " + constraint.name + " "
				} else if s.dialect == DialectSQLite {
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
	sql := "\nFOREIGN KEY (" + s.buildColumns(fk.columns) + ") REFERENCES " + fk.on + "(" + fk.references + ")"
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

func (s *Schema) ToString() string {
	return s.Build()
}

func (s *Schema) ToSQL() string {
	return s.Build()
}

func (s *Schema) ToDDL() string {
	return s.Build()
}

func (s *Schema) ToDialect(dialect string) string {
	s.dialect = dialect
	return s.Build()
}

func (s *Schema) ToDialectSQL(dialect string) string {
	s.dialect = dialect
	return s.Build()
}

func (s *Schema) ToDialectDDL(dialect string) string {
	s.dialect = dialect
	return s.Build()
}

func (s *Schema) ToPostgreSQL() string {
	s.dialect = "postgresql"
	return s.Build()
}

func (s *Schema) ToPostgreSQLSQL() string {
	s.dialect = "postgresql"
	return s.Build()
}
