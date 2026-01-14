package app

import (
	"github.com/sqweek/dialog"
)

// OpenFileDialog prompts the user to select a file to open.
// Returns the path and true if a file was selected, or empty string and false if cancelled.
func OpenFileDialog(title string, filters map[string]string) (string, bool, error) {
	b := dialog.File().Title(title)
	for ext, desc := range filters {
		b = b.Filter(desc, ext)
	}

	filename, err := b.Load()
	if err != nil {
		if err == dialog.ErrCancelled {
			return "", false, nil
		}
		return "", false, err
	}
	return filename, true, nil
}

// SaveFileDialog prompts the user to select a location to save a file.
// Returns the path and true if a path was selected, or empty string and false if cancelled.
func SaveFileDialog(title string, filters map[string]string) (string, bool, error) {
	b := dialog.File().Title(title)
	for ext, desc := range filters {
		b = b.Filter(desc, ext)
	}

	filename, err := b.Save()
	if err != nil {
		if err == dialog.ErrCancelled {
			return "", false, nil
		}
		return "", false, err
	}
	return filename, true, nil
}
