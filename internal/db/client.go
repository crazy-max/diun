package db

import (
	"reflect"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

// Client represents an active db object
type Client struct {
	*bolt.DB
	cfg      model.Db
	metadata Metadata
}

const (
	dbVersion      = 2
	bucketMetadata = "metadata"
	bucketManifest = "manifest"
)

// New creates new db instance
func New(cfg model.Db) (*Client, error) {
	db, err := bolt.Open(cfg.Path, 0600, &bolt.Options{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketMetadata))
		return err
	}); err != nil {
		return nil, err
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketManifest))
		return err
	}); err != nil {
		return nil, err
	}

	if err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketManifest))
		stats := b.Stats()
		log.Debug().Msgf("%d entries found in manifest bucket", stats.KeyN)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "cannot count entries in manifest bucket")
	}

	c := &Client{
		DB:       db,
		cfg:      cfg,
		metadata: Metadata{},
	}

	if err := c.ReadMetadata(); err != nil {
		return nil, err
	}
	if reflect.DeepEqual(c.metadata, Metadata{}) {
		c.metadata = Metadata{
			Version: 1,
		}
	}

	log.Debug().Msgf("Current database version: %d", c.metadata.Version)
	return c, nil
}

// Close closes db connection
func (c *Client) Close() error {
	return c.DB.Close()
}
