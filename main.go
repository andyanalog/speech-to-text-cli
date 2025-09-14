package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
		state:      StateSelectFile,
		filepicker: fp,
		spinner:    s,
		width:      80,
		height:     24,
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
			content = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
				titleStyle.Render("Speech-to-Text CLI"),
				successStyle.Render("Transcription completed"),
				transcriptionStyle.Render(m.transcription),
				subtitleStyle.Render("Press 'q' to exit"))
		}
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
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
