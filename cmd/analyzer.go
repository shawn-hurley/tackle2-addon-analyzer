package main

import (
	"github.com/konveyor/analyzer-lsp/dependency/dependency"
	"github.com/konveyor/analyzer-lsp/hubapi"
	"github.com/konveyor/tackle2-addon/command"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
)

//
// Analyzer application analyzer.
type Analyzer struct {
	*Data
}

//
// Run analyzer.
func (r *Analyzer) Run() (report Report, err error) {
	bin := path.Join(
		Dir,
		"opt",
		"konveyor-analyzer")
	output := path.Join(Dir, "report.yaml")
	cmd := command.Command{Path: bin}
	cmd.Options, err = r.options(output)
	if err != nil {
		return
	}
	err = cmd.Run()
	if err != nil {
		return
	}
	err = report.Read(output)
	return
}

//
// options builds Analyzer options.
func (r *Analyzer) options(output string) (options command.Options, err error) {
	settings := &Settings{}
	err = settings.Read(SettingsPath)
	if err != nil {
		return
	}
	options = command.Options{
		"--provider-settings",
		SettingsPath,
		"--output-file",
		output,
	}
	err = r.Tagger.AddOptions(&options)
	if err != nil {
		return
	}
	err = r.Mode.AddOptions(settings)
	if err != nil {
		return
	}
	if r.Rules != nil {
		err = r.Rules.AddOptions(&options)
		if err != nil {
			return
		}
	}
	if r.Labels != nil {
		err = r.Labels.AddOptions(&options)
		if err != nil {
			return
		}
	}
	err = r.Scope.AddOptions(&options)
	if err != nil {
		return
	}
	err = settings.Write(SettingsPath)
	if err != nil {
		return
	}
	return
}

//
// DepAnalyzer application analyzer.
type DepAnalyzer struct {
	*Data
}

//
// Run analyzer.
func (r *DepAnalyzer) Run() (deps Deps, err error) {
	bin := path.Join(
		Dir,
		"opt",
		"konveyor-analyzer-dep")
	output := path.Join(Dir, "report.yaml")
	cmd := command.Command{Path: bin}
	cmd.Options, err = r.options(output)
	if err != nil {
		return
	}
	err = cmd.Run()
	if err != nil {
		return
	}
	err = deps.Read(output)
	return
}

//
// options builds Analyzer options.
func (r *DepAnalyzer) options(output string) (options command.Options, err error) {
	settings := &Settings{}
	err = settings.Read(SettingsPath)
	if err != nil {
		return
	}
	options = command.Options{
		"--provider-settings",
		SettingsPath,
		"--output-file",
		output,
	}
	err = r.Mode.AddOptions(settings)
	if err != nil {
		return
	}
	err = settings.Write(SettingsPath)
	if err != nil {
		return
	}
	return
}

//
// Report analysis report file.
type Report []hubapi.RuleSet

//
// Read file.
func (r *Report) Read(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := io.ReadAll(f)
	err = yaml.Unmarshal(b, &r)
	return
}

//
// Deps analysis report file.
type Deps []dependency.Dep

//
// Read file.
func (r *Deps) Read(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := io.ReadAll(f)
	err = yaml.Unmarshal(b, &r)
	return
}
