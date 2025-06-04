package bot

import (
	"context"
    "log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
    // CommandStart is a command to start the bot.
    CommandStart = "/start"
    // CommandList is a command to list all assets.
    CommandAssets = "/assets"
)

type Bot struct {
    client *tgbotapi.BotAPI
}

func NewBot(token string) (*Bot, error) {
    bot, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        return nil, err
    }

	return &Bot{
        client: bot,
    }, nil
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	return nil
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
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! I'm a bot. I can help you with something.")

                    // Создайте клавиатуру с кнопками
                    keyboard := tgbotapi.NewReplyKeyboard(
                        tgbotapi.NewKeyboardButtonRow(
                            tgbotapi.NewKeyboardButton(CommandStart),
                            tgbotapi.NewKeyboardButton(CommandAssets),
                        ),
                    )
                    msg.ReplyMarkup = keyboard
                    b.client.Send(msg)

                case CommandAssets:
                    log.Println("Command /assets")
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Here is the list of all assets:")
                    b.client.Send(msg)
                }
        case <-ctx.Done():
            return
        }
    }
}