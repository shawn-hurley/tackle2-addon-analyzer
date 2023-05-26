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
	RuleSets   []api.Ref       `json:"rulesets"`
	Repository *api.Repository `json:"repository"`
	Identity   *api.Ref        `json:"identity"`
	Labels     struct {
		Included []string `json:"included,omitempty"`
		Excluded []string `json:"excluded,omitempty"`
	} `json:"labels"`
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
	err = r.addRuleSets()
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
// addRuleSets adds rulesets.
func (r *Rules) addRuleSets() (err error) {
	for _, ref := range r.RuleSets {
		var ruleset *api.RuleSet
		ruleset, err = addon.RuleSet.Get(ref.ID)
		if err != nil {
			return
		}
		err = r.addRules(ruleset)
		if err != nil {
			return
		}
		err = r.addRuleSetRepository(ruleset)
		if err != nil {
			return
		}
	}
	return
}

//
// addRules adds rules
func (r *Rules) addRules(ruleset *api.RuleSet) (err error) {
	ruleDir := path.Join(
		RuleDir,
		"/rulesets",
		strconv.Itoa(int(ruleset.ID)),
		"rules")
	err = nas.MkDir(ruleDir, 0755)
	if err != nil {
		return
	}
	n := len(ruleset.Rules)
	for _, ruleset := range ruleset.Rules {
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
// addRuleSetRepository adds ruleset repository.
func (r *Rules) addRuleSetRepository(ruleset *api.RuleSet) (err error) {
	if ruleset.Repository == nil {
		return
	}
	rootDir := path.Join(
		RuleDir,
		"/rulesets",
		strconv.Itoa(int(ruleset.ID)),
		"repository")
	err = nas.MkDir(rootDir, 0755)
	if err != nil {
		return
	}
	var ids []api.Ref
	if ruleset.Identity != nil {
		ids = []api.Ref{*ruleset.Identity}
	}
	rp, err := repository.New(
		rootDir,
		ruleset.Repository,
		ids)
	if err != nil {
		return
	}
	err = rp.Fetch()
	if err != nil {
		return
	}
	ruleDir := path.Join(rootDir, ruleset.Repository.Path)
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
