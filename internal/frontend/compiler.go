package frontend

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v5"
	. "sigs.k8s.io/yaml"
)

//go:embed schema.json
var schema []byte

var validator *jsonschema.Schema

func init() {
	c := jsonschema.NewCompiler()
	c.Draft = jsonschema.Draft7

	if err := c.AddResource("/schema.json", bytes.NewReader(schema)); err != nil {
		panic(err)
	}

	sch, err := c.Compile("/schema.json")
	if err != nil {
		panic(err)
	}

	validator = sch
}

var _ = jsonschema.Schema{}

func Compile(b []byte) ([]byte, error) {
	d := yaml.NewDecoder(bytes.NewReader(b),
		yaml.DisallowDuplicateKey(),
	)

	var yamlValue map[string]interface{}
	if err := d.Decode(&yamlValue); err != nil {
		return nil, err
	}

	jsonBytes, err := YAMLToJSON(b)
	if err != nil {
		panic(err)
	}

	var v map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &v); err != nil {
		panic(err)
	}

	if err := validator.Validate(v); err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}

	for _, instrument := range v["instruments"].([]interface{}) {
		inst := instrument.(map[string]interface{})
		writeIntCommand(buf, inst, "channel")
		writeIntCommand(buf, inst, "program")
		writeControl(buf, inst)
		writeAssignments(buf, inst)
	}

	for _, bar := range v["bars"].([]interface{}) {
		bar := bar.(map[string]interface{})

		buf.WriteString(fmt.Sprintf("\nbar \"%s\"\n", bar["name"]))

		writeTimeSig(buf, bar)

		for _, track := range bar["tracks"].([]interface{}) {
			track := track.(map[string]interface{})

			writeIntCommand(buf, track, "channel")
			writeIntCommand(buf, track, "program")
			writeControl(buf, track)

			if voices, ok := track["voices"].([]interface{}); ok {
				for _, voice := range voices {
					buf.WriteString(voice.(string) + "\n")
				}
			}
		}

		buf.WriteString("end\n")
	}

	for _, play := range v["play"].([]interface{}) {
		buf.WriteString(fmt.Sprintf("\nplay \"%s\"\n", play))
	}

	return buf.Bytes(), nil
}

type assignment struct {
	note string
	key  int
}

func writeAssignments(buf *bytes.Buffer, inst map[string]interface{}) {
	assign := inst["assign"].(map[string]interface{})
	assignments := make([]assignment, len(assign))
	i := 0
	for note, key := range inst["assign"].(map[string]interface{}) {
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
		buf.WriteString(fmt.Sprintf("assign %s %d\n", s.note, s.key))
	}
}

func writeIntCommand(buf *bytes.Buffer, inst map[string]interface{}, cmd string) {
	if v, ok := inst[cmd].(float64); ok {
		buf.WriteString(fmt.Sprintf("%s %d\n", cmd, int(v)))
	}
}

func writeTimeSig(buf *bytes.Buffer, bar map[string]interface{}) {
	if time, ok := bar["time"].(float64); ok {
		if sig, ok := bar["sig"].(float64); ok {
			buf.WriteString(fmt.Sprintf("timesig %d %d\n", int(time), int(sig)))
		}
	}
}

func writeControl(buf *bytes.Buffer, inst map[string]interface{}) {
	if control, ok := inst["control"].(float64); ok {
		if parameter, ok := inst["parameter"].(float64); ok {
			buf.WriteString(fmt.Sprintf("control %d %d\n", int(control), int(parameter)))
		}
	}
}
