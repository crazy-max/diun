// Package referrer is used for responses to the referrers to a manifest
package referrer

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"text/tabwriter"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
)

// ReferrerList contains the response to a request for referrers to a subject
type ReferrerList struct {
	Subject     ref.Ref                 `json:"subject"`               // subject queried
	Source      ref.Ref                 `json:"source"`                // source for referrers, if different from subject
	Descriptors []descriptor.Descriptor `json:"descriptors"`           // descriptors found in Index
	Annotations map[string]string       `json:"annotations,omitempty"` // annotations extracted from Index
	Manifest    manifest.Manifest       `json:"-"`                     // returned OCI Index
	Tags        []string                `json:"-"`                     // tags matched when fetching referrers
}

// Add appends an entry to rl.Manifest, used to modify the client managed Index
func (rl *ReferrerList) Add(m manifest.Manifest) error {
	rlM, ok := rl.Manifest.GetOrig().(v1.Index)
	if !ok {
		return fmt.Errorf("referrer list manifest is not an OCI index for %s", rl.Subject.CommonName())
	}
	// if entry already exists, return
	mDesc := m.GetDescriptor()
	for _, d := range rlM.Manifests {
		if d.Digest == mDesc.Digest {
			return nil
		}
	}
	// update descriptor, pulling up artifact type and annotations
	switch mOrig := m.GetOrig().(type) {
	case v1.ArtifactManifest:
		mDesc.Annotations = mOrig.Annotations
		mDesc.ArtifactType = mOrig.ArtifactType
	case v1.Manifest:
		mDesc.Annotations = mOrig.Annotations
		if mOrig.ArtifactType != "" {
			mDesc.ArtifactType = mOrig.ArtifactType
		} else {
			mDesc.ArtifactType = mOrig.Config.MediaType
		}
	case v1.Index:
		mDesc.Annotations = mOrig.Annotations
		mDesc.ArtifactType = mOrig.ArtifactType
	default:
		// other types are not supported
		return fmt.Errorf("invalid manifest for referrer \"%t\": %w", m.GetOrig(), errs.ErrUnsupportedMediaType)
	}
	// append descriptor to index
	rlM.Manifests = append(rlM.Manifests, mDesc)
	rl.Descriptors = rlM.Manifests
	err := rl.Manifest.SetOrig(rlM)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes an entry from rl.Manifest, used to modify the client managed Index
func (rl *ReferrerList) Delete(m manifest.Manifest) error {
	rlM, ok := rl.Manifest.GetOrig().(v1.Index)
	if !ok {
		return fmt.Errorf("referrer list manifest is not an OCI index for %s", rl.Subject.CommonName())
	}
	// delete matching entries from the list
	mDesc := m.GetDescriptor()
	found := false
	for i := len(rlM.Manifests) - 1; i >= 0; i-- {
		if rlM.Manifests[i].Digest == mDesc.Digest {
			rlM.Manifests = slices.Delete(rlM.Manifests, i, i+1)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("subject not found in referrer list%.0w", errs.ErrNotFound)
	}
	rl.Descriptors = rlM.Manifests
	err := rl.Manifest.SetOrig(rlM)
	if err != nil {
		return err
	}
	return nil
}

// IsEmpty reports if the returned Index contains no manifests
func (rl ReferrerList) IsEmpty() bool {
	rlM, ok := rl.Manifest.GetOrig().(v1.Index)
	if !ok || len(rlM.Manifests) == 0 {
		return true
	}
	return false
}

// MarshalPretty is used for printPretty template formatting
func (rl ReferrerList) MarshalPretty() ([]byte, error) {
	buf := &bytes.Buffer{}
	tw := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	var rRef ref.Ref
	if rl.Subject.IsSet() {
		rRef = rl.Subject
		fmt.Fprintf(tw, "Subject:\t%s\n", rl.Subject.CommonName())
	}
	if rl.Source.IsSet() {
		rRef = rl.Source
		fmt.Fprintf(tw, "Source:\t%s\n", rl.Source.CommonName())
	}
	fmt.Fprintf(tw, "\t\n")
	fmt.Fprintf(tw, "Referrers:\t\n")
	for _, d := range rl.Descriptors {
		fmt.Fprintf(tw, "\t\n")
		if rRef.IsSet() {
			fmt.Fprintf(tw, "  Name:\t%s\n", rRef.SetDigest(d.Digest.String()).CommonName())
		}
		err := d.MarshalPrettyTW(tw, "  ")
		if err != nil {
			return []byte{}, err
		}
	}
	if len(rl.Annotations) > 0 {
		fmt.Fprintf(tw, "Annotations:\t\n")
		keys := make([]string, 0, len(rl.Annotations))
		for k := range rl.Annotations {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			val := rl.Annotations[name]
			fmt.Fprintf(tw, "  %s:\t%s\n", name, val)
		}
	}
	err := tw.Flush()
	return buf.Bytes(), err
}

// FallbackTag returns the ref that should be used when the registry does not support the referrers API
func FallbackTag(r ref.Ref) (ref.Ref, error) {
	dig, err := digest.Parse(r.Digest)
	if err != nil {
		return r, fmt.Errorf("failed to parse digest for referrers: %w", err)
	}
	replaceRE := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	algo := replaceRE.ReplaceAllString(string(dig.Algorithm()), "-")
	hash := replaceRE.ReplaceAllString(string(dig.Hex()), "-")
	rOut := r.SetTag(fmt.Sprintf("%.32s-%.64s", algo, hash))
	return rOut, nil
}
