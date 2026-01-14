package app

import (
	"strings"

	"github.com/ncruces/zenity"
)

// OpenFileDialog prompts the user to select a file to open.
// Returns the path and true if a file was selected, or empty string and false if cancelled.
func OpenFileDialog(title string, filters map[string]string) (string, bool, error) {
	// Zenity FileFilters
	var zenFilters []zenity.FileFilter
	for ext, desc := range filters {
		if !strings.HasPrefix(ext, ".") {
			ext = "*." + ext
		} else {
			ext = "*" + ext
		}
		zenFilters = append(zenFilters, zenity.FileFilter{
			Name:     desc,
			Patterns: []string{ext},
		})
	}
	// Also add All Files for convenience?
	zenFilters = append(zenFilters, zenity.FileFilter{Name: "All Files", Patterns: []string{"*"}})

	filename, err := zenity.SelectFile(
		zenity.Title(title),
		zenity.FileFilters(zenFilters),
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			return "", false, nil
		}
		return "", false, err
	}
	return filename, true, nil
}

// SaveFileDialog prompts the user to select a location to save a file.
// Returns the path and true if a path was selected, or empty string and false if cancelled.
func SaveFileDialog(title string, filters map[string]string) (string, bool, error) {
	var zenFilters []zenity.FileFilter
	for ext, desc := range filters {
		if !strings.HasPrefix(ext, ".") {
			ext = "*." + ext
		} else {
			ext = "*" + ext
		}
		zenFilters = append(zenFilters, zenity.FileFilter{
			Name:     desc,
			Patterns: []string{ext},
		})
	}
	zenFilters = append(zenFilters, zenity.FileFilter{Name: "All Files", Patterns: []string{"*"}})

	filename, err := zenity.SelectFileSave(
		zenity.Title(title),
		zenity.FileFilters(zenFilters),
		zenity.ConfirmOverwrite(),
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			return "", false, nil
		}
		return "", false, err
	}
	return filename, true, nil
}
