package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inancgumus/screen"
	"github.com/ledongthuc/pdf"
)

type model struct {
	Files       []os.FileInfo
	CurrentIdx  int
	Content     string
	FontSize    int
	Loading     bool
	CurrentPage int
	TotalPages  int
	ReadingMode bool
}

func (m model) Init() tea.Cmd {
	return nil
}

type MsgType int

const (
	GoToPage MsgType = iota + 1
	Quit
	Select
	IncreaseFontSize
	DecreaseFontSize
	LoadingDone
)

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error starting program:", err)
		os.Exit(1)
	}
}

func initialModel() model {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory", err)
		os.Exit(1)
	}

	return model{
		Files:       files,
		CurrentIdx:  0,
		Content:     "Select a file to view its content",
		FontSize:    12,
		ReadingMode: false,
		CurrentPage: 1,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case LoadContentMsg:
		return m.handleLoadContentMsg(msg)
	case MsgType:
		return m.handleMsgType(msg)
	}
	return m, nil
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m.handleQuitKey()
	case "enter":
		return m.handleEnterKey()
	case "up", "w":
		return m.handleUpKey()
	case "down", "s":
		return m.handleDownKey()
	case "right", "d":
		return m.handleRightKey()
	case "left", "a":
		return m.handleLeftKey()
	case "+":
		return m.handlePlusKey()
	case "-":
		return m.handleMinusKey()
	}
	return m, nil
}

func (m model) handleLoadContentMsg(msg LoadContentMsg) (tea.Model, tea.Cmd) {
	content, totalPages, err := readPDFFile(msg.FileName, msg.Page)
	if err != nil {
		content = fmt.Sprintf("Error reading file: %v", err)
		totalPages = 0
	}
	m.Content = content
	m.TotalPages = totalPages
	m.ReadingMode = true
	return m, tea.Tick(time.Second/2, func(t time.Time) tea.Msg {
		return LoadingDone
	})
}

func (m model) handleMsgType(msg MsgType) (tea.Model, tea.Cmd) {
	switch msg {
	case GoToPage:
		return m.handleGoToPage()
	case LoadingDone:
		return m.handleLoadingDone()
	}
	return m, nil
}

func (m model) handleQuitKey() (tea.Model, tea.Cmd) {
	if m.ReadingMode {
		m.ReadingMode = false
		screen.Clear()
		return m, nil
	}
	return m, tea.Quit
}

func (m model) handleEnterKey() (tea.Model, tea.Cmd) {
	m.Loading = true
	screen.Clear()
	return m, func() tea.Msg {
		return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
	}
}

func (m model) handleUpKey() (tea.Model, tea.Cmd) {
	m.CurrentIdx--
	if m.CurrentIdx < 0 {
		m.CurrentIdx = len(m.Files) - 1
	}
	return m, nil
}

func (m model) handleDownKey() (tea.Model, tea.Cmd) {
	m.CurrentIdx++
	if m.CurrentIdx >= len(m.Files) {
		m.CurrentIdx = 0
	}
	return m, nil
}

func (m model) handlePlusKey() (tea.Model, tea.Cmd) {
	m.FontSize++
	// screen.Clear()
	setFontSize(m.FontSize * 2)
	return m, nil
}

func (m model) handleMinusKey() (tea.Model, tea.Cmd) {
	if m.FontSize > 1 {
		m.FontSize--
		setFontSize(m.FontSize * 2)
	}
	return m, nil
}

func setFontSize(size int) {
	cmd := exec.Command("resize", "-s", strconv.Itoa(size), strconv.Itoa(size*2))
	cmd.Run()
	time.Sleep(time.Millisecond * 100) // Give the terminal time to react
}

func (m model) handleRightKey() (tea.Model, tea.Cmd) {
	if m.ReadingMode && m.CurrentPage < m.TotalPages {
		m.CurrentPage++
		return m, func() tea.Msg {
			return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
		}
	}
	return m, nil
}

func (m model) handleLeftKey() (tea.Model, tea.Cmd) {
	if m.ReadingMode && m.CurrentPage > 1 {
		m.CurrentPage--
		return m, func() tea.Msg {
			return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
		}
	}
	return m, nil
}

func (m model) handleGoToPage() (tea.Model, tea.Cmd) {
	page, err := strconv.Atoi(m.Content)
	if err != nil || page < 1 || page > m.TotalPages {
		return m, nil
	}

	m.CurrentPage = page
	m.Loading = true

	return m, func() tea.Msg {
		return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
	}
}

func (m model) handleLoadingDone() (tea.Model, tea.Cmd) {
	m.Loading = false
	return m, nil
}

func (m model) View() string {
	if m.Loading {
		return "\nLoading..."
	}

	if m.ReadingMode {
		termWidth, _ := screen.Size()

		contentLines := divideTextIntoLines(m.Content, termWidth)

		contentText := strings.Join(contentLines, "\n")

		pageInfo := fmt.Sprintf("\nContent (Font Size %d, Page %d/%d):\n", m.FontSize, m.CurrentPage, m.TotalPages)

		return contentText + pageInfo

	}

	s := "Files:\n"
	for i, file := range m.Files {
		if i == m.CurrentIdx {
			s += fmt.Sprintf("> %s\n", file.Name())
		} else {
			s += fmt.Sprintf(" %s\n", file.Name())
		}
	}

	return s
}

func divideTextIntoLines(text string, maxWidth int) []string {
	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word) < maxWidth {
			currentLine += word + " "
		} else {
			lines = append(lines, currentLine)
			currentLine = word + " "
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

type LoadContentMsg struct {
	FileName string
	Page     int
}

func readPDFFile(fileName string, pageNum int) (string, int, error) {
	f, r, err := pdf.Open(fileName)
	if err != nil {
		return "", 0, err
	}
	defer func() {
		_ = f.Close()
	}()

	totalPages := r.NumPage()

	page := r.Page(pageNum)

	pageContent, err := page.GetPlainText(nil)
	if err != nil {
		return "", 0, err
	}

	return pageContent, totalPages, nil
}
