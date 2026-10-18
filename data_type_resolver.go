package migration

import (
	"fmt"
	"strings"
)

const (
	DriverSQLite   = "sqlite"
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"

	ColTypeIncrements    = "increments"
	ColTypeBigIncrements = "bigIncrements"
	ColTypeTinyInt       = "tinyInt"
	ColTypeBool          = "bool"
	ColTypeSmallInt      = "smallInt"
	ColTypeMediumInt     = "mediumInt"
	ColTypeInt           = "int"
	ColTypeBigInt        = "bigInt"
	ColTypeFloat         = "float"
	ColTypeDouble        = "double"
	ColTypeDecimal       = "decimal"
	ColTypeDate          = "date"
	ColTypeDateTime      = "dateTime"
	ColTypeDateTimeTz    = "dateTimeTz"
	ColTypeTime          = "time"
	ColTypeTimestamp     = "timestamp"
	ColTypeTimestampTz   = "timestampTz"
	ColTypeChar          = "char"
	ColTypeVarchar       = "varchar"
	ColTypeText          = "text"
	ColTypeJson          = "json"
	ColTypeTinyText      = "tinyText"
	ColTypeMediumText    = "mediumText"
	ColTypeLongText      = "longText"
	ColTypeBinary        = "binary"
	ColTypeVarBinary     = "varBinary"
	ColTypeBlob          = "blob"
	ColTypeTinyBlob      = "tinyBlob"
	ColTypeMediumBlob    = "mediumBlob"
	ColTypeLongBlob      = "longBlob"
	ColTypeEnum          = "enum"
	ColTypeSet           = "set"
	ColTypeUUID          = "uuid"
	ColTypeULID          = "ulid"
)

// DataType represents a column type
type DataType struct {
	columnName   string
	driver       string
	genericName  string
	sqliteName   string
	mysqlName    string
	postgresName string
	unsigned     bool
	length       uint
	precision    uint
	scale        uint
	prefix       string
	suffix       string

	enumValues []string
}

