package prize

import (
	"fmt"
	"log"
	"time"
	//"path"
	"bytes"
	//"runtime"
	"encoding/json"
	"encoding/gob"
	"github.com/boltdb/bolt"
)

var db *bolt.DB
var open bool

const PrizeBucketName = "prizes"

type Prize struct {
	ID          string
	Title       string
	Description string
	LicenseKey  string
	Claimed     bool
	Username    string
}

func Open() error {
	var err error
	config := &bolt.Options{Timeout: 1 * time.Second}
	db, err = bolt.Open("prize.db", 0600, config)
	if err != nil {
		log.Fatal(err)
	}
	open = true
	return nil
}

func Close() {
	open = false
	db.Close()
}

func (p *Prize) Save() error {
	if !open {
		return fmt.Errorf("db must be opened before saving!")
	}
	err := db.Update(func(tx *bolt.Tx) error {
		people, err := tx.CreateBucketIfNotExists([]byte(PrizeBucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		enc, err := p.encode()
		if err != nil {
			return fmt.Errorf("could not encode Person %s: %s", p.ID, err)
		}
		err = people.Put([]byte(p.ID), enc)
		return err
	})
	return err
}

func (p *Prize) gobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(p)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(data []byte) (*Prize, error) {
	var p *Prize
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Prize) encode() ([]byte, error) {
	enc, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decode(data []byte) (*Prize, error) {
	var p *Prize
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func GetPrize(id string) (*Prize, error) {
	if !open {
		return nil, fmt.Errorf("db must be opened before reading!")
	}
	var p *Prize
	err := db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(PrizeBucketName))
		k := []byte(id)
		p, err = decode(b.Get(k))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Could not get Prize ID %s", id)
		return nil, err
	}
	return p, nil
}

func NumberOfPrizes(bucket string) (int) {
	var numberOfPrizes int

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		numberOfPrizes = b.Stats().KeyN
		return nil
	})
	return numberOfPrizes
}

func List(bucket string) (string) {
	if !open {
		return "Prize list not available, please try again later..."
	}

	prizeList := "Prize List\n"

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			prizeList += fmt.Sprintf("Prize: %s\n", v)
		}
		return nil
	})
	return prizeList
}

func ListPrefix(bucket, prefix string) (string) {
	if !open {
		return "Prize list not available, please try again later..."
	}

	prizeList := "Prize List\n"

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		p := []byte(prefix)
		for k, v := c.Seek(p); bytes.HasPrefix(k, p); k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			prizeList += fmt.Sprintf("Prize: %s\n", v)
		}
		return nil
	})
	return prizeList
}

func ListRange(bucket, start, stop string) (string) {
	if !open {
		return "Prize list not available, please try again later..."
	}

	prizeList := "Prize List\n"

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		min := []byte(start)
		max := []byte(stop)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0;
		k, v = c.Next() {
			fmt.Printf("%s: %s\n", k, v)
			prizeList += fmt.Sprintf("Prize: %s\n", v)
		}
		return nil
	})
	return prizeList
}