package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestNewServer(t *testing.T) {
	connection, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })

	originalConnectDB := connectDB
	connectDB = func() (*sql.DB, error) {
		return connection, nil
	}
	t.Cleanup(func() { connectDB = originalConnectDB })

	server, err := newServer()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestRunReturnsDatabaseError(t *testing.T) {
	originalConnectDB := connectDB
	connectDB = func() (*sql.DB, error) {
		return nil, sql.ErrConnDone
	}
	t.Cleanup(func() { connectDB = originalConnectDB })

	if err := run(); err != sql.ErrConnDone {
		t.Fatalf("expected database error, got %v", err)
	}
}

func TestRunStartsServer(t *testing.T) {
	connection, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })

	originalConnectDB := connectDB
	originalRunServer := runServer
	connectDB = func() (*sql.DB, error) {
		return connection, nil
	}
	runServer = func(*gin.Engine) error {
		return nil
	}
	t.Cleanup(func() {
		connectDB = originalConnectDB
		runServer = originalRunServer
	})

	if err := run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMainEntrypoint(t *testing.T) {
	connection, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })

	originalConnectDB := connectDB
	originalRunServer := runServer
	connectDB = func() (*sql.DB, error) {
		return connection, nil
	}
	runServer = func(*gin.Engine) error {
		return nil
	}
	t.Cleanup(func() {
		connectDB = originalConnectDB
		runServer = originalRunServer
	})

	main()
}

func TestMainEntrypointReportsError(t *testing.T) {
	originalConnectDB := connectDB
	connectDB = func() (*sql.DB, error) {
		return nil, sql.ErrConnDone
	}
	t.Cleanup(func() { connectDB = originalConnectDB })

	main()
}
