// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/go-openapi/codescan"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"

	"github.com/jessevdk/go-flags"
	"go.yaml.in/yaml/v3"
)

const (
	basePath       = "../../../../fixtures/goparsing/spec"
	jsonResultFile = basePath + "/api_spec_go111.json"
	yamlResultFile = basePath + "/api_spec_go111.yml"

	jsonResultFileRef = basePath + "/api_spec_go111_ref.json"
	yamlResultFileRef = basePath + "/api_spec_go111_ref.yml"

	jsonResultFileTransparent = basePath + "/api_spec_go111_transparent.json"
	yamlResultFileTransparent = basePath + "/api_spec_go111_transparent.yml"
)

var enableSpecOutput bool

func init() {
	flag.BoolVar(&enableSpecOutput, "enable-spec-output", false, "enable spec gen test to write output to a file")
}

func TestSpecFileExecute(t *testing.T) {
	files := []string{"", "spec.json", "spec.yml", "spec.yaml"}

	for _, outputFile := range files {
		name := outputFile
		if outputFile == "" {
			name = "to stdout"
		}

		t.Run("should produce spec file "+name, func(t *testing.T) {
			spec := &SpecFile{
				WorkDir: basePath,
				Output:  flags.Filename(outputFile),
			}
			if outputFile == "" {
				defaultWriter = io.Discard
			}
			t.Cleanup(func() {
				if outputFile != "" {
					_ = os.Remove(outputFile)
				} else {
					defaultWriter = os.Stdout
				}
			})

			require.NoError(t, spec.Execute(nil))
		})
	}
}

func TestSpecFileExecuteRespectsSetXNullableForPointersOption(t *testing.T) {
	outputFileName := "spec.json"
	spec := &SpecFile{
		WorkDir:                 "../../../../fixtures/enhancements/pointers-nullable-by-default",
		Output:                  flags.Filename(outputFileName),
		ScanModels:              true,
		SetXNullableForPointers: true,
	}

	defer func() { _ = os.Remove(outputFileName) }()

	err := spec.Execute(nil)
	require.NoError(t, err)

	data, err := os.ReadFile(outputFileName)
	require.NoError(t, err)

	var got map[string]any
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	require.Len(t, got["definitions"], 2)
	require.Contains(t, got["definitions"], "Item")
	itemDefinition, ok := got["definitions"].(map[string]any)["Item"].(map[string]any)
	require.TrueT(t, ok)
	require.Contains(t, itemDefinition["properties"], "Value1")
	value1Property, ok := itemDefinition["properties"].(map[string]any)["Value1"].(map[string]any)
	require.TrueT(t, ok)
	require.MapContainsT(t, value1Property, "x-nullable")
	assert.Equal(t, true, value1Property["x-nullable"])
}

func TestGenerateJSONSpec(t *testing.T) {
	opts := codescan.Options{
		WorkDir:  basePath,
		Packages: []string{"./..."},
	}

	swspec, err := codescan.Run(&opts)
	require.NoError(t, err)

	data, err := marshalToJSONFormat(swspec, true)
	require.NoError(t, err)

	expected, err := os.ReadFile(jsonResultFile)
	require.NoError(t, err)

	verifyJSONData(t, data, expected)
}

func TestGenerateYAMLSpec(t *testing.T) {
	opts := codescan.Options{
		WorkDir:  basePath,
		Packages: []string{"./..."},
	}

	swspec, err := codescan.Run(&opts)
	require.NoError(t, err)

	data, err := marshalToYAMLFormat(swspec)
	require.NoError(t, err)

	expected, err := os.ReadFile(yamlResultFile)
	require.NoError(t, err)
	{
		var jsonObj any
		require.NoError(t, yaml.Unmarshal(expected, &jsonObj))

		rewritten, err := yaml.Marshal(jsonObj)
		require.NoError(t, err)
		expected = rewritten
	}

	if enableSpecOutput {
		require.NoError(t,
			os.WriteFile("expected.yaml", expected, 0o600),
		)
		require.NoError(t,
			os.WriteFile("generated.yaml", data, 0o600),
		)
	}

	verifyYAMLData(t, data, expected)
}

