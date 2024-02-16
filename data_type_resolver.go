package migration

import "fmt"

type DataType struct {
	columnName     string
	dialect        string
	genericName    string
	sqliteName     string
	mysqlName      string
	postgresqlName string
	length         uint
	precision      uint
	scale          uint
	bigIncrements  bool
	increments     bool
}

const (
	DialectSQLite   = "sqlite"
	DialectMySQL    = "mysql"
	DialectPostgres = "postgres"

	ColTypeIncrements     = "increments"
	ColTypeBigIncrements  = "bigIncrements"
	ColTypeTinyInt        = "tinyInt"
	ColTypeBool           = "bool"
	ColTypeSmallInt       = "smallInt"
	ColTypeMediumInt      = "mediumInt"
	ColTypeInt            = "int"
	ColTypeBigInt         = "bigInt"
	ColTypeUnsignedBigInt = "unsignedBigInt"
	ColTypeFloat          = "float"
	ColTypeDouble         = "double"
	ColTypeDecimal        = "decimal"
	ColTypeDate           = "date"
	ColTypeDateTime       = "dateTime"
	ColTypeTime           = "time"
	ColTypeTimestamp      = "timestamp"
	ColTypeChar           = "char"
	ColTypeVarchar        = "varchar"
	ColTypeText           = "text"
	ColTypeTinyText       = "tinyText"
	ColTypeMediumText     = "mediumText"
	ColTypeLongText       = "longText"
	ColTypeBinary         = "binary"
	ColTypeVarBinary      = "varBinary"
	ColTypeBlob           = "blob"
	ColTypeTinyBlob       = "tinyBlob"
	ColTypeMediumBlob     = "mediumBlob"
	ColTypeLongBlob       = "longBlob"
	ColTypeEnum           = "enum"
	ColTypeSet            = "set"
)

