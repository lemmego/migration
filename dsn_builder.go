package migration

import (
	"errors"
	"log"
	"regexp"
	"slices"
	"strings"
)

var (
	supportedDialects     = []string{"mysql", "postgres", "sqlite"}
	ErrUnsupportedDialect = errors.New("unsupported dialect")
)

type DsnBuilder struct {
	Dialect  string
	Host     string
	Port     string
	Username string
	Password string
	Name     string
	Params   string
}

type DSN struct {
	dialect  string
	host     string
	port     string
	username string
	password string
	name     string
	params   string
}

func (dsn *DSN) ToString() (string, error) {
	switch dsn.dialect {
	case DialectSQLite:
		return dsn.GetSqliteDSN(), nil
	case DialectMySQL:
		return dsn.GetMysqlDSN(), nil
	case DialectPostgres:
		return dsn.GetPostgresDSN(), nil
	// case "mssql":
	// 	return dsn.GetMssqlDSN(), nil
	default:
		return "", ErrUnsupportedDialect
	}
}

func (dsnBuilder *DsnBuilder) ToString() (string, error) {
	dsn, err := dsnBuilder.Build()
	if err != nil {
		return "", err
	}

	return dsn.ToString()
}
func NewDsnBuilder() *DsnBuilder {
	return &DsnBuilder{}
}

func (d *DsnBuilder) SetHost(host string) *DsnBuilder {
	if host == "" {
		log.Fatal("DB Host is required")
	}

	d.Host = host
	return d
}

func (d *DsnBuilder) SetPort(port string) *DsnBuilder {
	d.Port = port
	return d
}

func (d *DsnBuilder) SetUsername(username string) *DsnBuilder {
	if username == "" {
		log.Fatal("DB Username is required")
	}

	d.Username = username
	return d
}

func (d *DsnBuilder) SetPassword(password string) *DsnBuilder {
	d.Password = password
	return d
}

func (d *DsnBuilder) SetName(name string) *DsnBuilder {
	if name == "" {
		log.Fatal("DB name is required")
	}

	d.Name = name
	return d
}

func (d *DsnBuilder) SetDialect(dialect string) *DsnBuilder {
	d.Dialect = dialect
	return d
}

func (d *DsnBuilder) SetParams(params string) *DsnBuilder {
	d.Params = params
	return d
}

// Make sure the params field conforms to this format: param1=value1&paramN=valueN
// Or, (?:[a-zA-Z0-9]+=[a-zA-Z0-9]+)(?:&[a-zA-Z0-9]+=[a-zA-Z0-9]+)*
func (d *DsnBuilder) validateParams(params string) error {
	if regexp.MustCompile(`^(?:[a-zA-Z0-9]+=[a-zA-Z0-9]+)(?:&[a-zA-Z0-9]+=[a-zA-Z0-9]+)*$`).MatchString(params) {
		return nil
	}

	return errors.New("invalid params format")
}

func (d *DsnBuilder) Build() (*DSN, error) {
	dialect := strings.ToLower(d.Dialect)

	if d.Dialect == "" {
		return nil, errors.New("Dialect is required")
	}

	if !slices.Contains(supportedDialects, dialect) {
		return nil, ErrUnsupportedDialect
	}

	if dialect != "sqlite" && d.Host == "" {
		return nil, errors.New("DB Host is required")
	}

	if dialect != "sqlite" && d.Username == "" {
		return nil, errors.New("DB Username is required")
	}

	if d.Name == "" {
		return nil, errors.New("DB name is required")
	}

	if d.Dialect == "mysql" && d.Port == "" {
		d.Port = "3306"
	}

	if d.Dialect == "postgres" && d.Port == "" {
		d.Port = "5432"
	}

	if d.Dialect == "mssql" && d.Port == "" {
		d.Port = "1433"
	}

	if d.Dialect == "sqlite" {
		d.Host = d.Name
	}

	d.validateParams(d.Params)

	if d.Dialect == "mysql" || d.Dialect == "mssql" {
		d.Params = "?" + d.Params
	}

	if d.Dialect == "postgres" {
		split := strings.Split(d.Params, "&")
		d.Params = " " + strings.Join(split, " ")
	}

	return &DSN{
		dialect:  dialect,
		host:     d.Host,
		port:     d.Port,
		username: d.Username,
		password: d.Password,
		name:     d.Name,
		params:   d.Params,
	}, nil
}

func (d *DSN) GetMysqlDSN() string {
	return d.username + ":" + d.password + "@tcp(" + d.host + ":" + string(d.port) + ")/" + d.name + d.params
}

func (d *DSN) GetPostgresDSN() string {
	hostStr := ""
	portStr := ""
	userStr := ""
	passStr := ""
	dbStr := ""
	paramsStr := ""

	if d.host != "" {
		hostStr = "host=" + d.host
	}

	if d.port != "" {
		portStr = " port=" + d.port
	}

	if d.username != "" {
		userStr = " user=" + d.username
	}

	if d.password != "" {
		passStr = " password=" + d.password
	}

	if d.name != "" {
		dbStr = " dbname=" + d.name
	}

	if d.params != "" {
		paramsStr = d.params
	}

	return hostStr + portStr + userStr + passStr + dbStr + paramsStr
}

func (d *DSN) GetMssqlDSN() string {
	return "sqlserver://" + d.username + ":" + d.password + "@" + d.host + ":" + string(d.port) + "?database=" + d.name
}

func (d *DSN) GetSqliteDSN() string {
	return d.name
}
