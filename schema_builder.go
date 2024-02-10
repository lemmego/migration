package migration

import "fmt"

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
	columns    []string
	references string
	onDelete   string
	onUpdate   string
}

type Column struct {
	table        *Table
	name         string
	dataType     string
	nullable     bool
	defaultValue any
	unique       bool
	primary      bool
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
	tableName string
	operation string
	table     *Table
}

func NewSchema() *Schema {
	return &Schema{}
}

func (s *Schema) Create(tableName string, tableFunc func(t *Table) error) *Schema {
	s.tableName = tableName
	s.operation = "create"
	t := &Table{name: tableName}
	s.table = t
	if err := tableFunc(t); err != nil {
		panic(err)
	}
	return s
}

func (s *Schema) Table(tableName string, tableFunc func(t *Table) error) *Schema {
	s.tableName = tableName
	s.operation = "alter"
	t := &Table{name: tableName}
	s.table = t
	if err := tableFunc(t); err != nil {
		panic(err)
	}
	return s
}

func (s *Schema) Drop(tableName string) *Schema {
	s.tableName = tableName
	s.operation = "drop"
	s.table = &Table{name: tableName}
	return s
}

func (t *Table) tinyInt(name string) *Column {
	c := t.AddColumn(name, "tinyInt")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) bool(name string) *Column {
	c := t.AddColumn(name, "bool")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) smallInt(name string) *Column {
	c := t.AddColumn(name, "smallInt")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) mediumInt(name string) *Column {
	c := t.AddColumn(name, "mediumInt")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) int(name string) *Column {
	c := t.AddColumn(name, "int")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) bigInt(name string) *Column {
	c := t.AddColumn(name, "bigInt")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) float(name string) *Column {
	c := t.AddColumn(name, "float")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) double(name string) *Column {
	c := t.AddColumn(name, "double")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) decimal(name string) *Column {
	c := t.AddColumn(name, "decimal")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) date(name string) *Column {
	c := t.AddColumn(name, "date")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) dateTime(name string) *Column {
	c := t.AddColumn(name, "dateTime")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) time(name string) *Column {
	c := t.AddColumn(name, "time")
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) AddColumn(name string, dataType string) *Column {
	c := &Column{
		name:      name,
		table:     t,
		dataType:  ResolveType(dataType, t.dialect),
		operation: "add",
	}
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) DropColumn(name string) *Column {
	c := &Column{
		name:      name,
		operation: "drop",
	}
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) AlterColumn(name string, dataType string) *Column {
	c := &Column{
		name:      name,
		operation: "alter",
	}
	t.columns = append(t.columns, c)
	return c
}

