package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"libexchangebot/internal/model"
	"libexchangebot/internal/storage"
)

const (
	callbackMainMenu     = "main_menu"
	callbackLibGenres    = "lib_genres"
	callbackLibRecent    = "lib_recent"
	callbackLibAuthors   = "lib_authors"
	callbackLibPopular   = "lib_popular"
	callbackLibAvailable = "lib_available"

	prefixLibGenre    = "lib_genre:"
	prefixLibSubgenre = "lib_subgenre:"
	prefixLibBook     = "lib_book:"
	prefixLibQueue    = "lib_queue:"
)

// Subgenre represents a book subcategory within a genre.
type Subgenre struct {
	ID    string
	Title string
}

// Genre represents a top-level book category containing subgenres.
type Genre struct {
	ID        string
	Title     string
	Subgenres []Subgenre
}

var libraryGenres = []Genre{
	{
		ID:    "fiction",
		Title: "📖 Художественная литература",
		Subgenres: []Subgenre{
			{ID: "sci_fi", Title: "🚀 Фантастика и фэнтези"},
			{ID: "detective", Title: "🕵️ Детективы и триллеры"},
			{ID: "classics", Title: "🏛 Классика"},
			{ID: "modern", Title: "💌 Современная проза"},
			{ID: "poetry", Title: "🎭 Поэзия и драматургия"},
		},
	},
	{
		ID:    "science",
		Title: "🧠 Научпоп и наука",
		Subgenres: []Subgenre{
			{ID: "physics", Title: "🌌 Физика и космос"},
			{ID: "biology", Title: "🧬 Биология и медицина"},
			{ID: "psychology", Title: "🧩 Психология и нейробиология"},
			{ID: "math", Title: "📐 Математика и логика"},
		},
	},
	{
		ID:    "it",
		Title: "💻 IT и программирование",
		Subgenres: []Subgenre{
			{ID: "dev", Title: "⚙️ Разработка и архитектура"},
			{ID: "ai", Title: "🤖 ИИ и Data Science"},
			{ID: "security", Title: "🔒 Кибербезопасность и сети"},
			{ID: "devops", Title: "🐧 DevOps и администрирование"},
		},
	},
	{
		ID:    "business",
		Title: "💼 Бизнес и саморазвитие",
		Subgenres: []Subgenre{
			{ID: "management", Title: "📈 Менеджмент и стартапы"},
			{ID: "finance", Title: "💰 Финансы и инвестиции"},
			{ID: "productivity", Title: "🎯 Личная эффективность"},
			{ID: "marketing", Title: "🗣 Переговоры и маркетинг"},
		},
	},
	{
		ID:    "history",
		Title: "🏛 История и общество",
		Subgenres: []Subgenre{
			{ID: "world_history", Title: "🌍 Всемирная история"},
			{ID: "biographies", Title: "👤 Биографии и мемуары"},
			{ID: "philosophy", Title: "⚖️ Философия и социология"},
			{ID: "art", Title: "🎨 Искусство и культура"},
		},
	},
	{
		ID:    "kids",
		Title: "🧸 Детская и подростковая",
		Subgenres: []Subgenre{
			{ID: "fairy_tales", Title: "🧚 Сказки и приключения"},
			{ID: "ya", Title: "🎒 Подростковая проза (Young Adult)"},
			{ID: "kids_edu", Title: "🔬 Детские энциклопедии"},
		},
	},
}

func findGenre(id string) (Genre, bool) {
	for _, g := range libraryGenres {
		if g.ID == id {
			return g, true
		}
	}
	return Genre{}, false
}

func findSubgenre(genre Genre, subID string) (Subgenre, bool) {
	for _, sg := range genre.Subgenres {
		if sg.ID == subID {
			return sg, true
		}
	}
	return Subgenre{}, false
}

