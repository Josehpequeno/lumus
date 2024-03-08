package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
	"github.com/unidoc/unipdf/v3/extractor"
	unipdf "github.com/unidoc/unipdf/v3/model"
)

type model struct {
	Files         []os.FileInfo
	CurrentIdx    int
	Content       string
	FontSize      int
	FontStyle     tcell.Style
	DefaultStyle  tcell.Style
	SelectedStyle tcell.Style
	Loading       bool
	CurrentPage   int
	TotalPages    int
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

	style := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack)

	return model{
		Files:         files,
		CurrentIdx:    0,
		Content:       "Select a file to view its content",
		FontSize:      1,
		FontStyle:     style,
		DefaultStyle:  style,
		SelectedStyle: style.Foreground(tcell.ColorBlue),
	}
}

// funtions of model
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
		return m, tea.Quit
	case "enter":
		return m.handleEnterKey()
	case "up", "k":
		return m.handleUpKey()
	case "down", "j":
		return m.handleDownKey()
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

func (m model) handleEnterKey() (tea.Model, tea.Cmd) {
	// Lógica para lidar com a tecla "enter"
	m.Loading = true
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
	return m, nil
}

func (m model) handleMinusKey() (tea.Model, tea.Cmd) {
	if m.FontSize > 1 {
		m.FontSize--
	}
	return m, nil
}

func (m model) handleGoToPage() (tea.Model, tea.Cmd) {
	page, err := strconv.Atoi(m.Content)
	if err != nil || page < 1 || page > m.TotalPages {
		// Se a página for inválida, retorne o modelo atual sem comandos
		return m, nil
	}

	m.CurrentPage = page
	m.Loading = true

	// Retorna o modelo com um comando para carregar o conteúdo da página especificada
	return m, func() tea.Msg {
		return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
	}
}

func (m model) handleLoadingDone() (tea.Model, tea.Cmd) {
	// Define Loading como false para indicar que o carregamento foi concluído
	m.Loading = false

	// Retorna o modelo atual sem comandos adicionais
	return m, nil
}

func (m model) View() string {
	s := "Files:\n"
	for i, file := range m.Files {
		if i == m.CurrentIdx {
			s += fmt.Sprintf("> %s\n", file.Name())
		} else {
			s += fmt.Sprintf(" %s\n", file.Name())
		}
	}

	if m.Loading {
		s += "\nLoading..."
	} else {
		s += fmt.Sprintf("\nContent (Font Size %d, Page %d/%d):\n", m.FontSize, m.CurrentPage, m.TotalPages)
		s += m.Content
	}

	return s
}

type LoadContentMsg struct {
	FileName string
	Page     int
}

func readPDFFile(fileName string, pageNum int) (string, int, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	pdfReader, err := unipdf.NewPdfReader(f)
	if err != nil {
		return "", 0, err
	}

	totalPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", 0, err
	}

	page, err := pdfReader.GetPage(pageNum)
	if err != nil {
		return "", 0, err
	}

	ex, err := extractor.New(page)
	if err != nil {
		return "", 0, err
	}

	pageContent, err := ex.ExtractText()
	if err != nil {
		return "", 0, err
	}

	return pageContent, totalPages, nil
}
