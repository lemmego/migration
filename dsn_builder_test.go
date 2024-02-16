package migration

import "testing"

func TestDSNBuilderReturnsErrorForUnsupportedDialect(t *testing.T) {
	_, err := NewDsnBuilder().
		SetHost("localhost").
		SetPort("3306").
		SetUsername("root").
		SetPassword("password").
		SetName("test").
		SetParams("parseTime=true").
		SetDialect("unsupported").
		ToString()

	if err != ErrUnsupportedDialect {
		t.Errorf("Expected %s, got %s", ErrUnsupportedDialect, err)
	}
}

func TestDSNBuilderMySql(t *testing.T) {
	expected := "root:password@tcp(localhost:3306)/test?parseTime=true"
	dsn, _ := NewDsnBuilder().
		SetHost("localhost").
		SetPort("3306").
		SetUsername("root").
		SetPassword("password").
		SetName("test").
		SetParams("parseTime=true").
		SetDialect("mysql").
		ToString()

	if dsn != expected {
		t.Errorf("Expected %s, got %s", expected, dsn)
	}
}

func TestDSNBuilderPostgres(t *testing.T) {
	expected := "host=localhost port=5432 user=root password=password dbname=test sslmode=disable"
	dsn, err := NewDsnBuilder().
		SetHost("localhost").
		SetPort("5432").
		SetName("test").
		SetUsername("root").
		SetPassword("password").
		SetParams("sslmode=disable").
		SetDialect("postgres").
		ToString()

	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	if dsn != expected {
		t.Errorf("Expected dsn to be %s, got %s", expected, dsn)
	}
}
