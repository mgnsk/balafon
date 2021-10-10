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
	helpText, err := exec.Command("go", "run", "../../.", "--help").CombinedOutput()
	if err != nil {
		log.Fatal(err, string(helpText))
	}

	if output, err := exec.Command("go", "run", "../../.", "lint", "../../examples/bonham").CombinedOutput(); err != nil {
		log.Fatal(err, string(output))
	}

	if output, err := exec.Command("go", "run", "../../.", "lint", "../../examples/bach").CombinedOutput(); err != nil {
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

	data := readmeData{
		HelpSection:   string(helpText),
		BonhamExample: string(bonhamExample),
		BachExample:   string(bachExample),
	}

	return data
}
