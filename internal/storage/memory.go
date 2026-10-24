package storage

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"libexchangebot/internal/model"
)

var (
	ErrBookNotFound  = errors.New("book not found")
	ErrAlreadyQueued = errors.New("user is already in the queue")
	ErrIsOwner       = errors.New("owner cannot join the queue for their own book")
)

// MemoryStore is an in-memory thread-safe book repository.
type MemoryStore struct {
	mu    sync.RWMutex
	books []model.Book
}

// NewMemoryStore creates an in-memory book repository pre-populated with mock books.
func NewMemoryStore() *MemoryStore {
	now := time.Now()
	return &MemoryStore{
		books: []model.Book{
			{
				ID:         "1",
				Title:      "Дюна",
				ISBN:       "978-5-17-118366-0",
				Author:     "Фрэнк Герберт",
				GenreID:    "fiction",
				SubgenreID: "sci_fi",
				Owner:      "@id1",
				Queue:      []string{"@id2", "@id3"},
				AddedAt:    now.Add(-72 * time.Hour),
				Borrows:    12,
			},
			{
				ID:         "2",
				Title:      "Мастер и Маргарита",
				ISBN:       "978-5-389-01645-0",
				Author:     "Михаил Булгаков",
				GenreID:    "fiction",
				SubgenreID: "classics",
				Owner:      "@id2",
				Queue:      []string{},
				AddedAt:    now.Add(-60 * time.Hour),
				Borrows:    15,
			},
			{
				ID:         "3",
				Title:      "Чистый код",
				ISBN:       "978-5-4461-0960-9",
				Author:     "Роберт Мартин",
				GenreID:    "it",
				SubgenreID: "dev",
				Owner:      "@id3",
				Queue:      []string{"@id1", "@id4", "@id5"},
				AddedAt:    now.Add(-48 * time.Hour),
				Borrows:    20,
			},
			{
				ID:         "4",
				Title:      "Язык программирования Go",
				ISBN:       "978-5-8459-2051-5",
				Author:     "Алан Донован, Брайан Керниган",
				GenreID:    "it",
				SubgenreID: "dev",
				Owner:      "@id1",
				Queue:      []string{"@id2"},
				AddedAt:    now.Add(-36 * time.Hour),
				Borrows:    18,
			},
			{
				ID:         "5",
				Title:      "Краткая история времени",
				ISBN:       "978-5-17-091881-2",
				Author:     "Стивен Хокинг",
				GenreID:    "science",
				SubgenreID: "physics",
				Owner:      "@id4",
				Queue:      []string{},
				AddedAt:    now.Add(-24 * time.Hour),
				Borrows:    9,
			},
			{
				ID:         "6",
				Title:      "Думай медленно… решай быстро",
				ISBN:       "978-5-17-080053-7",
				Author:     "Даниэль Канеман",
				GenreID:    "science",
				SubgenreID: "psychology",
				Owner:      "@id2",
				Queue:      []string{"@id5"},
				AddedAt:    now.Add(-18 * time.Hour),
				Borrows:    14,
			},
			{
				ID:         "7",
				Title:      "Пиши, сокращай",
				ISBN:       "978-5-9614-6060-5",
				Author:     "Максим Ильяхов, Людмила Сарычева",
				GenreID:    "business",
				SubgenreID: "marketing",
				Owner:      "@id5",
				Queue:      []string{},
				AddedAt:    now.Add(-12 * time.Hour),
				Borrows:    11,
			},
			{
				ID:         "8",
				Title:      "Sapiens. Краткая история человечества",
				ISBN:       "978-5-906837-56-1",
				Author:     "Юваль Ной Харари",
				GenreID:    "history",
				SubgenreID: "world_history",
				Owner:      "@id3",
				Queue:      []string{"@id1", "@id2"},
				AddedAt:    now.Add(-6 * time.Hour),
				Borrows:    16,
			},
			{
				ID:         "9",
				Title:      "Задача трёх тел",
				ISBN:       "978-5-04-089167-2",
				Author:     "Лю Цысинь",
				GenreID:    "fiction",
				SubgenreID: "sci_fi",
				Owner:      "@id4",
				Queue:      []string{"@id1"},
				AddedAt:    now.Add(-3 * time.Hour),
				Borrows:    19,
			},
			{
				ID:         "10",
				Title:      "Высоконагруженные приложения",
				ISBN:       "978-5-4461-0512-0",
				Author:     "Мартин Клеппман",
				GenreID:    "it",
				SubgenreID: "dev",
				Owner:      "@id5",
				Queue:      []string{"@id2", "@id4"},
				AddedAt:    now.Add(-1 * time.Hour),
				Borrows:    22,
			},
		},
	}
}

