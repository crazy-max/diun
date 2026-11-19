package db

import (
	"bytes"
	"encoding/json"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	bolt "go.etcd.io/bbolt"
)

// ManifestEntry pairs the stored image reference with its latest observation.
// The observed tag and digest cannot be used to reconstruct the stored reference.
type ManifestEntry struct {
	Reference string
	Manifest  registry.Manifest
}

// First reports whether no reference history exists for the image's repository.
// Any stored tag or digest in that repository makes it return false, regardless
// of platform.
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

// ListManifest returns the latest observation for each stored image reference.
// Each reference appears once, with metadata from the last platform that wrote it.
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

// GetManifest returns the latest observation for the exact image reference,
// regardless of platform. If the reference is absent, it returns a zero Manifest
// and a nil error.
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

// PutManifest replaces the history entry for the exact image reference.
// It does not update platform state; use PutPlatformManifest to update both.
func (c *Client) PutManifest(image registry.Image, manifest registry.Manifest) error {
	entry, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	return c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketManifest))
		return b.Put([]byte(image.String()), entry)
	})
}

// GetPlatformManifest returns the last observation for the exact image reference
// and requested platform settings. If no platform state exists, it falls back to
// reference history so an existing reference is not treated as new. The fallback
// may contain another platform's metadata and must be resolved by the registry.
// If neither entry exists, it returns a zero Manifest and a nil error.
func (c *Client) GetPlatformManifest(image registry.Image, platform model.ImagePlatform) (registry.Manifest, error) {
	var manifest registry.Manifest
	err := c.View(func(tx *bolt.Tx) error {
		key := []byte(image.String())
		if bucket := tx.Bucket([]byte(bucketManifestPlatformState)).Bucket(key); bucket != nil {
			if entry := bucket.Get(platformSettingsKey(platform)); entry != nil {
				return json.Unmarshal(entry, &manifest)
			}
		}
		if entry := tx.Bucket([]byte(bucketManifest)).Get(key); entry != nil {
			return json.Unmarshal(entry, &manifest)
		}
		return nil
	})
	return manifest, err
}

// PutPlatformManifest atomically replaces the observation for the exact image
// reference and requested platform settings, along with its reference history.
// Other platforms retain their previous observations for update detection.
func (c *Client) PutPlatformManifest(image registry.Image, platform model.ImagePlatform, manifest registry.Manifest) error {
	entry, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	return c.Update(func(tx *bolt.Tx) error {
		key := []byte(image.String())
		bucket, err := tx.Bucket([]byte(bucketManifestPlatformState)).CreateBucketIfNotExists(key)
		if err != nil {
			return err
		}
		if err := bucket.Put(platformSettingsKey(platform), entry); err != nil {
			return err
		}
		return tx.Bucket([]byte(bucketManifest)).Put(key, entry)
	})
}

func platformSettingsKey(platform model.ImagePlatform) []byte {
	// Preserve empty settings: automatic selection and an explicitly requested
	// architecture can have different variant fallback rules in the registry.
	key, _ := json.Marshal([3]string{platform.OS, platform.Arch, platform.Variant})
	return key
}

// DeleteManifest atomically removes history and all platform state for the exact
// stored reference. Missing entries are ignored. It returns the total bytes of
// stored values removed, excluding keys and bbolt overhead. If the transaction
// fails, it returns zero and leaves the data intact.
func (c *Client) DeleteManifest(reference string) (int64, error) {
	var size int64
	err := c.Update(func(tx *bolt.Tx) error {
		key := []byte(reference)
		history := tx.Bucket([]byte(bucketManifest))
		size += int64(len(history.Get(key)))
		if err := history.Delete(key); err != nil {
			return err
		}
		platforms := tx.Bucket([]byte(bucketManifestPlatformState))
		if bucket := platforms.Bucket(key); bucket != nil {
			if err := bucket.ForEach(func(_, value []byte) error {
				size += int64(len(value))
				return nil
			}); err != nil {
				return err
			}
			if err := platforms.DeleteBucket(key); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return size, nil
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
