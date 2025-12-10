package oci

import (
	"net/http"
	"testing"

	"github.com/distribution/reference"
	"github.com/stretchr/testify/assert"
)

func TestCheckImage(t *testing.T) {
	image := "python:3.12"
	named, err := reference.ParseDockerRef(image)
	assert.NoError(t, err)

	rc, err := NewRegistryClient()
	assert.NoError(t, err)
	res, err := rc.CheckImage(t.Context(), named)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGetManifestURLForImage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in  string
		out string
	}{
		// Tests on Docker library images
		{
			in:  "python:3.13@sha256:2deb0891ec3f643b1d342f04cc22154e6b6a76b41044791b537093fae00b6884",
			out: "https://registry-1.docker.io/v2/library/python/manifests/sha256:2deb0891ec3f643b1d342f04cc22154e6b6a76b41044791b537093fae00b6884",
		},
		{
			in:  "python:3.13",
			out: "https://registry-1.docker.io/v2/library/python/manifests/3.13",
		},
		{
			in:  "python@sha256:2deb0891ec3f643b1d342f04cc22154e6b6a76b41044791b537093fae00b6884",
			out: "https://registry-1.docker.io/v2/library/python/manifests/sha256:2deb0891ec3f643b1d342f04cc22154e6b6a76b41044791b537093fae00b6884",
		},
		{
			in:  "python",
			out: "https://registry-1.docker.io/v2/library/python/manifests/latest",
		},
		// Tests on images hosted on Docker Hub
		{
			in:  "renku/amalthea-sessions:0.21.0@sha256:e0be19853aa5359039ea6ec2a2277b8dc4f404f14de5645e0cb604426b326ee3",
			out: "https://registry-1.docker.io/v2/renku/amalthea-sessions/manifests/sha256:e0be19853aa5359039ea6ec2a2277b8dc4f404f14de5645e0cb604426b326ee3",
		},
		{
			in:  "renku/amalthea-sessions:0.21.0",
			out: "https://registry-1.docker.io/v2/renku/amalthea-sessions/manifests/0.21.0",
		},
		{
			in:  "renku/amalthea-sessions@sha256:e0be19853aa5359039ea6ec2a2277b8dc4f404f14de5645e0cb604426b326ee3",
			out: "https://registry-1.docker.io/v2/renku/amalthea-sessions/manifests/sha256:e0be19853aa5359039ea6ec2a2277b8dc4f404f14de5645e0cb604426b326ee3",
		},
		{
			in:  "renku/amalthea-sessions",
			out: "https://registry-1.docker.io/v2/renku/amalthea-sessions/manifests/latest",
		},
		// Tests on images hosted in other registries
		{
			in:  "harbor.dev.renku.ch/renku-build/renku-build:renku-01k604nerkh0x9qjehbwmr8vyf",
			out: "https://harbor.dev.renku.ch/v2/renku-build/renku-build/manifests/renku-01k604nerkh0x9qjehbwmr8vyf",
		},
		{
			in:  "harbor.dev.renku.ch/deeply/nested/path:main",
			out: "https://harbor.dev.renku.ch/v2/deeply/nested/path/manifests/main",
		},
		{
			in:  "harbor.dev.renku.ch/renku-build/renku-build:renku-01k604nerkh0x9qjehbwmr8vyf@sha256:2deb0891ec3f643b1d342f04cc22154e6b6a76b41044791b537093fae00b6884",
			out: "https://harbor.dev.renku.ch/v2/renku-build/renku-build/manifests/sha256:2deb0891ec3f643b1d342f04cc22154e6b6a76b41044791b537093fae00b6884",
		},
		{
			in:  "harbor.dev.renku.ch/renku-build/renku-build@sha256:2deb0891ec3f643b1d342f04cc22154e6b6a76b41044791b537093fae00b6884",
			out: "https://harbor.dev.renku.ch/v2/renku-build/renku-build/manifests/sha256:2deb0891ec3f643b1d342f04cc22154e6b6a76b41044791b537093fae00b6884",
		},
		{
			in:  "harbor.dev.renku.ch/renku-build/renku-build",
			out: "https://harbor.dev.renku.ch/v2/renku-build/renku-build/manifests/latest",
		},
	}
	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			t.Parallel()
			t.Log(test.in)

			named, err := reference.ParseDockerRef(test.in)
			assert.NoError(t, err)
			result, err := GetManifestURLForImage(named)
			assert.NoError(t, err)
			assert.Equal(t, test.out, result.String())
		})
	}
}
