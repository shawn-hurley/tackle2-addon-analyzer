package main

import (
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
	"path"
	"strconv"
	"strings"
)

//
// Rules settings.
type Rules struct {
	Path       string          `json:"path"`
	Bundles    []api.Ref       `json:"bundles"`
	Repository *api.Repository `json:"repository"`
	Identity   *api.Ref        `json:"identity"`
	Labels     Labels          `json:"labels"`
	Tags       struct {
		Included []string `json:"included,omitempty"`
		Excluded []string `json:"excluded,omitempty"`
	} `json:"tags"`
	rules []string
}

//
// Build assets.
func (r *Rules) Build() (err error) {
	err = r.addFiles()
	if err != nil {
		return
	}
	err = r.addRepository()
	if err != nil {
		return
	}
	err = r.addBundles()
	if err != nil {
		return
	}
	return
}

//
// AddOptions adds analyzer options.
func (r *Rules) AddOptions(options *command.Options) (err error) {
	for _, path := range r.rules {
		options.Add("--rules", path)
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
func (r *Rules) addFiles() (err error) {
	if r.Path == "" {
		return
	}
	ruleDir := path.Join(RuleDir, "/files")
	err = nas.MkDir(ruleDir, 0755)
	if err != nil {
		return
	}
	r.rules = append(r.rules, ruleDir)
	bucket := addon.Bucket()
	err = bucket.Get(r.Path, ruleDir)
	if err != nil {
		return
	}
	return
}

//
// AddBundles adds bundles.
func (r *Rules) addBundles() (err error) {
	for _, ref := range r.Bundles {
		var bundle *api.RuleBundle
		bundle, err = addon.RuleBundle.Get(ref.ID)
		if err != nil {
			return
		}
		err = r.addRuleSets(bundle)
		if err != nil {
			return
		}
		err = r.addBundleRepository(bundle)
		if err != nil {
			return
		}
	}
	return
}

//
// addRuleSets adds ruleSets
func (r *Rules) addRuleSets(bundle *api.RuleBundle) (err error) {
	ruleDir := path.Join(
		RuleDir,
		"/bundles",
		strconv.Itoa(int(bundle.ID)),
		"rulesets")
	err = nas.MkDir(ruleDir, 0755)
	if err != nil {
		return
	}
	n := len(bundle.RuleSets)
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
		path := path.Join(ruleDir, name)
		addon.Activity("[FILE] Get rule: %s", path)
		err = addon.File.Get(ruleset.File.ID, path)
		if err != nil {
			break
		}
		if n == 1 {
			r.rules = append(r.rules, path)
		}
	}
	if n > 1 {
		r.rules = append(r.rules, ruleDir)
	}
	return
}

//
// addBundleRepository adds bundle repository.
func (r *Rules) addBundleRepository(bundle *api.RuleBundle) (err error) {
	if bundle.Repository == nil {
		return
	}
	rootDir := path.Join(
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
	ruleDir := path.Join(rootDir, bundle.Repository.Path)
	r.rules = append(r.rules, ruleDir)
	return
}

//
// addRepository adds custom repository.
func (r *Rules) addRepository() (err error) {
	if r.Repository == nil {
		return
	}
	rootDir := path.Join(
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
	ruleDir := path.Join(rootDir, r.Repository.Path)
	r.rules = append(r.rules, ruleDir)
	return
}
