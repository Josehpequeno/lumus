package main

import (
	// "bytes"
	"fmt"
	"image"
	// "image/png"
	"io"
	"unicode/utf8"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inancgumus/screen"
	"github.com/ledongthuc/pdf"
	// "github.com/nfnt/resize"
	"github.com/otiai10/gosseract/v2"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type model struct {
	Files        []os.DirEntry
	CurrentIdx   int
	Content      string
	Loading      bool
	CurrentPage  int
	TotalPages   int
	ReadingMode  bool
	GoToPageMode bool
	Viewport     viewport.Model
	List         list.Model
	FileName     string
	TextInput    textinput.Model
	Error        bool
	Ready        bool
}

var listHeight = screenHeight() - 2
var client *gosseract.Client
var pwd string

const useHighPerformanceRenderer = false

var (
	titleStyle         = lipgloss.NewStyle().MarginLeft(2)
	itemStyle          = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle  = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("FFFAE0")).Background(lipgloss.Color("002236"))
	paginationStyle    = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle          = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
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

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

type MsgType int

const (
	GoToPage MsgType = iota + 1
	Quit
	Select
	LoadingDone
)

var sourcePath string = "extract_text.py"
var targetPath string = "/usr/local/bin/extract_text_lumus"

func createSymlink(sourcePath, targetPath string) error {
	if _, err := os.Lstat(targetPath); err == nil {
		return nil
	}

	err := os.Symlink(sourcePath, targetPath)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	client = gosseract.NewClient()

	// Configurar os idiomas para inglês, espanhol e português brasileiro
	client.Languages = []string{"eng", "spa", "por+por"}
	defer client.Close()

	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error get current user:", err)
		os.Exit(1)
	}

	binPath := fmt.Sprintf("/home/%s/.local/bin/", currentUser.Username)
	if err := os.MkdirAll(binPath, os.ModePerm); err != nil {
		fmt.Println("Error creating bin folder:", err)
		os.Exit(1)
	}

	targetPath = binPath + "extract_text_lumus"

	pwd, err = os.Getwd()
	if err != nil {
		fmt.Println("Error reading directory path", err)
		os.Exit(1)
	}

	sourcePath = pwd + "/" + sourcePath

	err = createSymlink(sourcePath, targetPath)
	if err != nil {
		fmt.Println("Error creating link:", err)
		os.Exit(1)
	}
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

	ti := textinput.New()
	ti.Placeholder = "10"
	ti.Focus()
	ti.CharLimit = 10
	ti.Width = 20

	return model{
		Files:        filteredFiles,
		CurrentIdx:   0,
		Content:      "Select a file to view its content",
		ReadingMode:  false,
		CurrentPage:  1,
		GoToPageMode: false,
		List:         l,
		TextInput:    ti,
		Error:        false,
		Ready:        false,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var teaCmds []tea.Cmd
	var teaCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.List.SetWidth(msg.Width)
		headerHeight := lipgloss.Height(m.headerView(m.FileName))
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight 

		if !m.Ready {
			m.Viewport = viewport.New(screenWidth(), screenHeight()-verticalMarginHeight)
			m.Viewport.YPosition = headerHeight
			m.Viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.Viewport.SetContent(m.Content)
			m.Ready = true
			// Render the viewport one line below the header.
			m.Viewport.YPosition = headerHeight + 1
		} else {
			m.Viewport.Width = screenWidth()
			m.Viewport.Height = screenHeight() - verticalMarginHeight
		}
		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			teaCmds = append(teaCmds, viewport.Sync(m.Viewport))
		}
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case LoadContentMsg:
		return m.handleLoadContentMsg(msg)
	case MsgType:
		return m.handleMsgType(msg)
	}

	if m.GoToPageMode {
		m.TextInput, teaCmd = m.TextInput.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
		return m, tea.Batch(teaCmds...)
	}
	// Handle keyboard and mouse events in the viewport
	m.Viewport, teaCmd = m.Viewport.Update(msg)
	teaCmds = append(teaCmds, teaCmd)
	m.List, teaCmd = m.List.Update(msg)
	teaCmds = append(teaCmds, teaCmd)
	return m, tea.Batch(teaCmds...)
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var teaCmds []tea.Cmd
	var teaCmd tea.Cmd

	switch keypress := msg.String(); keypress {
	case "ctrl+c", "ctrl+q", "q", "esc":
		return m.handleQuitKey()
	case "enter":
		i, ok := m.List.SelectedItem().(item)
		if ok {
			m.FileName = string(i)
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
	case "p":
		return m.handleGoToPage(msg)
	}
	if m.GoToPageMode {
		m.TextInput, teaCmd = m.TextInput.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
		return m, tea.Batch(teaCmds...)
	}
	m.List, teaCmd = m.List.Update(msg)
	teaCmds = append(teaCmds, teaCmd)
	return m, teaCmd
}

