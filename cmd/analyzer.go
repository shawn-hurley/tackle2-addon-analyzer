package main

import (
	"github.com/konveyor/analyzer-lsp/hubapi"
	"github.com/konveyor/analyzer-lsp/provider/lib"
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	pathlib "path"
	"strconv"
	"strings"
	"time"
)

//
// Analyzer application analyzer.
type Analyzer struct {
	application *api.Application
	*Data
}

//
// Run analyzer.
func (r *Analyzer) Run() (err error) {
	path := pathlib.Join(
		Dir,
		"opt",
		"konveyor-analyzer")
	cmd := command.Command{Path: path}
	cmd.Options, err = r.options()
	if err != nil {
		return
	}
	err = cmd.Run()
	addon.Activity("SLEEP")
	time.Sleep(time.Minute * 10)
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
// Mode settings.
type Mode struct {
	Binary     bool   `json:"binary"`
	Artifact   string `json:"artifact"`
	WithDeps   bool   `json:"withDeps"`
	Repository repository.SCM
}

//
// AddOptions adds analyzer options.
func (r *Mode) AddOptions(settings *Settings) (err error) {
	if r.Binary {
		if r.Artifact != "" {
			bucket := addon.Bucket()
			err = bucket.Get(r.Artifact, BinDir)
			if err != nil {
				return
			}
			// TODO: options.Add("--input", BinDir)
		}
	} else {
		settings.Location(AppDir)
	}

	return
}

//
// Labels list of sources.
type Labels []string

//
// AddOptions add options.
func (r Labels) AddOptions(options *command.Options) (err error) {
	for _, source := range r {
		options.Add("--source", source)
	}
	return
}

//
// Scope settings.
type Scope struct {
	WithKnown bool `json:"withKnown"`
	Packages  struct {
		Included []string `json:"included,omitempty"`
		Excluded []string `json:"excluded,omitempty"`
	} `json:"packages"`
}

//
// AddOptions adds analyzer options.
func (r *Scope) AddOptions(options *command.Options) (err error) {
	if r.WithKnown {
		options.Add("--analyzeKnownLibraries")
	}
	if len(r.Packages.Included) > 0 {
		options.Add("--packages", r.Packages.Included...)
	}
	if len(r.Packages.Excluded) > 0 {
		options.Add("--excludePackages", r.Packages.Excluded...)
	}
	return
}

//
// Rules settings.
type Rules struct {
	Path       string          `json:"path" binding:"required"`
	Bundles    []api.Ref       `json:"bundles"`
	Repository *api.Repository `json:"repository"`
	Identity   *api.Ref        `json:"identity"`
	Tags       struct {
		Included []string `json:"included,omitempty"`
		Excluded []string `json:"excluded,omitempty"`
	} `json:"tags"`
}

//
// AddOptions adds analyzer options.
func (r *Rules) AddOptions(options *command.Options) (err error) {
	err = r.addFiles(options)
	if err != nil {
		return
	}
	err = r.addRepository(options)
	if err != nil {
		return
	}
	err = r.addBundles(options)
	if err != nil {
		return
	}
	if len(r.Tags.Included) > 0 {
		options.Add("--includeTags", r.Tags.Included...)
	}
	if len(r.Tags.Excluded) > 0 {
		options.Add("--excludeTags", r.Tags.Excluded...)
	}
	return
}

//
// addFiles add uploaded rules files.
func (r *Rules) addFiles(options *command.Options) (err error) {
	if r.Path == "" {
		return
	}
	ruleDir := pathlib.Join(RuleDir, "/files")
	err = nas.MkDir(ruleDir, 0755)
	if err != nil {
		return
	}
	options.Add("--rules", ruleDir)
	bucket := addon.Bucket()
	err = bucket.Get(r.Path, ruleDir)
	if err != nil {
		return
	}
	return
}

//
// AddBundles adds bundles.
func (r *Rules) addBundles(options *command.Options) (err error) {
	for _, ref := range r.Bundles {
		var bundle *api.RuleBundle
		bundle, err = addon.RuleBundle.Get(ref.ID)
		if err != nil {
			return
		}
		err = r.addRuleSets(options, bundle)
		if err != nil {
			return
		}
		err = r.addBundleRepository(options, bundle)
		if err != nil {
			return
		}
	}
	return
}

//
// addRuleSets adds ruleSets
func (r *Rules) addRuleSets(options *command.Options, bundle *api.RuleBundle) (err error) {
	ruleDir := pathlib.Join(
		RuleDir,
		"/bundles",
		strconv.Itoa(int(bundle.ID)),
		"rulesets")
	err = nas.MkDir(ruleDir, 0755)
	if err != nil {
		return
	}
	files := 0
	for _, ruleset := range bundle.RuleSets {
		fileRef := ruleset.File
		if fileRef == nil {
			continue
		}
		name := strings.Join(
			[]string{
				strconv.Itoa(int(ruleset.ID)),
				fileRef.Name},
			"-")
		path := pathlib.Join(ruleDir, name)
		addon.Activity("[FILE] Get rule: %s", path)
		err = addon.File.Get(ruleset.File.ID, path)
		if err != nil {
			break
		}
		files++
	}
	if files > 0 {
		options.Add("--rules", ruleDir)
	}
	return
}

//
// addBundleRepository adds bundle repository.
func (r *Rules) addBundleRepository(options *command.Options, bundle *api.RuleBundle) (err error) {
	if bundle.Repository == nil {
		return
	}
	rootDir := pathlib.Join(
		RuleDir,
		"/bundles",
		strconv.Itoa(int(bundle.ID)),
		"repository")
	err = nas.MkDir(rootDir, 0755)
	if err != nil {
		return
	}
	var ids []api.Ref
	if bundle.Identity != nil {
		ids = []api.Ref{*bundle.Identity}
	}
	rp, err := repository.New(
		rootDir,
		bundle.Repository,
		ids)
	if err != nil {
		return
	}
	err = rp.Fetch()
	if err != nil {
		return
	}
	ruleDir := pathlib.Join(rootDir, bundle.Repository.Path)
	options.Add("--rules", ruleDir)
	return
}

//
// addRepository adds custom repository.
func (r *Rules) addRepository(options *command.Options) (err error) {
	if r.Repository == nil {
		return
	}
	rootDir := pathlib.Join(
		RuleDir,
		"repository")
	err = nas.MkDir(rootDir, 0755)
	if err != nil {
		return
	}
	var ids []api.Ref
	if r.Identity != nil {
		ids = []api.Ref{*r.Identity}
	}
	rp, err := repository.New(
		rootDir,
		r.Repository,
		ids)
	if err != nil {
		return
	}
	err = rp.Fetch()
	if err != nil {
		return
	}
	ruleDir := pathlib.Join(rootDir, r.Repository.Path)
	options.Add("--rules", ruleDir)
	return
}

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
