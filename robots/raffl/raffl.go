package robots

import (
	"fmt"

	"github.com/trinchan/slackbot/robots"
	"math/rand"
)

type bot struct{}

func init() {
	p := &bot{}
	robots.RegisterRobot("raffl", p)
}

func (pb bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {
	go pb.DeferredAction(p)
	return "checking..."
}

func (pb bot) DeferredAction(p *robots.Payload) {

	reasons := make([]string, 0)
	reasons = append(reasons,
		"Raffle Item 1",
		"Raffle Item 2",
		"Raffle Item 3",
		"Raffle Item 4")

	pick := rand.Intn(100)

	message := ""
	if pick > 0 && pick < len(reasons) {
		message = fmt.Sprint("Your a winner! Here is your prize: %s", reasons[pick])
	} else {
		message = "Sorry, better luck next time!"
	}

	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     p.ChannelID,
		Username:    "raffl",
		Text:        fmt.Sprintf("Hi @%s\n. %s\n %s", p.UserName, "Let's see if you've won a prize.", message),
		IconEmoji:   ":gift:",
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
	}
	response.Send()
}

func (pb bot) Description() (description string) {
	return "Raffl bot!\n\tUsage: /raffl\n\tExpected Response: @user: Raffl this!"
}
