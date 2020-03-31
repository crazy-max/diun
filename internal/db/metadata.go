package db

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

// Metadata represents db metadata informations
type Metadata struct {
	Version int
}

const (
	metadataKey = "ID"
)

// ReadMetadata returns db metadata
func (c *Client) ReadMetadata() error {
	return c.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketMetadata))
		if entryBytes := b.Get([]byte(metadataKey)); entryBytes != nil {
			return json.Unmarshal(entryBytes, &c.metadata)
		}
		return nil
	})
}

// WriteMetadata writes db metadata
func (c *Client) WriteMetadata(metadata Metadata) error {
	entryBytes, _ := json.Marshal(metadata)

	err := c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketMetadata))
		return b.Put([]byte(metadataKey), entryBytes)
	})

	return err
}
