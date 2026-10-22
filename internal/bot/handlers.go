package bot

import (
	"context"
	"errors"
	"fmt"
	helpers "price-bot/internal/helpers"
	"price-bot/internal/storage"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

func (b *Bot) HandleStart(ctx context.Context, update tgbotapi.Update) {
	tgID := update.Message.From.ID
	username := update.Message.From.UserName

	user, err := b.users.EnsureUser(ctx, tgID, username)
	if err != nil {
		log.Errorf("ensure user: %v", err)
		b.SendText(update.Message.Chat.ID, "Ошибка. Попробуйте позже.")
		return
	}

	b.SendText(update.Message.Chat.ID, fmt.Sprintf("Hello %s!", user.Username))
}

func (b *Bot) HandleHelp(ctx context.Context, update tgbotapi.Update) {
	text := `
		Допустимые команды:
		/start - начать
		/help - справка
		/add - добавление товара
		/remove - удаление товара
		/list - показать список товаров
		/stat - показать статистику
	`
	b.SendText(update.Message.Chat.ID, text)
}

func (b *Bot) HandleAdd(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	tgID := update.Message.From.ID
	username := update.Message.From.UserName

	args := update.Message.CommandArguments()
	if args == "" {
		b.SendText(chatID, "Нужно минимум: /add <url> <price> [title]")
		return
	}

	parts := strings.Fields(args)
	if len(parts) < 2 {
		b.SendText(chatID, "Нужно минимум: /add <url> <price> [title]")
		return
	}

	rawURL := parts[0]
	rawPrice := parts[1]
	title := strings.Join(parts[2:], " ")

	if helpers.ValidateURL(rawURL) != nil {
		b.SendText(chatID, "Неверный URL. Должно начинаться с http:// или https://")
		return
	}

	price, err := helpers.ValidatePrice(rawPrice)
	if err != nil {
		b.SendText(chatID, "Неверный price. Price должен быть валидным и > 0")
		return
	}

	title = helpers.ValidateTitle(title)

	user, err := b.users.EnsureUser(ctx, tgID, username)
	if err != nil {
		log.Errorf("ensure user tg_id=%d: %v", tgID, err)
		b.SendText(chatID, "Ошибка. Попробуйте позже")
		return
	}

	item, err := b.items.Add(ctx, user.ID, rawURL, price, title)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicate) {
			b.SendText(chatID, "Данный URL уже существует. Введите другой")
			return
		}
		log.Errorf("error add item user_id=%d url=%s: %v", user.ID, rawURL, err)
		b.SendText(chatID, "Ошибка добавления. Попробуйте позже")
		return
	}

	b.SendText(chatID, fmt.Sprintf("Добавлен элемент [%d] %s - %.2f", item.ID, item.URL, item.TargetPrice))
}

func (b *Bot) HandleRemove(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	tgID := update.Message.From.ID

	args := update.Message.CommandArguments()
	if args == "" {
		b.SendText(chatID, "Использование: /remove <id>")
		return
	}

	user, err := b.users.GetByTgID(ctx, tgID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			b.SendText(chatID, "Вы еще не добавляли товары. /add <url> <price> [title]")
			return
		}
		log.Errorf("getting user by tg_id=%d: %v", tgID, err)
		b.SendText(chatID, "Ошибка. Попробуйте позже")
		return
	}

	idStr := strings.TrimSpace(strings.Fields(args)[0])
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		b.SendText(chatID, "ID товара должен быть положительным")
		return
	}

	err = b.items.Remove(ctx, user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrItemNotFound) {
			b.SendText(chatID, "Товар с таким ID не найден. /add <url> <price> [title]")
			return
		}
		log.Errorf("remove item item_id=%d: %v", id, err)
		b.SendText(chatID, "Ошибка удаления. Попробуйте позже")
		return
	}

	b.SendText(chatID, fmt.Sprintf("Удален товар [%d]", id))
}

func (b *Bot) HandleList(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	tgID := update.Message.From.ID

	user, err := b.users.GetByTgID(ctx, tgID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			b.SendText(chatID, "Вы еще не зарегистрированы/ /add <url> <price> [title]")
			return
		}
		log.Errorf("handle list tg_id=%d: %v", tgID, err)
		b.SendText(chatID, "Ошибка. Попробуйте позже")
		return
	}

	items, err := b.items.ListByUser(ctx, user.ID)
	if err != nil {
		log.Errorf("error")
		b.SendText(chatID, "Ошибка получения списка товаров. Попробуйте позже")
		return
	}

	if len(items) == 0 {
		b.SendText(chatID, "У вас еще нету товаров. /add <url> <price> [title]")
		return
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Список ваших товаров [%d]:\n\n", len(items)))
	for _, item := range items {
		sb.WriteString(helpers.FormatItem(item))
		sb.WriteString("\n")
	}

	b.SendText(chatID, sb.String())
}

func (b *Bot) HandleStat(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	tgID := update.Message.From.ID

	user, err := b.users.GetByTgID(ctx, tgID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			b.SendText(chatID, "Вы еще не добавили товары. /add <url> <price> [title]")
			return
		}
		log.Errorf("handle stat tg_id=%d: %v", tgID, err)
		b.SendText(chatID, "Ошибка. Попробуйте позже")
		return
	}

	count, err := b.items.CountByUser(ctx, user.ID)
	if err != nil {
		log.Errorf("count by user user_id=%d: %v", user.ID, err)
		b.SendText(chatID, "Ошибка. Попробуйте позже")
		return
	}

	b.SendText(chatID, fmt.Sprintf("Товаров: %d", count))
}