func (t *Table) RenameColumn(oldName string, newName string) *Column {
	c := &Column{
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

func (t *Table) ForeignKey(columns []string, references string, onDelete string, onUpdate string) {
	fk := &foreignKey{
		columns:    columns,
		references: references,
		onDelete:   onDelete,
		onUpdate:   onUpdate,
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
}

func (t *Table) DropForeignKey(name string) {
	c := &constraint{
		name:      name,
		operation: "drop",
	}
	t.constraints = append(t.constraints, c)
}

func (c *Column) Type(dataType string) *Column {
	c.dataType = dataType
	return c
}

func (c *Column) Nullable() *Column {
	c.nullable = true
	return c
}

func (c *Column) DefaultValue(defaultValue any) *Column {
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
	sql := "create table " + s.tableName + " (\n"
	for i, c := range s.table.columns {
		sql += c.name + " " + c.dataType
		if !c.nullable {
			sql += " not null"
		}
		if c.primary {
			sql += " primary key"
		}
		if c.unique {
			sql += " unique"
		}
		if c.defaultValue != nil {
			sql += " default " + s.toSQL(c.defaultValue)
		}
		if i < len(s.table.columns)-1 {
			sql += ",\n"
		}
	}
	if len(s.table.constraints) > 0 {
		sql += ",\n"
	}
	for i, c := range s.table.constraints {
		if c.primaryColumns != nil {
			sql += "constraint " + c.name + " primary key (" + s.toSQL(c.primaryColumns) + ")"
		}
		if c.index != nil {
			sql += "create index " + c.index.name + " on " + s.tableName + " (" + s.toSQL(c.index.columns) + ")"
		}
		if c.uniqueColumns != nil {
			sql += "constraint " + c.name + " unique (" + s.toSQL(c.uniqueColumns) + ")"
		}
		if c.foreignKey != nil {
			sql += "constraint " + c.name + " foreign key (" + s.toSQL(c.foreignKey.columns) + ") references " + c.foreignKey.references
			if c.foreignKey.onDelete != "" {
				sql += " on delete " + c.foreignKey.onDelete
			}
			if c.foreignKey.onUpdate != "" {
				sql += " on update " + c.foreignKey.onUpdate
			}
		}
		if i < len(s.table.constraints)-1 {
			sql += ",\n"
		}
	}
	sql += "\n);"
	return sql
}

func (s *Schema) buildAlter() string {
	sql := "alter table " + s.tableName + " "
	for _, c := range s.table.columns {
		switch c.operation {
		case "add":
			sql += "add column " + c.name + " " + c.dataType
			if !c.nullable {
				sql += " not null"
			}
			if c.primary {
				sql += " primary key"
			}
			if c.unique {
				sql += " unique"
			}
			if c.defaultValue != nil {
				sql += " default " + s.toSQL(c.defaultValue)
			}
			if len(c.foreignKeys) > 0 {
				for _, fk := range c.foreignKeys {
					sql += "constraint " + s.tableName + "_" + c.name + "_fkey foreign key (" + s.toSQL(fk.columns) + ") references " + fk.references
					if fk.onDelete != "" {
						sql += " on delete " + fk.onDelete
					}
					if fk.onUpdate != "" {
						sql += " on update " + fk.onUpdate
					}
				}
			}
		case "drop":
			sql += "drop column " + c.name
		case "alter":
			sql += "alter column " + c.name + " type " + c.dataType
		case "rename":
			sql += "rename column " + c.oldName + " to " + c.name
		}
	}
	if len(s.table.columns) > 0 && len(s.table.constraints) > 0 {
		sql += ",\n"
	}
	for i, c := range s.table.constraints {
		if c.primaryColumns != nil {
			if c.operation == "add" {
				sql += "add constraint " + c.name + " primary key (" + s.toSQL(c.primaryColumns) + ")"
			} else if c.operation == "drop" {
				sql += "drop constraint " + c.name
			}
		}
		if c.index != nil {
			if c.operation == "add" {
				sql += "create index " + c.index.name + " on " + s.tableName + " (" + s.toSQL(c.index.columns) + ")"
			} else if c.operation == "drop" {
				sql += "drop index " + c.index.name
			}
		}
		if c.uniqueColumns != nil {
			if c.operation == "add" {
				sql += "add constraint " + c.name + " unique (" + s.toSQL(c.uniqueColumns) + ")"
			} else if c.operation == "drop" {
				sql += "drop constraint " + c.name
			}
		}
		if c.foreignKey != nil {
			if c.operation == "add" {
				sql += "add constraint " + c.name + " foreign key (" + s.toSQL(c.foreignKey.columns) + ") references " + c.foreignKey.references
				if c.foreignKey.onDelete != "" {
					sql += " on delete " + c.foreignKey.onDelete
				}
				if c.foreignKey.onUpdate != "" {
					sql += " on update " + c.foreignKey.onUpdate
				}
			} else if c.operation == "drop" {
				sql += "drop constraint " + c.name
			}
		}
		if i < len(s.table.constraints)-1 {
			sql += ",\n"
		}
	}
	return sql
}

func (s *Schema) buildDrop() string {
	return "drop table if exists " + s.tableName + " cascade;"
}

func (s *Schema) toSQL(value any) string {
	switch v := value.(type) {
	case string:
		return "'" + v + "'"
	case int:
		return fmt.Sprintf("%d", v)
	case []string:
		sql := ""
		for i, s := range v {
			sql += s
			if i < len(v)-1 {
				sql += ", "
			}
		}
		return sql
	}
	return ""
}