func (m model) handleLoadContentMsg(msg LoadContentMsg) (tea.Model, tea.Cmd) {
	content, totalPages, err := readPDFFile(msg.FileName, msg.Page)
	if err != nil {
		content = fmt.Sprintf("Error reading file %s : %v", pwd+"/"+msg.FileName, err)
		totalPages = 0
	}
	m.Content = content
	m.Viewport.SetContent(content)
	m.TotalPages = totalPages
	m.ReadingMode = true
	m.Loading = false
	return m, tea.Tick(time.Second/5, func(t time.Time) tea.Msg {
		return LoadingDone
	})
}

func (m model) handleMsgType(msg MsgType) (tea.Model, tea.Cmd) {
	switch msg {
	case LoadingDone:
		return m.handleLoadingDone()
	}
	return m, nil
}

func (m model) handleQuitKey() (tea.Model, tea.Cmd) {
	if m.ReadingMode {
		m.ReadingMode = false
		m.GoToPageMode = false
		m.CurrentPage = 1
		return m, nil
	}
	if m.GoToPageMode {
		m.GoToPageMode = false
		m.ReadingMode = true
		return m, nil
	}
	return m, tea.Quit
}

func (m model) handleEnterKey() (tea.Model, tea.Cmd) {
	if m.GoToPageMode {
		pageStr := m.TextInput.Value()
		page, err := strconv.Atoi(pageStr)
		if err != nil || (page < 1 || page > m.TotalPages) {
			m.Error = true
			return m, nil
		}
		m.CurrentPage = page
		m.Loading = true
		m.TextInput.SetValue("")
		m.GoToPageMode = false
		return m, func() tea.Msg {
			return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
		}
	}
	selectedFile := m.Files[m.CurrentIdx]
	if selectedFile.IsDir() {
		err := os.Chdir(pwd + "/" + selectedFile.Name())
		if err != nil {
			fmt.Println("Error entering directory", err)
			os.Exit(1)
		}
		files, err := os.ReadDir(".")
		if err != nil {
			fmt.Println("Error reading directory", err)
			os.Exit(1)
		}
		pwd, err = os.Getwd()
		if err != nil {
			fmt.Println("Error reading directory path", err)
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

		m.List = l
		m.Files = filteredFiles
		m.CurrentIdx = 0
		m.Content = "Select a file to view its content"

		return m, nil
	}
	m.Loading = true
	return m, func() tea.Msg {
		return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
	}

}

func (m model) handleUpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd
	if m.ReadingMode {
		m.Viewport, teaCmd = m.Viewport.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
		return m, tea.Batch(teaCmds...)
	}
	m.CurrentIdx--
	if m.CurrentIdx < 0 {
		m.CurrentIdx = 0
	}
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m model) handleDownKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd
	if m.ReadingMode {
		m.Viewport, teaCmd = m.Viewport.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
		return m, tea.Batch(teaCmds...)
	}
	m.CurrentIdx++
	if m.CurrentIdx >= len(m.Files) {
		m.CurrentIdx = len(m.Files) - 1
	}
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m model) handleRightKey() (tea.Model, tea.Cmd) {
	if m.ReadingMode && m.CurrentPage < m.TotalPages {
		m.CurrentPage++
		// headerHeight := lipgloss.Height(m.headerView(m.FileName))
		m.Viewport.YPosition = 0
		return m, func() tea.Msg {
			return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
		}
	}
	return m, nil
}

