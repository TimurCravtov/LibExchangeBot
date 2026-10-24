package bot

import (
	"context"
	"log/slog"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (b *Bot) handleStart(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	_, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome to libexchangebot! Use /help to see available commands.",
	})
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send /start response",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

func (b *Bot) handleHelp(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	helpText := "Available commands:\n" +
		"/start - Start the bot\n" +
		"/help  - Show this help message"

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

func (b *Bot) handleDefault(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	_, err := client.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Unknown command or message. Use /help to see available commands.",
	})
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send default response",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}
