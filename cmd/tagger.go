package main

import (
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-hub/api"
	"math/rand"
	"strings"
)

//
// Tagger tags an application.
type Tagger struct {
	Enabled bool `json:"enabled"`
}

//
// AddOptions adds analyzer options.
func (r *Tagger) AddOptions(options *command.Options) (err error) {
	return
}

//
// Update updates application tags.
//   - Ensures categories exist.
//   - Endures tags exist.
//   - Replaces associated tags (by source).
func (r *Tagger) Update(appID uint, report Report) (err error) {
	addon.Activity("[TAG] Tagging Application %d.", appID)
	catMap, err := r.ensureCategories(report)
	if err != nil {
		return
	}
	wanted, err := r.ensureTags(catMap, report)
	if err != nil {
		return
	}
	err = r.ensureAssociated(appID, wanted)
	if err != nil {
		return
	}
	return
}

//
// ensureCategories ensures categories exist.
// Returns the map of category names to IDs.
func (r *Tagger) ensureCategories(report Report) (mp map[string]uint, err error) {
	mp = map[string]uint{}
	wanted := []api.TagCategory{}
	for _, ruleSet := range report {
		for _, s := range ruleSet.Tags {
			colon := strings.Index(s, ":")
			if colon > 0 {
				mp[s[:colon]] = 0
			}
		}
	}
	for name := range mp {
		wanted = append(
			wanted,
			api.TagCategory{
				Name: name,
				Rank: uint(rand.Intn(10)),
			})
	}
	for _, cat := range wanted {
		err = addon.TagCategory.Ensure(&cat)
		if err != nil {
			return
		}
		mp[cat.Name] = cat.ID
	}
	return
}

//
// ensureTags ensures tags exist.
// Returns the wanted tag IDs.
func (r *Tagger) ensureTags(catMap map[string]uint, report Report) (tags []uint, err error) {
	mp := map[TagRef]int{}
	wanted := []api.Tag{}
	for _, ruleSet := range report {
		for _, s := range ruleSet.Tags {
			colon := strings.Index(s, ":")
			if colon > 0 {
				ref := TagRef{
					Category: s[:colon],
					Name:     s[colon:],
				}
				mp[ref] = 0
			}
		}
	}
	for ref := range mp {
		catRef := api.Ref{
			ID: catMap[ref.Category],
		}
		wanted = append(
			wanted,
			api.Tag{
				Name:     ref.Name,
				Category: catRef,
			})
	}
	for _, tag := range wanted {
		err = addon.Tag.Ensure(&tag)
		if err != nil {
			return
		}
		tags = append(tags, tag.ID)
	}
	return
}

//
// ensureAssociated ensure wanted tags are associated.
func (r *Tagger) ensureAssociated(appID uint, wanted []uint) (err error) {
	tags := addon.Application.Tags(appID)
	tags.Source(Source)
	err = tags.Replace(wanted)
	return
}

//
// TagRef -
type TagRef struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}
