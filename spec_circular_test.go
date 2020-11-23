package loads_test

import (
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/stretchr/testify/assert"
)

func TestPathLoaderIssue(t *testing.T) {
	swaggerFile := "fixtures/json/resources/pathLoaderIssue.json"
	document, err := loads.Spec(swaggerFile)
	assert.NoError(t, err)
	assert.NotNil(t, document)
	validationErrs := validate.Spec(document, strfmt.Default)
	assert.NoError(t, validationErrs)
}
