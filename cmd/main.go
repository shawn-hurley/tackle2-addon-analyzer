package main

import (
	"github.com/konveyor/tackle2-addon/ssh"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
	"os"
	"path"
	"time"
)

var (
	addon        = hub.Addon
	BinDir       = ""
	SourceDir    = ""
	Dir          = ""
	M2Dir        = ""
	RuleDir      = ""
	SettingsPath = ""
	Source       = "Analysis"
)

func init() {
	Dir, _ = os.Getwd()
	SourceDir = path.Join(Dir, "source")
	BinDir = path.Join(Dir, "bin")
	RuleDir = path.Join(Dir, "rules")
	M2Dir = "/cache/m2"
	SettingsPath = path.Join(Dir, "opt", "settings.yaml")
}

type SoftError = hub.SoftError

//
// Data Addon data passed in the secret.
type Data struct {
	// Mode options.
	Mode Mode `json:"mode"`
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
		report, err := analyzer.Run()
		if err != nil {
			return
		}
		depAnalyzer := DepAnalyzer{}
		depAnalyzer.Data = d
		depReport, err := depAnalyzer.Run()
		if err != nil {
			return
		}
		//
		// Post report.
		p := path.Join(Dir, "issues.json")
		err = report.Write(p)
		if err != nil {
			return
		}
		mark := time.Now()
		issueFile, err := addon.File.Put(p)
		if err == nil {
			addon.Activity("[FILE] Uploaded: %s duration: %s", p, time.Since(mark))
		} else {
			return
		}
		p = path.Join(Dir, "deps.json")
		err = depReport.Write(p)
		if err != nil {
			return
		}
		mark = time.Now()
		depFile, err := addon.File.Put(p)
		if err == nil {
			addon.Activity("[FILE] Uploaded: %s duration: %s", p, time.Since(mark))
		} else {
			return
		}
		manifest := &api.AnalysisManifest{}
		manifest.Issues = api.Ref{ID: issueFile.ID}
		manifest.Dependencies = api.Ref{ID: depFile.ID}
		appAnalysis := addon.Application.Analysis(application.ID)
		mark = time.Now()
		err = appAnalysis.Create(manifest)
		if err == nil {
			addon.Activity("Analysis reported. duration: %s", time.Since(mark))
		} else {
			return
		}
		//
		// Tags.
		if d.Tagger.Enabled {
			err = d.Tagger.Update(application.ID, report)
			if err != nil {
				return
			}
		}
		//
		// Facts
		facts := addon.Application.Facts(application.ID)
		facts.Source(Source)
		err = facts.Replace(report.Facts())
		if err == nil {
			addon.Activity("Facts updated.")
		} else {
			return
		}

		addon.Activity("Done.")

		return
	})
}