func libraryMenuKeyboard() *models.InlineKeyboardMarkup {
	emptyQuery := ""
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🎭 По жанрам", CallbackData: callbackLibGenres},
			},
			{
				{Text: "🆕 Последние добавления", CallbackData: callbackLibRecent},
			},
			{
				{Text: "✍️ По авторам", CallbackData: callbackLibAuthors},
				{Text: "🔥 Популярные", CallbackData: callbackLibPopular},
			},
			{
				{Text: "✅ Доступные к обмену", CallbackData: callbackLibAvailable},
			},
			{
				{Text: "🔍 Быстрый поиск", SwitchInlineQueryCurrentChat: &emptyQuery},
			},
			{
				{Text: "⬅️ Главное меню", CallbackData: callbackMainMenu},
			},
		},
	}
}

func genresKeyboard() *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(libraryGenres)+1)
	for _, g := range libraryGenres {
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         g.Title,
				CallbackData: prefixLibGenre + g.ID,
			},
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "⬅️ Назад в библиотеку", CallbackData: callbackViewLibrary},
	})
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func subgenresKeyboard(genre Genre) *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(genre.Subgenres)+2)
	for _, sg := range genre.Subgenres {
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         sg.Title,
				CallbackData: fmt.Sprintf("%s%s:%s", prefixLibSubgenre, genre.ID, sg.ID),
			},
		})
	}
	rows = append(rows,
		[]models.InlineKeyboardButton{
			{
				Text:         "📚 Все книги жанра",
				CallbackData: fmt.Sprintf("%s%s:all", prefixLibSubgenre, genre.ID),
			},
		},
		[]models.InlineKeyboardButton{
			{Text: "⬅️ К списку жанров", CallbackData: callbackLibGenres},
		},
	)
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func booksListKeyboard(books []model.Book, backCallback, backLabel string) *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(books)+2)
	for _, bk := range books {
		statusIcon := "✅"
		if !bk.IsAvailable() {
			statusIcon = "⏳"
		}
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("%s %s — %s", statusIcon, bk.Title, bk.Author),
				CallbackData: prefixLibBook + bk.ID,
			},
		})
	}
	if backCallback != "" && backCallback != callbackViewLibrary {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: backLabel, CallbackData: backCallback},
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "⬅️ Назад в библиотеку", CallbackData: callbackViewLibrary},
	})
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func bookCardKeyboard(book model.Book) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         fmt.Sprintf("🙋 Встать в очередь (%d чел.)", len(book.Queue)),
					CallbackData: prefixLibQueue + book.ID,
				},
			},
			{
				{
					Text:         "⬅️ К поджанрам",
					CallbackData: prefixLibGenre + book.GenreID,
				},
				{
					Text:         "📚 В библиотеку",
					CallbackData: callbackViewLibrary,
				},
			},
		},
	}
}

func formatBooksSummary(header string, books []model.Book) string {
	if len(books) == 0 {
		return header + "\n\nСписок книг пока пуст."
	}

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n\n")
	for i, bk := range books {
		queueInfo := "свободна"
		if len(bk.Queue) > 0 {
			queueInfo = fmt.Sprintf("очередь: %s", strings.Join(bk.Queue, ", "))
		}
		fmt.Fprintf(&sb, "%d. «%s» — %s\n   ISBN: %s | Владелец: %s (%s)\n\n",
			i+1, bk.Title, bk.Author, bk.ISBN, bk.Owner, queueInfo)
	}
	sb.WriteString("Нажмите на книгу ниже для подробной информации или записи в очередь:")
	return sb.String()
}

func (b *Bot) handleViewLibrary(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		"📚 Выберите способ просмотра библиотеки:",
		libraryMenuKeyboard(),
	)
}

func (b *Bot) handleLibGenres(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		"🎭 Выберите жанр:",
		genresKeyboard(),
	)
}

func (b *Bot) handleLibGenreSelect(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	genreID := strings.TrimPrefix(update.CallbackQuery.Data, prefixLibGenre)
	genre, ok := findGenre(genreID)
	if !ok {
		b.respondOrEditCallback(ctx, client, update, "Жанр не найден.", genresKeyboard())
		return
	}

	text := fmt.Sprintf("%s\n\nВыберите поджанр:", genre.Title)
	b.respondOrEditCallback(ctx, client, update, text, subgenresKeyboard(genre))
}

