package bot

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (b *Bot) recoverMiddleware(next tgbot.HandlerFunc) tgbot.HandlerFunc {
	return func(ctx context.Context, client *tgbot.Bot, update *models.Update) {
		defer func() {
			if r := recover(); r != nil {
				b.logger.ErrorContext(ctx, "panic recovered in update handler",
					slog.Int64("update_id", update.ID),
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
			}
		}()
		next(ctx, client, update)
	}
}

func (b *Bot) loggingMiddleware(next tgbot.HandlerFunc) tgbot.HandlerFunc {
	return func(ctx context.Context, client *tgbot.Bot, update *models.Update) {
		start := time.Now()
		next(ctx, client, update)

		attrs := []any{
			slog.Int64("update_id", update.ID),
			slog.Duration("duration", time.Since(start)),
		}
		if update.Message != nil {
			attrs = append(attrs,
				slog.Int64("chat_id", update.Message.Chat.ID),
				slog.String("text", update.Message.Text),
			)
			if update.Message.From != nil {
				attrs = append(attrs, slog.Int64("user_id", update.Message.From.ID))
			}
		}

		b.logger.InfoContext(ctx, "processed telegram update", attrs...)
	}
}
