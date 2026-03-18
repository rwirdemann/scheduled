package telegram

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rwirdemann/scheduled"
	tea "charm.land/bubbletea/v2"
)

type Poller struct {
	bot    *tgbotapi.BotAPI
	sendFn func(tea.Msg)
}

// NewPoller creates a Poller from TELEGRAM_BOT_TOKEN env var.
// Returns nil (graceful degradation) if token is absent or bot init fails.
func NewPoller(sendFn func(tea.Msg)) *Poller {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Println("TELEGRAM_BOT_TOKEN not set, Telegram integration disabled")
		return nil
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("Failed to create Telegram bot: %v", err)
		return nil
	}
	return &Poller{bot: bot, sendFn: sendFn}
}

// Start launches the polling goroutine.
func (p *Poller) Start() {
	go p.poll()
}

func (p *Poller) poll() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	for update := range p.bot.GetUpdatesChan(u) {
		if update.Message == nil {
			continue
		}
		day, name := ParseWeekday(update.Message.Text)
		p.sendFn(scheduled.TelegramTaskMsg{Name: name, Day: day})
	}
}