var dataTypes = []DataType{
	{genericName: ColTypeIncrements, sqliteName: "INTEGER", mysqlName: "INT UNSIGNED", postgresqlName: "SERIAL"},
	{genericName: ColTypeBigIncrements, sqliteName: "INTEGER", mysqlName: "BIGINT UNSIGNED", postgresqlName: "BIGSERIAL"},
	{genericName: ColTypeTinyInt, sqliteName: "TINYINT", mysqlName: "TINYINT", postgresqlName: "SMALLINT"},
	{genericName: ColTypeBool, sqliteName: "BOOLEAN", mysqlName: "BOOLEAN", postgresqlName: "BOOLEAN"},
	{genericName: ColTypeSmallInt, sqliteName: "SMALLINT", mysqlName: "SMALLINT", postgresqlName: "SMALLINT"},
	{genericName: ColTypeMediumInt, sqliteName: "MEDIUMINT", mysqlName: "MEDIUMINT", postgresqlName: "INTEGER"},
	{genericName: ColTypeInt, sqliteName: "INT", mysqlName: "INT", postgresqlName: "INTEGER"},
	{genericName: ColTypeBigInt, sqliteName: "BIGINT", mysqlName: "BIGINT", postgresqlName: "BIGINT"},
	{genericName: ColTypeUnsignedBigInt, sqliteName: "BIGINT", mysqlName: "BIGINT UNSIGNED", postgresqlName: "BIGINT"},
	{genericName: ColTypeFloat, sqliteName: "FLOAT", mysqlName: "FLOAT", postgresqlName: "FLOAT"},
	{genericName: ColTypeDouble, sqliteName: "DOUBLE", mysqlName: "DOUBLE", postgresqlName: "DOUBLE PRECISION"},
	{genericName: ColTypeDecimal, sqliteName: "DECIMAL", mysqlName: "DECIMAL", postgresqlName: "DECIMAL"},
	{genericName: ColTypeDate, sqliteName: "DATE", mysqlName: "DATE", postgresqlName: "DATE"},
	{genericName: ColTypeDateTime, sqliteName: "DATETIME", mysqlName: "DATETIME", postgresqlName: "TIMESTAMP"},
	{genericName: ColTypeTime, sqliteName: "TIME", mysqlName: "TIME", postgresqlName: "TIME"},
	{genericName: ColTypeTimestamp, sqliteName: "TIMESTAMP", mysqlName: "TIMESTAMP", postgresqlName: "TIMESTAMP"},
	{genericName: ColTypeChar, sqliteName: "CHAR", mysqlName: "CHAR", postgresqlName: "CHAR"},
	{genericName: ColTypeVarchar, sqliteName: "VARCHAR", mysqlName: "VARCHAR", postgresqlName: "VARCHAR"},
	{genericName: ColTypeText, sqliteName: "TEXT", mysqlName: "TEXT", postgresqlName: "TEXT"},
	{genericName: ColTypeTinyText, sqliteName: "TINYTEXT", mysqlName: "TINYTEXT", postgresqlName: "TEXT"},
	{genericName: ColTypeMediumText, sqliteName: "MEDIUMTEXT", mysqlName: "MEDIUMTEXT", postgresqlName: "TEXT"},
	{genericName: ColTypeLongText, sqliteName: "LONGTEXT", mysqlName: "LONGTEXT", postgresqlName: "TEXT"},
	{genericName: ColTypeBinary, sqliteName: "BINARY", mysqlName: "BINARY", postgresqlName: "BYTEA"},
	{genericName: ColTypeVarBinary, sqliteName: "VARBINARY", mysqlName: "VARBINARY", postgresqlName: "BYTEA"},
	{genericName: ColTypeBlob, sqliteName: "BLOB", mysqlName: "BLOB", postgresqlName: "BYTEA"},
	{genericName: ColTypeTinyBlob, sqliteName: "TINYBLOB", mysqlName: "TINYBLOB", postgresqlName: "BYTEA"},
	{genericName: ColTypeMediumBlob, sqliteName: "MEDIUMBLOB", mysqlName: "MEDIUMBLOB", postgresqlName: "BYTEA"},
	{genericName: ColTypeLongBlob, sqliteName: "LONGBLOB", mysqlName: "LONGBLOB", postgresqlName: "BYTEA"},
	{genericName: ColTypeEnum, sqliteName: "ENUM", mysqlName: "ENUM", postgresqlName: "TEXT"},
	{genericName: ColTypeSet, sqliteName: "SET", mysqlName: "SET", postgresqlName: "TEXT"},
}

func NewDataType(columnName string, name string, dialect string) *DataType {
	for _, dataType := range dataTypes {
		if dataType.genericName == name {
			dataType.SetDialet(dialect)
			dataType.SetColumnName(columnName)
			return &dataType
		}
	}

	return nil
}

func (dataType *DataType) WithLength(length uint) *DataType {
	dataType.length = length
	return dataType
}

func (dataType *DataType) WithPrecision(precision uint) *DataType {
	dataType.precision = precision
	return dataType
}

func (dataType *DataType) WithScale(scale uint) *DataType {
	dataType.scale = scale
	return dataType
}

func (dataType *DataType) SetDialet(dialect string) {
	dataType.dialect = dialect
}

func (dataType *DataType) SetColumnName(columnName string) {
	dataType.columnName = columnName
}

func (dataType *DataType) ToString() string {
	if dataType.dialect == "" {
		panic("Dialect not set")
	}

	columnType := ""
	switch dataType.dialect {
	case DialectSQLite:
		columnType = dataType.sqliteName
	case DialectMySQL:
		columnType = dataType.mysqlName
	case DialectPostgres:
		columnType = dataType.postgresqlName
	default:
		panic("Unsupported dialect")
	}

	if dataType.columnName == "" {
		panic("Column name not set")
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

	return columnType
}

func (dataType *DataType) Increments() *DataType {
	dataType.increments = true
	return dataType
}

func (dataType *DataType) BigIncrements() *DataType {
	dataType.bigIncrements = true
	return dataType
}
