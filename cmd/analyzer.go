package main

import (
	"encoding/json"
	"github.com/konveyor/analyzer-lsp/dependency/dependency"
	"github.com/konveyor/analyzer-lsp/hubapi"
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-hub/api"
	"go.lsp.dev/uri"
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
func (r *DepAnalyzer) Run() (report DepReport, err error) {
	bin := path.Join(
		Dir,
		"opt",
		"konveyor-analyzer-dep")
	output := path.Join(Dir, "deps.yaml")
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
// Write issues file.
func (r *Report) Write(path string) (err error) {
	writer, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		_ = writer.Close()
	}()
	uriStr := func(in uri.URI) string {
		defer func() {
			recover()
		}()
		return in.Filename()
	}
	encoder := json.NewEncoder(writer)
	for _, ruleset := range *r {
		for ruleid, v := range ruleset.Violations {
			issue := api.Issue{
				RuleSet:     ruleset.Name,
				Rule:        ruleid,
				Description: v.Description,
				Labels:      v.Labels,
			}
			if v.Category != nil {
				issue.Category = string(*v.Category)
			}
			if v.Effort != nil {
				issue.Effort = *v.Effort
			}
			issue.Links = []api.Link{}
			for _, l := range v.Links {
				issue.Links = append(
					issue.Links,
					api.Link{
						URL:   l.URL,
						Title: l.Title,
					})
			}
			issue.Incidents = []api.Incident{}
			for _, i := range v.Incidents {
				incident := api.Incident{
					File:     uriStr(i.URI),
					Message:  i.Message,
					CodeSnip: i.CodeSnip,
					Facts:    i.Variables,
				}
				issue.Incidents = append(
					issue.Incidents,
					incident)
			}
			_ = encoder.Encode(&issue)
		}
	}
	return
}

//
// Facts builds facts.
func (r *Report) Facts() (facts []api.Fact) {
	for _, r := range *r {
		for _, v := range r.Violations {
			mp := make(map[string]interface{})
			_ = json.Unmarshal(v.Extras, &mp)
			for k, v := range mp {
				facts = append(
					facts,
					api.Fact{
						Key:   k,
						Value: v,
					})
			}
		}
	}
	return
}

//
// DepReport analysis report file.
type DepReport []dependency.Dep

//
// Read file.
func (r *DepReport) Read(path string) (err error) {
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
// Write deps file.
func (r *DepReport) Write(path string) (err error) {
	writer, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		_ = writer.Close()
	}()
	encoder := json.NewEncoder(writer)
	deps := []api.TechDependency{}
	for _, d := range *r {
		deps = append(
			deps,
			api.TechDependency{
				Indirect: d.Indirect,
				Name:     d.Name,
				Version:  d.Version,
				SHA:      d.SHA,
			})
	}
	_ = encoder.Encode(deps)
	return
}
