package migration

import (
	"log"
	"regexp"
	"slices"
	"strings"
)

var supportedDialects = []string{"mysql", "postgres", "sqlite", "mssql"}

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

func NewDsnBuilder(dialect string) *DsnBuilder {
	if !slices.Contains(supportedDialects, strings.ToLower(dialect)) {
		log.Fatal("Unsupported dialect")
	}
	return &DsnBuilder{Dialect: dialect}
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

func (d *DsnBuilder) SetParams(params string) *DsnBuilder {
	d.validateParams(params)

	if d.Dialect == "mysql" || d.Dialect == "mssql" {
		d.Params = "?" + params
	}

	if d.Dialect == "postgres" {
		split := strings.Split(params, "&")
		d.Params = " " + strings.Join(split, " ")
	}
	return d
}

// Make sure the params field conforms to this format: param1=value1&paramN=valueN
// Or, (?:[a-zA-Z0-9]+=[a-zA-Z0-9]+)(?:&[a-zA-Z0-9]+=[a-zA-Z0-9]+)*
func (d *DsnBuilder) validateParams(params string) {
	if regexp.MustCompile(`^(?:[a-zA-Z0-9]+=[a-zA-Z0-9]+)(?:&[a-zA-Z0-9]+=[a-zA-Z0-9]+)*$`).MatchString(params) {
		return
	}

	log.Fatal("Invalid params format")
}

func (d *DsnBuilder) Build() *DSN {
	return &DSN{
		host:     d.Host,
		port:     d.Port,
		username: d.Username,
		password: d.Password,
		name:     d.Name,
		params:   d.Params,
	}
}

func (d *DSN) GetMysqlDSN() string {
	return d.username + ":" + d.password + "@tcp(" + d.host + ":" + string(d.port) + ")/" + d.name + d.params
}

func (d *DSN) GetPostgresDSN() string {
	return "host=" + d.host + " port=" + d.port + " user=" + d.username + " password=" + d.password + " dbname=" + d.name + d.params
}

func (d *DSN) GetMssqlDSN() string {
	return "sqlserver://" + d.username + ":" + d.password + "@" + d.host + ":" + string(d.port) + "?database=" + d.name
}

func (d *DSN) GetSqliteDSN() string {
	return d.name
}
