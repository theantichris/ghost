package ui

import (
	"time"

	"github.com/theantichris/ghost/v3/internal/storage"
)

type threadItem struct {
	thread storage.Thread
}

// Title returns the thread's title.
func (item threadItem) Title() string {
	return item.thread.Title
}

// Description returns the thread's formatted update timestamp.
func (item threadItem) Description() string {
	return item.thread.UpdatedAt.Format(time.ANSIC)
}

// FilterValue returns the title for the filter to search against.
func (item threadItem) FilterValue() string {
	return item.thread.Title
}
