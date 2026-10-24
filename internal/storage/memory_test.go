package storage

import (
	"errors"
	"strings"
	"testing"

	"libexchangebot/internal/model"
)

func TestMemoryStore_AddBook(t *testing.T) {
	s := NewMemoryStore()
	before := len(s.All())

	added := s.AddBook(model.Book{
		Title:  "Новая книга",
		Author: "Иван Иванов",
		ISBN:   "978-5-00-000000-1",
		Owner:  "@id7",
	})
	if added.ID == "" {
		t.Fatal("expected generated book ID")
	}
	if len(s.All()) != before+1 {
		t.Fatalf("expected %d books, got %d", before+1, len(s.All()))
	}
}

func TestNewMemoryStore_MockBooks(t *testing.T) {
	s := NewMemoryStore()
	books := s.All()
	if len(books) < 5 || len(books) > 10 {
		t.Fatalf("expected 5..10 mock books, got %d", len(books))
	}

	for _, b := range books {
		if b.Title == "" || b.ISBN == "" || b.Author == "" {
			t.Errorf("book %+v has empty required fields", b)
		}
		if !strings.HasPrefix(b.Owner, "@id") {
			t.Errorf("expected owner to start with @id, got %q", b.Owner)
		}
		for _, q := range b.Queue {
			if !strings.HasPrefix(q, "@id") {
				t.Errorf("expected queue member to start with @id, got %q", q)
			}
		}
	}
}

func TestMemoryStore_JoinQueue(t *testing.T) {
	s := NewMemoryStore()

	book, err := s.JoinQueue("2", "@id9")
	if err != nil {
		t.Fatalf("JoinQueue unexpected error: %v", err)
	}
	if len(book.Queue) != 1 || book.Queue[0] != "@id9" {
		t.Fatalf("unexpected queue: %v", book.Queue)
	}

	if _, err := s.JoinQueue("2", "@id9"); !errors.Is(err, ErrAlreadyQueued) {
		t.Errorf("expected ErrAlreadyQueued, got %v", err)
	}

	if _, err := s.JoinQueue("2", "@id2"); !errors.Is(err, ErrIsOwner) {
		t.Errorf("expected ErrIsOwner, got %v", err)
	}
}
