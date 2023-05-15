package main

import (
	"github.com/konveyor/analyzer-lsp/provider/lib"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

//
// Settings - provider settings file.
type Settings []lib.Config

//
// Read file.
func (r *Settings) Read(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := io.ReadAll(f)
	err = yaml.Unmarshal(b, r)
	return
}

//
// Write file.
func (r *Settings) Write(path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := yaml.Marshal(r)
	if err != nil {
		return
	}
	_, err = f.Write(b)
	return
}

//
// Location update the location on each provider.
func (r *Settings) Location(path string) {
	for i := range *r {
		p := &(*r)[i]
		p.Location = path
	}
}
