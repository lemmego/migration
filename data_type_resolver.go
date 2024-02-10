package migration

type DataTypeResolver interface {
	ResolveType(string) string
}

type DataType struct {
	genericName    string
	sqliteName     string
	mysqlName      string
	postgresqlName string
}

var genericNames = []string{
	"tinyInt", "bool", "smallInt", "mediumInt", "int", "bigInt", "float", "double", "decimal", "date", "dateTime", "time", "timestamp", "char", "varchar", "text", "tinyText", "mediumText", "longText", "binary", "varBinary", "blob", "tinyBlob", "mediumBlob", "longBlob", "enum", "set",
}

var dataTypes = []DataType{
	{genericName: "increments", sqliteName: "UNSIGNED INTEGER", mysqlName: "UNSIGNED INTEGER", postgresqlName: "UNSIGNED INTEGER"},
	{genericName: "tinyInt", sqliteName: "TINYINT", mysqlName: "TINYINT", postgresqlName: "SMALLINT"},
	{genericName: "bool", sqliteName: "BOOLEAN", mysqlName: "BOOLEAN", postgresqlName: "BOOLEAN"},
	{genericName: "smallInt", sqliteName: "SMALLINT", mysqlName: "SMALLINT", postgresqlName: "SMALLINT"},
	{genericName: "mediumInt", sqliteName: "MEDIUMINT", mysqlName: "MEDIUMINT", postgresqlName: "INTEGER"},
	{genericName: "int", sqliteName: "INT", mysqlName: "INT", postgresqlName: "INTEGER"},
	{genericName: "bigInt", sqliteName: "BIGINT", mysqlName: "BIGINT", postgresqlName: "BIGINT"},
	{genericName: "float", sqliteName: "FLOAT", mysqlName: "FLOAT", postgresqlName: "FLOAT"},
	{genericName: "double", sqliteName: "DOUBLE", mysqlName: "DOUBLE", postgresqlName: "DOUBLE PRECISION"},
	{genericName: "decimal", sqliteName: "DECIMAL", mysqlName: "DECIMAL", postgresqlName: "DECIMAL"},
	{genericName: "date", sqliteName: "DATE", mysqlName: "DATE", postgresqlName: "DATE"},
	{genericName: "dateTime", sqliteName: "DATETIME", mysqlName: "DATETIME", postgresqlName: "TIMESTAMP"},
	{genericName: "time", sqliteName: "TIME", mysqlName: "TIME", postgresqlName: "TIME"},
	{genericName: "timestamp", sqliteName: "TIMESTAMP", mysqlName: "TIMESTAMP", postgresqlName: "TIMESTAMP"},
	{genericName: "char", sqliteName: "CHAR", mysqlName: "CHAR", postgresqlName: "CHAR"},
	{genericName: "varchar", sqliteName: "VARCHAR", mysqlName: "VARCHAR", postgresqlName: "VARCHAR"},
	{genericName: "text", sqliteName: "TEXT", mysqlName: "TEXT", postgresqlName: "TEXT"},
	{genericName: "tinyText", sqliteName: "TINYTEXT", mysqlName: "TINYTEXT", postgresqlName: "TEXT"},
	{genericName: "mediumText", sqliteName: "MEDIUMTEXT", mysqlName: "MEDIUMTEXT", postgresqlName: "TEXT"},
	{genericName: "longText", sqliteName: "LONGTEXT", mysqlName: "LONGTEXT", postgresqlName: "TEXT"},
	{genericName: "binary", sqliteName: "BINARY", mysqlName: "BINARY", postgresqlName: "BYTEA"},
	{genericName: "varBinary", sqliteName: "VARBINARY", mysqlName: "VARBINARY", postgresqlName: "BYTEA"},
	{genericName: "blob", sqliteName: "BLOB", mysqlName: "BLOB", postgresqlName: "BYTEA"},
	{genericName: "tinyBlob", sqliteName: "TINYBLOB", mysqlName: "TINYBLOB", postgresqlName: "BYTEA"},
	{genericName: "mediumBlob", sqliteName: "MEDIUMBLOB", mysqlName: "MEDIUMBLOB", postgresqlName: "BYTEA"},
	{genericName: "longBlob", sqliteName: "LONGBLOB", mysqlName: "LONGBLOB", postgresqlName: "BYTEA"},
	{genericName: "enum", sqliteName: "ENUM", mysqlName: "ENUM", postgresqlName: "TEXT"},
	{genericName: "set", sqliteName: "SET", mysqlName: "SET", postgresqlName: "TEXT"},
}

func ResolveType(name string, dialect string) string {
	for _, dataType := range dataTypes {
		if dataType.genericName == name {
			switch dialect {
			case "sqlite":
				return dataType.sqliteName
			case "mysql":
				return dataType.mysqlName
			case "postgresql":
				return dataType.postgresqlName
			case "postgres":
				return dataType.postgresqlName
			default:
				panic("Unsupported dialect")
			}
		}
	}

	return name
}
