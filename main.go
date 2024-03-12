package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inancgumus/screen"
	"github.com/ledongthuc/pdf"
	"github.com/muesli/reflow/indent"
)

type model struct {
	Files         []os.DirEntry
	CurrentIdx    int
	Content       string
	Loading       bool
	CurrentPage   int
	TotalPages    int
	ReadingMode   bool
	ContentOffset int // Offset de rolagem do conteúdo
	viewport      viewport.Model
	list          list.Model
	choice        string
}

var listHeight = screenHeight() -2
const useHighPerformanceRenderer = false

var (
	titleStyle         = lipgloss.NewStyle().MarginLeft(2)
	itemStyle          = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle  = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("FFFAE0")).Background(lipgloss.Color("002236"))
	paginationStyle    = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle          = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle      = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	titleStyleViewport = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s",  i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func (m model) Init() tea.Cmd {
	return nil
}

type MsgType int

const (
	GoToPage MsgType = iota + 1
	Quit
	Select
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
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory", err)
		os.Exit(1)
	}

	var filteredFiles []os.DirEntry
	for _, file := range files {
		if file.IsDir() || strings.HasSuffix(file.Name(), ".pdf") {
			filteredFiles = append(filteredFiles, file)
		}
	}

	items := []list.Item{}
	for _, file := range filteredFiles {
		items = append(items, item(file.Name()))
	}

	const defaultWidth = 20
	
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Lumus"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return model{
		Files:       filteredFiles,
		CurrentIdx:  0,
		Content:     "Select a file to view its content",
		ReadingMode: false,
		CurrentPage: 1,
		list: l,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var teaCmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case LoadContentMsg:
		return m.handleLoadContentMsg(msg)
	case MsgType:
		return m.handleMsgType(msg)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m.handleQuitKey()
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m.handleEnterKey()
		case "up", "w":
			return m.handleUpKey(msg)
		case "down", "s":
			return m.handleDownKey(msg)
		case "right", "d":
			return m.handleRightKey()
		case "left", "a":
			return m.handleLeftKey()
		case "backspace":
			return m.handleBackspaceKey(msg)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
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
		m.CurrentPage = 1
		screen.Clear()
		return m, nil
	}
	return m, tea.Quit
}

func (m model) handleEnterKey() (tea.Model, tea.Cmd) {
	// i, ok := m.list.SelectedItem().(item)
	// if ok {
		// m.CurrentIdx = int(i)
	// }
	selectedFile := m.Files[m.CurrentIdx]
	if selectedFile.IsDir() {
		os.Chdir(selectedFile.Name())
		files, err := os.ReadDir(".")
		if err != nil {
			fmt.Println("Error reading directory", err)
			os.Exit(1)
		}
		var filteredFiles []os.DirEntry
		for _, file := range files {
			if file.IsDir() || strings.HasSuffix(file.Name(), ".pdf") {
				filteredFiles = append(filteredFiles, file)
			}
		}
		items := []list.Item{}
		for _, file := range filteredFiles {
			items = append(items, item(file.Name()))
		}
	
		const defaultWidth = 20
		
		l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
		l.Title = "Lumus"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.Styles.Title = titleStyle
		l.Styles.PaginationStyle = paginationStyle
		l.Styles.HelpStyle = helpStyle
	
		return model{
			Files:       filteredFiles,
			CurrentIdx:  0,
			Content:     "Select a file to view its content",
			ReadingMode: false,
			CurrentPage: 1,
			list: l,
		},nil
	}

	m.Loading = true
	screen.Clear()
	return m, func() tea.Msg {
		return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
	}
}

func (m model) handleUpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ContentOffset > 0 && m.ReadingMode {
		m.ContentOffset--
	}
	m.CurrentIdx--
	if m.CurrentIdx < 0 {
		m.CurrentIdx = len(m.Files) - 1
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) handleDownKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ContentOffset < len(strings.Split(m.Content, "\n"))-screenHeight() && m.ReadingMode {
		m.ContentOffset++
	}
	m.CurrentIdx++
	if m.CurrentIdx >= len(m.Files) {
		m.CurrentIdx = 0
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
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

func (m model) handleBackspaceKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.ReadingMode {
		os.Chdir("../")
		files, err := os.ReadDir(".")
		if err != nil {
			fmt.Println("Error reading directory", err)
			os.Exit(1)
		}
		var filteredFiles []os.DirEntry
		for _, file := range files {
			if file.IsDir() || strings.HasSuffix(file.Name(), ".pdf") {
				filteredFiles = append(filteredFiles, file)
			}
		}
		items := []list.Item{}
		for _, file := range filteredFiles {
			items = append(items, item(file.Name()))
		}
	
		const defaultWidth = 20
		
		l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
		l.Title = "Lumus"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.Styles.Title = titleStyle
		l.Styles.PaginationStyle = paginationStyle
		l.Styles.HelpStyle = helpStyle
	
		return model{
			Files:       filteredFiles,
			CurrentIdx:  0,
			Content:     "Select a file to view its content",
			ReadingMode: false,
			CurrentPage: 1,
			list: l,
		}, nil
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
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
		contentWidth := screenWidth() - 4
		contentLines := strings.Split(m.Content, "\n")

		// Renderizar apenas o conteúdo visível na tela
		start := m.ContentOffset
		end := start + screenHeight() - 2 // -2 para o cabeçalho e rodapé
		if end > len(contentLines) {
			end = len(contentLines)
		}
		visibleContent := strings.Join(contentLines[start:end], "\n")

		pageInfo := fmt.Sprintf("\nContent (Page %d/%d):\n", m.CurrentPage, m.TotalPages)
		content := lipgloss.NewStyle().
			Width(contentWidth).
			Render(visibleContent)

		pageInfoStyle := lipgloss.NewStyle().Width(contentWidth).Render(pageInfo)

		return indent.String(content, 2) + pageInfoStyle

	}

	// header := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800")).Bold(true).Render("Lumus\n")

	// s := header + "\nFiles:\n"
	// for i, file := range m.Files {
	// if i == m.CurrentIdx {
	// s += fmt.Sprintf("→ %s\n", file.Name())
	// } else {
	// s += fmt.Sprintf(" %s\n", file.Name())
	// }
	// }

	// footer := "\nInstructions:\n" +
	// "  - Use arrow keys to navigate\n" +
	// "  - Press Enter to view a file\n" +
	// "  - Press q to quit\n" +
	// "  - Press backspace to go back one directory\n"

	// items := []list.Item{}
	// for _, file := range m.Files {
		// items = append(items, item(file.Name()))
	// }
// 
	// const defaultWidth = 20
// 
	// // return s + footer
	// l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	// l.Title = "Lumus"
	// l.SetShowStatusBar(false)
	// l.SetFilteringEnabled(false)
	// l.Styles.Title = titleStyle
	// l.Styles.PaginationStyle = paginationStyle
	// l.Styles.HelpStyle = helpStyle
	// // model := model{list: l}
	// m.list = l
	return "\n" + m.list.View()
}

func screenWidth() int {
	width, _ := screen.Size()
	return width
}

func screenHeight() int {
	_, height := screen.Size()
	return height
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) headerView(name string) string {
	title := titleStyleViewport.Render(name)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
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
