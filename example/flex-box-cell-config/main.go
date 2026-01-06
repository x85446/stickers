package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/76creates/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	colors = []string{
		"#fc5c65", "#fd9644", "#fed330", "#26de81", "#2bcbba",
		"#eb3b5a", "#fa8231", "#f7b731", "#20bf6b", "#0fb9b1",
		"#45aaf2", "#4b7bec", "#a55eea", "#d1d8e0", "#778ca3",
		"#2d98da", "#3867d6", "#8854d0", "#a5b1c2", "#4b6584",
	}

	// Cell labels (20 cells, skipping H and T which are used for commands)
	cellLabels = []string{
		"A", "B", "C", // Row 1
		"D", "E", "F", "G", "I", "J", "K", // Row 2 (skip H - used for hide text)
		"L", "M", "N", "O", "P", // Row 3
		"Q", "R", "S", "U", "V", // Row 4 (skip T - used for border toggle)
	}
)

// CellConfig stores configuration for each cell
type CellConfig struct {
	UseFixedWidth  bool
	UseFixedHeight bool
	FixedWidth     int
	FixedHeight    int
	RatioX         int
	RatioY         int
}

type ConfigMode int

const (
	ModeNormal ConfigMode = iota
	ModeConfiguring
	ModeSettingValue
)

type model struct {
	flexBox        *flexbox.FlexBox
	borderType     int               // 0=none, 1=normal, 2=rounded, 3=thick, 4=double
	hideText       bool              // Toggle text visibility
	cellConfigs    map[string]*CellConfig // Per-cell configuration
	configMode     ConfigMode        // Current configuration mode
	selectedCell   string            // Currently selected cell (A-V, skip H/T)
	settingType    string            // What we're setting (width, height, ratioX, ratioY)
	inputBuffer    string            // Buffer for numeric input
	firstInput     bool              // True if we haven't typed yet in value mode
	termWidth      int
	termHeight     int
}

var borderTypes = []lipgloss.Border{
	{}, // no border
	lipgloss.NormalBorder(),
	lipgloss.RoundedBorder(),
	lipgloss.ThickBorder(),
	lipgloss.DoubleBorder(),
}

var borderNames = []string{
	"Filled Background",
	"Normal Border",
	"Rounded Border",
	"Thick Border",
	"Double Border",
}

