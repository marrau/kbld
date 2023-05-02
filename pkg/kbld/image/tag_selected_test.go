// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctlimg "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/image"
	ctlreg "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/registry"
)

// TestMatchesTagSelection runs test cases on the matchesPlatform function which verifies
// whether the given platform can run on the required platform by checking the
// compatibility of architecture, OS, OS version, OS features, variant and features.
// Adapted from https://github.com/google/go-containerregistry/blob/570ba6c88a5041afebd4599981d849af96f5dba9/pkg/v1/remote/index_test.go#L251
func TestMatchesTagSelection(t *testing.T) {
	registry, _ := ctlreg.NewRegistry(ctlreg.Opts{
		VerifyCerts:   true,
		Insecure:      false,
		EnvAuthPrefix: "KBLD_TEST_",
	})

	tests := []struct {
		image   string
		pattern string
		want    string
	}{
		{
			image:   "docker.io/library/alpine",
			pattern: "edge",
			want:    "edge",
		},
		{
			image:   "docker.io/library/alpine",
			pattern: ".+_rc1",
			want:    "3.17.0_rc1",
		},
		{
			image:   "docker.io/library/alpine",
			pattern: "3.13.[0-9]+",
			want:    "3.13.12",
		},
	}

	for _, test := range tests {
		repository, err := regname.NewRepository(test.image)
		if err != nil {
			t.Errorf("failed on creating repo, %v", err)
			continue
		}

		tagSelectedImage := ctlimg.NewTagSelectedImage(
			test.image,
			&ctlconf.VersionSelection{
				Regex: &ctlconf.VersionSelectionRegex{
					Pattern: test.pattern,
				},
			},
			registry,
		)

		resultVersion, _, err := tagSelectedImage.RegexTagSelect(repository)
		if err != nil {
			t.Errorf("failed on selecting tags, %v", err)
		} else {
			if test.want != resultVersion {
				t.Errorf("matchesPattern(%v, %v); got %v, want %v", test.image, test.pattern, resultVersion, test.want)
			}
		}
	}
}
