package bot

import (
	"context"
	"fmt"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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

func backToLibraryKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "⬅️ Назад в библиотеку", CallbackData: callbackViewLibrary},
			},
		},
	}
}

func backToSubgenresKeyboard(genreID string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "⬅️ Назад к поджанрам", CallbackData: prefixLibGenre + genreID},
			},
			{
				{Text: "📚 В меню библиотеки", CallbackData: callbackViewLibrary},
			},
		},
	}
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
		text := fmt.Sprintf("📚 Все книги в жанре «%s»:\n\nСписок книг пока пуст.", genre.Title)
		b.respondOrEditCallback(ctx, client, update, text, backToSubgenresKeyboard(genre.ID))
		return
	}

	subgenre, ok := findSubgenre(genre, subID)
	if !ok {
		b.respondOrEditCallback(ctx, client, update, "Поджанр не найден.", subgenresKeyboard(genre))
		return
	}

	text := fmt.Sprintf("%s → %s\n\nСписок книг в этом разделе пока пуст.", genre.Title, subgenre.Title)
	b.respondOrEditCallback(ctx, client, update, text, backToSubgenresKeyboard(genre.ID))
}

func (b *Bot) handleLibRecent(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		"🆕 Последние добавленные книги:\n\nНовых поступлений пока нет.",
		backToLibraryKeyboard(),
	)
}

func (b *Bot) handleLibAuthors(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		"✍️ Поиск по авторам:\n\nСписок авторов пока пуст.",
		backToLibraryKeyboard(),
	)
}

func (b *Bot) handleLibPopular(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		"🔥 Популярные книги:\n\nСписок популярных книг пока пуст.",
		backToLibraryKeyboard(),
	)
}

func (b *Bot) handleLibAvailable(ctx context.Context, client *tgbot.Bot, update *models.Update) {
	b.respondOrEditCallback(
		ctx,
		client,
		update,
		"✅ Книги, доступные к обмену прямо сейчас:\n\nСписок пока пуст.",
		backToLibraryKeyboard(),
	)
}
