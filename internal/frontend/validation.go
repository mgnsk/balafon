package frontend

import (
	"bytes"
	_ "embed"
	"errors"

	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed schema.json
var schema []byte

var interpretMeta = jsonschema.MustCompileString("interpretMeta.json", `{
	"properties": {
		"interpret": true
	}
}`)

type interpreterCompiler struct{}

func (c interpreterCompiler) Compile(_ jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if _, ok := m["interpret"]; ok {
		return interpreterValidator{}, nil
	}
	return nil, nil
}

type interpreterValidator struct {
}

func (ext interpreterValidator) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	switch doc := v.(type) {
	case map[string]interface{}:
		it := interpreter.New()

		var (
			verr    error
			errInst invalidInstrumentError
		)

		lines, err := render(doc)
		if err != nil {
			if errors.As(err, &errInst) {
				return &jsonschema.ValidationError{
					InstanceLocation: errInst.path,
					Message:          errInst.Error(),
				}
			}
			panic(err)
		}

		for _, line := range lines {
			if _, err := it.Eval(line.output); err != nil {
				perr := &jsonschema.ValidationError{
					InstanceLocation: line.path,
					Message:          err.Error(),
				}
				if verr != nil {
					verr = jsonschema.ValidationError{}.Group(verr.(*jsonschema.ValidationError), perr)
				} else {
					verr = perr
				}
			}
		}

		if verr != nil {
			return verr
		}
	default:
	}
	return nil
}

var validator *jsonschema.Schema

func init() {
	c := jsonschema.NewCompiler()
	c.Draft = jsonschema.Draft7

	c.RegisterExtension("interpret", interpretMeta, interpreterCompiler{})

	if err := c.AddResource("/schema.json", bytes.NewReader(schema)); err != nil {
		panic(err)
	}

	sch, err := c.Compile("/schema.json")
	if err != nil {
		panic(err)
	}

	validator = sch
}
