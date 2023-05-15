package main

import (
	"github.com/konveyor/tackle2-addon/ssh"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/nas"
	"os"
	"path"
)

var (
	addon        = hub.Addon
	HomeDir      = ""
	BinDir       = ""
	SourceDir    = ""
	Dir          = ""
	M2Dir        = ""
	ReportPath   = ""
	DepsPath     = ""
	RuleDir      = ""
	SettingsPath = ""
)

func init() {
	Dir, _ = os.Getwd()
	HomeDir, _ = os.UserHomeDir()
	SourceDir = path.Join(Dir, "source")
	BinDir = path.Join(Dir, "bin")
	ReportPath = path.Join(Dir, "report.yaml")
	DepsPath = path.Join(Dir, "deps.yaml")
	RuleDir = path.Join(Dir, "rules")
	M2Dir = "/cache/m2"
	SettingsPath = path.Join(Dir, "opt", "settings.yaml")
}

type SoftError = hub.SoftError

//
// Data Addon data passed in the secret.
type Data struct {
	// Output directory within application bucket.
	Output string `json:"output" binding:"required"`
	// Mode options.
	Mode Mode `json:"mode"`
	// Labels list.
	Labels Labels `json:"labels"`
	// Scope options.
	Scope Scope `json:"scope"`
	// Rules options.
	Rules *Rules `json:"rules"`
	// Tagger options.
	Tagger Tagger `json:"tagger"`
}

//
// main
func main() {
	addon.Run(func() (err error) {
		//
		// Get the addon data associated with the task.
		d := &Data{}
		err = addon.DataWith(d)
		if err != nil {
			err = &SoftError{Reason: err.Error()}
			return
		}
		//
		// Create directories.
		for _, dir := range []string{BinDir, M2Dir, RuleDir} {
			err = nas.MkDir(dir, 0755)
			if err != nil {
				return
			}
		}
		//
		// Fetch application.
		addon.Activity("Fetching application.")
		application, err := addon.Task.Application()
		if err != nil {
			return
		}
		//
		// SSH
		agent := ssh.Agent{}
		err = agent.Start()
		if err != nil {
			return
		}
		//
		// Build assets.
		err = d.Mode.Build(application)
		if err != nil {
			return
		}
		err = d.Rules.Build()
		if err != nil {
			return
		}
		//
		// Run analysis.
		analyzer := Analyzer{}
		analyzer.Data = d
		err = analyzer.Run()
		if err != nil {
			return
		}
		//
		// Tagging.
		if d.Tagger.Enabled {
			var report Report
			err = report.Read(ReportPath)
			if err != nil {
				return
			}
			err = d.Tagger.Update(application.ID, report)
			if err != nil {
				return
			}
		}
		//
		// Find deps.
		depAnalyzer := DepAnalyzer{}
		depAnalyzer.Data = d
		err = depAnalyzer.Run()
		if err != nil {
			return
		}
		return
	})
}
