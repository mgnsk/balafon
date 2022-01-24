package frontend

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v5"
	. "sigs.k8s.io/yaml"
)

// Compile YAML bytes to gong script.
func Compile(b []byte) ([]byte, error) {
	var yamlDoc map[string]interface{}

	if err := yaml.UnmarshalWithOptions(b, &yamlDoc, yaml.Strict()); err != nil {
		return nil, fmt.Errorf(yaml.FormatError(err, true, true))
	}

	jsonBytes, err := YAMLToJSON(b)
	if err != nil {
		panic(err)
	}

	var jsonDoc map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonDoc); err != nil {
		panic(err)
	}

	if err := validator.Validate(jsonDoc); err != nil {
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
						panic(err)
					}

					res, err := path.AnnotateSource(b, true)
					if err != nil {
						panic(err)
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

	var buf strings.Builder
	for _, line := range render(jsonDoc) {
		buf.WriteString(line.output)
	}

	return []byte(buf.String()), nil
}

type outputLine struct {
	path   string
	output string
}

type assignment struct {
	note string
	key  int
}

// render renders the JSON document. Invalid keys are skipped.
func render(doc map[string]interface{}) []outputLine {
	var lines []outputLine

	if instruments, ok := doc["instruments"].([]interface{}); ok {
		for instIndex, instrument := range instruments {
			if inst, ok := instrument.(map[string]interface{}); ok {
				if v, ok := inst["channel"].(float64); ok {
					lines = append(lines, outputLine{
						path:   fmt.Sprintf("/instruments/%d/channel", instIndex),
						output: fmt.Sprintf("channel %d\n", int(v)),
					})
				}

				if assign, ok := inst["assign"].(map[string]interface{}); ok {
					assignments := make([]assignment, len(assign))
					i := 0
					for note, key := range assign {
						assignments[i] = assignment{
							note: note,
							key:  int(key.(float64)),
						}
						i++
					}

					sort.Slice(assignments, func(i, j int) bool {
						return assignments[i].key < assignments[j].key
					})

					for _, s := range assignments {
						lines = append(lines, outputLine{
							path:   fmt.Sprintf("/instruments/%d/assign/%s", instIndex, s.note),
							output: fmt.Sprintf("assign %s %d\n", s.note, s.key),
						})
					}
				}
			}
		}

		if bars, ok := doc["bars"].([]interface{}); ok {
			for barIndex, bar := range bars {
				if bar, ok := bar.(map[string]interface{}); ok {
					lines = append(lines, outputLine{
						path:   fmt.Sprintf("/bars/%d", barIndex),
						output: fmt.Sprintf("\nbar \"%v\"\n", bar["name"]),
					})

					if time, ok := bar["time"].(float64); ok {
						if sig, ok := bar["sig"].(float64); ok {
							lines = append(lines, outputLine{
								path:   fmt.Sprintf("/bars/%d/time", barIndex),
								output: fmt.Sprintf("timesig %d %d\n", int(time), int(sig)),
							})
						}
					}

					if params, ok := bar["params"].([]interface{}); ok {
						for paramIndex, param := range params {
							if param, ok := param.(map[string]interface{}); ok {
								if v, ok := param["channel"].(float64); ok {
									lines = append(lines, outputLine{
										path:   fmt.Sprintf("/bars/%d/params/%d/channel", barIndex, paramIndex),
										output: fmt.Sprintf("channel %d\n", int(v)),
									})
								}

								if v, ok := param["tempo"].(float64); ok {
									lines = append(lines, outputLine{
										path:   fmt.Sprintf("/bars/%d/params/%d/tempo", barIndex, paramIndex),
										output: fmt.Sprintf("tempo %d\n", int(v)),
									})
								}

								if v, ok := param["program"].(float64); ok {
									lines = append(lines, outputLine{
										path:   fmt.Sprintf("/bars/%d/params/%d/program", barIndex, paramIndex),
										output: fmt.Sprintf("program %d\n", int(v)),
									})
								}

								if control, ok := param["control"].(float64); ok {
									if parameter, ok := param["parameter"].(float64); ok {
										lines = append(lines, outputLine{
											path:   fmt.Sprintf("/bars/%d/params/%d/control", barIndex, paramIndex),
											output: fmt.Sprintf("control %d %d\n", int(control), int(parameter)),
										})
									}
								}
							}
						}
					}

					if tracks, ok := bar["tracks"].([]interface{}); ok {
						for trackIndex, track := range tracks {
							if track, ok := track.(map[string]interface{}); ok {
								if v, ok := track["channel"].(float64); ok {
									lines = append(lines, outputLine{
										path:   fmt.Sprintf("/bars/%d/tracks/%d/channel", barIndex, trackIndex),
										output: fmt.Sprintf("channel %d\n", int(v)),
									})
								}

								if voices, ok := track["voices"].([]interface{}); ok {
									for voiceIndex, voice := range voices {
										lines = append(lines, outputLine{
											path:   fmt.Sprintf("/bars/%d/tracks/%d/voices/%d", barIndex, trackIndex, voiceIndex),
											output: voice.(string) + "\n",
										})
									}
								}
							}
						}
					}

					lines = append(lines, outputLine{
						path:   fmt.Sprintf("/bars/%d", barIndex),
						output: "end\n",
					})
				}
			}
		}
	}

	if playList, ok := doc["play"].([]interface{}); ok {
		for i, play := range playList {
			lines = append(lines, outputLine{
				path:   fmt.Sprintf("/play/%d", i),
				output: fmt.Sprintf("\nplay \"%v\"\n", play),
			})
		}
	}

	return lines
}
