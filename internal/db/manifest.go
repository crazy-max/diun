package db

import (
	"bytes"
	"encoding/json"

	"github.com/crazy-max/diun/v4/pkg/registry"
	bolt "go.etcd.io/bbolt"
)

// ManifestEntry pairs the stored image reference with its latest observation.
// The observed tag and digest cannot be used to reconstruct the stored reference.
type ManifestEntry struct {
	Reference string
	Manifest  registry.Manifest
}

// First checks if a Docker image has ever been analyzed
func (c *Client) First(image registry.Image) (bool, error) {
	found := false

	err := c.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketManifest)).Cursor()
		for _, separator := range []string{":", "@"} {
			prefix := []byte(image.Name() + separator)
			if k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix) {
				found = true
				return nil
			}
		}
		return nil
	})

	return !found, err
}

// ListManifest return a list of Docker images manifests
func (c *Client) ListManifest() ([]registry.Manifest, error) {
	var manifests []registry.Manifest

	err := c.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketManifest)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var manifest registry.Manifest
			if err := json.Unmarshal(v, &manifest); err != nil {
				return err
			}
			manifests = append(manifests, manifest)
		}
		return nil
	})

	return manifests, err
}

// GetManifest returns Docker image manifest
func (c *Client) GetManifest(image registry.Image) (registry.Manifest, error) {
	var manifest registry.Manifest

	err := c.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketManifest))
		if entryBytes := b.Get([]byte(image.String())); entryBytes != nil {
			return json.Unmarshal(entryBytes, &manifest)
		}
		return nil
	})

	return manifest, err
}

// PutManifest add Docker image manifest in db
func (c *Client) PutManifest(image registry.Image, manifest registry.Manifest) error {
	entryBytes, _ := json.Marshal(manifest)
	return c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketManifest))
		return b.Put([]byte(image.String()), entryBytes)
	})
}

// DeleteManifest removes the exact stored reference. Missing entries are ignored.
func (c *Client) DeleteManifest(reference string) error {
	return c.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketManifest)).Delete([]byte(reference))
	})
}

// ListImage groups stored references and their observations by repository name.
func (c *Client) ListImage() (map[string][]ManifestEntry, error) {
	images := make(map[string][]ManifestEntry)

	err := c.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketManifest)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var manifest registry.Manifest
			if err := json.Unmarshal(v, &manifest); err != nil {
				return err
			}
			images[manifest.Name] = append(images[manifest.Name], ManifestEntry{
				Reference: string(k),
				Manifest:  manifest,
			})
		}
		return nil
	})

	return images, err
}
