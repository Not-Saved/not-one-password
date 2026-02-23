package config

import (
	"os"
	"testing"
)

func setDBEnvVars(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_PASSWORD", "testpass")
	t.Setenv("DB_NAME", "testdb")
}

func TestLoad_AllEnvVarsSet(t *testing.T) {
	setDBEnvVars(t)
	t.Setenv("APP_PORT", "9090")

	cfg := Load()

	if cfg.DB.Host != "localhost" {
		t.Errorf("expected DB.Host 'localhost', got %q", cfg.DB.Host)
	}
	if cfg.DB.Port != "5432" {
		t.Errorf("expected DB.Port '5432', got %q", cfg.DB.Port)
	}
	if cfg.DB.User != "testuser" {
		t.Errorf("expected DB.User 'testuser', got %q", cfg.DB.User)
	}
	if cfg.DB.Password != "testpass" {
		t.Errorf("expected DB.Password 'testpass', got %q", cfg.DB.Password)
	}
	if cfg.DB.Name != "testdb" {
		t.Errorf("expected DB.Name 'testdb', got %q", cfg.DB.Name)
	}
	if cfg.AppPort != "9090" {
		t.Errorf("expected AppPort '9090', got %q", cfg.AppPort)
	}
}

func TestLoad_AppPortDefault(t *testing.T) {
	setDBEnvVars(t)
	os.Unsetenv("APP_PORT")

	cfg := Load()

	if cfg.AppPort != "8080" {
		t.Errorf("expected default AppPort '8080', got %q", cfg.AppPort)
	}
}

func TestConnString(t *testing.T) {
	dbCfg := DBConfig{
		Host:     "myhost",
		Port:     "5433",
		User:     "admin",
		Password: "secret",
		Name:     "mydb",
	}

	connStr := dbCfg.ConnString()
	expected := "host=myhost port=5433 user=admin password=secret dbname=mydb sslmode=disable"

	if connStr != expected {
		t.Errorf("expected conn string %q, got %q", expected, connStr)
	}
}

func TestGetEnv_WithValue(t *testing.T) {
	t.Setenv("TEST_GET_ENV_KEY", "myvalue")

	val := getEnv("TEST_GET_ENV_KEY", "fallback")
	if val != "myvalue" {
		t.Errorf("expected 'myvalue', got %q", val)
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("TEST_GET_ENV_MISSING")

	val := getEnv("TEST_GET_ENV_MISSING", "default_val")
	if val != "default_val" {
		t.Errorf("expected 'default_val', got %q", val)
	}
}

func TestGetEnv_EmptyUseFallback(t *testing.T) {
	t.Setenv("TEST_GET_ENV_EMPTY", "")

	val := getEnv("TEST_GET_ENV_EMPTY", "fallback")
	if val != "fallback" {
		t.Errorf("expected 'fallback' for empty env var, got %q", val)
	}
}
