package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	StateSelectFile = iota
	StateProcessing
	StateComplete
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B"))

	transcriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8F8F2")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#6272A4")).
				Padding(1, 2).
				Width(80)
)

type model struct {
	state         int
	filepicker    filepicker.Model
	spinner       spinner.Model
	selectedFile  string
	transcription string
	error         string
	width         int
	height        int
	scrollOffset  int
	maxScroll     int
}

func initialModel() model {
	// Initialize file picker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".mp3", ".wav", ".m4a", ".flac"}
	fp.CurrentDirectory, _ = os.Getwd()

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot

	return model{
		state:        StateSelectFile,
		filepicker:   fp,
		spinner:      s,
		width:        80,
		height:       24,
		scrollOffset: 0,
		maxScroll:    0,
	}
}

func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.state == StateComplete && m.transcription != "" {
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			}
		case "down", "j":
			if m.state == StateComplete && m.transcription != "" {
				if m.scrollOffset < m.maxScroll {
					m.scrollOffset++
				}
			}
		case "home":
			if m.state == StateComplete && m.transcription != "" {
				m.scrollOffset = 0
			}
		case "end":
			if m.state == StateComplete && m.transcription != "" {
				m.scrollOffset = m.maxScroll
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.state == StateSelectFile {
			m.filepicker.Height = msg.Height - 4
		}

	case processCompleteMsg:
		m.state = StateComplete
		m.transcription = string(msg)
		m.scrollOffset = 0

		// Calculate max scroll based on transcription length and available space
		transcriptionHeight := m.height - 10 // Leave space for title and instructions
		if transcriptionHeight < 5 {
			transcriptionHeight = 5
		}

		// Wrap text and count lines
		wrappedText := m.wrapText(m.transcription, m.width-8) // Account for padding and border
		totalLines := len(strings.Split(wrappedText, "\n"))

		if totalLines > transcriptionHeight {
			m.maxScroll = totalLines - transcriptionHeight
		} else {
			m.maxScroll = 0
		}

		return m, nil

	case processErrorMsg:
		m.error = string(msg)
		m.state = StateComplete
		return m, nil

	case spinner.TickMsg:
		if m.state == StateProcessing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	switch m.state {
	case StateSelectFile:
		m.filepicker, cmd = m.filepicker.Update(msg)

		// Check if user selected a file
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.selectedFile = path
			m.state = StateProcessing
			return m, tea.Batch(m.spinner.Tick, m.startProcessing())
		}

	case StateProcessing:
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var content string

	switch m.state {
	case StateSelectFile:
		content = fmt.Sprintf("%s\n\n%s\n\n%s",
			titleStyle.Render("Speech-to-Text CLI"),
			subtitleStyle.Render("Select a video or audio file to transcribe:"),
			m.filepicker.View())

	case StateProcessing:
		content = fmt.Sprintf("%s\n\n%s %s\n%s\n\n%s",
			titleStyle.Render("Speech-to-Text CLI"),
			m.spinner.View(),
			"Processing audio...",
			subtitleStyle.Render(fmt.Sprintf("File: %s", filepath.Base(m.selectedFile))),
			subtitleStyle.Render("Extracting audio and transcribing... This may take a few minutes..."))

	case StateComplete:
		if m.error != "" {
			content = fmt.Sprintf("%s\n\n%s\n\n%s",
				titleStyle.Render("Speech-to-Text CLI"),
				errorStyle.Render("Error occurred:"),
				errorStyle.Render(m.error))
		} else {
			scrollInstructions := ""
			if m.maxScroll > 0 {
				scrollInstructions = subtitleStyle.Render(fmt.Sprintf("Use ↑/↓ or j/k to scroll • Line %d-%d of %d • Press 'q' to exit",
					m.scrollOffset+1,
					min(m.scrollOffset+(m.height-10), len(strings.Split(m.wrapText(m.transcription, m.width-8), "\n"))),
					len(strings.Split(m.wrapText(m.transcription, m.width-8), "\n"))))
			} else {
				scrollInstructions = subtitleStyle.Render("Press 'q' to exit")
			}

			content = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
				titleStyle.Render("Speech-to-Text CLI"),
				successStyle.Render("Transcription completed"),
				m.renderScrollableTranscription(),
				scrollInstructions)
		}
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// wrapText wraps text to fit within the specified width
func (m model) wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// renderScrollableTranscription renders the transcription with scrolling
func (m model) renderScrollableTranscription() string {
	transcriptionHeight := m.height - 10 // Leave space for title and instructions
	if transcriptionHeight < 5 {
		transcriptionHeight = 5
	}

	// Wrap text to fit the display width
	wrappedText := m.wrapText(m.transcription, m.width-8) // Account for padding and border
	lines := strings.Split(wrappedText, "\n")

	// Extract visible lines based on scroll offset
	startLine := m.scrollOffset
	endLine := min(startLine+transcriptionHeight, len(lines))

	if startLine >= len(lines) {
		startLine = max(0, len(lines)-transcriptionHeight)
		endLine = len(lines)
	}

	visibleText := strings.Join(lines[startLine:endLine], "\n")

	// Pad with empty lines if needed to maintain consistent height
	visibleLines := strings.Split(visibleText, "\n")
	for len(visibleLines) < transcriptionHeight {
		visibleLines = append(visibleLines, "")
	}
	visibleText = strings.Join(visibleLines, "\n")

	return transcriptionStyle.
		Width(m.width - 4).
		Height(transcriptionHeight + 2). // +2 for padding
		Render(visibleText)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type processCompleteMsg string
type processErrorMsg string

func (m model) startProcessing() tea.Cmd {
	return func() tea.Msg {
		transcription, err := processAudioSTT(m.selectedFile)
		if err != nil {
			return processErrorMsg(err.Error())
		}
		return processCompleteMsg(transcription)
	}
}

func main() {
	fmt.Println("Speech-to-Text CLI")
	fmt.Println("A tool to extract audio and transcribe speech from video/audio files")
	fmt.Println("")

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
