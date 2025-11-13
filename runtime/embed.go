package runtimec

import _ "embed"

// Source holds the contents of runtime.c embedded for use by the CLI.
//
//go:embed runtime.c
var Source []byte
