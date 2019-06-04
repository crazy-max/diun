package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/registry"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

// Client represents an active db object
type Client struct {
	*bolt.DB
	cfg model.Db
}

const bucket = "analysis"

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

// GetAnalysis returns Docker image analysis
func (c *Client) GetAnalysis(image registry.Image) (registry.Inspect, error) {
	var ana registry.Inspect

	err := c.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if entryBytes := b.Get([]byte(image.String())); entryBytes != nil {
			return json.Unmarshal(entryBytes, &ana)
		}
		return nil
	})

	return ana, err
}

// PutAnalysis add Docker image analysis in db
func (c *Client) PutAnalysis(image registry.Image, analysis registry.Inspect) error {
	entryBytes, _ := json.Marshal(analysis)

	err := c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Put([]byte(image.String()), entryBytes)
	})

	return err
}