func (b *Bot) handleLibSubgenreSelect(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	payload := strings.TrimPrefix(update.CallbackQuery.Data, prefixLibSubgenre)
	parts := strings.SplitN(payload, ":", 2)
	if len(parts) != 2 {
		b.respondOrEditCallback(ctx, client, update, "Некорректный выбор категории.", genresKeyboard())
		return
	}

	genreID, subID := parts[0], parts[1]
	genre, ok := findGenre(genreID)
	if !ok {
		b.respondOrEditCallback(ctx, client, update, "Жанр не найден.", genresKeyboard())
		return
	}

	if subID == "all" {
		books := b.store.ByGenre(genre.ID)
		header := fmt.Sprintf("📚 Все книги в жанре «%s»:", genre.Title)
		b.respondOrEditCallback(
			ctx,
			client,
			update,
			formatBooksSummary(header, books),
			booksListKeyboard(books, prefixLibGenre+genre.ID, "⬅️ Назад к поджанрам"),
		)
		return
	}

	subgenre, ok := findSubgenre(genre, subID)
	if !ok {
		b.respondOrEditCallback(ctx, client, update, "Поджанр не найден.", subgenresKeyboard(genre))
		return
	}

	books := b.store.BySubgenre(genre.ID, subgenre.ID)
	header := fmt.Sprintf("%s → %s", genre.Title, subgenre.Title)
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		formatBooksSummary(header, books),
		booksListKeyboard(books, prefixLibGenre+genre.ID, "⬅️ Назад к поджанрам"),
	)
}

func (b *Bot) handleLibBookSelect(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	bookID := strings.TrimPrefix(update.CallbackQuery.Data, prefixLibBook)
	book, ok := b.store.GetByID(bookID)
	if !ok {
		b.respondOrEditCallback(ctx, client, update, "Книга не найдена.", libraryMenuKeyboard())
		return
	}

	b.respondOrEditCallback(ctx, client, update, book.FormatCard(), bookCardKeyboard(book))
}

func (b *Bot) handleLibJoinQueue(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	bookID := strings.TrimPrefix(update.CallbackQuery.Data, prefixLibQueue)
	userHandle := formatUserHandle(&update.CallbackQuery.From)

	book, err := b.store.JoinQueue(bookID, userHandle)
	var alertText string
	switch {
	case err == nil:
		alertText = fmt.Sprintf("Вы (%s) добавлены в очередь!", userHandle)
	case errors.Is(err, storage.ErrAlreadyQueued):
		alertText = "Вы уже находитесь в очереди на эту книгу."
	case errors.Is(err, storage.ErrIsOwner):
		alertText = "Вы являетесь владельцем этой книги."
	default:
		alertText = "Книга не найдена."
	}

	if _, ansErr := client.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            alertText,
		ShowAlert:       err != nil,
	}); ansErr != nil {
		b.logger.ErrorContext(ctx, "failed to answer queue callback", slog.Any("error", ansErr))
	}

	if err == nil && update.CallbackQuery.Message.Message != nil {
		msg := update.CallbackQuery.Message.Message
		if _, editErr := client.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.ID,
			Text:        book.FormatCard(),
			ReplyMarkup: bookCardKeyboard(book),
		}); editErr != nil {
			b.logger.ErrorContext(ctx, "failed to update book card after joining queue", slog.Any("error", editErr))
		}
	}
}

func (b *Bot) handleLibRecent(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	books := b.store.Recent(10)
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		formatBooksSummary("🆕 Последние добавленные книги:", books),
		booksListKeyboard(books, "", ""),
	)
}

func (b *Bot) handleLibAuthors(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	books := b.store.All()
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		formatBooksSummary("✍️ Книги по авторам:", books),
		booksListKeyboard(books, "", ""),
	)
}

func (b *Bot) handleLibPopular(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	books := b.store.Popular(10)
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		formatBooksSummary("🔥 Популярные книги:", books),
		booksListKeyboard(books, "", ""),
	)
}

func (b *Bot) handleLibAvailable(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	books := b.store.Available()
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		formatBooksSummary("✅ Книги, доступные к обмену (без очереди):", books),
		booksListKeyboard(books, "", ""),
	)
}

func formatUserHandle(u *models.User) string {
	if u == nil {
		return "@id0"
	}
	if u.Username != "" {
		return "@" + u.Username
	}
	return fmt.Sprintf("@id%d", u.ID)
}
