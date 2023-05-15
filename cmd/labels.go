package main

import "github.com/konveyor/tackle2-addon/command"

//
// Labels list of sources.
type Labels []string

//
// AddOptions add options.
func (r Labels) AddOptions(options *command.Options) (err error) {
	for _, source := range r {
		options.Add("--source", source)
	}
	return
}
