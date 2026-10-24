package model

import (
	"fmt"
	"strings"
	"time"
)

// Book represents a book in the exchange library.
type Book struct {
	ID         string
	Title      string
	ISBN       string
	Author     string
	GenreID    string
	SubgenreID string
	Owner      string   // Telegram handle, e.g. "@id1"
	Queue      []string // Queue of users waiting/holding, e.g. ["@id2", "@id3"]
	AddedAt    time.Time
	Borrows    int
}

// IsAvailable reports whether the book has an empty borrowing queue.
func (b Book) IsAvailable() bool {
	return len(b.Queue) == 0
}

// FormatCard returns a formatted Russian text card with full book details.
func (b Book) FormatCard() string {
	queueText := "свободна (очередь пуста)"
	if len(b.Queue) > 0 {
		queueText = strings.Join(b.Queue, " → ")
	}

	return fmt.Sprintf(
		"📖 %s\n"+
			"✍️ Автор: %s\n"+
			"🔢 ISBN: %s\n"+
			"👤 Владелец: %s\n"+
			"👥 Очередь: %s",
		b.Title,
		b.Author,
		b.ISBN,
		b.Owner,
		queueText,
	)
}
