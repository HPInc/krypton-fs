package config

import (
	"os"
	"strconv"
	"testing"
)

func init() {
	InitTestLogger()
}

func TestDoesNotOverrideWhenNoEnvVariablePresent(t *testing.T) {
	c := Config{}
	expected := c.Server.Host
	c.OverrideFromEnvironment()
	if c.Server.Host != "" {
		t.Fatalf(
			`OverrideFromEnvironment:, Expected Server.Host = %s,  found: %s`,
			expected, c.Server.Host)
	}
}

func TestOverridesStringEnvVariable(t *testing.T) {
	c := Config{}
	expected := "localhost"
	os.Setenv("FS_SERVER", expected)
	c.OverrideFromEnvironment()
	if c.Server.Host != expected {
		t.Fatalf(
			`OverrideFromEnvironment:, Expected Server.Host = %s,  found: %s`,
			expected, c.Server.Host)
	}
}

func TestOverridesIntEnvVariable(t *testing.T) {
	c := Config{}
	expected := 123
	os.Setenv("FS_PORT", strconv.Itoa(expected))
	c.OverrideFromEnvironment()
	if c.Server.Port != expected {
		t.Fatalf(
			`OverrideFromEnvironment:, Expected Server.Port = %d,  found: %d`,
			expected, c.Server.Port)
	}
}

func TestOverridesSkipsBadEnvVariable(t *testing.T) {
	c := Config{}
	expected := 123
	c.Server.Port = expected
	os.Setenv("FS_PORT", "not_an_int")
	c.OverrideFromEnvironment()
	if c.Server.Port != expected {
		t.Fatalf(
			`OverrideFromEnvironment:, Expected Server.Port = %d,  found: %d`,
			expected, c.Server.Port)
	}
}
