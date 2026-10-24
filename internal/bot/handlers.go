package bot

import (
	"context"
	"log/slog"
	"strings"

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
	b.clearSuggestSession(update.Message.Chat.ID)

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
	if update.CallbackQuery != nil && update.CallbackQuery.Message.Message != nil {
		b.clearSuggestSession(update.CallbackQuery.Message.Message.Chat.ID)
	}
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

	if update.CallbackQuery.InlineMessageID != "" {
		if _, err := client.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			InlineMessageID: update.CallbackQuery.InlineMessageID,
			Text:            text,
			ReplyMarkup:     markup,
		}); err != nil && !strings.Contains(err.Error(), "message is not modified") {
			b.logger.ErrorContext(ctx, "failed to edit inline callback message",
				slog.String("inline_message_id", update.CallbackQuery.InlineMessageID),
				slog.Any("error", err),
			)
		}
		return
	}

	msg := update.CallbackQuery.Message.Message
	if msg == nil {
		if update.CallbackQuery.From.ID != 0 {
			_, _ = client.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      update.CallbackQuery.From.ID,
				Text:        text,
				ReplyMarkup: markup,
			})
		}
		return
	}

	_, err := client.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		if strings.Contains(err.Error(), "message is not modified") {
			return
		}
		b.logger.WarnContext(ctx, "edit callback message failed, sending new message",
			slog.Int64("chat_id", msg.Chat.ID),
			slog.Int("message_id", msg.ID),
			slog.Any("error", err),
		)
		if _, sendErr := client.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      msg.Chat.ID,
			Text:        text,
			ReplyMarkup: markup,
		}); sendErr != nil {
			b.logger.ErrorContext(ctx, "failed to send fallback callback message",
				slog.Int64("chat_id", msg.Chat.ID),
				slog.Any("error", sendErr),
			)
		}
	}
}

func (b *Bot) handleDefault(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.InlineQuery != nil {
		b.handleInlineSearch(ctx, client, update.InlineQuery)
		return
	}

	if update.CallbackQuery != nil {
		b.respondOrEditCallback(
			ctx,
			client,
			update,
			"🔄 Бот был обновлён или перезапущен. Пожалуйста, выберите действие в главном меню (или нажмите /start):",
			mainMenuKeyboard(),
		)
		return
	}

	if update.Message == nil {
		return
	}

	if b.tryHandleSuggestInput(ctx, client, update) {
		return
	}

	_, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Используйте кнопки меню ниже или команду /start:",
		ReplyMarkup: mainMenuKeyboard(),
	})
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send default response",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

func (b *Bot) handleInlineSearch(ctx context.Context, client *tgbot.Bot, iq *models.InlineQuery) {
	books := b.store.Search(iq.Query)
	results := make([]models.InlineQueryResult, 0, len(books))

	for _, bk := range books {
		desc := bk.Author + " | ISBN: " + bk.ISBN + " | Владелец: " + bk.Owner
		results = append(results, &models.InlineQueryResultArticle{
			ID:          bk.ID,
			Title:       bk.Title,
			Description: desc,
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: bk.FormatCard(),
			},
			ReplyMarkup: bookCardKeyboard(bk),
		})
	}

	if _, err := client.AnswerInlineQuery(ctx, &tgbot.AnswerInlineQueryParams{
		InlineQueryID: iq.ID,
		Results:       results,
		CacheTime:     1,
	}); err != nil {
		b.logger.ErrorContext(ctx, "failed to answer inline query",
			slog.String("inline_query_id", iq.ID),
			slog.Any("error", err),
		)
	}
}
