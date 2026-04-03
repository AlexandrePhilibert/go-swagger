// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/require"

	"github.com/go-openapi/analysis"
	"github.com/go-openapi/loads"
)

func TestSpec_Issue1429(t *testing.T) {
	defer discardOutput()()

	// acknowledge fix in go-openapi/spec
	specPath := filepath.Join("..", "fixtures", "bugs", "1429", "swagger-1429.yaml")
	_, err := loads.Spec(specPath)
	require.NoError(t, err)

	opts := testGenOpts()
	opts.Spec = specPath
	_, err = opts.validateAndFlattenSpec()
	require.NoError(t, err)

	// more aggressive fixture on $refs, with validation errors, but flatten ok
	specPath = filepath.Join("..", "fixtures", "bugs", "1429", "swagger.yaml")
	specDoc, err := loads.Spec(specPath)
	require.NoError(t, err)

	opts.Spec = specPath
	opts.FlattenOpts.BasePath = specDoc.SpecFilePath()
	opts.FlattenOpts.Spec = analysis.New(specDoc.Spec())
	opts.FlattenOpts.Minimal = true
	err = analysis.Flatten(*opts.FlattenOpts)
	require.NoError(t, err)

	specDoc, _ = loads.Spec(specPath) // needs reload
	opts.FlattenOpts.Spec = analysis.New(specDoc.Spec())
	opts.FlattenOpts.Minimal = false
	err = analysis.Flatten(*opts.FlattenOpts)
	require.NoError(t, err)
}

func TestSpec_Issue2527(t *testing.T) {
	defer discardOutput()()

	t.Run("spec should be detected as invalid", func(t *testing.T) {
		specPath := filepath.Join("..", "fixtures", "bugs", "2527", "swagger.yml")
		_, err := loads.Spec(specPath)
		require.NoError(t, err)

		opts := testGenOpts()
		opts.Spec = specPath
		opts.ValidateSpec = true // test options skip validation by default
		_, err = opts.validateAndFlattenSpec()
		require.Error(t, err)
	})

	t.Run("fixed spec should be detected as valid", func(t *testing.T) {
		specPath := filepath.Join("..", "fixtures", "bugs", "2527", "swagger-fixed.yml")
		_, err := loads.Spec(specPath)
		require.NoError(t, err)

		opts := testGenOpts()
		opts.Spec = specPath
		opts.ValidateSpec = true
		_, err = opts.validateAndFlattenSpec()
		require.NoError(t, err)
	})
}

func TestSpec_FindSwaggerSpec(t *testing.T) {
	keepErr := func(_ string, err error) error { return err }
	require.Error(t, keepErr(findSwaggerSpec("")))
	require.Error(t, keepErr(findSwaggerSpec("nowhere")))
	require.Error(t, keepErr(findSwaggerSpec(filepath.Join("..", "fixtures"))))
	require.NoError(t, keepErr(findSwaggerSpec(filepath.Join("..", "fixtures", "codegen", "shipyard.yml"))))
}

func TestSpec_Issue1621(t *testing.T) {
	defer discardOutput()()

	// acknowledge fix in go-openapi/spec
	specPath := filepath.Join("..", "fixtures", "bugs", "1621", "fixture-1621.yaml")
	_, err := loads.Spec(specPath)
	require.NoError(t, err)

	opts := testGenOpts()
	opts.Spec = specPath
	opts.ValidateSpec = true
	_, err = opts.validateAndFlattenSpec()
	require.NoError(t, err)
}

func TestShared_Issue1614(t *testing.T) {
	defer discardOutput()()

	// acknowledge fix in go-openapi/spec
	specPath := filepath.Join("..", "fixtures", "bugs", "1614", "gitea.json")
	_, err := loads.Spec(specPath)
	require.NoError(t, err)

	opts := testGenOpts()
	opts.Spec = specPath
	opts.ValidateSpec = true
	_, err = opts.validateAndFlattenSpec()
	require.NoError(t, err)
}

func Test_AnalyzeSpec_Issue2216(t *testing.T) {
	defer discardOutput()()

	t.Run("single-swagger-file", func(t *testing.T) {
		specPath := filepath.Join("..", "fixtures", "bugs", "2216", "swagger-single.yml")

		opts := testGenOpts()
		opts.Spec = specPath
		opts.ValidateSpec = true
		opts.PropertiesSpecOrder = true
		_, _, err := opts.analyzeSpec()
		require.NoError(t, err)
	})

	t.Run("splitted-swagger-file", func(t *testing.T) {
		specPath := filepath.Join("..", "fixtures", "bugs", "2216", "swagger.yml")

		opts := testGenOpts()
		opts.Spec = specPath
		opts.ValidateSpec = true
		opts.PropertiesSpecOrder = true
		_, _, err := opts.analyzeSpec()
		require.NoError(t, err)
	})
}

