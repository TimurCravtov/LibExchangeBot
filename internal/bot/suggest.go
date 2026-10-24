package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"libexchangebot/internal/model"
)

const callbackSuggestCancel = "suggest_cancel"

type suggestStep int

const (
	stepWaitTitle suggestStep = iota + 1
	stepWaitAuthor
	stepWaitISBN
)

type suggestSession struct {
	Step   suggestStep
	Title  string
	Author string
}

func suggestCancelKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "❌ Отмена", CallbackData: callbackSuggestCancel},
			},
		},
	}
}

func suggestDoneKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "📖 Предложить ещё книгу", CallbackData: callbackSuggestBook},
			},
			{
				{Text: "📚 Посмотреть библиотеку", CallbackData: callbackViewLibrary},
			},
			{
				{Text: "⬅️ Главное меню", CallbackData: callbackMainMenu},
			},
		},
	}
}

func (b *Bot) clearSuggestSession(chatID int64) {
	b.suggestMu.Lock()
	delete(b.suggestSessions, chatID)
	b.suggestMu.Unlock()
}

func (b *Bot) handleSuggestBook(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
		return
	}
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	b.suggestMu.Lock()
	if b.suggestSessions == nil {
		b.suggestSessions = make(map[int64]*suggestSession)
	}
	b.suggestSessions[chatID] = &suggestSession{Step: stepWaitTitle}
	b.suggestMu.Unlock()

	text := "📖 Предложить книгу (шаг 1 из 3)\n\n" +
		"Пожалуйста, введите точное название книги:"
	b.respondOrEditCallback(ctx, client, update, text, suggestCancelKeyboard())
}

func (b *Bot) handleSuggestCancel(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery != nil && update.CallbackQuery.Message.Message != nil {
		b.clearSuggestSession(update.CallbackQuery.Message.Message.Chat.ID)
	}
	b.respondOrEditCallback(ctx, client, update, welcomeMessageText, mainMenuKeyboard())
}

func (b *Bot) tryHandleSuggestInput(ctx context.Context, client *tgbot.Bot, update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	chatID := update.Message.Chat.ID

	b.suggestMu.Lock()
	session, ok := b.suggestSessions[chatID]
	if !ok {
		b.suggestMu.Unlock()
		return false
	}

	input := strings.TrimSpace(update.Message.Text)
	if input == "" {
		b.suggestMu.Unlock()
		_, _ = client.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Пожалуйста, отправьте текстовое значение или нажмите «❌ Отмена».",
			ReplyMarkup: suggestCancelKeyboard(),
		})
		return true
	}

	switch session.Step {
	case stepWaitTitle:
		session.Title = input
		session.Step = stepWaitAuthor
		b.suggestMu.Unlock()

		reply := fmt.Sprintf(
			"✍️ Предложить книгу (шаг 2 из 3)\n\n"+
				"📖 Название: «%s»\n\n"+
				"Пожалуйста, введите точное имя автора книги (например, Михаил Булгаков):",
			input,
		)
		if _, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        reply,
			ReplyMarkup: suggestCancelKeyboard(),
		}); err != nil {
			b.logger.ErrorContext(ctx, "failed to send suggest step 2", slog.Any("error", err))
		}
		return true

	case stepWaitAuthor:
		session.Author = input
		title := session.Title
		session.Step = stepWaitISBN
		b.suggestMu.Unlock()

		reply := fmt.Sprintf(
			"🔢 Предложить книгу (шаг 3 из 3)\n\n"+
				"📖 Название: «%s»\n"+
				"✍️ Автор: %s\n\n"+
				"Теперь введите ISBN книги (например, 978-5-17-118366-0):",
			title,
			input,
		)
		if _, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        reply,
			ReplyMarkup: suggestCancelKeyboard(),
		}); err != nil {
			b.logger.ErrorContext(ctx, "failed to send suggest step 3", slog.Any("error", err))
		}
		return true

	case stepWaitISBN:
		title := session.Title
		author := session.Author
		delete(b.suggestSessions, chatID)
		b.suggestMu.Unlock()

		owner := formatUserHandle(update.Message.From)
		created := b.store.AddBook(model.Book{
			Title:  title,
			Author: author,
			ISBN:   input,
			Owner:  owner,
			Queue:  []string{},
		})

		reply := fmt.Sprintf("✅ Книга успешно предложена и добавлена в библиотеку!\n\n%s", created.FormatCard())
		if _, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        reply,
			ReplyMarkup: suggestDoneKeyboard(),
		}); err != nil {
			b.logger.ErrorContext(ctx, "failed to send suggest completion", slog.Any("error", err))
		}
		return true
	}

	b.suggestMu.Unlock()
	return false
}
