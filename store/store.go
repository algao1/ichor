package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gocarina/gocsv"
	"go.uber.org/zap"
)

type Store struct {
	DB     *bolt.DB
	logger *zap.Logger
	mu     sync.Mutex
}

func Create(logger *zap.Logger) (*Store, error) {
	db, err := bolt.Open("data/ichor.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create store: %w", err)
	}

	logger.Info("created bolt database")

	return &Store{DB: db, logger: logger}, nil
}

// Initialize stands up the necessary buckets for future transactions.
// TODO: This requires an overhaul, manually configuring bucket creation
//			 is a bit of a hassle.
func (s *Store) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.DB.Update(func(tx *bolt.Tx) error {
		for _, field := range Fields {
			_, err := tx.CreateBucketIfNotExists([]byte(field))
			if err != nil {
				return fmt.Errorf("unable to create bucket: %w", err)
			}
			s.logger.Info("ensured bucket exists",
				zap.String("bucket", field),
			)
		}

		return nil
	})
}

func (s *Store) exportSingle(filepath, field string, in interface{}) error {
	dest := fmt.Sprintf("%s/%s.csv", filepath, field)
	file, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	err = gocsv.MarshalFile(in, file)
	if err != nil {
		return err
	}

	s.logger.Info("successfully exported bucket as csv",
		zap.String("bucket", field),
		zap.String("filepath", dest),
	)

	return nil
}

func (s *Store) Export(filepath string) error {
	var gl []TimePoint
	if err := s.GetPoints(time.Unix(0, 0), time.Now(), FieldGlucose, &gl); err != nil {
		return err
	}
	if err := s.exportSingle(filepath, FieldGlucose, gl); err != nil {
		return err
	}

	var carbs []Carbohydrate
	if err := s.GetPoints(time.Unix(0, 0), time.Now(), FieldCarbohydrate, &carbs); err != nil {
		return err
	}
	if err := s.exportSingle(filepath, FieldCarbohydrate, carbs); err != nil {
		return err
	}

	var insulin []Insulin
	if err := s.GetPoints(time.Unix(0, 0), time.Now(), FieldInsulin, &insulin); err != nil {
		return err
	}
	if err := s.exportSingle(filepath, FieldInsulin, insulin); err != nil {
		return err
	}

	s.logger.Info("completed export of database")

	return nil
}

// AddPoint adds a singular TimePoint under a field.
// Returns an error if the field does not exist.
func (s *Store) AddPoint(field string, t time.Time, pt interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		encoded, err := json.Marshal(pt)
		if err != nil {
			return err
		}

		s.logger.Debug("added point",
			zap.String("field", field),
			zap.Time("time", t),
			zap.Any("point", pt),
		)

		return b.Put(timeToBytes(t), encoded)
	})
}

// GetPoints retrieves a series of TimePoints for a given field,
// between two dates. Returns an error if the field does not exist.
func (s *Store) GetPoints(start, end time.Time, field string, ptsPtr interface{}) error {
	min := timeToBytes(start)
	max := timeToBytes(end)

	values := make([][]byte, 0)

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(field))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", field)
		}

		c := b.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			values = append(values, v)
		}

		s.logger.Debug("found points",
			zap.String("field", field),
			zap.Time("start", start),
			zap.Time("end", end),
			zap.Int("found", len(values)),
		)

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

	s.mu.Lock()
	defer s.mu.Unlock()

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

		s.logger.Debug("found points",
			zap.String("field", field),
			zap.Int("last", last),
			zap.Int("found", len(values)),
		)

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
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FieldObject))
		if b == nil {
			return fmt.Errorf("unable to find bucket: %s", FieldObject)
		}

		encoded, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		s.logger.Debug("added object",
			zap.String("index", index),
			zap.Any("object", obj),
		)

		return b.Put([]byte(index), encoded)
	})
}

func (s *Store) GetObject(index string, obj interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

		s.logger.Debug("found object",
			zap.String("index", index),
			zap.Any("object", obj),
		)

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