func main() {
	m := model{
		flexBox:     flexbox.New(0, 0),
		borderType:  1, // Start with normal border
		cellConfigs: initCellConfigs(),
		configMode:  ModeNormal,
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func initCellConfigs() map[string]*CellConfig {
	configs := make(map[string]*CellConfig)

	// Default configurations for all cells (skip H and T - used for commands)
	defaultRatios := map[string][2]int{
		// Row 1
		"A": {1, 3}, "B": {3, 3}, "C": {1, 3},
		// Row 2 (skip H)
		"D": {2, 4}, "E": {2, 4}, "F": {3, 4}, "G": {3, 4},
		"I": {3, 4}, "J": {4, 4}, "K": {4, 4},
		// Row 3
		"L": {2, 5}, "M": {3, 5}, "N": {10, 5}, "O": {3, 5}, "P": {2, 5},
		// Row 4 (skip T)
		"Q": {1, 4}, "R": {1, 4}, "S": {1, 4}, "U": {1, 4}, "V": {1, 4},
	}

	for label, ratios := range defaultRatios {
		configs[label] = &CellConfig{
			UseFixedWidth:  false,
			UseFixedHeight: false,
			FixedWidth:     20,
			FixedHeight:    5,
			RatioX:         ratios[0],
			RatioY:         ratios[1],
		}
	}

	return configs
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
		m.rebuildFlexBox()
	case tea.KeyMsg:
		switch m.configMode {
		case ModeNormal:
			return m.handleNormalMode(msg)
		case ModeConfiguring:
			return m.handleConfiguringMode(msg)
		case ModeSettingValue:
			return m.handleSettingValueMode(msg)
		}
	}
	return m, nil
}

func (m *model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := strings.ToUpper(msg.String())

	switch key {
	case "ctrl+c", "Q":
		return m, tea.Quit
	case "T":
		m.borderType = (m.borderType + 1) % len(borderTypes)
		m.rebuildFlexBox()
	case "H":
		m.hideText = !m.hideText
		m.rebuildFlexBox()
	default:
		// Check if it's a cell label (A-T)
		for _, label := range cellLabels {
			if key == label {
				m.configMode = ModeConfiguring
				m.selectedCell = label
				m.inputBuffer = ""
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *model) handleConfiguringMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	config := m.cellConfigs[m.selectedCell]

	switch strings.ToLower(msg.String()) {
	case "esc":
		m.configMode = ModeNormal
		m.rebuildFlexBox()
	case "w":
		// Toggle width mode and immediately go to value input
		config.UseFixedWidth = !config.UseFixedWidth
		m.configMode = ModeSettingValue
		m.firstInput = true
		if config.UseFixedWidth {
			m.settingType = "width"
			m.inputBuffer = strconv.Itoa(config.FixedWidth)
		} else {
			m.settingType = "ratioX"
			m.inputBuffer = strconv.Itoa(config.RatioX)
		}
		m.rebuildFlexBox()
	case "h":
		// Toggle height mode and immediately go to value input
		config.UseFixedHeight = !config.UseFixedHeight
		m.configMode = ModeSettingValue
		m.firstInput = true
		if config.UseFixedHeight {
			m.settingType = "height"
			m.inputBuffer = strconv.Itoa(config.FixedHeight)
		} else {
			m.settingType = "ratioY"
			m.inputBuffer = strconv.Itoa(config.RatioY)
		}
		m.rebuildFlexBox()
	case "1":
		if config.UseFixedWidth {
			m.configMode = ModeSettingValue
			m.settingType = "width"
			m.inputBuffer = strconv.Itoa(config.FixedWidth)
			m.firstInput = true
		}
	case "2":
		if config.UseFixedHeight {
			m.configMode = ModeSettingValue
			m.settingType = "height"
			m.inputBuffer = strconv.Itoa(config.FixedHeight)
			m.firstInput = true
		}
	case "3":
		if !config.UseFixedWidth {
			m.configMode = ModeSettingValue
			m.settingType = "ratioX"
			m.inputBuffer = strconv.Itoa(config.RatioX)
			m.firstInput = true
		}
	case "4":
		if !config.UseFixedHeight {
			m.configMode = ModeSettingValue
			m.settingType = "ratioY"
			m.inputBuffer = strconv.Itoa(config.RatioY)
			m.firstInput = true
		}
	}
	return m, nil
}

func (m *model) handleSettingValueMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	config := m.cellConfigs[m.selectedCell]

	switch msg.String() {
	case "esc":
		m.configMode = ModeConfiguring
	case "enter":
		// Enter accepts the current value
		if val, err := strconv.Atoi(m.inputBuffer); err == nil && val > 0 {
			switch m.settingType {
			case "width":
				config.FixedWidth = val
			case "height":
				config.FixedHeight = val
			case "ratioX":
				config.RatioX = val
			case "ratioY":
				config.RatioY = val
			}
		}
		m.configMode = ModeConfiguring
		m.rebuildFlexBox()
	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
		m.firstInput = false
	default:
		// Check if it's a digit
		if len(msg.String()) == 1 && msg.String()[0] >= '0' && msg.String()[0] <= '9' {
			if m.firstInput {
				// First digit overwrites the entire buffer
				m.inputBuffer = msg.String()
				m.firstInput = false
			} else {
				// Subsequent digits append
				m.inputBuffer += msg.String()
			}
		}
	}
	return m, nil
}

func (m *model) rebuildFlexBox() {
	// Clear and rebuild
	m.flexBox.SetRows([]*flexbox.Row{})

	// Row 1: A, B, C
	row1 := m.flexBox.NewRow()
	row1Cells := []string{"A", "B", "C"}
	for i, label := range row1Cells {
		cell := m.createCell(label, i)
		row1.AddCells(cell)
	}

	// Row 2: D-K (7 cells, skip H)
	row2 := m.flexBox.NewRow()
	row2Cells := []string{"D", "E", "F", "G", "I", "J", "K"}
	for i, label := range row2Cells {
		cell := m.createCell(label, i+3)
		row2.AddCells(cell)
	}

	// Row 3: L-P (5 cells)
	row3 := m.flexBox.NewRow()
	row3Cells := []string{"L", "M", "N", "O", "P"}
	for i, label := range row3Cells {
		cell := m.createCell(label, i+10)
		row3.AddCells(cell)
	}

	// Row 4: Q-V (5 cells, skip T)
	row4 := m.flexBox.NewRow()
	row4Cells := []string{"Q", "R", "S", "U", "V"}
	for i, label := range row4Cells {
		cell := m.createCell(label, i+15)
		row4.AddCells(cell)
	}

	m.flexBox.AddRows([]*flexbox.Row{row1, row2, row3, row4})
}

func (m *model) createCell(label string, colorIndex int) *flexbox.Cell {
	config := m.cellConfigs[label]

	// Create cell with dynamic ratios
	cell := flexbox.NewCell(config.RatioX, config.RatioY)

	// Apply fixed dimensions if configured
	if config.UseFixedWidth {
		cell.SetFixedWidth(config.FixedWidth)
	}
	if config.UseFixedHeight {
		cell.SetFixedHeight(config.FixedHeight)
	}

	// Apply style
	cell.SetStyle(m.getCellStyle(colorIndex))

	// Set content generator
	cell.SetContentGenerator(func(w, h int) string {
		if m.hideText && m.configMode == ModeNormal {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}

		// Build content based on mode
		if m.configMode == ModeConfiguring && m.selectedCell == label {
			return m.renderConfigMenu(w, h)
		} else if m.configMode == ModeSettingValue && m.selectedCell == label {
			return m.renderValueInput(w, h)
		} else {
			return m.renderCellInfo(label, w, h)
		}
	})

	return cell
}

func (m *model) renderCellInfo(label string, w, h int) string {
	config := m.cellConfigs[label]

	lines := []string{
		fmt.Sprintf("[%s]", label),
	}

	// Width info
	if config.UseFixedWidth {
		lines = append(lines, fmt.Sprintf("W:fix(%d)", config.FixedWidth))
	} else {
		lines = append(lines, fmt.Sprintf("W:dyn(×%d)", config.RatioX))
	}

	// Height info
	if config.UseFixedHeight {
		lines = append(lines, fmt.Sprintf("H:fix(%d)", config.FixedHeight))
	} else {
		lines = append(lines, fmt.Sprintf("H:dyn(×%d)", config.RatioY))
	}

	// Actual dimensions
	lines = append(lines, fmt.Sprintf("Actual:%dx%d", w+m.borderOffset(), h+m.borderOffset()))

	content := strings.Join(lines, "\n")

	// Highlight if selected
	style := lipgloss.NewStyle().Width(w).Height(h).Align(lipgloss.Center, lipgloss.Center)
	if m.configMode != ModeNormal && m.selectedCell == label {
		style = style.Background(lipgloss.Color("#333333"))
	}

	return style.Render(content)
}

func (m *model) renderConfigMenu(w, h int) string {
	config := m.cellConfigs[m.selectedCell]

	wMode := "Dynamic"
	wValue := fmt.Sprintf("(×%d)", config.RatioX)
	if config.UseFixedWidth {
		wMode = "Fixed"
		wValue = fmt.Sprintf("(%d)", config.FixedWidth)
	}

	hMode := "Dynamic"
	hValue := fmt.Sprintf("(×%d)", config.RatioY)
	if config.UseFixedHeight {
		hMode = "Fixed"
		hValue = fmt.Sprintf("(%d)", config.FixedHeight)
	}

	lines := []string{
		fmt.Sprintf("[%s] CONFIG", m.selectedCell),
		"",
		fmt.Sprintf("Width:  %s %s", wMode, wValue),
		fmt.Sprintf("Height: %s %s", hMode, hValue),
		"",
		"[w] toggle & set width",
		"[h] toggle & set height",
		"",
		"[ESC] to exit",
	}

	content := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Width(w).Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color("#1a1a1a")).
		Foreground(lipgloss.Color("#ffffff")).
		Render(content)
}

func (m *model) renderValueInput(w, h int) string {
	settingLabel := m.settingType
	switch m.settingType {
	case "width":
		settingLabel = "Fixed Width"
	case "height":
		settingLabel = "Fixed Height"
	case "ratioX":
		settingLabel = "Width Ratio"
	case "ratioY":
		settingLabel = "Height Ratio"
	}

	lines := []string{
		fmt.Sprintf("[%s] SETTING", m.selectedCell),
		"",
		fmt.Sprintf("%s:", settingLabel),
		m.inputBuffer + "_",
		"",
		"[ENTER] accept",
		"[0-9] type new value",
		"[ESC] cancel",
	}

	content := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Width(w).Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color("#1a1a1a")).
		Foreground(lipgloss.Color("#00ff00")).
		Render(content)
}

