package migration

import "testing"

func TestDSNBuilderReturnsErrorForUnsupportedDialect(t *testing.T) {
	_, err := (&DataSource{Dialect: "unsupported"}).String()

	if err != ErrUnsupportedDialect {
		t.Errorf("Expected %s, got %s", ErrUnsupportedDialect, err)
	}
}

func TestToStringSQLite(t *testing.T) {
	expected := "file::memory:?cache=shared"
	dsn, err := (&DataSource{
		Dialect: "sqlite",
		Name:    ":memory:",
		Params:  "cache=shared",
	}).String()

	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	if dsn != expected {
		t.Errorf("Expected %s, got %s", expected, dsn)
	}
}

func TestToStringMySql(t *testing.T) {
	expected := "root:password@tcp(localhost:3306)/test?parseTime=true"
	dsn, _ := (&DataSource{
		Host:     "localhost",
		Port:     "3306",
		Username: "root",
		Password: "password",
		Name:     "test",
		Params:   "parseTime=true",
		Dialect:  "mysql",
	}).String()

	if dsn != expected {
		t.Errorf("Expected %s, got %s", expected, dsn)
	}
}

func TestToStringPostgres(t *testing.T) {
	expected := "host=localhost port=5432 user=root password=password dbname=test sslmode=disable"
	dsn, err := (&DataSource{
		Host:     "localhost",
		Port:     "5432",
		Username: "root",
		Password: "password",
		Name:     "test",
		Params:   "sslmode=disable",
		Dialect:  "postgres",
	}).String()

	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	if dsn != expected {
		t.Errorf("Expected dsn to be %s, got %s", expected, dsn)
	}
}
