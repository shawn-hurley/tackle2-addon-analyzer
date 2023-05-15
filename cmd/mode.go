package main

import (
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-hub/api"
	"path"
	"strings"
)

//
// Mode settings.
type Mode struct {
	Binary     bool   `json:"binary"`
	Artifact   string `json:"artifact"`
	WithDeps   bool   `json:"withDeps"`
	Repository repository.SCM
	appDir     string
}

//
// Build assets.
func (r *Mode) Build(application *api.Application) (err error) {
	if !r.Binary {
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
		r.appDir = path.Join(SourceDir, application.Repository.Path)
		var r repository.SCM
		r, err = repository.New(
			SourceDir,
			application.Repository,
			application.Identities)
		if err != nil {
			return
		}
		err = r.Fetch()
	} else {
		if r.Artifact != "" {
			bucket := addon.Bucket()
			err = bucket.Get(r.Artifact, BinDir)
		}
	}
	return
}

//
// AddOptions adds analyzer options.
func (r *Mode) AddOptions(settings *Settings) (err error) {
	if r.Binary {
		// settings.Binary(BinDir)
	} else {
		settings.Location(r.appDir)
	}

	return
}
