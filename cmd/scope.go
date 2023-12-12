package main

import (
	"fmt"

	"github.com/konveyor/tackle2-addon/command"
)

const (
	packageVar = "package"
)

// Scope settings.
type Scope struct {
	WithKnownLibs bool `json:"withKnownLibs"`
	Packages      struct {
		Included []string `json:"included,omitempty"`
		Excluded []string `json:"excluded,omitempty"`
	} `json:"packages"`
}

// AddOptions adds analyzer options.
func (r *Scope) AddOptions(options *command.Options) (err error) {
	if !r.WithKnownLibs {
		options.Add(
			"--dep-label-selector",
			"!konveyor.io/dep-source=open-source")
	}
	addon.Activity("adding packages includeted and excluded", "included", r.Packages.Included, "excluded", r.Packages.Excluded)
	if len(r.Packages.Excluded) > 0 || len(r.Packages.Included) > 0 {
		filterString := ""
		packageIncluded := false
		if len(r.Packages.Included) > 0 {
			filterString = fmt.Sprintf("(!%s", packageVar)
			for _, i := range r.Packages.Included {
				filterString = fmt.Sprintf("%s || %s=%s", filterString, packageVar, i)

			}
			filterString = fmt.Sprintf("%s)", filterString)
			packageIncluded = true
		}
		if packageIncluded && len(r.Packages.Excluded) > 0 {
			filterString = fmt.Sprintf("%s && (!%s", filterString, packageVar)
			for _, i := range r.Packages.Excluded {
				filterString = fmt.Sprintf("%s || !%s=%s", filterString, packageVar, i)
			}
		} else if len(r.Packages.Excluded) > 0 {
			filterString = fmt.Sprintf("(!%s", packageVar)
			for _, i := range r.Packages.Included {
				filterString = fmt.Sprintf("%s || %s=%s", filterString, packageVar, i)

			}
			filterString = fmt.Sprintf("%s)", filterString)
		}
		if filterString != "" {
			options.Add("--incident-selector", filterString)
		}
	}
	return
}
