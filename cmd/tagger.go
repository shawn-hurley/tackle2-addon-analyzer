package main

import (
	"encoding/json"
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-hub/api"
	"io"
	"math/rand"
	"os"
	pathlib "path"
)

//
// Tagger tags an application.
type Tagger struct {
	Enabled bool `json:"enabled"`
}

//
// AddOptions adds analyzer options.
func (r *Tagger) AddOptions(options *command.Options) (err error) {
	if r.Enabled {
		options.Add("--exportSummary")
	}
	return
}

//
// Update updates application tags.
//   - Ensures categories exist.
//   - Endures tags exist.
//   - Replaces associated tags (by source).
func (r *Tagger) Update(appID uint) (err error) {
	addon.Activity("[TAG] Tagging Application %d.", appID)
	report, err := r.report()
	if err != nil {
		return
	}
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
// report reads and returns the json summary.
// Returns the summary.
func (r *Tagger) report() (report []Summary, err error) {
	path := pathlib.Join(ReportDir, "analysisSummary.json")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := io.ReadAll(f)
	if err != nil {
		return
	}
	report = []Summary{}
	err = json.Unmarshal(b, &report)
	if err != nil {
		return
	}
	return
}

//
// ensureCategories ensures categories exist.
// Returns the map of category names to IDs.
func (r *Tagger) ensureCategories(report []Summary) (mp map[string]uint, err error) {
	mp = map[string]uint{}
	wanted := []api.TagCategory{}
	for i := range report {
		for _, tag := range report[i].Tags {
			mp[tag.Category] = 0
		}
		for name := range mp {
			wanted = append(
				wanted,
				api.TagCategory{
					Name: name,
					Rank: uint(rand.Intn(10)),
				})
		}
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
func (r *Tagger) ensureTags(catMap map[string]uint, report []Summary) (tags []uint, err error) {
	mp := map[TagRef]int{}
	wanted := []api.Tag{}
	for i := range report {
		for _, ref := range report[i].Tags {
			mp[ref] = 0
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
	tags.Source("Analysis")
	err = tags.Replace(wanted)
	return
}

//
// Summary analyzer object.
type Summary struct {
	Tags []TagRef `json:"technologyTags"`
}

//
// TagRef is a tag name & category.
type TagRef struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}
