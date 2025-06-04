package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lenin884/cryptobot/internal/market"
	"github.com/lenin884/cryptobot/internal/storage"
)

const (
	// CommandStart is a command to start the bot.
	CommandStart = "/start"
	// CommandList is a command to list all assets.
	CommandAssets = "/assets"
	// CommandSync syncs trade history.
	CommandSync = "/sync"
)

type Bot struct {
	client  *tgbotapi.BotAPI
	market  *market.Bybit
	storage *storage.Storage
}

func NewBot(token string, m *market.Bybit, s *storage.Storage) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		client:  bot,
		market:  m,
		storage: s,
	}, nil
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.client.Send(msg)
	return err
}

func (b *Bot) GetUpdatesChan() (tgbotapi.UpdatesChannel, error) {
	config := tgbotapi.NewUpdate(0)
	config.Timeout = 60
	return b.client.GetUpdatesChan(config)
}

func (b *Bot) HandleUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case update := <-updates:
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			// Захардкодим только для пользователя os_eugene
			if update.Message.From.UserName != "os_eugene" {
				continue
			}

			// Обработка сообщения
			switch update.Message.Text {
			case CommandStart:
				log.Println("Command /start")
				keyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(CommandStart),
						tgbotapi.NewKeyboardButton(CommandAssets),
						tgbotapi.NewKeyboardButton(CommandSync),
					),
				)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! I'm a bot.")
				msg.ReplyMarkup = keyboard
				b.client.Send(msg)

			case CommandAssets:
				log.Println("Command /assets")
				assets, err := b.storage.Assets()
				if err != nil {
					b.SendMessage(update.Message.Chat.ID, "error: "+err.Error())
					continue
				}
				var text string
				for _, a := range assets {
					price, _ := b.market.GetCurrentPrice(ctx, a.Symbol)
					text += fmt.Sprintf("%s qty: %.4f avg: %.4f current: %.4f\n", a.Symbol, a.Qty, a.AvgPrice, price)
				}
				if text == "" {
					text = "no assets"
				}
				b.SendMessage(update.Message.Chat.ID, text)

			case CommandSync:
				log.Println("Command /sync")
				var all []storage.Trade
				spotTrades, err := b.market.GetTradeHistory(ctx, "spot")
				if err == nil {
					for _, t := range spotTrades {
						all = append(all, storage.Trade(t))
					}
				}
				futureTrades, err := b.market.GetTradeHistory(ctx, "linear")
				if err == nil {
					for _, t := range futureTrades {
						all = append(all, storage.Trade(t))
					}
				}
				if err := b.storage.SaveTrades(all); err != nil {
					b.SendMessage(update.Message.Chat.ID, "sync error: "+err.Error())
				} else {
					b.SendMessage(update.Message.Chat.ID, "history synced")
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
