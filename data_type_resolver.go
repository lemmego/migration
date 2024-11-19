package migration

import "fmt"

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
		return fmt.Sprintf("%s(%s)", columnType, dataType.enumValues)
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

// AddSuffixes adds suffixes to the column type
func (dataType *DataType) AddSuffixes() *DataType {
	// If the driver is postgres and the column type is unsigned, add a check constraint
	if dataType.driver == DriverPostgres && dataType.unsigned {
		dataType.AppendSufix(fmt.Sprintf("CHECK (%s > 0)", dataType.columnName))
	}

	// If the driver is postgres and the column type is enum or set, add a check constraint
	if dataType.driver == DriverPostgres && len(dataType.enumValues) > 0 && (dataType.genericName == ColTypeEnum || dataType.genericName == ColTypeSet) {
		dataType.AppendSufix(fmt.Sprintf("CHECK (%s IN (%s))", dataType.columnName, dataType.enumValues))
	}

	return dataType
}

// AppendSufix appends a suffix to the column type
func (dataType *DataType) AppendSufix(suffix string) *DataType {
	dataType.suffix = dataType.suffix + " " + suffix
	return dataType
}
