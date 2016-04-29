package robots

import (
	"fmt"

	"github.com/trinchan/slackbot/robots"
	"math/rand"
	"time"
	"github.com/trinchan/slackbot/robots/raffl/db"
	"log"
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

var botInitialized bool = false;

func init() {
	p := &bot{}
	robots.RegisterRobot("raffl", p)

/*
		// Get a prize from the database by their ID.
		for _, id := range []string{"100", "101"} {
			p, err := person.GetPerson(id)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(p)
		}

		person.ListPrefix(prize.PrizeBucketName, "20")         // ... with key prefix `20`
		person.ListRange(prize.PrizeBucketName, "101", "103")  // ... within range `101` to `103`
		*/

}

func InitDb (p *robots.Payload) (err error) {

	if(botInitialized) {
		return nil;
	}

	log.Println("Initializing the Database")

	prizes := []*prize.Prize{
		{"100", "JetBrains product license", "1 year subscription to any product", "11112222", false, ""},
	}

	fmt.Println("Prizes: %s", prizes)


	// Persist prizes in the database.
	for _, p := range prizes {
		p.Save()
	}


	return nil;
}

func (pb bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {

	log.Printf("[DEBUG] Payload: %s", p)
	prize.Open()
	defer prize.Close()

	status := "checking..."

	if(!botInitialized && p.Text != "init") {
		return "raffle needs to be initialized before continuing..."
	}

	switch p.Text {
	case "init":
		status = "initializing"
		go pb.InitializeDeferred(p)
		break;
	case "status":
		status = "running status"
		prize.List(prize.PrizeBucketName)                     // each key/val in people bucket
		break;
	default:
		status = "checking for winner"
		go pb.CheckForPrizeWinDeferred(p)
	}

	return status
}

func (pb bot) InitializeDeferred(p *robots.Payload) {

	message := ""
	err := InitDb(p)
	if(err != nil) {
		botInitialized = false;
		message = "Initialization failed"
	} else {
		botInitialized = true;
		message = "Initialization complete"
	}
	SendResponse(p, message)
}

var SendResponse = func(p *robots.Payload, message string) {
	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     "@" + p.UserName,
		Username:    "raffl",
		Text:        message,
		IconEmoji:   ":gift:",
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
	}
	response.Send()
}

func (pb bot) CheckForPrizeWinDeferred(p *robots.Payload) {

	rafflePrizes := make([]prize.Prize, 0)

	pick := rand.Intn(8)

	outcome := ""
	if pick >= 0 && pick < len(rafflePrizes) {
		outcome = fmt.Sprintf("Your a winner! Here is your prize: %s", rafflePrizes[pick])
	} else {
		outcome = "Sorry, better luck next time!"
	}
	message := fmt.Sprintf("Hi @%s!\n %s\n %s", p.UserName, "Let's see if you've won a prize...", outcome)
	SendResponse(p, message)
}

func (pb bot) Description() (description string) {
	return "Raffl bot!\n\tUsage: /raffl\n\tExpected Response: @user: Raffl this!"
}
