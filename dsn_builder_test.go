package migration

import "testing"

func TestDSNBuilderLogsFatalOnInvalidDriver(t *testing.T) {
	NewDsnBuilder("invalid")
}

func TestDSNBuilderMySql(t *testing.T) {
	expected := "root:password@tcp(localhost:3306)/test?parseTime=true"
	dsn := NewDsnBuilder("mysql").
		SetHost("localhost").
		SetPort("3306").
		SetUsername("root").
		SetPassword("password").
		SetName("test").
		SetParams("parseTime=true").
		Build().
		GetMysqlDSN()

	if dsn != expected {
		t.Errorf("Expected %s, got %s", expected, dsn)
	}
}

func TestDSNBuilderPostgres(t *testing.T) {
	expected := "host=localhost port=5432 user=root password=password dbname=test sslmode=disable"
	dsn := NewDsnBuilder("postgres").
		SetHost("localhost").
		SetPort("5432").
		SetName("test").
		SetUsername("root").
		SetPassword("password").
		SetParams("sslmode=disable").
		Build().
		GetPostgresDSN()

	if dsn != expected {
		t.Errorf("Expected %s, got %s", expected, dsn)
	}
}
