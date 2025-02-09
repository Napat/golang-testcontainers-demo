package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	content := []byte(`
mysql:
  host: localhost
  port: 3306
  user: test
  password: test
  database: testdb
`)
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	// Test loading config
	cfg, err := Load(tmpfile.Name())
	require.NoError(t, err)

	// Verify config values
	assert.Equal(t, "localhost", cfg.MySQL.Host)
	assert.Equal(t, "3306", cfg.MySQL.Port)
	assert.Equal(t, "test", cfg.MySQL.User)
	assert.Equal(t, "test", cfg.MySQL.Password)
	assert.Equal(t, "testdb", cfg.MySQL.Database)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	assert.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	content := []byte(`invalid: yaml: content`)
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	_, err = Load(tmpfile.Name())
	assert.Error(t, err)
}
