package db

import (
	"bytes"
	"encoding/json"

	"github.com/crazy-max/diun/pkg/registry"
	bolt "go.etcd.io/bbolt"
)

// First checks if a Docker image has ever been analyzed
func (c *Client) First(image registry.Image) (bool, error) {
	found := false

	err := c.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketManifest)).Cursor()
		name := []byte(image.Name())
		for k, _ := c.Seek(name); k != nil && bytes.HasPrefix(k, name); k, _ = c.Next() {
			found = true
			return nil
		}
		return nil
	})

	return !found, err
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

	err := c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketManifest))
		return b.Put([]byte(image.String()), entryBytes)
	})

	return err
}
