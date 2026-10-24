package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"libexchangebot/internal/model"
	"libexchangebot/internal/storage"
)

const (
	booksPerPage = 3

	callbackMainMenu     = "main_menu"
	callbackLibGenres    = "lib_genres"
	callbackLibAll       = "lib_all"
	callbackLibRecent    = "lib_recent"
	callbackLibAuthors   = "lib_authors"
	callbackLibPopular   = "lib_popular"
	callbackLibAvailable = "lib_available"

	prefixLibGenre    = "lib_genre:"
	prefixLibSubgenre = "lib_subgenre:"
	prefixLibBook     = "lib_book:"
	prefixLibQueue    = "lib_queue:"
	prefixLibNoop     = "lib_noop:"
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
				{Text: "📚 Все книги", CallbackData: callbackLibAll},
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

func (b *Bot) genresKeyboard() *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(libraryGenres)+1)
	for _, g := range libraryGenres {
		count := len(b.store.ByGenre(g.ID))
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("%s (%d)", g.Title, count),
				CallbackData: prefixLibGenre + g.ID,
			},
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "⬅️ Назад в библиотеку", CallbackData: callbackViewLibrary},
	})
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (b *Bot) subgenresKeyboard(genre Genre) *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(genre.Subgenres)+2)
	for _, sg := range genre.Subgenres {
		count := len(b.store.BySubgenre(genre.ID, sg.ID))
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("%s (%d)", sg.Title, count),
				CallbackData: fmt.Sprintf("%s%s:%s:0", prefixLibSubgenre, genre.ID, sg.ID),
			},
		})
	}
	genreTotal := len(b.store.ByGenre(genre.ID))
	rows = append(rows,
		[]models.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("📚 Все книги жанра (%d)", genreTotal),
				CallbackData: fmt.Sprintf("%s%s:all:0", prefixLibSubgenre, genre.ID),
			},
		},
		[]models.InlineKeyboardButton{
			{Text: "⬅️ К списку жанров", CallbackData: callbackLibGenres},
		},
	)
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func paginateBooks(books []model.Book, page, perPage int) (pageBooks []model.Book, curPage, totalPages, startIdx int) {
	if perPage <= 0 {
		perPage = booksPerPage
	}
	if len(books) == 0 {
		return nil, 0, 1, 0
	}

	totalPages = (len(books) + perPage - 1) / perPage
	curPage = page
	if curPage < 0 {
		curPage = 0
	}
	if curPage >= totalPages {
		curPage = totalPages - 1
	}

	startIdx = curPage * perPage
	endIdx := startIdx + perPage
	if endIdx > len(books) {
		endIdx = len(books)
	}
	return books[startIdx:endIdx], curPage, totalPages, startIdx
}

