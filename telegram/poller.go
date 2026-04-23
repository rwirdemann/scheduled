package telegram

import (
	"log"
	"os"
	"strings"
	"time"

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
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 30
	for {
		updates, err := p.bot.GetUpdates(cfg)
		if err != nil {
			if !isTransientError(err) {
				log.Printf("Telegram polling error: %v", err)
			}
			time.Sleep(3 * time.Second)
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= cfg.Offset {
				cfg.Offset = update.UpdateID + 1
			}
			if update.Message == nil {
				continue
			}
			day, name := ParseWeekday(update.Message.Text)
			p.sendFn(scheduled.TelegramTaskMsg{Name: name, Day: day})
		}
	}
}

func isTransientError(err error) bool {
	s := err.Error()
	return strings.Contains(s, "connection reset by peer") ||
		strings.Contains(s, "EOF") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "no such host")
}
