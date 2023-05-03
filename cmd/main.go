package main

import (
	"github.com/konveyor/analyzer-lsp/hubapi"
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-addon/ssh"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/nas"
	"os"
	"path"
	"strings"
)

var (
	addon     = hub.Addon
	HomeDir   = ""
	DepDir    = ""
	BinDir    = ""
	SourceDir = ""
	AppDir    = ""
	Dir       = ""
	M2Dir     = ""
	ReportDir = ""
	RuleDir   = ""
)

func init() {
	Dir, _ = os.Getwd()
	HomeDir, _ = os.UserHomeDir()
	SourceDir = path.Join(Dir, "source")
	DepDir = path.Join(Dir, "deps")
	BinDir = path.Join(Dir, "bin")
	ReportDir = path.Join(Dir, "report")
	RuleDir = path.Join(Dir, "rules")
	M2Dir = "/cache/m2"
}

type SoftError = hub.SoftError

//
// Data Addon data passed in the secret.
type Data struct {
	// Output directory within application bucket.
	Output string `json:"output" binding:"required"`
	// Mode options.
	Mode Mode `json:"mode"`
	// Sources list.
	Sources Sources `json:"sources"`
	// Targets list.
	Targets Targets `json:"targets"`
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
		for _, dir := range []string{BinDir, M2Dir, RuleDir, ReportDir} {
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
			report := &hubapi.RuleSet{}
			err = d.Tagger.Update(application.ID, report)
			if err != nil {
				return
			}
		}
		return
	})
}
