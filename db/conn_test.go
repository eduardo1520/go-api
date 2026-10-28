package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestConnectDBReturnsConnection(t *testing.T) {
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	originalOpenDB := openDB
	openDB = func(string, string) (*sql.DB, error) {
		return connection, nil
	}
	defer func() { openDB = originalOpenDB }()

	db, err := ConnectDB()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db != connection {
		t.Fatal("expected the opened database connection")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConnectDBReturnsOpenError(t *testing.T) {
	originalOpenDB := openDB
	openDB = func(string, string) (*sql.DB, error) {
		return nil, errors.New("open failed")
	}
	defer func() { openDB = originalOpenDB }()

	db, err := ConnectDB()
	if db != nil || err == nil || err.Error() != "open failed" {
		t.Fatalf("expected open error, got db=%v err=%v", db, err)
	}
}

func TestConnectDBReturnsPingError(t *testing.T) {
	connection, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	connection.Close()

	originalOpenDB := openDB
	openDB = func(string, string) (*sql.DB, error) {
		return connection, nil
	}
	defer func() { openDB = originalOpenDB }()

	db, err := ConnectDB()
	if db != nil || err == nil {
		t.Fatalf("expected ping error, got db=%v err=%v", db, err)
	}
}