func (m *model) getCellStyle(colorIndex int) lipgloss.Style {
	if m.borderType == 0 {
		// Filled background
		return lipgloss.NewStyle().Background(lipgloss.Color(colors[colorIndex]))
	}
	// Border with type
	return lipgloss.NewStyle().
		Border(borderTypes[m.borderType]).
		BorderForeground(lipgloss.Color(colors[colorIndex]))
}

func (m *model) borderOffset() int {
	if m.borderType == 0 {
		return 0
	}
	return 2
}

func (m *model) View() string {
	// Build status text based on mode
	var status string
	switch m.configMode {
	case ModeNormal:
		status = fmt.Sprintf("%s (T) | Text (H) | Select cell (A-V, skip H/T)", borderNames[m.borderType])
	case ModeConfiguring:
		config := m.cellConfigs[m.selectedCell]
		status = fmt.Sprintf("[%s] W:%v H:%v | (w) width (h) height (ESC) exit",
			m.selectedCell, config.UseFixedWidth, config.UseFixedHeight)
	case ModeSettingValue:
		status = fmt.Sprintf("Setting %s for [%s]: %s | (Enter) confirm (ESC) cancel",
			m.settingType, m.selectedCell, m.inputBuffer)
	}

	header := lipgloss.NewStyle().Bold(true).Render(status)
	return header + "\n" + m.flexBox.Render()
}