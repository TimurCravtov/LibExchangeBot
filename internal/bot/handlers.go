package bot

import (
	"context"
	"log/slog"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	callbackViewLibrary = "view_library"
	callbackSuggestBook = "suggest_book"
	callbackAdminPanel  = "admin_panel"

	welcomeMessageText = "Добро пожаловать в библиотеку обмена книгами! Выберите действие:"
)

func mainMenuKeyboard() *models.InlineKeyboardMarkup {
	emptyQuery := ""
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "📚 Посмотреть библиотеку", CallbackData: callbackViewLibrary},
			},
			{
				{Text: "🔍 Поиск книги", SwitchInlineQueryCurrentChat: &emptyQuery},
			},
			{
				{Text: "📖 Предложить книгу", CallbackData: callbackSuggestBook},
			},
			{
				{Text: "⚙️ Панель администратора", CallbackData: callbackAdminPanel},
			},
		},
	}
}

func (b *Bot) handleStart(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	_, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        welcomeMessageText,
		ReplyMarkup: mainMenuKeyboard(),
	})
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send /start response",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

func (b *Bot) handleMainMenuCallback(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(ctx, client, update, welcomeMessageText, mainMenuKeyboard())
}

func (b *Bot) handleHelp(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	helpText := "Доступные команды:\n" +
		"/start - Главное меню\n" +
		"/help  - Показать справку"

	_, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   helpText,
	})
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send /help response",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

func (b *Bot) handleSuggestBook(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(ctx, client, update, "📖 Раздел «Предложить книгу»", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "⬅️ Главное меню", CallbackData: callbackMainMenu}},
		},
	})
}

func (b *Bot) handleAdminPanel(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(ctx, client, update, "⚙️ Раздел «Панель администратора»", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "⬅️ Главное меню", CallbackData: callbackMainMenu}},
		},
	})
}

func (b *Bot) respondOrEditCallback(
	ctx context.Context,
	client *tgbot.Bot,
	update *models.Update,
	text string,
	markup *models.InlineKeyboardMarkup,
) {
	if update.CallbackQuery == nil {
		return
	}

	if _, err := client.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	}); err != nil {
		b.logger.ErrorContext(ctx, "failed to answer callback query",
			slog.String("callback_id", update.CallbackQuery.ID),
			slog.Any("error", err),
		)
	}

	msg := update.CallbackQuery.Message.Message
	if msg == nil {
		return
	}

	_, err := client.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to edit callback message",
			slog.Int64("chat_id", msg.Chat.ID),
			slog.Int("message_id", msg.ID),
			slog.Any("error", err),
		)
	}
}

func (b *Bot) handleDefault(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	_, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Неизвестная команда. Используйте /start для открытия главного меню.",
	})
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send default response",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}
