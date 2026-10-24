package bot

import (
	"fmt"
	"testing"

	"libexchangebot/internal/model"
)

func TestPaginateBooks(t *testing.T) {
	books := make([]model.Book, 8)
	for i := range books {
		books[i] = model.Book{ID: fmt.Sprintf("%d", i+1), Title: fmt.Sprintf("Book %d", i+1)}
	}

	page0, curPage, totalPages, startIdx := paginateBooks(books, 0, 3)
	if len(page0) != 3 || curPage != 0 || totalPages != 3 || startIdx != 0 {
		t.Fatalf("page 0 mismatch: len=%d cur=%d total=%d start=%d", len(page0), curPage, totalPages, startIdx)
	}

	page2, curPage, totalPages, startIdx := paginateBooks(books, 2, 3)
	if len(page2) != 2 || curPage != 2 || totalPages != 3 || startIdx != 6 {
		t.Fatalf("page 2 mismatch: len=%d cur=%d total=%d start=%d", len(page2), curPage, totalPages, startIdx)
	}

	kb := paginatedBooksKeyboard(page0, 0, totalPages, "lib_all", "", "")
	if len(kb.InlineKeyboard) != 5 { // 3 books + 1 pagination row + 1 back row
		t.Fatalf("expected 5 keyboard rows, got %d", len(kb.InlineKeyboard))
	}
	navRow := kb.InlineKeyboard[3]
	if len(navRow) != 3 || navRow[0].Text != "<" || navRow[1].Text != "1 / 3" || navRow[2].Text != ">" {
		t.Fatalf("unexpected navigation row: %+v", navRow)
	}
	if navRow[0].CallbackData != "lib_noop:first" {
		t.Errorf("expected first-page noop callback on '<', got %q", navRow[0].CallbackData)
	}
	if navRow[2].CallbackData != "lib_all:1" {
		t.Errorf("expected next-page callback 'lib_all:1' on '>', got %q", navRow[2].CallbackData)
	}
}