var dataTypes = []DataType{
	{genericName: ColTypeIncrements, sqliteName: "INTEGER", mysqlName: "INT", postgresName: "SERIAL", unsigned: true},
	{genericName: ColTypeBigIncrements, sqliteName: "INTEGER", mysqlName: "BIGINT", postgresName: "BIGSERIAL", unsigned: true},
	{genericName: ColTypeTinyInt, sqliteName: "TINYINT", mysqlName: "TINYINT", postgresName: "SMALLINT"},
	{genericName: ColTypeBool, sqliteName: "BOOLEAN", mysqlName: "BOOLEAN", postgresName: "BOOLEAN"},
	{genericName: ColTypeSmallInt, sqliteName: "SMALLINT", mysqlName: "SMALLINT", postgresName: "SMALLINT"},
	{genericName: ColTypeMediumInt, sqliteName: "MEDIUMINT", mysqlName: "MEDIUMINT", postgresName: "INTEGER"},
	{genericName: ColTypeInt, sqliteName: "INTEGER", mysqlName: "INT", postgresName: "INTEGER"},
	{genericName: ColTypeBigInt, sqliteName: "BIGINT", mysqlName: "BIGINT", postgresName: "BIGINT"},
	{genericName: ColTypeFloat, sqliteName: "FLOAT", mysqlName: "FLOAT", postgresName: "FLOAT"},
	{genericName: ColTypeDouble, sqliteName: "DOUBLE", mysqlName: "DOUBLE", postgresName: "DOUBLE PRECISION"},
	{genericName: ColTypeDecimal, sqliteName: "DECIMAL", mysqlName: "DECIMAL", postgresName: "DECIMAL"},
	{genericName: ColTypeDate, sqliteName: "DATE", mysqlName: "DATE", postgresName: "DATE"},
	{genericName: ColTypeDateTime, sqliteName: "DATETIME", mysqlName: "DATETIME", postgresName: "TIMESTAMP"},
	{genericName: ColTypeDateTimeTz, sqliteName: "DATETIME", mysqlName: "DATETIME", postgresName: "TIMESTAMP WITH TIME ZONE"},
	{genericName: ColTypeTime, sqliteName: "TIME", mysqlName: "TIME", postgresName: "TIME"},
	{genericName: ColTypeTimestamp, sqliteName: "TIMESTAMP", mysqlName: "TIMESTAMP", postgresName: "TIMESTAMP"},
	{genericName: ColTypeTimestampTz, sqliteName: "TIMESTAMP", mysqlName: "TIMESTAMP", postgresName: "TIMESTAMP WITH TIME ZONE"},
	{genericName: ColTypeChar, sqliteName: "CHAR", mysqlName: "CHAR", postgresName: "CHAR"},
	{genericName: ColTypeVarchar, sqliteName: "VARCHAR", mysqlName: "VARCHAR", postgresName: "VARCHAR"},
	{genericName: ColTypeText, sqliteName: "TEXT", mysqlName: "TEXT", postgresName: "TEXT"},
	// JSONB on Postgres rather than JSON: it stores a parsed form, so it is
	// cheaper to read than to write and it can carry a GIN index, which is
	// the trade a document column in a table almost always wants. The cost is
	// that key order is not preserved and duplicate keys collapse — both
	// properties of JSON objects rather than of this choice.
	//
	// SQLite has no JSON type. It stores JSON as TEXT and its JSON1 functions
	// operate on TEXT, so TEXT is not a fallback here, it is the answer.
	{genericName: ColTypeJson, sqliteName: "TEXT", mysqlName: "JSON", postgresName: "JSONB"},
	{genericName: ColTypeTinyText, sqliteName: "TINYTEXT", mysqlName: "TINYTEXT", postgresName: "TEXT"},
	{genericName: ColTypeMediumText, sqliteName: "MEDIUMTEXT", mysqlName: "MEDIUMTEXT", postgresName: "TEXT"},
	{genericName: ColTypeLongText, sqliteName: "LONGTEXT", mysqlName: "LONGTEXT", postgresName: "TEXT"},
	{genericName: ColTypeBinary, sqliteName: "BINARY", mysqlName: "BINARY", postgresName: "BYTEA"},
	{genericName: ColTypeVarBinary, sqliteName: "VARBINARY", mysqlName: "VARBINARY", postgresName: "BYTEA"},
	{genericName: ColTypeBlob, sqliteName: "BLOB", mysqlName: "BLOB", postgresName: "BYTEA"},
	{genericName: ColTypeTinyBlob, sqliteName: "TINYBLOB", mysqlName: "TINYBLOB", postgresName: "BYTEA"},
	{genericName: ColTypeMediumBlob, sqliteName: "MEDIUMBLOB", mysqlName: "MEDIUMBLOB", postgresName: "BYTEA"},
	{genericName: ColTypeLongBlob, sqliteName: "LONGBLOB", mysqlName: "LONGBLOB", postgresName: "BYTEA"},
	{genericName: ColTypeEnum, sqliteName: "TEXT", mysqlName: "ENUM", postgresName: "TEXT"},
	{genericName: ColTypeSet, sqliteName: "TEXT", mysqlName: "SET", postgresName: "TEXT"},
	{genericName: ColTypeUUID, sqliteName: "TEXT", mysqlName: "CHAR", postgresName: "UUID"},
	{genericName: ColTypeULID, sqliteName: "TEXT", mysqlName: "CHAR", postgresName: "CHAR"},
}

// NewDataType creates a new DataType
func NewDataType(columnName string, name string, dialect string) *DataType {
	for _, dataType := range dataTypes {
		if dataType.genericName == name {
			dataType.SetDialect(dialect)
			dataType.SetColumnName(columnName)
			dataType.AddSuffixes()
			return &dataType
		}
	}

	return nil
}

// WithLength sets the length of the column
func (dataType *DataType) WithLength(length uint) *DataType {
	dataType.length = length
	return dataType
}

// WithPrecision sets the precision of the column
func (dataType *DataType) WithPrecision(precision uint) *DataType {
	dataType.precision = precision
	return dataType
}

// WithScale sets the scale of the column
func (dataType *DataType) WithScale(scale uint) *DataType {
	dataType.scale = scale
	return dataType
}

// WithEnumValues sets the enum values of the column
func (dataType *DataType) WithEnumValues(enumValues []string) *DataType {
	dataType.enumValues = enumValues
	// The CHECK is appended here rather than in AddSuffixes, because
	// AddSuffixes has already run: NewDataType calls it before the caller
	// gets a chance to chain this on, so enumValues was always nil at that
	// point and the constraint was never emitted. A column declared as an
	// enum that accepts any string is worse than no column, since the schema
	// claims it is constrained.
	if dataType.driver == DriverPostgres && len(enumValues) > 0 &&
		(dataType.genericName == ColTypeEnum || dataType.genericName == ColTypeSet) {
		dataType.AppendSufix(fmt.Sprintf("CHECK (%s IN (%s))",
			dataType.columnName, dataType.quotedEnumValues()))
	}
	return dataType
}

// SetDialect sets the driver of the column
func (dataType *DataType) SetDialect(dialect string) {
	dataType.driver = dialect
}

// SetColumnName sets the column name of the column
func (dataType *DataType) SetColumnName(columnName string) {
	dataType.columnName = columnName
}