func (m model) handleLeftKey() (tea.Model, tea.Cmd) {
	if m.ReadingMode && m.CurrentPage > 1 {
		m.CurrentPage--
		// headerHeight := lipgloss.Height(m.headerView(m.FileName))
		m.Viewport.YPosition = 0
		return m, func() tea.Msg {
			return LoadContentMsg{FileName: m.Files[m.CurrentIdx].Name(), Page: m.CurrentPage}
		}
	}
	return m, nil
}

func (m model) handleBackspaceKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.ReadingMode && !m.GoToPageMode {
		dir, _ := filepath.Split(pwd)
		parentDir := filepath.Dir(dir)
		err := os.Chdir(parentDir)
		if err != nil {
			fmt.Println("Error entering directory", err)
			os.Exit(1)
		}
		files, err := os.ReadDir(".")
		if err != nil {
			fmt.Println("Error reading directory", err)
			os.Exit(1)
		}
		pwd, err = os.Getwd()
		if err != nil {
			fmt.Println("Error reading directory path", err)
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

		m.List = l
		m.Files = filteredFiles
		m.CurrentIdx = 0
		m.Content = "Select a file to view its content"
		return m, nil
	}
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	if m.GoToPageMode {
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd
	}
	return m, cmd
}

func (m model) handleGoToPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.ReadingMode && !m.GoToPageMode {
		return m, nil
	}
	m.GoToPageMode = true
	m.ReadingMode = false

	return m, nil
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
		return fmt.Sprintf("%s\n%s\n%s", m.headerView(m.FileName), m.Viewport.View(), m.footerView())
	}

	if m.GoToPageMode {
		if !m.Error {
			return fmt.Sprintf("Go to Page: \n%s\n%s", m.TextInput.View(), "(q to quit)")
		}
		return fmt.Sprintf("Go to Page: \n%s\n%s\n%s", m.TextInput.View(), "(q to quit)", "Non-existent page")
	}

	return "\n" + m.List.View()
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
	title := titleStyleViewport.Render(pwd + "/" + name)
	line := strings.Repeat("─", max(0, m.Viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%% Page %d/%d ", m.Viewport.ScrollPercent()*100, m.CurrentPage, m.TotalPages))
	str := "Press 'p' to Go To Page. Arrow Keys to change of page. "
	line := str + strings.Repeat(" ", max(0, m.Viewport.Width-(lipgloss.Width(info)+len(str))))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

type LoadContentMsg struct {
	FileName string
	Page     int
}

func readPDFFile(fileName string, pageNum int) (string, int, error) {
	f, r, err := pdf.Open(pwd + "/" + fileName)
	if err != nil {
		return err.Error(), 0, err
	}

	totalPages := r.NumPage()
	_ = f.Close()

	// pythonExecutable := "./pdf_extractor"

	//call python function
	// cmd := exec.Command(pythonExecutable, pwd+fileName, fmt.Sprintf("%d", pageNum))
	cmd := exec.Command("python3", targetPath, pwd+"/"+fileName, fmt.Sprintf("%d", pageNum))

	output, err := cmd.Output()
	if err != nil {
		return err.Error(), 0, err
	}
	pageContent := string(output)

	parts := strings.Split(pageContent, "\n")

	widthPage := parts[0]
	heightPage := parts[1]

	pageContent = strings.Join(parts[2:], "\n")
	pageContent =  textWithWidth(pageContent)
	// eofLine := strings.Repeat("#", max(0, screenWidth()))
	// pageContent += "\n" + eofLine + eofLine
	outputDir := "lumus_images_extract"

	//extrair imagens
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err.Error(), 0, err
	}
	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	//configurar as opções de extração de imagens
	pageSelection := []string{strconv.Itoa(pageNum)}

	if err := api.ExtractImagesFile(fileName, outputDir, pageSelection, nil); err != nil {
		return err.Error(), 0, err
	}

	imagePath, err := getImageFile(outputDir)
	if err != nil || imagePath == "" {
		return pageContent, totalPages, nil
	}

	imgFile, err := os.Open(imagePath)
	if err != nil {
		// return err.Error(), 0, err
		return pageContent, totalPages, nil
	}
	defer imgFile.Close()

	// Decodificar a imagem
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return pageContent, totalPages, nil
	}

	bounds := img.Bounds()
	widthImg := bounds.Max.X
	heightImg := bounds.Max.Y
	//
	// // Redimensionar a imagem para melhorar o desempenho do OCR
	// resizedImg := resize.Resize(0, 1100, img, resize.Lanczos3)
	// // A imagem resultante está vazia ou tem dimensões inesperadas
	// pageContent += strconv.Itoa(widthImg) + " x " + strconv.Itoa(heightImg) + "\n"
	// var buf bytes.Buffer
	//
	// // Codifique a imagem para o buffer de bytes
	// if err = png.Encode(&buf, resizedImg); err != nil {
	// return pageContent, totalPages, nil
	// }
	//
	// // Passe os bytes para a função SetImageFromBytes
	// err = client.SetImageFromBytes(buf.Bytes())
	// if err != nil {
	// return pageContent, totalPages, nil
	// }

	err = client.SetImage(imagePath)
	if err != nil {
		// return err.Error(), 0, err
		return pageContent, totalPages, nil
	}

	text, err := client.Text()
	if err != nil {
		return pageContent, totalPages, nil
	}

	if widthPage == "" || heightPage == "" {
		return text, totalPages, nil
	}

	heightPageFloat, err := strconv.ParseFloat(heightPage, 64)
	if err != nil {
		return pageContent, totalPages, nil
	}
	widthPageFloat, err := strconv.ParseFloat(widthPage, 64)
	if err != nil {
		return pageContent, totalPages, nil
	}

	perHeight := float64(heightImg) / heightPageFloat
	perWidth := float64(widthImg) / widthPageFloat

	if perHeight > 0.8 && perWidth > 0.8 {
		return text, totalPages, nil
	}

	similarity := levenshteinDistance(pageContent, text)

	// Definir um limite para determinar se os textos são suficientemente semelhantes
	threshold := 125

	if similarity <= threshold {
		return text, totalPages, nil
	}
	return text + "\n" + pageContent, totalPages, nil
}