// All returns a copy of all books.
func (s *MemoryStore) All() []model.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]model.Book, len(s.books))
	for i, b := range s.books {
		out[i] = cloneBook(b)
	}
	return out
}

// GetByID finds a book by its ID.
func (s *MemoryStore) GetByID(id string) (model.Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, b := range s.books {
		if b.ID == id {
			return cloneBook(b), true
		}
	}
	return model.Book{}, false
}

// ByGenre returns all books in a given genre.
func (s *MemoryStore) ByGenre(genreID string) []model.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []model.Book
	for _, b := range s.books {
		if b.GenreID == genreID {
			out = append(out, cloneBook(b))
		}
	}
	return out
}

// BySubgenre returns all books in a given genre and subgenre.
func (s *MemoryStore) BySubgenre(genreID, subgenreID string) []model.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []model.Book
	for _, b := range s.books {
		if b.GenreID == genreID && b.SubgenreID == subgenreID {
			out = append(out, cloneBook(b))
		}
	}
	return out
}

// Recent returns books sorted from newest to oldest (up to limit).
func (s *MemoryStore) Recent(limit int) []model.Book {
	books := s.All()
	slices.SortFunc(books, func(a, b model.Book) int {
		return b.AddedAt.Compare(a.AddedAt)
	})
	if limit > 0 && len(books) > limit {
		return books[:limit]
	}
	return books
}

// Popular returns books sorted by borrow count descending (up to limit).
func (s *MemoryStore) Popular(limit int) []model.Book {
	books := s.All()
	slices.SortFunc(books, func(a, b model.Book) int {
		return b.Borrows - a.Borrows
	})
	if limit > 0 && len(books) > limit {
		return books[:limit]
	}
	return books
}

// Available returns books that currently have an empty queue.
func (s *MemoryStore) Available() []model.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []model.Book
	for _, b := range s.books {
		if b.IsAvailable() {
			out = append(out, cloneBook(b))
		}
	}
	return out
}

// Search finds books matching query in title, author, ISBN, or owner.
func (s *MemoryStore) Search(query string) []model.Book {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return s.All()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []model.Book
	for _, b := range s.books {
		if strings.Contains(strings.ToLower(b.Title), q) ||
			strings.Contains(strings.ToLower(b.Author), q) ||
			strings.Contains(strings.ToLower(b.ISBN), q) ||
			strings.Contains(strings.ToLower(b.Owner), q) {
			out = append(out, cloneBook(b))
		}
	}
	return out
}

// JoinQueue adds userHandle (e.g. "@id4") to the book's borrowing queue.
func (s *MemoryStore) JoinQueue(bookID, userHandle string) (model.Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.books {
		if s.books[i].ID != bookID {
			continue
		}
		if strings.EqualFold(s.books[i].Owner, userHandle) {
			return cloneBook(s.books[i]), ErrIsOwner
		}
		for _, u := range s.books[i].Queue {
			if strings.EqualFold(u, userHandle) {
				return cloneBook(s.books[i]), ErrAlreadyQueued
			}
		}
		s.books[i].Queue = append(s.books[i].Queue, userHandle)
		return cloneBook(s.books[i]), nil
	}
	return model.Book{}, ErrBookNotFound
}

// AddBook adds a new book to the in-memory store and assigns it an ID and timestamp.
func (s *MemoryStore) AddBook(book model.Book) model.Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	book.ID = fmt.Sprintf("%d", len(s.books)+1)
	if book.AddedAt.IsZero() {
		book.AddedAt = time.Now()
	}
	if book.Queue == nil {
		book.Queue = []string{}
	}
	if book.GenreID == "" {
		book.GenreID = "fiction"
		book.SubgenreID = "modern"
	}

	s.books = append(s.books, cloneBook(book))
	return cloneBook(book)
}

func cloneBook(b model.Book) model.Book {
	cp := b
	cp.Queue = slices.Clone(b.Queue)
	return cp
}
