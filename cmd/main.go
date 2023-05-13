package main

import (
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-addon/ssh"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/nas"
	"os"
	"path"
	"strings"
)

var (
	addon        = hub.Addon
	HomeDir      = ""
	BinDir       = ""
	SourceDir    = ""
	AppDir       = ""
	Dir          = ""
	M2Dir        = ""
	ReportPath   = ""
	DepsPath     = ""
	RuleDir      = ""
	SettingsPath = ""
)

func init() {
	Dir, _ = os.Getwd()
	addonDir := os.Getenv("ADDON")
	if addonDir != "" {
		Dir = addonDir
	}
	HomeDir, _ = os.UserHomeDir()
	SourceDir = path.Join(Dir, "source")
	BinDir = path.Join(Dir, "bin")
	ReportPath = path.Join(Dir, "report.yaml")
	DepsPath = path.Join(Dir, "deps.yaml")
	RuleDir = path.Join(Dir, "rules")
	M2Dir = "/cache/m2"
	SettingsPath = path.Join(Dir, "opt", "settings.json")
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
		// analyzer
		analyzer := Analyzer{}
		analyzer.Data = d
		//
		// Fetch application.
		addon.Activity("Fetching application.")
		application, err := addon.Task.Application()
		if err == nil {
			analyzer.application = application
		} else {
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
		// Fetch repository.
		if !d.Mode.Binary {
			addon.Total(2)
			if application.Repository == nil {
				err = &SoftError{Reason: "Application repository not defined."}
				return
			}
			SourceDir = path.Join(
				SourceDir,
				strings.Split(
					path.Base(
						application.Repository.URL),
					".")[0])
			AppDir = path.Join(SourceDir, application.Repository.Path)
			var r repository.SCM
			r, err = repository.New(
				SourceDir,
				application.Repository,
				application.Identities)
			if err != nil {
				return
			}
			err = r.Fetch()
			if err == nil {
				addon.Increment()
				analyzer.Mode.Repository = r
			} else {
				return
			}
		}
		//
		// Run analyzer.
		err = analyzer.Run()
		if err == nil {
			addon.Increment()
		} else {
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
		return
	})
}