func getImageFile(dir string) (string, error) {
	var imageFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isImageFile(info.Name()) {
			imageFiles = append(imageFiles, path)
		}
		return nil
	})

	if err != nil || len(imageFiles) == 0 {
		return "", nil
	}
	return imageFiles[0], nil
}

func isImageFile(name string) bool {
	ext := filepath.Ext(name)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return true
	default:
		return false
	}
}

func levenshteinDistance(s1, s2 string) int {
	m, n := len(s1), len(s2)
	if m == 0 {
		return n
	}

	if n == 0 {
		return m
	}

	matrix := make([][]int, m+1)
	for i := range matrix {
		matrix[i] = make([]int, n+1)
	}

	//inicializar linhas e colunas
	for i := 0; i <= m; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= n; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(matrix[i-1][j]+1, matrix[i][j-1]+1, matrix[i-1][j-1]+cost)
		}
	}

	return matrix[m][n]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}


func textWithWidth(s string) string {
	if len(s) == 0 {
		return s
	}
	
	text := ""
	i := 0
	width := screenWidth() - 1

	if len(s) > width {
		for (i+width) < len(s)-1 {
			if utf8.ValidRune(rune(s[i+width])) {
				text += s[i:i+width] +"\n"
				i += width
			} else {
				j := 1
				for utf8.ValidRune(rune(s[i+width-j])) && j < width {	
					j++
				}
				i += width-(j+2)
				text += s[i:i+width-(j+2)] +"\n"
			}
		}
		text += s[i:len(s)-1] +"\n"
		return text
	} else {
		return s
	}
}
