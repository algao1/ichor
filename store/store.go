package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
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
func (s *Store) Initialize() error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(FieldGlucose))
		if err != nil {
			return fmt.Errorf("unable to create bucket: %w", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(FieldGlucosePred))
		if err != nil {
			return fmt.Errorf("unable to create bucket: %w", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(FieldCarbohydrate))
		if err != nil {
			return fmt.Errorf("unable to create bucket: %w", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(FieldObject))
		if err != nil {
			return fmt.Errorf("unable to create bucket: %w", err)
		}

		return nil
	})
}

// AddPoint adds a singular TimePoint under a field.
// Returns an error if the field does not exist.
func (s *Store) AddPoint(field string, t time.Time, pt interface{}) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		encoded, err := json.Marshal(pt)
		if err != nil {
			return err
		}

		return b.Put(timeToBytes(t), encoded)
	})
}

// GetPoints retrieves a series of TimePoints for a given field,
// between two dates. Returns an error if the field does not exist.
func (s *Store) GetPoints(start, end time.Time, field string, ptsPtr interface{}) error {
	min := timeToBytes(start)
	max := timeToBytes(end)

	values := make([][]byte, 0)

	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		c := b.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			values = append(values, v)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Reflection weirdness below.

	if reflect.TypeOf(ptsPtr).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, but got: %v", reflect.TypeOf(ptsPtr).Kind())
	}
	slice := reflect.ValueOf(ptsPtr).Elem()

	if reflect.ValueOf(ptsPtr).Elem().Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, but got: %v", reflect.ValueOf(ptsPtr).Elem().Kind())
	}
	etype := reflect.TypeOf(ptsPtr).Elem().Elem()

	for _, v := range values {
		eint := reflect.New(etype).Interface()
		err := json.Unmarshal(v, &eint)
		if err != nil {
			return err
		}
		elem := reflect.ValueOf(eint).Elem()
		slice.Set(reflect.Append(slice, elem))
	}

	return nil
}

func (s *Store) GetLastPoints(field string, last int, ptsPtr interface{}) error {
	values := make([][]byte, 0)

	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		c := b.Cursor()
		for k, v := c.Last(); k != nil && last > 0; k, v = c.Prev() {
			values = append(values, v)
			last--
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Reverse the list so it goes from earliest -> latest.
	// Generics would be a good use case for this.
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}

	// Reflection weirdness below.

	if reflect.TypeOf(ptsPtr).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, but got: %v", reflect.TypeOf(ptsPtr).Kind())
	}
	slice := reflect.ValueOf(ptsPtr).Elem()

	if reflect.ValueOf(ptsPtr).Elem().Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, but got: %v", reflect.ValueOf(ptsPtr).Elem().Kind())
	}
	etype := reflect.TypeOf(ptsPtr).Elem().Elem()

	for _, v := range values {
		eint := reflect.New(etype).Interface()
		err := json.Unmarshal(v, &eint)
		if err != nil {
			return err
		}
		elem := reflect.ValueOf(eint).Elem()
		slice.Set(reflect.Append(slice, elem))
	}

	return nil
}

func (s *Store) AddObject(index string, obj interface{}) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FieldObject))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", FieldObject)
		}

		encoded, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		return b.Put([]byte(index), encoded)
	})
}

func (s *Store) GetObject(index string, obj interface{}) error {
	var found []byte
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FieldObject))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", FieldObject)
		}

		found = b.Get([]byte(index))
		if found == nil {
			return fmt.Errorf("unable to find key: %s", index)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err = json.Unmarshal(found, &obj); err != nil {
		return fmt.Errorf("unable to unmarshal object: %w", err)
	}

	return nil
}

func timeToBytes(t time.Time) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(t.Unix()))
	return buf[:]
}
