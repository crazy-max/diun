package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

// Client represents an active db object
type Client struct {
	*bolt.DB
	cfg model.Db
}

const bucket = "manifest"

// New creates new db instance
func New(cfg model.Db) (*Client, error) {
	db, err := bolt.Open(cfg.Path, 0600, &bolt.Options{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		return nil, err
	}

	if err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		stats := b.Stats()
		log.Debug().Msgf("%d entries found in database", stats.KeyN)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("cannot count entries in database, %v", err)
	}

	return &Client{db, cfg}, nil
}

// Close closes db connection
func (c *Client) Close() error {
	return c.DB.Close()
}

// First checks if a Docker image has ever been analyzed
func (c *Client) First(image registry.Image) (bool, error) {
	found := false

	err := c.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
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
func (c *Client) GetManifest(image registry.Image) (docker.Manifest, error) {
	var manifest docker.Manifest

	err := c.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if entryBytes := b.Get([]byte(image.String())); entryBytes != nil {
			return json.Unmarshal(entryBytes, &manifest)
		}
		return nil
	})

	return manifest, err
}

// PutManifest add Docker image manifest in db
func (c *Client) PutManifest(image registry.Image, manifest docker.Manifest) error {
	entryBytes, _ := json.Marshal(manifest)

	err := c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Put([]byte(image.String()), entryBytes)
	})

	return err
}
