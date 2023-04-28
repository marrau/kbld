// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"regexp"
	"sort"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctlreg "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/registry"
	"github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions"
)

// TagSelectedImage represents an image that will be resolved into url+digest
type TagSelectedImage struct {
	url       string
	selection *ctlconf.VersionSelection
	registry  ctlreg.Registry
}

func NewTagSelectedImage(url string, selection *ctlconf.VersionSelection,
	registry ctlreg.Registry) TagSelectedImage {

	return TagSelectedImage{url, selection, registry}
}

func (i TagSelectedImage) URL() (string, []ctlconf.Origin, error) {
	repo, err := regname.NewRepository(i.url, regname.WeakValidation)
	if err != nil {
		return "", nil, err
	}

	var tag string

	switch {
	case i.selection.Semver != nil:
		tag, _, err = i.SemverTagSelect(repo)
		if err != nil {
			return "", nil, err
		}
	case i.selection.Regex != nil:
		{
			tag, _, err = i.RegexTagSelect(repo)
			if err != nil {
				return "", nil, err
			}
		}
	default:
		return "", nil, fmt.Errorf("Unknown tag selection strategy")
	}

	// tag value is included by ResolvedImage
	return NewResolvedImage(i.url+":"+tag, i.registry).URL()
}

func (i TagSelectedImage) RegexTagSelect(repo regname.Repository) (string, []ctlconf.Origin, error) {
	tags, err := i.registry.ListTags(repo)
	if err != nil {
		return "", nil, err
	}

	regex, err := regexp.Compile(i.selection.Regex.Pattern)
	if err != nil {
		return "", nil, err
	}

	matchingTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		if regex.MatchString(tag) {
			matchingTags = append(matchingTags, tag)
		}
	}

	if len(matchingTags) == 0 {
		return "", nil, fmt.Errorf("expected to find at least one version, but did not")
	}

	// return last item of a sorted slice which should be the highest "version" found for your regex-pattern
	sort.Strings(matchingTags)
	highestVersion := matchingTags[len(matchingTags)-1]

	return highestVersion, nil, nil
}

func (i TagSelectedImage) SemverTagSelect(repo regname.Repository) (string, []ctlconf.Origin, error) {
	tags, err := i.registry.ListTags(repo)
	if err != nil {
		return "", nil, err
	}

	matchedVers := versions.NewRelaxedSemversNoErr(tags).FilterPrereleases(i.selection.Semver.Prereleases)

	if len(i.selection.Semver.Constraints) > 0 {
		matchedVers, err = matchedVers.FilterConstraints(i.selection.Semver.Constraints)
		if err != nil {
			return "", nil, fmt.Errorf("Selecting versions: %s", err)
		}
	}

	highestVersion, found := matchedVers.Highest()
	if !found {
		return "", nil, fmt.Errorf("Expected to find at least one version, but did not")
	}

	return highestVersion, nil, nil
}
