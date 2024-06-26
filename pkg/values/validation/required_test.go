package validation_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/flant/addon-operator/pkg/module_manager/models/modules"
	"github.com/flant/addon-operator/pkg/utils"
)

func Test_Transform_Required(t *testing.T) {
	g := NewWithT(t)
	var err error

	configValuesYaml := `
type: object
required:
- param1
properties:
  param1:
    type: string
    enum:
    - val1
  param2:
    type: string
`
	valuesYaml := `
x-extend:
  schema: "config-values.yaml"
x-required-for-helm:
- internal
type: object
properties:
  internal:
    type: object
    required:
    - param1
    x-required-for-helm:
      - param2
      - param3
    properties:
      param1:
        type: string
      param2:
        type: string
      param3:
        type: string
`

	valuesStorage, err := modules.NewValuesStorage("moduleName", nil, []byte(configValuesYaml), []byte(valuesYaml))
	g.Expect(err).ShouldNot(HaveOccurred())

	var moduleValues utils.Values

	// Intermediate values after hook execution.
	moduleValues, err = utils.NewValuesFromBytes([]byte(`
moduleName:
  param1: val1
  internal:
    param1: val1
    param2: val2
`))
	g.Expect(err).ShouldNot(HaveOccurred())

	// Values contract is satisfied, param1 is present.
	mErr := valuesStorage.GetSchemaStorage().ValidateValues("moduleName", moduleValues)
	g.Expect(mErr).ShouldNot(HaveOccurred())

	// Helm contract is not satisfied — no internal.param3 field.
	mErr = valuesStorage.GetSchemaStorage().ValidateModuleHelmValues("moduleName", moduleValues)
	g.Expect(mErr).Should(HaveOccurred())

	// Intermediate values after another hook execution.
	moduleValues, err = utils.NewValuesFromBytes([]byte(`
moduleName:
  param1: val1
  internal:
    param1: val1
    param3: val2
`))
	g.Expect(err).ShouldNot(HaveOccurred())

	// Values contract is satisfied, param1 is present.
	mErr = valuesStorage.GetSchemaStorage().ValidateValues("moduleName", moduleValues)
	g.Expect(mErr).ShouldNot(HaveOccurred())

	// Helm contract is not satisfied — no internal.param2 field.
	mErr = valuesStorage.GetSchemaStorage().ValidateModuleHelmValues("moduleName", moduleValues)
	g.Expect(mErr).Should(HaveOccurred())

	// Effective values before helm execution.
	moduleValues, err = utils.NewValuesFromBytes([]byte(`
moduleName:
  param1: val1
  internal:
    param1: val1
    param2: val44
    param3: val2
`))
	g.Expect(err).ShouldNot(HaveOccurred())

	// Values contract is satisfied, param1 is present.
	mErr = valuesStorage.GetSchemaStorage().ValidateValues("moduleName", moduleValues)
	g.Expect(mErr).ShouldNot(HaveOccurred())

	// Helm contract is now satisfied.
	mErr = valuesStorage.GetSchemaStorage().ValidateModuleHelmValues("moduleName", moduleValues)
	g.Expect(mErr).ShouldNot(HaveOccurred())
}
