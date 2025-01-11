package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ALF-ONE/telegram-samp-rcon-bot/samp"
	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
)

// https://sampwiki.blast.hk/wiki/Controlling_Your_Server#RCON_Commands
var commands = []string{
	"/exit",
	"/echo",
	"/hostname",
	"/gamemodetext",
	"/mapname",
	"/exec",
	"/kick",
	"/ban",
	"/changemode",
	"/gmx",
	"/reloadbans",
	"/reloadlog",
	"/say",
	"/players",
	"/banip",
	"/unbanip",
	"/gravity",
	"/weather",
	"/loadfs",
	"/weburl",
	"/unloadfs",
	"/reloadfs",
	"/rcon_password",
	"/password",
	"/messageslimit",
	"/ackslimit",
	"/messageholelimit",
	"/playertimeout",
	"/language",
}

type RconBot struct {
	telegramToken string
	botPassword   string
	serverHost    string
	rconPassword  string
	tg            *telebot.Bot
	sessionTime   map[int64]time.Time
}

func main() {
	godotenv.Load()

	rcon := RconBot{}

	rcon.readEnvironment()
	rcon.startBot()
}

func (bot *RconBot) startBot() {
	tg, err := telebot.NewBot(telebot.Settings{
		Token:  bot.telegramToken,
		Poller: &telebot.LongPoller{Timeout: time.Second * 10},
	})
	if err != nil {
		panic(err)
	}

	bot.tg = tg
	bot.sessionTime = make(map[int64]time.Time)

	for _, cmd := range commands {
		tg.Handle(cmd, bot.rconHandler)
	}

	tg.Handle("/cmdlist", bot.cmdlistHandler)
	tg.Handle("/login", bot.loginHandler)

	tg.Start()
}

func (bot *RconBot) readEnvironment() {
	bot.telegramToken = os.Getenv("SAMP_RCON_TG_TOKEN")
	if len(strings.TrimSpace(bot.telegramToken)) == 0 {
		fmt.Println("environment 'SAMP_RCON_TG_TOKEN' is empty")
		return
	}

	bot.botPassword = os.Getenv("SAMP_RCON_BOT_PASSWORD")
	if len(strings.TrimSpace(bot.botPassword)) == 0 {
		fmt.Println("environment 'SAMP_RCON_BOT_PASSWORD' is empty")
		return
	}

	bot.serverHost = os.Getenv("SAMP_RCON_SERVER_HOST")
	if len(strings.TrimSpace(bot.serverHost)) == 0 {
		fmt.Println("environment 'SAMP_RCON_SERVER_HOST' is empty")
		return
	}

	bot.rconPassword = os.Getenv("SAMP_RCON_SERVER_PASSWORD")
	if len(strings.TrimSpace(bot.rconPassword)) == 0 {
		fmt.Println("environment 'SAMP_RCON_SERVER_PASSWORD' is empty")
		return
	}
}

func (bot *RconBot) isValidSession(userid int64) bool {
	return time.Now().Before(bot.sessionTime[userid])
}

func (bot *RconBot) sendRconCommand(message *telebot.Message) {
	replyMessage, _ := bot.tg.Reply(message, "wait...")

	if _, err := samp.Send(bot.serverHost, bot.rconPassword, message.Text[1:]); err != nil {
		bot.tg.Edit(replyMessage, err.Error())
		return
	}

	bot.tg.Edit(replyMessage, "done!")
}

func (bot *RconBot) loginHandler(ctx telebot.Context) error {
	user := ctx.Sender()

	if !bot.isValidSession(user.ID) {
		password := strings.TrimPrefix(ctx.Message().Text, "/login ")

		if password != bot.botPassword {
			return ctx.Reply("invalid password!")
		}

		bot.sessionTime[user.ID] = time.Now().Add(1 * time.Minute)
	}

	return ctx.Reply("success!")
}

func (bot *RconBot) cmdlistHandler(ctx telebot.Context) error {
	if !bot.isValidSession(ctx.Sender().ID) {
		return ctx.Reply("use: /login [bot password]")
	}

	var text string
	for _, cmd := range commands {
		text += cmd + "\n"
	}

	return ctx.Reply(text)
}

func (bot *RconBot) rconHandler(ctx telebot.Context) error {
	if !bot.isValidSession(ctx.Sender().ID) {
		return ctx.Reply("use: /login [bot password]")
	}

	bot.sendRconCommand(ctx.Message())
	return nil
}
