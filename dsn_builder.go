package migration

import "strconv"

type DsnBuilder struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	Charset  string
}

type DSN struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	Charset  string
}

func NewDsnBuilder() *DsnBuilder {
	return &DsnBuilder{}
}

func (d *DsnBuilder) SetHost(host string) *DsnBuilder {
	if host == "" {
		panic("DB Host is required")
	}

	d.Host = host
	return d
}

func (d *DsnBuilder) SetPort(port string) *DsnBuilder {
	if port == "" {
		panic("DB Port is required")
	}

	if intPort, err := strconv.Atoi(port); err != nil {
		panic(err)
	} else {
		d.Port = intPort
	}
	return d
}

func (d *DsnBuilder) SetUsername(username string) *DsnBuilder {
	if username == "" {
		panic("DB Username is required")
	}

	d.Username = username
	return d
}

func (d *DsnBuilder) SetPassword(password string) *DsnBuilder {
	if password == "" {
		panic("DB Password is required")
	}

	d.Password = password
	return d
}

func (d *DsnBuilder) SetName(name string) *DsnBuilder {
	if name == "" {
		panic("DB name is required")
	}

	d.Name = name
	return d
}

func (d *DsnBuilder) SetCharset(charset string) *DsnBuilder {
	d.Charset = charset
	return d
}

func (d *DsnBuilder) Build() *DSN {
	return &DSN{
		Host:     d.Host,
		Port:     d.Port,
		Username: d.Username,
		Password: d.Password,
		Name:     d.Name,
		Charset:  d.Charset,
	}
}

func (d *DSN) GetMysqlDSN() string {
	return d.Username + ":" + d.Password + "@tcp(" + d.Host + ":" + string(d.Port) + ")/" + d.Name + "?charset=" + d.Charset
}

func (d *DSN) GetPostgresDSN() string {
	return "host=" + d.Host + " port=" + string(d.Port) + " user=" + d.Username + " password=" + d.Password + " dbname=" + d.Name + " sslmode=disable"
}

func (d *DSN) GetSqliteDSN() string {
	return d.Name
}

func (d *DSN) GetMssqlDSN() string {
	return "sqlserver://" + d.Username + ":" + d.Password + "@" + d.Host + ":" + string(d.Port) + "?database=" + d.Name
}
