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
[void][System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")
$f = New-Object System.Windows.Forms.OpenFileDialog
$f.Title = "%s"
if ("%s" -ne "") { try { $f.InitialDirectory = "%s" } catch {} }
$f.Filter = "%s"
# IMPORTANT: ShowDialog() needs to run on a thread with STA (Single Thread Apartment) state.
# PowerShell by default is MTA. We can force STA by starting powershell with -Sta flag,
# but since we are running inside -Command, we rely on the parent process or use a runspace.
# However, the simplest fix for "no dialog appearing" in many contexts is to ensure
# the form actually gets focus and runs the message loop.
# System.Windows.Forms.DialogResult.OK is 1.

$result = $f.ShowDialog()
if ($result -eq "OK" -or $result -eq 1) { 
    Write-Host $f.FileName 
}
`, title, initialDir, initialDir, filter)

	// Add -Sta flag to force Single Threaded Apartment mode which is required for OLE/WinForms dialogs
	cmd := exec.Command("powershell", "-Sta", "-NoProfile", "-NonInteractive", "-Command", psScript)
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

// ShowInfoDialog shows an information dialog with the given title and message.
func ShowInfoDialog(title, message string) error {
	err := zenity.Info(message, zenity.Title(title), zenity.Width(400))
	if err != nil {
		fmt.Printf("Zenity Info error: %v. Attempting fallback...\n", err)
		return showInfoDialogFallback(title, message)
	}
	return nil
}

func showInfoDialogFallback(title, message string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("fallback only supported on Windows")
	}

	// Escape quotes in message and title for PowerShell string
	message = strings.ReplaceAll(message, "\"", "`\"")
	title = strings.ReplaceAll(title, "\"", "`\"")

	psScript := fmt.Sprintf(`
[void][System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")
[System.Windows.Forms.MessageBox]::Show("%s", "%s", [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information)
`, message, title)

	cmd := exec.Command("powershell", "-Sta", "-NoProfile", "-NonInteractive", "-Command", psScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fallback failed: %w", err)
	}
	return nil
}

// SaveFileDialog prompts the user to select a location to save a file.
// Returns the path and true if a path was selected, or empty string and false if cancelled.
func SaveFileDialog(title string, initialDir string, filters map[string]string) (string, bool, error) {
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

	options := []zenity.Option{
		zenity.Title(title),
		zenity.FileFilters(zenFilters),
		zenity.ConfirmOverwrite(),
	}
	if initialDir != "" {
		options = append(options, zenity.Filename(initialDir))
	}

	filename, err := zenity.SelectFileSave(options...)
	if err != nil {
		if err == zenity.ErrCanceled {
			return "", false, nil
		}
		fmt.Printf("Zenity Save error: %v. Attempting fallback...\n", err)
		return saveFileDialogFallback(title, initialDir, strings.Join(psFilters, "|"))
	}
	return filename, true, nil
}

func saveFileDialogFallback(title string, initialDir string, filter string) (string, bool, error) {
	if runtime.GOOS != "windows" {
		return "", false, fmt.Errorf("fallback only supported on Windows")
	}

	psScript := fmt.Sprintf(`
[void][System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")
$f = New-Object System.Windows.Forms.SaveFileDialog
$f.Title = "%s"
if ("%s" -ne "") { try { $f.InitialDirectory = "%s" } catch {} }
$f.Filter = "%s"
$result = $f.ShowDialog()
if ($result -eq "OK" -or $result -eq 1) { 
    Write-Host $f.FileName 
}
`, title, initialDir, initialDir, filter)

	cmd := exec.Command("powershell", "-Sta", "-NoProfile", "-NonInteractive", "-Command", psScript)
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
