package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var readmeTpl = template.Must(template.New("").Funcs(template.FuncMap{
	"indent": func(places int, input string) string {
		lines := strings.Split(input, "\n")
		result := make([]string, len(lines))
		for i, line := range lines {
			if len(line) > 0 {
				result[i] = strings.Repeat(" ", places) + line
			}
		}
		return strings.Join(result, "\n")
	},
	"trim_trailing_newlines": func(input string) string {
		return strings.TrimRight(input, "\n")
	},
}).ParseFiles("README.md.tpl"))

type readmeData struct {
	HelpSection   string
	BonhamExample string
	BachExample   string
	MultiExample  string
	YAMLExample   string
}

func main() {
	root := &cobra.Command{
		Short: "tools",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	root.AddCommand(&cobra.Command{
		Use:   "gen-readme",
		Short: "Generate README.md",
		RunE: func(c *cobra.Command, _ []string) error {
			f, err := os.Create("../../README.md")
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			return readmeTpl.ExecuteTemplate(f, "README.md.tpl", createReadmeData())
		},
	})

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func createReadmeData() readmeData {
	helpText, err := exec.Command("go", "run", "../../cmd/gong/.", "--help").CombinedOutput()
	if err != nil {
		log.Fatal(err, string(helpText))
	}

	if output, err := exec.Command("go", "run", "../../cmd/gonglint/.", "../../examples/bonham").CombinedOutput(); err != nil {
		log.Fatal(err, string(output))
	}

	if output, err := exec.Command("go", "run", "../../cmd/gonglint/.", "../../examples/bach").CombinedOutput(); err != nil {
		log.Fatal(err, string(output))
	}

	if output, err := exec.Command("go", "run", "../../cmd/gonglint/.", "../../examples/multichannel").CombinedOutput(); err != nil {
		log.Fatal(err, string(output))
	}

	if output, err := exec.Command("bash", "-c", "set -eou pipefail; go run ../../cmd/yaml2gong/. ../../examples/example.yml | go run ../../cmd/gonglint/. -").CombinedOutput(); err != nil {
		log.Fatal(err, string(output))
	}

	bonhamExample, err := ioutil.ReadFile("../../examples/bonham")
	if err != nil {
		log.Fatal(err)
	}

	bachExample, err := ioutil.ReadFile("../../examples/bach")
	if err != nil {
		log.Fatal(err)
	}

	multiExample, err := ioutil.ReadFile("../../examples/multichannel")
	if err != nil {
		log.Fatal(err)
	}

	yamlExample, err := ioutil.ReadFile("../../examples/example.yml")
	if err != nil {
		log.Fatal(err)
	}

	data := readmeData{
		HelpSection:   string(helpText),
		BonhamExample: string(bonhamExample),
		BachExample:   string(bachExample),
		MultiExample:  string(multiExample),
		YAMLExample:   string(yamlExample),
	}

	return data
}
