package prize

import (
	"fmt"
	"log"
	"time"
	"bytes"
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
	Userid	    string
	Link	    string
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

func SelectAndClaimPrize(index int, userName string, userID string) (string, error) {

	if !open {
		return "", fmt.Errorf("db must be opened before reading!")
	}

	var p *Prize

	err := db.View(func(tx *bolt.Tx) error {
		var err error
		count := 0;

		c := tx.Bucket([]byte(PrizeBucketName)).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			count++;
			if count == index {
				p, err = decode(v);
				if err != nil {
					return err
				}
				break;
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("Could not select Prize, please try again later... (%v)", index)
	}

	prizeInfo := ""
	if(userID != p.Userid) {
		p.Claimed = true;
		p.Username = userName
		p.Userid = userID
		err = p.Save()
		if err != nil {
			return "", fmt.Errorf("Could not claim Prize, please try again later... %v", p.ID)
		}
		prizeInfo = fmt.Sprintf("You're a winner!\nTitle: %v\nDetails: %v\nFor more info go to: %v\n", p.Title, p.Description, p.Link)
	}  else {
		prizeInfo = "Sorry, you've already claimed a raffle prize for this round. Please try again in te next round."
	}

	return prizeInfo, nil
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

func Count(bucket string) (int) {
	var numberOfPrizes int

	db.View(func(tx *bolt.Tx) error {
		numberOfPrizes = 0
		c := tx.Bucket([]byte(PrizeBucketName)).Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			numberOfPrizes++
		}
		return nil
	})
	return numberOfPrizes
}

func numberOfPrizes(bucket string, claimed bool) (int) {
	var numberOfPrizes int

	db.View(func(tx *bolt.Tx) error {
		numberOfPrizes = 0
		c := tx.Bucket([]byte(PrizeBucketName)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			p, err := decode(v);
			if err == nil && p.Claimed == claimed {
				numberOfPrizes++
			}
		}
		return nil
	})
	return numberOfPrizes
}

func NumberOfUnclaimedPrizes(bucket string) (int) {
	return numberOfPrizes(bucket, false)
}

func NumberOfClaimedPrizes(bucket string) (int) {
	return numberOfPrizes(bucket, true)
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

func ListUnclaimed(bucket string) (string) {
	if !open {
		return "Prize list not available, please try again later..."
	}

	prizeList := "Prize List\n"

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			p, err := decode(v);
			if err == nil && p.Claimed == false {
				fmt.Printf("key=%s, value=%s\n", k, v)
				prizeList += fmt.Sprintf("Prize: %s\n", v)
			}
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