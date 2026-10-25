package bot

import (
	"context"
	"fmt"
	"price-bot/internal/models"
	"price-bot/internal/storage"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

type Notifier struct {
	api          *tgbotapi.BotAPI
	items        *storage.ItemRepo
	users        *storage.UserRepo
	pollInterval time.Duration
}

func NewNotifier(api *tgbotapi.BotAPI, items *storage.ItemRepo, users *storage.UserRepo, pollInterval time.Duration) *Notifier {
	return &Notifier{
		api:          api,
		items:        items,
		users:        users,
		pollInterval: pollInterval,
	}
}

func (n *Notifier) Run(ctx context.Context) {
	ticker := time.NewTicker(n.pollInterval)
	defer ticker.Stop()

	n.NotifyOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info("notifier: shutting down")
			return
		case <-ticker.C:
			n.NotifyOnce(ctx)
		}
	}
}

func (n *Notifier) NotifyOnce(ctx context.Context) {
	const limit = 100
	items, err := n.items.ItemsToNotify(ctx, limit)
	if err != nil {
		log.Errorf("get items to notify: %w", err)
		return
	}

	if len(items) == 0 {
		return
	}

	log.Infof("notifier: %d items to notify", len(items))

	for _, it := range items {
		n.NotifyItem(ctx, it)
	}
}

func (n *Notifier) NotifyItem(ctx context.Context, item models.Item) {
	user, err := n.users.GetByID(ctx, item.UserID)
	if err != nil {
		log.Errorf("get user by id=%d: %w", item.UserID, err)
		return
	}

	text := fmt.Sprintf("Цена упала!\n\nЦель: %.2f\nТекущая: %.2f %s\n%s", item.TargetPrice, *item.LastPrice, *item.LastCurrency, item.URL)

	msg := tgbotapi.NewMessage(user.TgID, text)
	if _, err := n.api.Send(msg); err != nil {
		log.Errorf("send notification to user tg_id=%d: %v", user.TgID, err)
		return
	}

	if err := n.items.MarkNotified(ctx, item.ID, *item.LastPrice); err != nil {
		log.Errorf("mark notified item_id=%d: %v", item.ID, err)
		return
	}

	log.Infof("notified user tg_id=%d item_id=%d price=%.2f", user.TgID, item.ID, *item.LastPrice)
}
