package slushy

import (
	"embed"
	"testing"

	"github.com/taubyte/tau/pkg/starlark"
	"gotest.tools/v3/assert"
)

// Embed the necessary Starlark scripts for testing.
//
//go:embed testdata/*
var testFiles embed.FS

func TestBasicExec(t *testing.T) {
	// Create a VM with the embedded file system.
	vm, err := starlark.New(testFiles)
	assert.NilError(t, err, "Failed to create VM")

	// Register the testModule with its native Go Add function as a built-in.
	vm.Module(New())

	// Load the context to call functions.
	ctx, err := vm.File("testdata/cmd.star")
	assert.NilError(t, err, "Failed to load cmd.star")

	// Call the native Go Add function directly.
	result, err := ctx.CallWithNative("execute", "echo howdy")
	assert.NilError(t, err, "Failed to call execute function")

	// Check the result.
	if val, ok := result.(string); ok {
		assert.Equal(t, val, "howdy\n", "Expected 8 as the result, got %d", val)
	} else {
		t.Errorf("Expected result to be a string, got %T", result)
	}
}
