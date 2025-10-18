package view

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/ui"
	"github.com/rivo/tview"
)

// ProfileSwitcher represents the profile switcher modal
type ProfileSwitcher struct {
	app         *App
	profiles    []string
	inputField  *tview.InputField
	suggestions *tview.List
	container   *tview.Flex
}

// NewProfileSwitcher creates a new profile switcher modal
func NewProfileSwitcher(app *App) *ProfileSwitcher {
	inputField := tview.NewInputField()
	suggestions := tview.NewList()
	container := tview.NewFlex()

	ps := &ProfileSwitcher{
		app:         app,
		inputField:  inputField,
		suggestions: suggestions,
		container:   container,
		profiles:    []string{},
	}

	ps.loadProfiles()
	ps.setupComponents()

	return ps
}

// loadProfiles reads AWS config file and extracts profile names
func (ps *ProfileSwitcher) loadProfiles() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	var profiles []string
	profileMap := make(map[string]bool) // To avoid duplicates

	// Read config file
	configPath := os.Getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		configPath = filepath.Join(homeDir, ".aws", "config")
	}
	if file, err := os.Open(configPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
				// Extract profile name from [profile name]
				profileName := strings.TrimSpace(line[9 : len(line)-1])
				if !profileMap[profileName] {
					profiles = append(profiles, profileName)
					profileMap[profileName] = true
				}
			} else if strings.HasPrefix(line, "[default]") {
				if !profileMap["default"] {
					profiles = append(profiles, "default")
					profileMap["default"] = true
				}
			}
		}
	}

	// Read credentials file for additional profiles
	credentialsPath := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")

	if credentialsPath == "" {
		credentialsPath = filepath.Join(homeDir, ".aws", "credentials")
	}
	filepath.Join(homeDir, ".aws", "credentials")
	if file, err := os.Open(credentialsPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				// Extract profile name from [profile_name]
				profileName := strings.TrimSpace(line[1 : len(line)-1])
				if !profileMap[profileName] {
					profiles = append(profiles, profileName)
					profileMap[profileName] = true
				}
			}
		}
	}

	// Add "default" if not already present
	if !profileMap["default"] {
		profiles = append(profiles, "default")
	}

	sort.Strings(profiles)
	ps.profiles = profiles
}

// setupComponents configures the input field and autocomplete functionality
func (ps *ProfileSwitcher) setupComponents() {
	ps.setupInputField()
	ps.setupSuggestions()
	ps.setupContainer()
}

// setupInputField configures the input field
func (ps *ProfileSwitcher) setupInputField() {
	currentProfile := ps.app.profile
	if currentProfile == "" {
		currentProfile = "default"
	}

	ps.inputField.SetLabel("AWS Profile: ")
	ps.inputField.SetText(currentProfile)
	ps.inputField.SetBackgroundColor(color.Color(theme.BgColor))
	ps.inputField.SetFieldBackgroundColor(color.Color(theme.BgColor))
	ps.inputField.SetLabelColor(color.Color(theme.FgColor))
	ps.inputField.SetFieldTextColor(color.Color(theme.Cyan))
	ps.inputField.SetBorderPadding(1, 0, 2, 0)

	// Handle input changes for autocomplete
	ps.inputField.SetChangedFunc(ps.updateSuggestions)

	// Handle key events
	ps.inputField.SetInputCapture(ps.handleInputKeyEvents)

	// Handle completion
	ps.inputField.SetDoneFunc(ps.handleInputDone)
}

// setupSuggestions configures the suggestions list
func (ps *ProfileSwitcher) setupSuggestions() {
	ps.suggestions.SetBackgroundColor(color.Color(theme.BgColor))
	ps.suggestions.SetMainTextColor(color.Color(theme.FgColor))
	ps.suggestions.SetSelectedTextColor(color.Color(theme.BgColor))
	ps.suggestions.SetSelectedBackgroundColor(color.Color(theme.Cyan))
	ps.suggestions.SetBorder(false)
	ps.suggestions.ShowSecondaryText(false)
	ps.suggestions.SetBorderPadding(1, 0, 2, 0)

	// Initially show all profiles
	ps.updateSuggestionsList("")

	// Handle selection from suggestions
	ps.suggestions.SetInputCapture(ps.handleSuggestionKeyEvents)
}

