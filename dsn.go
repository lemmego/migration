package migration

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

var (
	supportedDialects     = []string{DriverSQLite, DriverMySQL, DriverPostgres /*, "mssql"*/}
	ErrUnsupportedDialect = errors.New("unsupported driver")
)

// DataSource holds the necessary fields for a DSN (data source name)
type DataSource struct {
	Driver   string
	Host     string
	Port     string
	Username string
	Password string
	Name     string
	Params   string
}

// String returns the string representation of the data source
func (ds *DataSource) String() (string, error) {
	dialect := strings.ToLower(ds.Driver)

	if ds.Driver == "" {
		return "", errors.New("driver is required")
	}

	if !slices.Contains(supportedDialects, dialect) {
		return "", ErrUnsupportedDialect
	}

	if dialect != DriverSQLite && ds.Host == "" {
		return "", errors.New("DB Host is required")
	}

	if dialect != DriverSQLite && ds.Username == "" {
		return "", errors.New("DB Username is required")
	}

	if ds.Name == "" {
		return "", errors.New("DB Name is required")
	}

	if ds.Driver == DriverMySQL && ds.Port == "" {
		ds.Port = "3306"
	}

	if ds.Driver == DriverPostgres && ds.Port == "" {
		ds.Port = "5432"
	}

	// if d.Driver == "mssql" && d.Port == "" {
	// 	d.Port = "1433"
	// }

	if ds.Driver == DriverSQLite {
		ds.Host = ds.Name
	}

	ds.validateParams(ds.Params)

	if ds.Driver == DriverMySQL /*|| d.Driver == "mssql"*/ {
		ds.Params = "?" + ds.Params
	}

	if ds.Driver == DriverPostgres {
		split := strings.Split(ds.Params, "&")
		ds.Params = " " + strings.Join(split, " ")
	}

	switch ds.Driver {
	case DriverSQLite:
		return ds.getSqliteDSN(), nil
	case DriverMySQL:
		return ds.getMysqlDSN(), nil
	case DriverPostgres:
		return ds.getPostgresDSN(), nil
	// case "mssql":
	// 	return dsn.getMssqlDSN(), nil
	default:
		return "", ErrUnsupportedDialect
	}
}

// Make sure the params field conforms to this format: param1=value1&paramN=valueN
func (d *DataSource) validateParams(params string) error {
	if regexp.MustCompile(`^(?:[a-zA-Z0-9]+=[a-zA-Z0-9]+)(?:&[a-zA-Z0-9]+=[a-zA-Z0-9]+)*$`).MatchString(params) {
		return nil
	}

	return errors.New("invalid params format")
}

// Example: file::memory:?cache=shared
func (d *DataSource) getSqliteDSN() string {
	return "file:" + d.Name + "?" + d.Params
}

// Example: root:password@tcp(localhost:3306)/test?parseTime=true
func (d *DataSource) getMysqlDSN() string {
	return d.Username + ":" + d.Password + "@tcp(" + d.Host + ":" + string(d.Port) + ")/" + d.Name + d.Params
}

// Example: host=localhost port=5432 user=root password=password dbname=test sslmode=disable
func (d *DataSource) getPostgresDSN() string {
	hostStr := ""
	portStr := ""
	userStr := ""
	passStr := ""
	dbStr := ""
	paramsStr := ""

	if d.Host != "" {
		hostStr = "host=" + d.Host
	}

	if d.Port != "" {
		portStr = " port=" + d.Port
	}

	if d.Username != "" {
		userStr = " user=" + d.Username
	}

	if d.Password != "" {
		passStr = " password=" + d.Password
	}

	if d.Name != "" {
		dbStr = " dbname=" + d.Name
	}

	if d.Params != "" {
		paramsStr = d.Params
	}

	return hostStr + portStr + userStr + passStr + dbStr + paramsStr
}

// func (d *DSN) getMssqlDSN() string {
// 	return "sqlserver://" + d.username + ":" + d.password + "@" + d.host + ":" + string(d.port) + "?database=" + d.name
// }
