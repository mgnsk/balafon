package strictyaml

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v5"
	kyaml "sigs.k8s.io/yaml"
)

// UnmarshalToJSON unmarshals YAML to JSON with schema validation.
func UnmarshalToJSON(b []byte, sch *jsonschema.Schema) (map[string]interface{}, error) {
	var yamlDoc map[string]interface{}

	if err := Unmarshal(b, &yamlDoc, sch); err != nil {
		return nil, err
	}

	jsonBytes, err := kyaml.YAMLToJSON(b)
	if err != nil {
		return nil, err
	}

	var jsonDoc map[string]interface{}

	if err := json.Unmarshal(jsonBytes, &jsonDoc); err != nil {
		return nil, err
	}

	return jsonDoc, nil
}

// Unmarshal unmarshals YAML with JSON schema validation and annotated source errors.
func Unmarshal(b []byte, target interface{}, sch *jsonschema.Schema) error {
	if err := yaml.UnmarshalWithOptions(b, target, yaml.Strict()); err != nil {
		return fmt.Errorf(yaml.FormatError(err, true, true))
	}

	jsonBytes, err := kyaml.YAMLToJSON(b)
	if err != nil {
		return err
	}

	var jsonDoc map[string]interface{}

	if err := json.Unmarshal(jsonBytes, &jsonDoc); err != nil {
		return err
	}

	if err := sch.Validate(jsonDoc); err != nil {
		var verr *jsonschema.ValidationError

		if errors.As(err, &verr) {
			errs := verr.BasicOutput().Errors

			// Get the deepest error.
			sort.Slice(errs, func(i, j int) bool {
				return len(errs[i].InstanceLocation) < len(errs[j].InstanceLocation)
			})

			berr := errs[len(errs)-1]

			if match := additionalPropertiesPattern.FindStringSubmatch(berr.Error); len(match) == 2 {
				// Add the invalid path element for annotation.
				berr.InstanceLocation = berr.InstanceLocation + "/" + match[1]
			}

			return NewError(b, jsonPathToYAML(berr.InstanceLocation), berr.Error)
		}

		return err
	}

	return nil
}

var additionalPropertiesPattern = regexp.MustCompile(`additionalProperties '(.*)' not allowed`)

func jsonPathToYAML(path string) *yaml.Path {
	builder := &yaml.PathBuilder{}
	builder = builder.Root()

	for _, elem := range strings.Split(path, "/")[1:] {
		num, err := strconv.Atoi(elem)
		if err != nil {
			builder.Child(elem)
		} else {
			builder.Index(uint(num))
		}
	}

	return builder.Build()
}