func paginatedBooksKeyboard(
	pageBooks []model.Book,
	curPage int,
	totalPages int,
	pageCallbackPrefix string,
	backCallback string,
	backLabel string,
) *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(pageBooks)+3)
	for _, bk := range pageBooks {
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

	prevCallback := prefixLibNoop + "first"
	if curPage > 0 {
		prevCallback = fmt.Sprintf("%s:%d", pageCallbackPrefix, curPage-1)
	}

	nextCallback := prefixLibNoop + "last"
	if curPage+1 < totalPages {
		nextCallback = fmt.Sprintf("%s:%d", pageCallbackPrefix, curPage+1)
	}

	rows = append(rows, []models.InlineKeyboardButton{
		{
			Text:         "<",
			CallbackData: prevCallback,
		},
		{
			Text:         fmt.Sprintf("%d / %d", curPage+1, totalPages),
			CallbackData: fmt.Sprintf("%spage:%d:%d", prefixLibNoop, curPage+1, totalPages),
		},
		{
			Text:         ">",
			CallbackData: nextCallback,
		},
	})

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

func formatBooksPage(header string, pageBooks []model.Book, startIdx, totalBooks, curPage, totalPages int) string {
	if totalBooks == 0 {
		return header + "\n\nСписок книг пока пуст."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s (стр. %d из %d, всего: %d)\n\n", header, curPage+1, totalPages, totalBooks)
	for i, bk := range pageBooks {
		queueInfo := "свободна"
		if len(bk.Queue) > 0 {
			queueInfo = fmt.Sprintf("очередь: %s", strings.Join(bk.Queue, ", "))
		}
		fmt.Fprintf(&sb, "%d. «%s» — %s\n   ISBN: %s | Владелец: %s (%s)\n\n",
			startIdx+i+1, bk.Title, bk.Author, bk.ISBN, bk.Owner, queueInfo)
	}
	sb.WriteString("Нажмите на книгу ниже для подробной информации или записи в очередь:")
	return sb.String()
}

func (b *Bot) renderPaginatedBooks(
	ctx context.Context,
	client *tgbot.Bot,
	update *models.Update,
	header string,
	books []model.Book,
	page int,
	pageCallbackPrefix string,
	backCallback string,
	backLabel string,
) {
	pageBooks, curPage, totalPages, startIdx := paginateBooks(books, page, booksPerPage)
	text := formatBooksPage(header, pageBooks, startIdx, len(books), curPage, totalPages)
	kb := paginatedBooksKeyboard(pageBooks, curPage, totalPages, pageCallbackPrefix, backCallback, backLabel)
	b.respondOrEditCallback(ctx, client, update, text, kb)
}

func parsePageSuffix(data, prefix string) int {
	rest := strings.TrimPrefix(data, prefix)
	rest = strings.TrimPrefix(rest, ":")
	if rest == "" {
		return 0
	}
	page, err := strconv.Atoi(rest)
	if err != nil || page < 0 {
		return 0
	}
	return page
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
		b.genresKeyboard(),
	)
}

func (b *Bot) handleLibGenreSelect(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	genreID := strings.TrimPrefix(update.CallbackQuery.Data, prefixLibGenre)
	genre, ok := findGenre(genreID)
	if !ok {
		b.respondOrEditCallback(ctx, client, update, "Жанр не найден.", b.genresKeyboard())
		return
	}

	genreCount := len(b.store.ByGenre(genre.ID))
	text := fmt.Sprintf("%s (%d)\n\nВыберите поджанр:", genre.Title, genreCount)
	b.respondOrEditCallback(ctx, client, update, text, b.subgenresKeyboard(genre))
}

func (b *Bot) handleLibSubgenreSelect(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	payload := strings.TrimPrefix(update.CallbackQuery.Data, prefixLibSubgenre)
	parts := strings.Split(payload, ":")
	if len(parts) < 2 {
		b.respondOrEditCallback(ctx, client, update, "Некорректный выбор категории.", b.genresKeyboard())
		return
	}

	genreID, subID := parts[0], parts[1]
	page := 0
	if len(parts) >= 3 {
		if p, err := strconv.Atoi(parts[2]); err == nil && p >= 0 {
			page = p
		}
	}

	genre, ok := findGenre(genreID)
	if !ok {
		b.respondOrEditCallback(ctx, client, update, "Жанр не найден.", b.genresKeyboard())
		return
	}

	pagePrefix := prefixLibSubgenre + genre.ID + ":" + subID

	if subID == "all" {
		books := b.store.ByGenre(genre.ID)
		header := fmt.Sprintf("📚 Все книги в жанре «%s»", genre.Title)
		b.renderPaginatedBooks(
			ctx,
			client,
			update,
			header,
			books,
			page,
			pagePrefix,
			prefixLibGenre+genre.ID,
			"⬅️ Назад к поджанрам",
		)
		return
	}

	subgenre, ok := findSubgenre(genre, subID)
	if !ok {
		b.respondOrEditCallback(ctx, client, update, "Поджанр не найден.", b.subgenresKeyboard(genre))
		return
	}

	books := b.store.BySubgenre(genre.ID, subgenre.ID)
	header := fmt.Sprintf("%s → %s", genre.Title, subgenre.Title)
	b.renderPaginatedBooks(
		ctx,
		client,
		update,
		header,
		books,
		page,
		pagePrefix,
		prefixLibGenre+genre.ID,
		"⬅️ Назад к поджанрам",
	)
}

func (b *Bot) handleLibNoop(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	reason := strings.TrimPrefix(update.CallbackQuery.Data, prefixLibNoop)
	var text string
	switch {
	case reason == "first":
		text = "Это первая страница"
	case reason == "last":
		text = "Это последняя страница"
	case strings.HasPrefix(reason, "page:"):
		parts := strings.Split(strings.TrimPrefix(reason, "page:"), ":")
		if len(parts) == 2 {
			text = fmt.Sprintf("Страница %s из %s", parts[0], parts[1])
		}
	}

	if _, err := client.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            text,
	}); err != nil {
		b.logger.ErrorContext(ctx, "failed to answer noop callback", slog.Any("error", err))
	}
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

	if err == nil {
		if update.CallbackQuery.InlineMessageID != "" {
			if _, editErr := client.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				InlineMessageID: update.CallbackQuery.InlineMessageID,
				Text:            book.FormatCard(),
				ReplyMarkup:     bookCardKeyboard(book),
			}); editErr != nil && !strings.Contains(editErr.Error(), "message is not modified") {
				b.logger.ErrorContext(ctx, "failed to update inline book card after joining queue", slog.Any("error", editErr))
			}
		} else if update.CallbackQuery.Message.Message != nil {
			msg := update.CallbackQuery.Message.Message
			if _, editErr := client.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				ChatID:      msg.Chat.ID,
				MessageID:   msg.ID,
				Text:        book.FormatCard(),
				ReplyMarkup: bookCardKeyboard(book),
			}); editErr != nil && !strings.Contains(editErr.Error(), "message is not modified") {
				b.logger.ErrorContext(ctx, "failed to update book card after joining queue", slog.Any("error", editErr))
			}
		}
	}
}

func (b *Bot) handleLibAll(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	page := parsePageSuffix(update.CallbackQuery.Data, callbackLibAll)
	books := b.store.All()
	b.renderPaginatedBooks(ctx, client, update, "📚 Все книги библиотеки", books, page, callbackLibAll, "", "")
}

func (b *Bot) handleLibRecent(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	page := parsePageSuffix(update.CallbackQuery.Data, callbackLibRecent)
	books := b.store.Recent(0)
	b.renderPaginatedBooks(ctx, client, update, "🆕 Последние добавленные книги", books, page, callbackLibRecent, "", "")
}

func (b *Bot) handleLibAuthors(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	page := parsePageSuffix(update.CallbackQuery.Data, callbackLibAuthors)
	books := b.store.All()
	b.renderPaginatedBooks(ctx, client, update, "✍️ Книги по авторам", books, page, callbackLibAuthors, "", "")
}

func (b *Bot) handleLibPopular(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	page := parsePageSuffix(update.CallbackQuery.Data, callbackLibPopular)
	books := b.store.Popular(0)
	b.renderPaginatedBooks(ctx, client, update, "🔥 Популярные книги", books, page, callbackLibPopular, "", "")
}

func (b *Bot) handleLibAvailable(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	page := parsePageSuffix(update.CallbackQuery.Data, callbackLibAvailable)
	books := b.store.Available()
	b.renderPaginatedBooks(ctx, client, update, "✅ Книги, доступные к обмену", books, page, callbackLibAvailable, "", "")
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
