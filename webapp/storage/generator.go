//+build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	testCaseBlobFileName    string = "testcases_blob.go"
	expectationBlobFileName string = "expectation_blob.go"
	stageBlobFileName       string = "stage_blob.go"
	embedTestCaseFolder     string = "../../default-e2e-test-data/testcases"
	embedExpectationFolder  string = "../../default-e2e-test-data/expectations"
	embedStageFolder        string = "../../default-e2e-test-data/stages"
)

var conv = map[string]interface{}{"conv": fmtByteSlice}

var testCaseTmpl = template.Must(template.New("").Funcs(conv).Parse(`package storage
// Code generated by go generate; DO NOT EDIT.
func init() {
	{{- range $name, $file := . }}
    	DefaultTestCaseStore.Add("{{ $name }}", []byte{ {{ conv $file }} })
	{{- end }}
}`),
)

var expectationTmpl = template.Must(template.New("").Funcs(conv).Parse(`package storage
// Code generated by go generate; DO NOT EDIT.
func init() {
	{{- range $name, $file := . }}
    	DefaultExpectationStore.Add("{{ $name }}", []byte{ {{ conv $file }} })
	{{- end }}
}`),
)

var stageTmpl = template.Must(template.New("").Funcs(conv).Parse(`package storage
// Code generated by go generate; DO NOT EDIT.
func init() {
	{{- range $name, $file := . }}
    	DefaultStageStore.Add("{{ $name }}", []byte{ {{ conv $file }} })
	{{- end }}
}`),
)

func fmtByteSlice(s []byte) string {
	builder := strings.Builder{}

	for _, v := range s {
		builder.WriteString(fmt.Sprintf("%d,", int(v)))
	}

	return builder.String()
}

type executor interface {
	Execute(builder io.Writer, configs interface{}) error
}

func generateBlob(fpath string, bpath string, tmpl executor) error {
	// Checking directory with files
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		log.Fatal("Static directory does not exists!")
	}

	// Create map for filenames
	configs := make(map[string][]byte)

	// Walking through embed directory
	err := filepath.Walk(fpath, func(path string, info os.FileInfo, err error) error {
		relativePath := filepath.ToSlash(strings.TrimPrefix(path, fpath))

		if info.IsDir() {
			// Skip directories
			log.Println(path, "is a directory, skipping...")
			return nil
		} else {
			// If element is a simple file, embed
			log.Println(path, "is a file, packing in...")

			b, err := ioutil.ReadFile(path)
			if err != nil {
				// If file not reading
				log.Printf("Error reading %s: %s", path, err)
				return err
			}

			// Add file name to map
			configs[relativePath] = b
		}

		return nil
	})
	if err != nil {
		log.Fatal("Error walking through embed directory:", err)
	}

	// Create blob file
	f, err := os.Create(bpath)
	if err != nil {
		log.Fatal("Error creating blob file:", err)
	}
	defer f.Close()

	// Create buffer
	builder := &bytes.Buffer{}

	// Execute template
	if err = tmpl.Execute(builder, configs); err != nil {
		log.Fatal("Error executing template", err)
	}

	// Formatting generated code
	data, err := format.Source(builder.Bytes())
	if err != nil {
		log.Fatal("Error formatting generated code", err)
	}

	// Writing blob file
	if err = ioutil.WriteFile(bpath, data, os.ModePerm); err != nil {
		log.Fatal("Error writing blob file", err)
	}

	return nil
}

func main() {
	if err := generateBlob(embedTestCaseFolder, testCaseBlobFileName, testCaseTmpl); err != nil {
		log.Fatal(err)
	}

	if err := generateBlob(embedExpectationFolder, expectationBlobFileName, expectationTmpl); err != nil {
		log.Fatal(err)
	}

	if err := generateBlob(embedStageFolder, stageBlobFileName, stageTmpl); err != nil {
		log.Fatal(err)
	}
}