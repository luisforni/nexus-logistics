package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "this-is-a-32-char-minimum-secret!!")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("BACKEND_PORT", "9090")
	t.Setenv("APP_ENV", "development")
	t.Setenv("REDIS_URL", "redis://localhost:6379")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURL)
	assert.Equal(t, "redis://localhost:6379", cfg.RedisURL)
	assert.NotEmpty(t, cfg.JWTSecret)
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestLoad_ShortJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "tooshort")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "32")
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("JWT_SECRET", "this-is-a-32-char-minimum-secret!!")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "this-is-a-32-char-minimum-secret!!")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")

	t.Setenv("BACKEND_PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("OPTIMIZER_HOST", "")
	t.Setenv("OPTIMIZER_PORT", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "production", cfg.AppEnv)
	assert.Equal(t, "optimizer", cfg.OptimizerHost)
	assert.Equal(t, "9090", cfg.OptimizerPort)
}
