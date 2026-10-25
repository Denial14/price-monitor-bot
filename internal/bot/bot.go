package bot

import (
	"context"
	"fmt"
	"price-bot/internal/parser"
	"price-bot/internal/storage"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	users    *storage.UserRepo
	items    *storage.ItemRepo
	registry *parser.Registry
	notifier *Notifier
}

func New(token string, pool *pgxpool.Pool, registry *parser.Registry, pollInterval time.Duration) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("error creating bot: %w", err)
	}
	log.Infof("Authorized as @%s", bot.Self.UserName)

	users := storage.NewUserRepo(pool)
	items := storage.NewItemRepo(pool)
	notifier := NewNotifier(bot, items, users, pollInterval)

	return &Bot{
		api:      bot,
		users:    users,
		items:    items,
		registry: registry,
		notifier: notifier,
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.notifier.Run(ctx)
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	defer b.api.StopReceivingUpdates()

	for {
		select {
		case <-ctx.Done():
			log.Info("Bot: shutting down")
			return nil
		case update, ok := <-updates:
			if !ok {
				wg.Wait()
				return nil
			}
			b.HandleUpdate(ctx, update)
		}
	}
}

func (b *Bot) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	if !update.Message.IsCommand() {
		return
	}

	switch update.Message.Command() {
	case "start":
		b.HandleStart(ctx, update)
	case "help":
		b.HandleHelp(ctx, update)
	case "add":
		b.HandleAdd(ctx, update)
	case "remove":
		b.HandleRemove(ctx, update)
	case "list":
		b.HandleList(ctx, update)
	case "stat":
		b.HandleStat(ctx, update)
	case "check":
		b.HandleCheck(ctx, update)
	default:
		b.SendText(update.Message.Chat.ID, "Неизвестная команда. Попробуй /help")
	}
}

func (b *Bot) SendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Errorf("send message: %v", err)
	}
}
