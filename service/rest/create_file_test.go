// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
)

type testTableResult struct {
	desc   string
	result bool
}

func TestMain(m *testing.M) {
	fsLogger, _ = zap.NewProduction(zap.AddCaller())
	os.Exit(m.Run())
}

// validate file name
func TestFileNameValidation(t *testing.T) {
	m := map[string]testTableResult{
		strings.Repeat(`a`, minFileNameLength-1): {`not enough length`, false},
		strings.Repeat(`a`, maxFileNameLength+1): {`name too long`, false},
		`!_is_invalid`:                           {`! is not allowed`, false},
		`<_is_invalid`:                           {`< is not allowed`, false},
		`>_is_invalid`:                           {`> is not allowed`, false},
		`/_is_invalid`:                           {`/ is not allowed`, false},
		`_is_valid`:                              {`_ is allowed`, true},
		`._is_valid`:                             {`. is allowed`, true},
		strings.Repeat(`a`, minFileNameLength):   {`name is minimal allowed length`, true},
		strings.Repeat(`a`, maxFileNameLength):   {`name is maximum allowed length`, true},
	}

	for k, v := range m {
		if isValidFileName(k) != v.result {
			t.Fatalf(
				"File name validation error: %s - %s, expected: %v, got: %v",
				k, v.desc, v.result, !v.result)
		}
	}
}

// validate checksum
func TestChecksumValidation(t *testing.T) {
	m := map[string]testTableResult{
		strings.Repeat(`a`, minChecksumLength-1):                      {`not enough length`, false},
		strings.Repeat(`a`, maxChecksumLength+1):                      {`too long`, false},
		base64.RawStdEncoding.EncodeToString([]byte("hello")):         {`invalid no padding`, false},
		base64.StdEncoding.EncodeToString([]byte("a")):                {`minimal allowed length`, true},
		base64.StdEncoding.EncodeToString([]byte("overflowmaximuml")): {`maximum allowed length`, true},
	}

	for k, v := range m {
		if isValidChecksum(k) != v.result {
			t.Fatalf(
				"Checksum validation error: %s - %s, expected: %v, got: %v",
				k, v.desc, v.result, !v.result)
		}
	}
}
