package robots

import (
	"fmt"

	"github.com/trinchan/slackbot/robots"
	"github.com/boltdb/bolt"
	"math/rand"
	"log"
	"time"
	"encoding/json"
	"github.com/trinchan/slackbot/robots/raffl/db"
	"boldtb/person"
)

type Prize struct {
	Created time.Time
	Title   string
	Description string
	LicenseKey string
	Claimed bool
	Username string
}

type bot struct{}

func init() {
	p := &bot{}
	robots.RegisterRobot("raffl", p)

	prize.Open()
	defer prize.Close()

	// A Person struct consists of ID, Name, Age, Job.
	prizes := []*prize.Prize{
		{"100", "JetBrains product license", "1 year subscription to any product", "11112222", false, ""},
	}

	// Persist people in the database.
	for _, p := range prizes {
		p.Save()
	}

	// Get a prize from the database by their ID.
	for _, id := range []string{"100", "101"} {
		p, err := person.GetPerson(id)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(p)
	}


	person.List(prize.PrizeBucketName)                     // each key/val in people bucket
	person.ListPrefix(prize.PrizeBucketName, "20")         // ... with key prefix `20`
	person.ListRange(prize.PrizeBucketName, "101", "103")  // ... within range `101` to `103`

}

func (pb bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {
	go pb.DeferredAction(p)
	return "checking..."
}

func (pb bot) DeferredAction(p *robots.Payload) {


	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rafflePrizes := make([]Prize, 0)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("prizes"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("key=%s, value=%s\n", k, v)
				var p *Prize;
			err := json.Unmarshal(v, &p)
			if err != nil {
				return err
			}
			rafflePrizes = append(rafflePrizes, p)
			return nil
		})
		return nil
	})


	pick := rand.Intn(8)

	message := ""
	if pick > 0 && pick < len(rafflePrizes) {
		message = fmt.Sprintf("Your a winner! Here is your prize: %s", rafflePrizes[pick])
	} else {
		message = "Sorry, better luck next time!"
	}

	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     "@" + p.UserName,
		Username:    "raffl",
		Text:        fmt.Sprintf("Hi @%s!\n %s\n %s", p.UserName, "Let's see if you've won a prize...", message),
		IconEmoji:   ":gift:",
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
	}
	response.Send()
}

func (pb bot) Description() (description string) {
	return "Raffl bot!\n\tUsage: /raffl\n\tExpected Response: @user: Raffl this!"
}
