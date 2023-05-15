package main

import (
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
func (r *Analyzer) Run() (err error) {
	path := path.Join(
		Dir,
		"opt",
		"konveyor-analyzer")
	cmd := command.Command{Path: path}
	cmd.Options, err = r.options()
	if err != nil {
		return
	}
	err = cmd.Run()
	return
}

//
// options builds Analyzer options.
func (r *Analyzer) options() (options command.Options, err error) {
	settings := &Settings{}
	err = settings.Read(SettingsPath)
	if err != nil {
		return
	}
	options = command.Options{
		"--provider-settings",
		SettingsPath,
		"--output-file",
		ReportPath,
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
func (r *DepAnalyzer) Run() (err error) {
	path := path.Join(
		Dir,
		"opt",
		"konveyor-analyzer-dep")
	cmd := command.Command{Path: path}
	cmd.Options, err = r.options()
	if err != nil {
		return
	}
	err = cmd.Run()
	return
}

//
// options builds Analyzer options.
func (r *DepAnalyzer) options() (options command.Options, err error) {
	settings := &Settings{}
	err = settings.Read(SettingsPath)
	if err != nil {
		return
	}
	options = command.Options{
		"--provider-settings",
		SettingsPath,
		"--output-file",
		DepsPath,
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
