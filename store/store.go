package store

import (
	"bytes"
	"encoding/binary"
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

type TimePoint struct {
	Time  time.Time
	Value float64
}

func Create() (*Store, error) {
	db, err := bolt.Open("ichor.db", 0600, nil)
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
func (s *Store) AddPoint(field string, tp TimePoint) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		var vbuf [8]byte
		binary.BigEndian.PutUint64(vbuf[:], math.Float64bits(tp.Value))

		return b.Put(timeToBytes(tp.Time), vbuf[:])
	})
}

// GetPoints retrieves a series of TimePoints for a given field,
// between two dates. Returns an error if the field does not exist.
func (s *Store) GetPoints(start, end time.Time, field string) ([]TimePoint, error) {
	tps := make([]TimePoint, 0)
	min := timeToBytes(start)
	max := timeToBytes(end)

	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		c := b.Cursor()

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			tps = append(tps, TimePoint{bytesToTime(k), bytesToFloat64(v)})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return tps, nil
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
