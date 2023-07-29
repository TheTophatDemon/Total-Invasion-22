//go:build ignore

package main

import (
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func snakeToPascalCase(str string) string {
	var b strings.Builder
	for _, word := range strings.Split(str, "_") {
		var capitalized string
		if len(word) > 1 {
			capitalized = strings.ToUpper(string(word[0])) + word[1:]
		} else {
			capitalized = strings.ToUpper(word)
		}
		b.WriteString(capitalized)
	}
	return b.String()
}

func main() {
	// Get template file
	inFile, err := os.Open("declarations.go.tmpl")
	panicOnErr(err)
	defer inFile.Close()

	// Parse into string
	templateString, err := ioutil.ReadAll(inFile)
	panicOnErr(err)

	// Make template
	funcs := template.FuncMap{
		"toCamel": func(str string) string {
			return strings.ToLower(string(str[0])) + str[1:]
		},
	}
	tmpl := template.Must(template.New("list").Funcs(funcs).Parse(string(templateString)))

	// Generated output file
	outFile, err := os.Create("declarations.go")
	panicOnErr(err)
	defer outFile.Close()

	// Get a list of all the component code files in this package
	fileEntries, err := os.ReadDir(".")
	panicOnErr(err)

	componentNames := make([]string, 0, len(fileEntries))

	for _, entry := range fileEntries {
		if strings.HasSuffix(entry.Name(), ".go") && strings.HasPrefix(entry.Name(), "comp_") {
			// Files that have the pattern comp_*.go are considered components and must contain a type with the remaining file name in PascalCase
			componentNames = append(componentNames,
				snakeToPascalCase(
					strings.TrimPrefix(
						strings.TrimSuffix(entry.Name(), ".go"), "comp_")))
		}
	}

	// Execute template
	err = tmpl.Execute(outFile, componentNames)
	panicOnErr(err)
}
