package frontend

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mgnsk/gong/strictyaml"
)

// Compile YAML bytes to gong script.
func Compile(b []byte) (string, error) {
	jsonDoc, err := strictyaml.UnmarshalToJSON(b, validator)
	if err != nil {
		return "", err
	}

	var buf strings.Builder

	lines, err := render(jsonDoc)
	if err != nil {
		// Already caught by validation.
		panic(err)
	}

	for _, line := range lines {
		buf.WriteString(line.output)
	}

	return buf.String(), nil
}

type outputLine struct {
	path   string
	output string
}

type assignment struct {
	note string
	key  int
}

type invalidInstrumentError struct {
	instName string
	path     string
}

func (e invalidInstrumentError) Error() string {
	return fmt.Sprintf("instrument '%s' not defined", e.instName)
}

// render renders the JSON document. Invalid keys are skipped.
func render(doc map[string]interface{}) ([]outputLine, error) {
	var lines []outputLine
	channels := map[string]int{}

	if instruments, ok := doc["instruments"].(map[string]interface{}); ok {
		names := make([]string, 0, len(instruments))
		for instName := range instruments {
			names = append(names, instName)
		}
		sort.Strings(names)

		for _, instName := range names {
			if inst, ok := instruments[instName].(map[string]interface{}); ok {
				if v, ok := inst["channel"].(float64); ok {
					channels[instName] = int(v)

					lines = append(lines, outputLine{
						path:   fmt.Sprintf("/instruments/%s/channel", instName),
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
							path:   fmt.Sprintf("/instruments/%s/assign/%s", instName, s.note),
							output: fmt.Sprintf("assign %s %d\n", s.note, s.key),
						})
					}
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

				if v, ok := bar["tempo"].(float64); ok {
					lines = append(lines, outputLine{
						path:   fmt.Sprintf("/bars/%d/tempo", barIndex),
						output: fmt.Sprintf("tempo %d\n", int(v)),
					})
				}

				if params, ok := bar["params"].(map[string]interface{}); ok {
					names := make([]string, 0, len(params))
					for instName := range params {
						names = append(names, instName)
					}
					sort.Strings(names)

					for _, instName := range names {
						path := fmt.Sprintf("/bars/%d/params/%s", barIndex, instName)

						channel, ok := channels[instName]
						if !ok {
							return nil, invalidInstrumentError{instName, path}
						}

						lines = append(lines, outputLine{
							path:   path,
							output: fmt.Sprintf("channel %d\n", channel),
						})

						if param, ok := params[instName].(map[string]interface{}); ok {
							if v, ok := param["program"].(float64); ok {
								lines = append(lines, outputLine{
									path:   fmt.Sprintf("/bars/%d/params/%s/program", barIndex, instName),
									output: fmt.Sprintf("program %d\n", int(v)),
								})
							}

							if control, ok := param["control"].(float64); ok {
								if parameter, ok := param["parameter"].(float64); ok {
									lines = append(lines, outputLine{
										path:   fmt.Sprintf("/bars/%d/params/%s/control", barIndex, instName),
										output: fmt.Sprintf("control %d %d\n", int(control), int(parameter)),
									})
								}
							}
						}
					}
				}

				if tracks, ok := bar["tracks"].(map[string]interface{}); ok {
					names := make([]string, 0, len(tracks))
					for instName := range tracks {
						names = append(names, instName)
					}
					sort.Strings(names)

					for _, instName := range names {
						path := fmt.Sprintf("/bars/%d/tracks/%s", barIndex, instName)

						channel, ok := channels[instName]
						if !ok {
							return nil, invalidInstrumentError{instName, path}
						}

						lines = append(lines, outputLine{
							path:   fmt.Sprintf("/bars/%d/tracks/%s", barIndex, instName),
							output: fmt.Sprintf("channel %d\n", channel),
						})

						if voices, ok := tracks[instName].([]interface{}); ok {
							for voiceIndex, voice := range voices {
								lines = append(lines, outputLine{
									path:   fmt.Sprintf("/bars/%d/tracks/%s/%d", barIndex, instName, voiceIndex),
									output: voice.(string) + "\n",
								})
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

	if playList, ok := doc["play"].([]interface{}); ok {
		for i, play := range playList {
			lines = append(lines, outputLine{
				path:   fmt.Sprintf("/play/%d", i),
				output: fmt.Sprintf("\nplay \"%v\"\n", play),
			})
		}
	}

	return lines, nil
}
