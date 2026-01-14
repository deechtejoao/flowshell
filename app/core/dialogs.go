package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ncruces/zenity"
)

// OpenFileDialog prompts the user to select a file to open.
// Returns the path and true if a file was selected, or empty string and false if cancelled.
func OpenFileDialog(title string, initialDir string, filters map[string]string) (string, bool, error) {
	// Zenity FileFilters
	var zenFilters []zenity.FileFilter
	var psFilters []string

	for ext, desc := range filters {
		origExt := ext
		if !strings.HasPrefix(ext, ".") {
			ext = "*." + ext
		} else {
			ext = "*" + ext
		}
		zenFilters = append(zenFilters, zenity.FileFilter{
			Name:     desc,
			Patterns: []string{ext},
		})

		// PowerShell filter: "Description (*.ext)|*.ext"
		psFilters = append(psFilters, fmt.Sprintf("%s (*.%s)|*.%s", desc, origExt, origExt))
	}
	// Also add All Files for convenience?
	zenFilters = append(zenFilters, zenity.FileFilter{Name: "All Files", Patterns: []string{"*"}})
	psFilters = append(psFilters, "All Files (*.*)|*.*")

	fmt.Printf("Opening file dialog: %s (initialDir: %s)\n", title, initialDir)

	options := []zenity.Option{
		zenity.Title(title),
		zenity.FileFilters(zenFilters),
	}
	if initialDir != "" {
		options = append(options, zenity.Filename(initialDir))
	}

	filename, err := zenity.SelectFile(options...)
	if err != nil {
		if err == zenity.ErrCanceled {
			return "", false, nil
		}
		fmt.Printf("Zenity error: %v. Attempting fallback...\n", err)
		return openFileDialogFallback(title, initialDir, strings.Join(psFilters, "|"))
	}
	fmt.Printf("File selected: %s\n", filename)
	return filename, true, nil
}

func openFileDialogFallback(title string, initialDir string, filter string) (string, bool, error) {
	if runtime.GOOS != "windows" {
		return "", false, fmt.Errorf("fallback only supported on Windows")
	}

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$f = New-Object System.Windows.Forms.OpenFileDialog
$f.Title = "%s"
if ("%s" -ne "") { $f.InitialDirectory = "%s" }
$f.Filter = "%s"
if ($f.ShowDialog() -eq "OK") { Write-Host $f.FileName }
`, title, initialDir, initialDir, filter)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", false, fmt.Errorf("fallback failed: %w", err)
	}

	path := strings.TrimSpace(out.String())
	if path == "" {
		return "", false, nil
	}
	return path, true, nil
}

// SaveFileDialog prompts the user to select a location to save a file.
// Returns the path and true if a path was selected, or empty string and false if cancelled.
func SaveFileDialog(title string, filters map[string]string) (string, bool, error) {
	var zenFilters []zenity.FileFilter
	var psFilters []string

	for ext, desc := range filters {
		origExt := ext
		if !strings.HasPrefix(ext, ".") {
			ext = "*." + ext
		} else {
			ext = "*" + ext
		}
		zenFilters = append(zenFilters, zenity.FileFilter{
			Name:     desc,
			Patterns: []string{ext},
		})
		psFilters = append(psFilters, fmt.Sprintf("%s (*.%s)|*.%s", desc, origExt, origExt))
	}
	zenFilters = append(zenFilters, zenity.FileFilter{Name: "All Files", Patterns: []string{"*"}})
	psFilters = append(psFilters, "All Files (*.*)|*.*")

	filename, err := zenity.SelectFileSave(
		zenity.Title(title),
		zenity.FileFilters(zenFilters),
		zenity.ConfirmOverwrite(),
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			return "", false, nil
		}
		fmt.Printf("Zenity Save error: %v. Attempting fallback...\n", err)
		return saveFileDialogFallback(title, "", strings.Join(psFilters, "|"))
	}
	return filename, true, nil
}

func saveFileDialogFallback(title string, initialDir string, filter string) (string, bool, error) {
	if runtime.GOOS != "windows" {
		return "", false, fmt.Errorf("fallback only supported on Windows")
	}

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$f = New-Object System.Windows.Forms.SaveFileDialog
$f.Title = "%s"
if ("%s" -ne "") { $f.InitialDirectory = "%s" }
$f.Filter = "%s"
if ($f.ShowDialog() -eq "OK") { Write-Host $f.FileName }
`, title, initialDir, initialDir, filter)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", false, fmt.Errorf("fallback failed: %w", err)
	}

	path := strings.TrimSpace(out.String())
	if path == "" {
		return "", false, nil
	}
	return path, true, nil
}
