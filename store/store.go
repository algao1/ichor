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

type TimeSeriesStore struct {
	DB *bolt.DB
}

type TimePoint struct {
	Time  time.Time
	Value float64
}

func Create() (*TimeSeriesStore, error) {
	db, err := bolt.Open("ichor.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create store: %w", err)
	}
	return &TimeSeriesStore{db}, nil
}

func (tss *TimeSeriesStore) Initialize() error {
	return tss.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("glucose"))
		if err != nil {
			return fmt.Errorf("unable to create bucket: %s", err)
		}
		return nil
	})
}

func (tss *TimeSeriesStore) AddPoint(field string, tp TimePoint) error {
	return tss.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))

		var vbuf [8]byte
		binary.BigEndian.PutUint64(vbuf[:], math.Float64bits(tp.Value))

		return b.Put(timeToBytes(tp.Time), vbuf[:])
	})
}

func (tss *TimeSeriesStore) GetPoints(start, end time.Time, field string) ([]TimePoint, error) {
	tps := make([]TimePoint, 0)
	min := timeToBytes(start)
	max := timeToBytes(end)

	err := tss.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
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