// ToString returns the string representation of the column type
func (dataType *DataType) ToString() string {
	if dataType.driver == "" {
		panic("Driver not set")
	}

	columnType := ""
	switch dataType.driver {
	case DriverSQLite:
		columnType = dataType.sqliteName
	case DriverMySQL:
		columnType = dataType.mysqlName
	case DriverPostgres:
		columnType = dataType.postgresName
	default:
		panic("Unsupported driver")
	}

	if dataType.columnName == "" {
		panic("Column name not set")
	}

	if dataType.driver == DriverMySQL && dataType.unsigned {
		columnType = columnType + " UNSIGNED"
	}

	if dataType.driver == DriverMySQL && dataType.genericName == ColTypeEnum {
		return fmt.Sprintf("%s(%s)", columnType, dataType.quotedEnumValues())
	}

	if dataType.length > 0 {
		return fmt.Sprintf("%s(%d)", columnType, dataType.length)
	}

	if dataType.precision > 0 && dataType.scale > 0 {
		return fmt.Sprintf("%s(%d,%d)", columnType, dataType.precision, dataType.scale)
	}

	if dataType.precision > 0 {
		return fmt.Sprintf("%s(%d)", columnType, dataType.precision)
	}

	if dataType.scale > 0 {
		return fmt.Sprintf("%s(%d)", columnType, dataType.scale)
	}

	return columnType
}

// autoIncrements reports whether this type is a generated surrogate key.
func (dataType *DataType) autoIncrements() bool {
	return dataType.genericName == ColTypeIncrements ||
		dataType.genericName == ColTypeBigIncrements
}

// AddSuffixes adds suffixes to the column type
func (dataType *DataType) AddSuffixes() *DataType {
	// An unsigned column on Postgres becomes a CHECK, because Postgres has no
	// unsigned integers — except for an auto-increment key, which gets
	// nothing.
	//
	// SERIAL and BIGSERIAL already draw from a sequence that starts at 1 and
	// ascends, so the CHECK is a per-insert validation that cannot fail. The
	// reason to drop it is not the cost, though: it is that one declaration
	// was producing three different schemas. The same BigIncrements("id")
	// gave Postgres a CHECK, MySQL an UNSIGNED column type and SQLite
	// nothing — so a project moving between them did not get the table it
	// wrote, and a Postgres table built by this package could not be matched
	// by one written by hand. For a surrogate key the portable answer is the
	// dialect's own auto-increment type and no extra constraint.
	//
	// This is in practice the only path that reached the CHECK at all, which
	// is worth recording because it is not what it looks like. An explicit
	// UnsignedInt("views") sets the flag through Column.Unsigned, which runs
	// *after* NewDataType has already called AddSuffixes — so the suffix is
	// decided before the column is marked, and no CHECK is ever appended.
	// MySQL is unaffected because its UNSIGNED is read at render time instead.
	// It is the same ordering bug Enum has, and it is left alone here rather
	// than fixed in passing: deciding what an unsigned column should assert
	// on Postgres is a design question — a counter wants >= 0 and the suffix
	// says > 0 — and answering it would add a constraint to every table
	// already built with one of those columns.
	if dataType.driver == DriverPostgres && dataType.unsigned && !dataType.autoIncrements() {
		dataType.AppendSufix(fmt.Sprintf("CHECK (%s > 0)", dataType.columnName))
	}

	// The enum CHECK is appended by WithEnumValues, not here: this runs
	// before the values are supplied, so the condition could never be true.

	return dataType
}

// AppendSufix appends a suffix to the column type
func (dataType *DataType) AppendSufix(suffix string) *DataType {
	dataType.suffix = dataType.suffix + " " + suffix
	return dataType
}

// Precision returns the column's declared precision, or zero when it has none.
func (dataType *DataType) Precision() uint {
	if dataType == nil {
		return 0
	}
	return dataType.precision
}

// quotedEnumValues renders the permitted values as a SQL list.
//
// Both call sites used to interpolate the []string with %s, which is Go's
// slice syntax — `[draft published]`, unquoted and space-separated. On MySQL
// that produced an ENUM type no server would accept; on Postgres it would
// have produced an unparsable CHECK, had the CHECK been reached at all.
//
// A single quote inside a value is doubled, which is SQL's own escape. These
// values come from a migration rather than from a request, so this is not the
// package's injection boundary — but a value like "it's fine" is ordinary
// enough that silently truncating the literal at the apostrophe would be a
// nasty surprise.
func (dataType *DataType) quotedEnumValues() string {
	quoted := make([]string, 0, len(dataType.enumValues))
	for _, value := range dataType.enumValues {
		quoted = append(quoted, "'"+strings.ReplaceAll(value, "'", "''")+"'")
	}
	return strings.Join(quoted, ", ")
}