func TestGenerateJSONSpecWithSpec(t *testing.T) {
	opts := codescan.Options{
		WorkDir:    basePath,
		Packages:   []string{"./..."},
		RefAliases: true,
	}

	swspec, err := codescan.Run(&opts)
	require.NoError(t, err)

	data, err := marshalToJSONFormat(swspec, true)
	require.NoError(t, err)

	expected, err := os.ReadFile(jsonResultFileRef)
	require.NoError(t, err)

	verifyJSONData(t, data, expected)
}

func TestGenerateYAMLSpecWithRefAliases(t *testing.T) {
	opts := codescan.Options{
		WorkDir:    basePath,
		Packages:   []string{"./..."},
		RefAliases: true,
	}

	swspec, err := codescan.Run(&opts)
	require.NoError(t, err)

	data, err := marshalToYAMLFormat(swspec)
	require.NoError(t, err)

	expected, err := os.ReadFile(yamlResultFileRef)
	require.NoError(t, err)
	{
		var jsonObj any
		require.NoError(t, yaml.Unmarshal(expected, &jsonObj))

		rewritten, err := yaml.Marshal(jsonObj)
		require.NoError(t, err)
		expected = rewritten
	}

	if enableSpecOutput {
		require.NoError(t,
			os.WriteFile("expected_ref.yaml", expected, 0o600),
		)
		require.NoError(t,
			os.WriteFile("generated_ref.yaml", data, 0o600),
		)
	}

	verifyYAMLData(t, data, expected)
}

func TestGenerateJSONSpecWithTransparentAliases(t *testing.T) {
	opts := codescan.Options{
		WorkDir:            basePath,
		Packages:           []string{"./..."},
		TransparentAliases: true,
	}

	swspec, err := codescan.Run(&opts)
	require.NoError(t, err)

	data, err := marshalToJSONFormat(swspec, true)
	require.NoError(t, err)

	expected, err := os.ReadFile(jsonResultFileTransparent)
	require.NoError(t, err)

	verifyJSONData(t, data, expected)
}

func TestGenerateYAMLSpecWithTransparentAliases(t *testing.T) {
	opts := codescan.Options{
		WorkDir:            basePath,
		Packages:           []string{"./..."},
		TransparentAliases: true,
	}

	swspec, err := codescan.Run(&opts)
	require.NoError(t, err)

	data, err := marshalToYAMLFormat(swspec)
	require.NoError(t, err)

	expected, err := os.ReadFile(yamlResultFileTransparent)
	require.NoError(t, err)
	{
		var jsonObj any
		require.NoError(t, yaml.Unmarshal(expected, &jsonObj))

		rewritten, err := yaml.Marshal(jsonObj)
		require.NoError(t, err)
		expected = rewritten
	}

	if enableSpecOutput {
		require.NoError(t,
			os.WriteFile("expected_transparent.yaml", expected, 0o600),
		)
		require.NoError(t,
			os.WriteFile("generated_transparent.yaml", data, 0o600),
		)
	}

	verifyYAMLData(t, data, expected)
}

func TestStripXOrderFromJSON(t *testing.T) {
	t.Run("strips x-order from properties", func(t *testing.T) {
		input := `{"definitions":{"Foo":{"properties":{"zeta":{"type":"string","x-order":0},"alpha":{"type":"string","x-order":1}}}}}`
		result, err := stripXOrderFromJSON([]byte(input))
		require.NoError(t, err)

		var got map[string]any
		require.NoError(t, json.Unmarshal(result, &got))

		defs := got["definitions"].(map[string]any)
		foo := defs["Foo"].(map[string]any)
		props := foo["properties"].(map[string]any)
		zeta := props["zeta"].(map[string]any)
		alpha := props["alpha"].(map[string]any)

		assert.NotContains(t, zeta, "x-order")
		assert.NotContains(t, alpha, "x-order")
		assert.Equal(t, "string", zeta["type"])
		assert.Equal(t, "string", alpha["type"])
	})

	t.Run("preserves key order", func(t *testing.T) {
		input := `{"b":1,"a":2,"c":3}`
		result, err := stripXOrderFromJSON([]byte(input))
		require.NoError(t, err)
		// The output should preserve order: b, a, c
		assert.Equal(t, `{"b":1,"a":2,"c":3}`, string(result))
	})

	t.Run("handles empty objects", func(t *testing.T) {
		input := `{"x-order":0}`
		result, err := stripXOrderFromJSON([]byte(input))
		require.NoError(t, err)
		assert.Equal(t, `{}`, string(result))
	})
}

