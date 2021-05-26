package db

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/crazy-max/diun/v4/pkg/registry"
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

// DeleteManifest deletes a Docker image manifest
func (c *Client) DeleteManifest(manifest registry.Manifest) error {
	return c.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketManifest)).Delete([]byte(fmt.Sprintf("%s:%s", manifest.Name, manifest.Tag)))
	})
}

// ListImage return a list of Docker images with their linked manifests
func (c *Client) ListImage() (map[string][]registry.Manifest, error) {
	images := make(map[string][]registry.Manifest)

	err := c.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketManifest)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var manifest registry.Manifest
			if err := json.Unmarshal(v, &manifest); err != nil {
				return err
			}
			if _, ok := images[manifest.Name]; !ok {
				images[manifest.Name] = []registry.Manifest{}
			}
			images[manifest.Name] = append(images[manifest.Name], manifest)
		}
		return nil
	})

	return images, err
}
