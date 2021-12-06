package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/boltdb/bolt"
)

// See https://itnext.io/storing-time-series-in-rocksdb-a-cookbook-e873fcb117e4
// for inspiration.

type Store struct {
	DB *bolt.DB
}

func Create() (*Store, error) {
	db, err := bolt.Open("data/ichor.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create store: %w", err)
	}
	return &Store{db}, nil
}

// Initialize stands up the necessary buckets for future transactions.
//		TODO: Initialize from a manifest file.
func (s *Store) Initialize() error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("glucose"))
		if err != nil {
			return fmt.Errorf("unable to create bucket: %w", err)
		}
		return nil
	})
}

// AddPoint adds a singular TimePoint under a field.
// Returns an error if the field does not exist.
func (s *Store) AddPoint(field string, pt *TimePoint) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		encoded, err := json.Marshal(pt)
		if err != nil {
			return err
		}

		return b.Put(timeToBytes(pt.Time), encoded)
	})
}

// GetPoints retrieves a series of TimePoints for a given field,
// between two dates. Returns an error if the field does not exist.
func (s *Store) GetPoints(start, end time.Time, field string) ([]*TimePoint, error) {
	pts := make([]*TimePoint, 0)
	min := timeToBytes(start)
	max := timeToBytes(end)

	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		c := b.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			var pt TimePoint

			err := json.Unmarshal(v, &pt)
			if err != nil {
				return err
			}

			pts = append(pts, &pt)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pts, nil
}

func (s *Store) GetLastPoints(field string, last int) ([]*TimePoint, error) {
	pts := make([]*TimePoint, 0)

	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		c := b.Cursor()
		for k, v := c.Last(); k != nil && last > 0; k, v = c.Prev() {
			var pt TimePoint

			err := json.Unmarshal(v, &pt)
			if err != nil {
				return err
			}

			pts = append(pts, &pt)

			last--
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pts, nil
}

func timeToBytes(t time.Time) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(t.Unix()))
	return buf[:]
}

func bytesToTime(b []byte) time.Time {
	unsigned := binary.BigEndian.Uint64(b)
	return time.Unix(int64(unsigned), 0)
}

func bytesToFloat64(b []byte) float64 {
	bits := binary.BigEndian.Uint64(b)
	return math.Float64frombits(bits)
}
