package bot

import (
	"context"
	"fmt"
	"log/slog"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"libexchangebot/internal/config"
)

// Bot wraps the Telegram bot client and application dependencies.
type Bot struct {
	client *tgbot.Bot
	logger *slog.Logger
}

// New creates and configures a new Telegram Bot instance.
func New(cfg *config.Config, logger *slog.Logger) (*Bot, error) {
	b := &Bot{
		logger: logger,
	}

	opts := []tgbot.Option{
		tgbot.WithMiddlewares(
			b.recoverMiddleware,
			b.loggingMiddleware,
		),
		tgbot.WithDefaultHandler(b.handleDefault),
		tgbot.WithErrorsHandler(func(err error) {
			logger.Error("telegram bot error", slog.Any("error", err))
		}),
	}

	if cfg.Debug {
		opts = append(opts, tgbot.WithDebugHandler(func(format string, args ...any) {
			logger.Debug(fmt.Sprintf(format, args...))
		}))
	}

	client, err := tgbot.New(cfg.TelegramToken, opts...)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}

	b.client = client
	b.registerHandlers()

	return b, nil
}

// Start registers bot commands with Telegram and starts long polling until ctx is canceled.
func (b *Bot) Start(ctx context.Context) error {
	if _, err := b.client.SetMyCommands(ctx, &tgbot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "start", Description: "Главное меню"},
			{Command: "help", Description: "Помощь"},
		},
	}); err != nil {
		b.logger.WarnContext(ctx, "failed to set bot commands", slog.Any("error", err))
	}

	b.logger.InfoContext(ctx, "telegram bot started (long polling)")
	b.client.Start(ctx)
	b.logger.InfoContext(ctx, "telegram bot stopped")
	return nil
}

func (b *Bot) registerHandlers() {
	b.client.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, b.handleStart)
	b.client.RegisterHandler(tgbot.HandlerTypeMessageText, "/help", tgbot.MatchTypeExact, b.handleHelp)

	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackMainMenu, tgbot.MatchTypeExact, b.handleMainMenuCallback)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackViewLibrary, tgbot.MatchTypeExact, b.handleViewLibrary)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackSuggestBook, tgbot.MatchTypeExact, b.handleSuggestBook)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackAdminPanel, tgbot.MatchTypeExact, b.handleAdminPanel)

	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackLibGenres, tgbot.MatchTypeExact, b.handleLibGenres)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, prefixLibGenre, tgbot.MatchTypePrefix, b.handleLibGenreSelect)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, prefixLibSubgenre, tgbot.MatchTypePrefix, b.handleLibSubgenreSelect)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackLibRecent, tgbot.MatchTypeExact, b.handleLibRecent)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackLibAuthors, tgbot.MatchTypeExact, b.handleLibAuthors)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackLibPopular, tgbot.MatchTypeExact, b.handleLibPopular)
	b.client.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, callbackLibAvailable, tgbot.MatchTypeExact, b.handleLibAvailable)
}
