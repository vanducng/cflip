package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vanducng/cflip/internal/config"
)

var (
	quitTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9B9B9B"))
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)
	indentString = "  "
)

// compactDelegate is a minimal item delegate for compact rendering
type compactDelegate struct{}

func (d compactDelegate) Height() int                               { return 1 }
func (d compactDelegate) Spacing() int                              { return 0 }
func (d compactDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d compactDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	cursor := indentString
	text := i.title

	if index == m.Index() {
		cursor = "▶ "
		text = selectedStyle.Render(i.title)
	}
	fmt.Fprintf(w, "%s%s", cursor, text)
}

// item represents a provider choice
type item struct {
	providerName string
	title        string
	desc         string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// model represents the interactive menu
type model struct {
	list        list.Model
	choices     []item
	quitting    bool
	selected    string
	selectedIdx int
}

// initialModel creates the initial model
func initialModel(cfg *config.Config) model {
	// Always include anthropic as first option
	providerNames := []string{anthropicProvider}

	// Collect all unique external providers
	providerSet := make(map[string]bool)
	for name := range cfg.Providers {
		if name != anthropicProvider {
			providerSet[name] = true
		}
	}

	// Always include known providers
	providerSet[claudeCodeProvider] = true
	providerSet[glmProvider] = true

	// Convert to slice and sort
	var externalProviders []string
	for name := range providerSet {
		externalProviders = append(externalProviders, name)
	}
	sort.Strings(externalProviders)
	providerNames = append(providerNames, externalProviders...)

	// Convert to items
	var items []item
	for _, name := range providerNames {
		provider := cfg.Providers[name]
		displayName, statusText := getProviderDisplayInfo(name, provider)

		title := displayName
		if statusText == "OAuth" {
			title += " (OAuth)"
		} else {
			title += " (API)"
		}

		if cfg.Provider == name {
			title += currentMarker
		}

		items = append(items, item{
			providerName: name,
			title:        title,
			desc:         "",
		})
	}

	// Create the list
	const defaultWidth = 40
	const listHeight = 8

	// Convert []item to []list.Item
	listItems := make([]list.Item, len(items))
	for i, it := range items {
		listItems[i] = it
	}

	l := list.New(listItems, compactDelegate{}, defaultWidth, listHeight)
	l.Title = titleStyle.Render("Select Provider")
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	// Find current provider and set as selected
	for i, it := range items {
		if it.providerName == cfg.Provider {
			l.Select(i)
			break
		}
	}

	return model{
		list:    l,
		choices: items,
	}
}

// Init implements tea.Model
func (m model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			m.selected = ""
			return m, tea.Quit

		case "enter":
			selectedItem := m.list.SelectedItem()
			if i, ok := selectedItem.(item); ok {
				m.selected = i.providerName
				m.selectedIdx = m.list.Index()
				m.quitting = true
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m model) View() string {
	if m.quitting {
		if m.selected == "" {
			return quitTextStyle.Render("No provider selected")
		}
		return ""
	}
	return docStyle.Render(m.list.View())
}

var docStyle = lipgloss.NewStyle().
	Margin(0, 1)

// RunInteractiveSelection runs the interactive provider selection
func RunInteractiveSelection(cfg *config.Config) (string, error) {
	// Check if we're in a terminal
	if !isTerminal() {
		return "", fmt.Errorf("interactive mode requires a terminal")
	}

	p := tea.NewProgram(initialModel(cfg))

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run interactive selection: %w", err)
	}

	if model, ok := m.(model); ok {
		return model.selected, nil
	}

	return "", fmt.Errorf("no provider selected")
}

// isTerminal checks if we're running in a terminal
func isTerminal() bool {
	// Simple check - in a real implementation, you might want to use
	// something more sophisticated
	return true
}