func TestExtractPropertyOrder(t *testing.T) {
	t.Run("extracts order from YAML spec", func(t *testing.T) {
		specPath := filepath.Join("..", "fixtures", "codegen", "auto-property-order.yml")
		orderMap, err := extractPropertyOrder(specPath)
		require.NoError(t, err)

		// orderedModel: zeta, alpha, mu, beta, epsilon
		require.Contains(t, orderMap, "orderedModel")
		require.Equal(t, []string{"zeta", "alpha", "mu", "beta", "epsilon"}, orderMap["orderedModel"])

		// nestedModel: charlie, able, innerObj
		require.Contains(t, orderMap, "nestedModel")
		require.Equal(t, []string{"charlie", "able", "innerObj"}, orderMap["nestedModel"])

		// nested properties: innerObj has zebra, apple, mango
		require.Contains(t, orderMap, "nestedModel.innerObj")
		require.Equal(t, []string{"zebra", "apple", "mango"}, orderMap["nestedModel.innerObj"])

		// mixedModel: noOrderZ, explicitSecond, noOrderA, explicitFirst
		require.Contains(t, orderMap, "mixedModel")
		require.Equal(t, []string{"noOrderZ", "explicitSecond", "noOrderA", "explicitFirst"}, orderMap["mixedModel"])
	})

	t.Run("extracts order from keep-spec-order fixture", func(t *testing.T) {
		specPath := filepath.Join("..", "fixtures", "codegen", "keep-spec-order.yml")
		orderMap, err := extractPropertyOrder(specPath)
		require.NoError(t, err)

		require.Contains(t, orderMap, "abctype")
		require.Equal(t, []string{"ccc", "bbb", "aaa", "inner-object"}, orderMap["abctype"])

		require.Contains(t, orderMap, "abctype.inner-object")
		require.Equal(t, []string{"inner-ccc", "inner-bbb", "inner-aaa"}, orderMap["abctype.inner-object"])
	})
}

func TestApplyPropertyOrder(t *testing.T) {
	specPath := filepath.Join("..", "fixtures", "codegen", "auto-property-order.yml")
	orderMap, err := extractPropertyOrder(specPath)
	require.NoError(t, err)

	specDoc, err := loads.Spec(specPath)
	require.NoError(t, err)

	applyPropertyOrder(specDoc.Spec(), orderMap)

	t.Run("x-order is set on properties", func(t *testing.T) {
		schema := specDoc.Spec().Definitions["orderedModel"]
		// zeta=0, alpha=1, mu=2, beta=3, epsilon=4
		xo, ok := schema.Properties["zeta"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 0, xo)

		xo, ok = schema.Properties["alpha"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 1, xo)

		xo, ok = schema.Properties["mu"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 2, xo)

		xo, ok = schema.Properties["beta"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 3, xo)

		xo, ok = schema.Properties["epsilon"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 4, xo)
	})

	t.Run("document order overrides explicit x-order", func(t *testing.T) {
		schema := specDoc.Spec().Definitions["mixedModel"]
		// document order: noOrderZ(0), explicitSecond(1), noOrderA(2), explicitFirst(3)
		// explicit x-order values are overridden by document position
		xo, ok := schema.Properties["noOrderZ"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 0, xo)

		xo, ok = schema.Properties["explicitSecond"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 1, xo)

		xo, ok = schema.Properties["noOrderA"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 2, xo)

		xo, ok = schema.Properties["explicitFirst"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 3, xo)
	})

	t.Run("nested properties get x-order", func(t *testing.T) {
		schema := specDoc.Spec().Definitions["nestedModel"]
		innerObj := schema.Properties["innerObj"]

		xo, ok := innerObj.Properties["zebra"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 0, xo)

		xo, ok = innerObj.Properties["apple"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 1, xo)

		xo, ok = innerObj.Properties["mango"].Extensions.GetInt(xOrder)
		require.True(t, ok)
		require.Equal(t, 2, xo)
	})
}

func TestAnalyzeSpec_AutoPropertyOrder(t *testing.T) {
	defer discardOutput()()

	specPath := filepath.Join("..", "fixtures", "codegen", "auto-property-order.yml")

	opts := testGenOpts()
	opts.Spec = specPath
	opts.ValidateSpec = false
	specDoc, _, err := opts.analyzeSpec()
	require.NoError(t, err)

	// Verify that analyzeSpec automatically applied property order
	schema := specDoc.Spec().Definitions["orderedModel"]

	xo, ok := schema.Properties["zeta"].Extensions.GetInt(xOrder)
	require.True(t, ok)
	require.Equal(t, 0, xo)

	xo, ok = schema.Properties["epsilon"].Extensions.GetInt(xOrder)
	require.True(t, ok)
	require.Equal(t, 4, xo)
}
