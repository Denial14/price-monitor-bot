package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

func (b *Bot) HandleStart(ctx context.Context, update tgbotapi.Update) {
	tgID := update.Message.From.ID
	username := update.Message.From.UserName

	user, err := b.users.EnsureUser(ctx, tgID, username)
	if err != nil {
		log.Errorf("ensure user: %v", err)
		b.SendText(update.Message.Chat.ID, "Ошибка. Попробуте позже.")
		return
	}

	b.SendText(update.Message.Chat.ID, fmt.Sprintf("Hello %s!", user.Username))
}

func (b *Bot) HandleHelp(ctx context.Context, update tgbotapi.Update) {
	text := `
		Допустимые команды:
		/start - начать
		/help - справка
	`
	b.SendText(update.Message.Chat.ID, text)
}