func TestSpecPropertyOrder(t *testing.T) {
	specCmd := &SpecFile{
		WorkDir:    "../../../../fixtures/goparsing/property-order",
		ScanModels: true,
		Format:     "json",
	}

	err := specCmd.Execute([]string{"./..."})
	require.NoError(t, err)

	// Re-run to capture the output
	opts := codescan.Options{
		WorkDir:    "../../../../fixtures/goparsing/property-order",
		Packages:   []string{"./..."},
		ScanModels: true,
	}
	swspec, err := codescan.Run(&opts)
	require.NoError(t, err)

	setPropertyOrderFromGoStructs(swspec, []string{"./..."}, "../../../../fixtures/goparsing/property-order", "")

	data, err := marshalToJSONFormat(swspec, true)
	require.NoError(t, err)

	// Verify x-order is NOT in the output
	assert.NotContains(t, string(data), "x-order")

	// Verify property order in JSON matches Go struct field order.
	// For OrderedItem: zeta, alpha, mu, beta, epsilon (NOT alphabetical)
	jsonStr := string(data)

	t.Run("OrderedItem properties follow struct field order", func(t *testing.T) {
		foundZeta := findJSONKeyPosition(t, jsonStr, "OrderedItem", "zeta")
		foundAlpha := findJSONKeyPosition(t, jsonStr, "OrderedItem", "alpha")
		foundMu := findJSONKeyPosition(t, jsonStr, "OrderedItem", "mu")
		foundBeta := findJSONKeyPosition(t, jsonStr, "OrderedItem", "beta")
		foundEpsilon := findJSONKeyPosition(t, jsonStr, "OrderedItem", "epsilon")

		assert.LessT(t, foundZeta, foundAlpha)
		assert.LessT(t, foundAlpha, foundMu)
		assert.LessT(t, foundMu, foundBeta)
		assert.LessT(t, foundBeta, foundEpsilon)
	})

	t.Run("AnotherModel properties follow struct field order", func(t *testing.T) {
		foundCharlie := findJSONKeyPosition(t, jsonStr, "AnotherModel", "charlie")
		foundAble := findJSONKeyPosition(t, jsonStr, "AnotherModel", "able")
		foundBaker := findJSONKeyPosition(t, jsonStr, "AnotherModel", "baker")

		assert.LessT(t, foundCharlie, foundAble)
		assert.LessT(t, foundAble, foundBaker)
	})
}

// findJSONKeyPosition returns the byte offset of a property key
// within a specific definition in the JSON string.
func findJSONKeyPosition(t *testing.T, jsonStr, defName, propKey string) int {
	t.Helper()

	// Find the definition section first
	defStart := strings.Index(jsonStr, `"`+defName+`"`)
	require.NotEqual(t, -1, defStart, "definition %q not found", defName)

	// Within the definition, find the property key
	section := jsonStr[defStart:]
	pos := strings.Index(section, `"`+propKey+`"`)
	require.NotEqual(t, -1, pos, "property %q not found in definition %q", propKey, defName)

	return defStart + pos
}

func verifyJSONData(t *testing.T, data, expectedJSON []byte) {
	t.Helper()

	var got, expected any

	require.NoError(t, json.Unmarshal(data, &got))
	require.NoError(t, json.Unmarshal(expectedJSON, &expected))
	assert.Equal(t, expected, got)
}

func verifyYAMLData(t *testing.T, data, expectedYAML []byte) {
	t.Helper()

	var got, expected any

	require.NoError(t, yaml.Unmarshal(data, &got))
	require.NoError(t, yaml.Unmarshal(expectedYAML, &expected))
	assert.Equal(t, expected, got)
}
