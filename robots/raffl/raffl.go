package robots

import (
	"fmt"
	"github.com/trinchan/slackbot/robots"
	"math/rand"
	"github.com/trinchan/slackbot/robots/raffl/db"
	"log"
)


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

func InitDb(payload *robots.Payload) (err error) {

	if (botInitialized) {
		return nil;
	}

	log.Println("Initializing the Database")

	prize.Open()
	defer prize.Close()

	prizes := []*prize.Prize{
		{
			"100",
			"JetBrains Product License",
			"1 year subscription to any product\nLicense Key: CH3DZ-3EEVS-727UJ-2P4KK-7IHL8",
			"CH3DZ-3EEVS-727UJ-2P4KK-7IHL8",
			false,
			"",
			"",
			"https://www.jetbrains.com/products.html",
		},
		{
			"101",
			"JetBrains product license",
			"1 year subscription to any product\nLicense Key: 7WIF8-AKIWA-CX0QD-A7BY8-6A1EH",
			"7WIF8-AKIWA-CX0QD-A7BY8-6A1EH",
			false,
			"",
			"",
			"https://www.jetbrains.com/products.html",
		},
	}

	// Persist prizes in the database.
	for _, p := range prizes {
		p.Save()
	}

	prize.List(prize.PrizeBucketName)                     // each key/val in people bucket

	numberOfPrizes := prize.Count(prize.PrizeBucketName)
	log.Printf("Number of prizes: %d", numberOfPrizes)

	return nil;
}

func (pb bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {

	log.Printf("[DEBUG] Payload: %s", p)
	prize.Open()
	defer prize.Close()

	status := "checking..."

	if (!botInitialized && p.Text != "init") {
		return "raffle needs to be initialized before continuing..."
	}

	switch p.Text {
	case "init":
		status = "initializing"
		go pb.InitializeDeferred(p)
		break;
	case "status":
		status = "running status"
		go pb.PrizeStatusDeferred(p, false)
		break;
	case "astatus":
		status = "running admin status"
		go pb.PrizeStatusDeferred(p, true)
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
	if (err != nil) {
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

func (pb bot) PrizeStatusDeferred(p *robots.Payload, admin bool) {
	prize.Open()
	defer prize.Close()
	if admin {
		SendResponse(p, prize.List(prize.PrizeBucketName))
	} else {
		SendResponse(p, prize.ListUnclaimed(prize.PrizeBucketName))
	}
}

func (pb bot) CheckForPrizeWinDeferred(p *robots.Payload) {

	prize.Open()
	defer prize.Close()
	numberOfUnclaimedPrizes := prize.NumberOfUnclaimedPrizes(prize.PrizeBucketName)

	if numberOfUnclaimedPrizes == 0 {
		SendResponse(p, "All prizes have been claimed for this round. An new round will open up soon...")
		return
	}

	pick := rand.Intn(4)
	outcome := ""

	log.Printf("Number of prizes: %v - pick %v", numberOfUnclaimedPrizes, pick)

	if pick > 0 && pick <= numberOfUnclaimedPrizes {
		prizeInfo, err := prize.SelectAndClaimPrize(pick, p.UserName, p.UserID)
		if err != nil {
			outcome = fmt.Sprintf("Something went wrong, our crack dev team will check this out: %v", err)
		} else {
			outcome = prizeInfo
		}
	} else {
		outcome = "Sorry, better luck next time!"
	}
	message := fmt.Sprintf("Hi @%s!\n%s\n%s", p.UserName, "Let's see if you've won a prize...", outcome)
	SendResponse(p, message)
}

func (pb bot) Description() (description string) {
	return "Raffl bot!\n\tUsage: /raffl\n\tExpected Response: @user: Raffl this!"
}
