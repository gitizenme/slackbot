package robots

import (
	"fmt"

	"github.com/trinchan/slackbot/robots"
)

type bot struct{}

func init() {
	p := &bot{}
	robots.RegisterRobot("raffl", p)
}

func (pb bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {
	go pb.DeferredAction(p)
	//return "raffl this!"
	return ""
}

func (pb bot) DeferredAction(p *robots.Payload) {
	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     p.ChannelID,
		Username:    "raffl",
		Text:        fmt.Sprintf("@%s Raffl!", p.UserName),
		IconEmoji:   ":gift:",
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
	}
	response.Send()
}

func (pb bot) Description() (description string) {
	return "Raffl bot!\n\tUsage: /raffl\n\tExpected Response: @user: Raffl this!"
}
