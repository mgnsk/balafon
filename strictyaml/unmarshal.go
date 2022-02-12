package strictyaml

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v5"
	kyaml "sigs.k8s.io/yaml"
)

// UnmarshalToJSON unmarshals YAML to JSON with schema validation.
func UnmarshalToJSON(b []byte, sch *jsonschema.Schema) (map[string]interface{}, error) {
	var yamlDoc map[string]interface{}

	if err := yaml.UnmarshalWithOptions(b, &yamlDoc, yaml.Strict()); err != nil {
		return nil, fmt.Errorf(yaml.FormatError(err, true, true))
	}

	jsonBytes, err := kyaml.YAMLToJSON(b)
	if err != nil {
		return nil, err
	}

	var jsonDoc map[string]interface{}

	if err := json.Unmarshal(jsonBytes, &jsonDoc); err != nil {
		return nil, err
	}

	if err := sch.Validate(jsonDoc); err != nil {
		var verr *jsonschema.ValidationError
		if errors.As(err, &verr) {
			var format strings.Builder

			for _, e := range verr.BasicOutput().Errors {
				// Skip generic jsonschema errors.
				if !strings.HasPrefix(e.Error, "doesn't validate with") {
					if match := additionalPropertiesPattern.FindStringSubmatch(e.Error); len(match) == 2 {
						// Add the invalid path element for annotation.
						e.InstanceLocation = e.InstanceLocation + "/" + match[1]
					}

					path, err := yaml.PathString(jsonPathToYAML(e.InstanceLocation))
					if err != nil {
						return nil, err
					}

					res, err := path.AnnotateSource(b, true)
					if err != nil {
						return nil, err
					}

					format.WriteString(fmt.Sprintf("%s:\n%s\n", e.Error, string(res)))
				}
			}

			if format.Len() == 0 {
				panic("invalid jsonschema error")
			}

			return nil, fmt.Errorf("%s", format.String())
		}

		return nil, err
	}

	return jsonDoc, nil
}

var additionalPropertiesPattern = regexp.MustCompile(`additionalProperties '(.*)' not allowed`)

func jsonPathToYAML(path string) string {
	var format strings.Builder
	format.WriteString("$")

	for _, elem := range strings.Split(path, "/")[1:] {
		num, err := strconv.ParseUint(elem, 10, 64)
		if err != nil {
			format.WriteString("." + elem)
		} else {
			format.WriteString(fmt.Sprintf("[%d]", num))
		}
	}

	return format.String()
}