// setupContainer configures the main container layout
func (ps *ProfileSwitcher) setupContainer() {
	ps.container.SetDirection(tview.FlexRow)
	ps.container.SetBackgroundColor(color.Color(theme.BgColor))
	ps.container.SetBorder(true)
	ps.container.SetBorderColor(color.Color(theme.BorderColor))
	ps.container.SetTitle(" AWS Profile Switcher ")

	// Add components to container
	ps.container.AddItem(ps.inputField, 3, 0, true)
	ps.container.AddItem(ps.suggestions, 0, 1, false)

	// Add instructions at the bottom
	instructions := tview.NewTextView()
	instructions.SetText(fmt.Sprintf("[%s]Tab: Move to suggestions | Enter: Select | Esc: Cancel", theme.Gray))
	instructions.SetTextAlign(tview.AlignCenter)
	instructions.SetBackgroundColor(color.Color(theme.BgColor))
	instructions.SetDynamicColors(true)

	ps.container.AddItem(instructions, 1, 0, false)
}

// updateSuggestions filters and updates the suggestions list based on input
func (ps *ProfileSwitcher) updateSuggestions(text string) {
	ps.updateSuggestionsList(text)
}

// updateSuggestionsList updates the suggestions list with filtered profiles
func (ps *ProfileSwitcher) updateSuggestionsList(filter string) {
	ps.suggestions.Clear()

	currentProfile := ps.app.profile
	if currentProfile == "" {
		currentProfile = "default"
	}

	for _, profile := range ps.profiles {
		if filter == "" || strings.Contains(strings.ToLower(profile), strings.ToLower(filter)) {
			displayText := profile
			if profile == currentProfile {
				displayText = fmt.Sprintf("• %s (current)", profile)
			}
			ps.suggestions.AddItem(displayText, "", 0, nil)
		}
	}
}

// handleInputKeyEvents handles keyboard events for the input field
func (ps *ProfileSwitcher) handleInputKeyEvents(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		ps.close()
		return nil
	case tcell.KeyTab:
		// Move focus to suggestions
		ps.app.SetFocus(ps.suggestions)
		return nil
	case tcell.KeyDown:
		// Move to suggestions and select first item
		ps.app.SetFocus(ps.suggestions)
		ps.suggestions.SetCurrentItem(0)
		return nil
	}
	return event
}

// handleSuggestionKeyEvents handles keyboard events for the suggestions list
func (ps *ProfileSwitcher) handleSuggestionKeyEvents(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		ps.close()
		return nil
	case tcell.KeyTab:
		// Move focus back to input
		ps.app.SetFocus(ps.inputField)
		return nil
	case tcell.KeyEnter:
		ps.selectFromSuggestions()
		return nil
	}
	return event
}

// handleInputDone handles when Enter is pressed in the input field
func (ps *ProfileSwitcher) handleInputDone(key tcell.Key) {
	if key == tcell.KeyEnter {
		text := strings.TrimSpace(ps.inputField.GetText())
		if text != "" {
			ps.switchToProfile(text)
		}
	}
}

// selectFromSuggestions selects the highlighted profile from suggestions
func (ps *ProfileSwitcher) selectFromSuggestions() {
	selectedIndex := ps.suggestions.GetCurrentItem()
	if selectedIndex < 0 {
		return
	}

	// Get the actual profile name (remove current marker if present)
	mainText, _ := ps.suggestions.GetItemText(selectedIndex)
	profileName := strings.TrimPrefix(mainText, "• ")
	profileName = strings.TrimSuffix(profileName, " (current)")

	ps.switchToProfile(profileName)
}

// switchToProfile switches to the specified profile
func (ps *ProfileSwitcher) switchToProfile(profileName string) {
	// Validate profile exists
	profileExists := false
	for _, profile := range ps.profiles {
		if profile == profileName {
			profileExists = true
			break
		}
	}

	if !profileExists {
		ps.app.Notice.Error(fmt.Sprintf("Profile '%s' not found", profileName))
		return
	}

	if err := ps.app.switchProfile(profileName); err != nil {
		ps.app.Notice.Error(fmt.Sprintf("Failed to switch profile: %v", err))
	} else {
		ps.app.Notice.Info(fmt.Sprintf("Switched to profile: %s", profileName))
	}

	ps.close()
}

// close closes the profile switcher modal
func (ps *ProfileSwitcher) close() {
	ps.app.Pages.RemovePage("profile-switcher")
}

// Show displays the profile switcher modal
func (ps *ProfileSwitcher) Show() {
	if len(ps.profiles) == 0 {
		ps.app.Notice.Warn("No AWS profiles found in ~/.aws/config or ~/.aws/credentials")
		return
	}

	// Use the same ui.Modal pattern as search modal
	modal := ui.Modal(ps.container, 80, 15, 7, ps.close)
	ps.app.Pages.AddPage("profile-switcher", modal, true, true)
	ps.app.SetFocus(ps.inputField)
}
