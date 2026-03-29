package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taxpilot/internal/calc"
	"taxpilot/internal/forms"
	"taxpilot/internal/interview"
	"taxpilot/internal/tui"
)

// Styles specific to rollforward view
var (
	rfFlaggedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Bold(true)
	rfChangedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF"))
	rfInputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379"))
	rfComputedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
	rfCursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3E4451")).
			Foreground(lipgloss.Color("#FFFFFF"))
	rfEditStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#4A90D9"))
	rfFlashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C678DD")).
			Bold(true)
	rfHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABB2BF")).
			Bold(true)
)

// flashClearMsg is sent to clear the flash highlight after a delay.
type flashClearMsg struct{}

// RollforwardView is the interactive spreadsheet-like editor for rollforward mode.
type RollforwardView struct {
	rf *interview.Rollforward

	// UI state
	cursor       int // selected field row index
	scrollOffset int
	width        int
	height       int

	// Editing
	editing    bool
	editBuffer string
	editCursor int
	editErr    string

	// Filters
	showOnlyFlagged bool
	showOnlyInputs  bool

	// Flash: keys that just changed from an edit
	flashKeys map[string]bool

	// Calculator sub-mode
	calcMode      bool
	calcInput     string
	calcResult    string
	calcResultVal float64
	calcHasResult bool
	calcRates     map[string]float64
	calcRatesErr  string
	calcLoading   bool

	// Dependency info overlay
	showDeps    bool
	depInfoText string

	// Status message
	statusMsg string
}

// NewRollforwardView creates a new rollforward view.
func NewRollforwardView(rf *interview.Rollforward) RollforwardView {
	return RollforwardView{
		rf:        rf,
		flashKeys: make(map[string]bool),
	}
}

// Init satisfies tea.Model.
func (m RollforwardView) Init() tea.Cmd {
	return nil
}

// visibleFields returns the fields to display based on current filter settings.
func (m RollforwardView) visibleFields() []interview.RollforwardField {
	var result []interview.RollforwardField
	for _, f := range m.rf.Fields {
		if m.showOnlyFlagged && !f.Flagged && !f.Changed {
			continue
		}
		if m.showOnlyInputs && f.FieldType != forms.UserInput {
			continue
		}
		result = append(result, f)
	}
	return result
}

// findVisibleIndex returns the index of a field key in the visible fields list.
func (m RollforwardView) findVisibleIndex(visible []interview.RollforwardField, key string) int {
	for i, f := range visible {
		if f.Key == key {
			return i
		}
	}
	return -1
}

// Update satisfies tea.Model.
func (m RollforwardView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case flashClearMsg:
		m.flashKeys = make(map[string]bool)
		return m, nil

	case tui.ExportPDFResultMsg:
		if msg.Err != nil {
			m.statusMsg = fmt.Sprintf("Export error: %v", msg.Err)
		} else {
			m.statusMsg = fmt.Sprintf("Exported %d file(s)", len(msg.Paths))
		}
		return m, nil

	case tui.ExchangeRatesMsg:
		m.calcLoading = false
		if msg.Err != nil {
			m.calcRatesErr = msg.Err.Error()
		} else {
			m.calcRates = msg.Rates
			m.calcRatesErr = ""
		}
		return m, nil

	case tea.KeyMsg:
		if m.calcMode {
			return m.updateCalcMode(msg)
		}
		if m.editing {
			return m.updateEditing(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m RollforwardView) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visible := m.visibleFields()
	maxIdx := len(visible) - 1

	switch msg.String() {
	case "q", "ctrl+c":
		_ = m.rf.SaveState()
		return m, tea.Quit

	case "esc":
		m.showDeps = false
		m.statusMsg = ""
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
		return m, nil

	case "down", "j":
		if m.cursor < maxIdx {
			m.cursor++
			m.ensureVisible()
		}
		return m, nil

	case "pgup", "ctrl+u":
		half := m.viewableLines() / 2
		m.cursor -= half
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, nil

	case "pgdown", "ctrl+d":
		half := m.viewableLines() / 2
		m.cursor += half
		if m.cursor > maxIdx {
			m.cursor = maxIdx
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, nil

	case "home", "g":
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil

	case "end", "G":
		m.cursor = maxIdx
		m.ensureVisible()
		return m, nil

	case "enter":
		m.showDeps = false
		if len(visible) > 0 && m.cursor <= maxIdx {
			field := visible[m.cursor]
			if field.FieldType == forms.UserInput {
				// Edit the input field
				m.editing = true
				m.editErr = ""
				if field.IsString {
					m.editBuffer = field.StrValue
				} else if field.Value != 0 {
					m.editBuffer = fmt.Sprintf("%.2f", field.Value)
				} else {
					m.editBuffer = ""
				}
				m.editCursor = len(m.editBuffer)
			} else {
				// Computed field: jump to first input source
				sources := m.rf.GetDepInfo(field.Key).InputSources
				if len(sources) > 0 {
					// Find in visible fields; try each source until one is visible
					for _, src := range sources {
						idx := m.findVisibleIndex(visible, src)
						if idx >= 0 {
							m.cursor = idx
							m.ensureVisible()
							m.statusMsg = fmt.Sprintf("Jumped to input: %s", m.rf.FieldLabel(src))
							break
						}
					}
				} else {
					m.statusMsg = "No editable input sources for this field"
				}
			}
		}
		return m, nil

	case "d":
		// Show dependency info for current field
		m.showDeps = false
		if len(visible) > 0 && m.cursor <= maxIdx {
			field := visible[m.cursor]
			info := m.rf.GetDepInfo(field.Key)

			var lines []string
			lines = append(lines, fmt.Sprintf("Dependencies for: %s [%s]", field.Label, field.Key))
			lines = append(lines, "")

			if len(info.DirectDeps) > 0 {
				lines = append(lines, "Direct dependencies:")
				for _, dep := range info.DirectDeps {
					label := m.rf.FieldLabel(dep)
					lines = append(lines, fmt.Sprintf("  %s  (%s)", label, dep))
				}
			} else {
				lines = append(lines, "No dependencies (standalone field)")
			}

			if len(info.InputSources) > 0 {
				lines = append(lines, "")
				lines = append(lines, "Input sources (editable):")
				for _, src := range info.InputSources {
					label := m.rf.FieldLabel(src)
					val := m.rf.Computed[src]
					strVal := m.rf.StrInputs[src]
					if strVal != "" {
						lines = append(lines, fmt.Sprintf("  %s = %s  (%s)", label, strVal, src))
					} else {
						lines = append(lines, fmt.Sprintf("  %s = %s  (%s)", label, formatDollar(val), src))
					}
				}
				lines = append(lines, "")
				lines = append(lines, "Press Enter to jump to first input source")
			}

			m.depInfoText = strings.Join(lines, "\n")
			m.showDeps = true
		}
		return m, nil

	case "f":
		m.showOnlyFlagged = !m.showOnlyFlagged
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil

	case "i":
		m.showOnlyInputs = !m.showOnlyInputs
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil

	case "s":
		if err := m.rf.SaveState(); err != nil {
			m.statusMsg = fmt.Sprintf("Save error: %v", err)
		} else {
			m.statusMsg = "State saved"
		}
		return m, nil

	case "e":
		return m, func() tea.Msg {
			return tui.ExportPDFMsg{
				Results:   m.rf.Computed,
				StrInputs: m.rf.StrInputs,
				TaxYear:   m.rf.TaxYear,
			}
		}

	case "c":
		m.calcMode = true
		m.calcInput = ""
		m.calcResult = ""
		m.calcHasResult = false
		if m.calcRates == nil && !m.calcLoading {
			m.calcLoading = true
			return m, func() tea.Msg {
				rates, err := calc.FetchRates()
				return tui.ExchangeRatesMsg{Rates: rates, Err: err}
			}
		}
		return m, nil

	case "tab":
		// Jump to next form
		if len(visible) == 0 {
			return m, nil
		}
		currentForm := visible[m.cursor].FormID
		for i := m.cursor + 1; i < len(visible); i++ {
			if visible[i].FormID != currentForm {
				m.cursor = i
				m.ensureVisible()
				return m, nil
			}
		}
		// Wrap to beginning
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil
	}
	return m, nil
}

func (m RollforwardView) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visible := m.visibleFields()
	if m.cursor >= len(visible) {
		m.editing = false
		return m, nil
	}
	field := visible[m.cursor]

	switch msg.String() {
	case "enter":
		// Commit edit
		m.editing = false
		m.editErr = ""

		if field.IsString {
			changed, err := m.rf.UpdateStrInput(field.Key, m.editBuffer)
			if err != nil {
				m.editErr = err.Error()
				return m, nil
			}
			return m.flashChanged(changed)
		}

		// Numeric
		val := 0.0
		if m.editBuffer != "" {
			// Strip commas and dollar signs
			clean := strings.ReplaceAll(m.editBuffer, ",", "")
			clean = strings.ReplaceAll(clean, "$", "")
			parsed, err := strconv.ParseFloat(clean, 64)
			if err != nil {
				m.editErr = "Invalid number"
				m.editing = true
				return m, nil
			}
			val = parsed
		}

		changed, err := m.rf.UpdateInput(field.Key, val)
		if err != nil {
			m.editErr = err.Error()
			return m, nil
		}
		return m.flashChanged(changed)

	case "esc":
		m.editing = false
		m.editErr = ""
		return m, nil

	case "backspace":
		if m.editCursor > 0 {
			m.editBuffer = m.editBuffer[:m.editCursor-1] + m.editBuffer[m.editCursor:]
			m.editCursor--
		}
		return m, nil

	case "left":
		if m.editCursor > 0 {
			m.editCursor--
		}
		return m, nil

	case "right":
		if m.editCursor < len(m.editBuffer) {
			m.editCursor++
		}
		return m, nil

	default:
		// Insert character
		if len(msg.String()) == 1 {
			ch := msg.String()[0]
			if (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' || ch == ',' ||
				(field.IsString && ch >= ' ' && ch <= '~') {
				m.editBuffer = m.editBuffer[:m.editCursor] + string(ch) + m.editBuffer[m.editCursor:]
				m.editCursor++
			}
		}
		return m, nil
	}
}

// updateCalcMode handles key events in calculator sub-mode.
func (m RollforwardView) updateCalcMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.calcMode = false
		m.calcInput = ""
		m.calcResult = ""
		m.calcHasResult = false
		return m, nil

	case tea.KeyEnter:
		if m.calcInput == "" && m.calcHasResult {
			// Use the result: apply to currently selected field if it's an editable input
			visible := m.visibleFields()
			if m.cursor < len(visible) {
				field := visible[m.cursor]
				if field.FieldType == forms.UserInput && !field.IsString {
					changed, err := m.rf.UpdateInput(field.Key, m.calcResultVal)
					if err != nil {
						m.editErr = err.Error()
					} else {
						m.calcMode = false
						m.calcInput = ""
						m.calcResult = ""
						m.calcHasResult = false
						return m.flashChanged(changed)
					}
				}
			}
			m.calcMode = false
			m.calcInput = ""
			m.calcResult = ""
			m.calcHasResult = false
			return m, nil
		}
		if m.calcInput != "" {
			result, breakdown, err := calc.Eval(m.calcInput, m.calcRates)
			if err != nil {
				m.calcResult = "Error: " + err.Error()
				m.calcHasResult = false
			} else {
				m.calcResultVal = result
				m.calcHasResult = true
				m.calcInput = ""
				if breakdown != "" {
					m.calcResult = breakdown
				} else {
					m.calcResult = fmt.Sprintf("= %.2f", result)
				}
			}
		}
		return m, nil

	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.calcInput) > 0 {
			m.calcInput = m.calcInput[:len(m.calcInput)-1]
			m.calcHasResult = false
			m.calcResult = ""
		}
		return m, nil

	case tea.KeyRunes:
		m.calcInput += msg.String()
		m.calcHasResult = false
		m.calcResult = ""
		return m, nil

	case tea.KeySpace:
		m.calcInput += " "
		return m, nil
	}

	return m, nil
}

func (m RollforwardView) flashChanged(changed []string) (tea.Model, tea.Cmd) {
	m.flashKeys = make(map[string]bool, len(changed))
	for _, k := range changed {
		m.flashKeys[k] = true
	}
	return m, tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
		return flashClearMsg{}
	})
}

// displayRow is a single line in the rendered field table.
type displayRow struct {
	text     string // rendered line
	fieldIdx int    // index into visible fields, or -1 for headers/separators
}

// buildDisplayRows pre-computes all rows including form headers and separators.
func (m RollforwardView) buildDisplayRows() ([]displayRow, int) {
	visible := m.visibleFields()
	var rows []displayRow
	cursorRow := 0
	currentForm := forms.FormID("")

	for idx, field := range visible {
		// Form separator
		if field.FormID != currentForm {
			currentForm = field.FormID
			if len(rows) > 0 {
				rows = append(rows, displayRow{text: "", fieldIdx: -1})
			}
			rows = append(rows, displayRow{
				text:     tui.HighlightStyle.Render(fmt.Sprintf("  %s", field.FormName)),
				fieldIdx: -1,
			})
		}

		if idx == m.cursor {
			cursorRow = len(rows)
		}
		rows = append(rows, displayRow{
			text:     m.renderFieldRow(idx, field),
			fieldIdx: idx,
		})
	}

	return rows, cursorRow
}

func (m *RollforwardView) ensureVisible() {
	_, cursorRow := m.buildDisplayRows()
	maxLines := m.viewableLines()
	if cursorRow < m.scrollOffset {
		m.scrollOffset = cursorRow
	} else if cursorRow >= m.scrollOffset+maxLines {
		m.scrollOffset = cursorRow - maxLines + 1
	}
}

func (m RollforwardView) viewableLines() int {
	// Reserve: border (2) + padding (2) + header (1) + blank (1) + column header (1)
	// + separator (1) + footer blank (1) + footer (1) + status (1) = ~11
	lines := m.height - 11
	if lines < 5 {
		lines = 20
	}
	return lines
}

// View satisfies tea.Model.
func (m RollforwardView) View() string {
	if m.calcMode {
		return m.viewCalc()
	}

	var sections []string
	visible := m.visibleFields()

	// Header
	flagCount := m.rf.CountFlagged()
	header := tui.TitleStyle.Render(fmt.Sprintf(
		"Rollforward: %d \u2192 %d  |  %d flagged  |  %d total fields",
		m.rf.PriorYear, m.rf.TaxYear, flagCount, len(visible),
	))
	sections = append(sections, header)

	// Filter indicator
	var filters []string
	if m.showOnlyFlagged {
		filters = append(filters, "flagged only")
	}
	if m.showOnlyInputs {
		filters = append(filters, "inputs only")
	}
	if len(filters) > 0 {
		sections = append(sections, tui.HelpStyle.Render("Filter: "+strings.Join(filters, ", ")))
	}
	sections = append(sections, "")

	// Parameter changes summary
	if len(m.rf.ParamChanges) > 0 {
		sections = append(sections, rfFlaggedStyle.Render("Tax law changes:"))
		for _, pc := range m.rf.ParamChanges {
			sections = append(sections, fmt.Sprintf("  %s: $%.0f \u2192 $%.0f (%+.0f)",
				pc.Name, pc.OldValue, pc.NewValue, pc.Delta))
		}
		sections = append(sections, "")
	}

	// Column header
	colHeader := fmt.Sprintf("  %-42s %-7s %14s %14s %10s", "Field", "Type", "Value", "Prior Year", "Delta")
	sections = append(sections, rfHeaderStyle.Render(colHeader))
	sections = append(sections, "  "+strings.Repeat("\u2500", 91))

	// Build all display rows and slice the visible window
	displayRows, _ := m.buildDisplayRows()
	maxLines := m.viewableLines()

	endOffset := m.scrollOffset + maxLines
	if endOffset > len(displayRows) {
		endOffset = len(displayRows)
	}
	startOffset := m.scrollOffset
	if startOffset > len(displayRows) {
		startOffset = len(displayRows)
	}

	for _, row := range displayRows[startOffset:endOffset] {
		sections = append(sections, row.text)
	}

	// Scroll position indicator
	if len(displayRows) > maxLines {
		pct := 0
		if len(displayRows)-maxLines > 0 {
			pct = m.scrollOffset * 100 / (len(displayRows) - maxLines)
		}
		sections = append(sections, tui.HelpStyle.Render(
			fmt.Sprintf("  (%d/%d rows, %d%% scrolled)", endOffset, len(displayRows), pct),
		))
	}

	// Dependency info overlay
	if m.showDeps && m.depInfoText != "" {
		depBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#61AFEF")).
			Padding(0, 1).
			Width(tui.ContentWidth(m.width) - 4)

		sections = append(sections, "")
		sections = append(sections, depBox.Render(m.depInfoText))
	}

	// Edit popup overlay
	if m.editing {
		visible := m.visibleFields()
		if m.cursor < len(visible) {
			field := visible[m.cursor]

			editBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4A90D9")).
				Padding(0, 1).
				Width(tui.ContentWidth(m.width) - 4)

			// Build the edit buffer display with cursor
			editDisplay := m.editBuffer
			if m.editCursor < len(editDisplay) {
				editDisplay = editDisplay[:m.editCursor] + "\u2588" + editDisplay[m.editCursor:]
			} else {
				editDisplay += "\u2588"
			}

			var priorDisplay string
			if field.IsString {
				if len(field.Options) > 0 {
					priorDisplay = formatOptionLabel(field.PriorStr)
				} else {
					priorDisplay = field.PriorStr
				}
			} else {
				priorDisplay = formatDollar(field.PriorValue)
			}

			popupLines := []string{
				tui.TitleStyle.Render("Editing: " + field.Label),
				tui.HelpStyle.Render("Prior year: " + priorDisplay),
				"",
				tui.HighlightStyle.Render("\u25b8 ") + tui.InputStyle.Render(editDisplay),
			}
			if m.editErr != "" {
				popupLines = append(popupLines, tui.ErrorStyle.Render(m.editErr))
			}
			popupLines = append(popupLines, tui.HelpStyle.Render("[Enter] confirm  [Esc] cancel"))

			sections = append(sections, "")
			sections = append(sections, editBox.Render(
				lipgloss.JoinVertical(lipgloss.Left, popupLines...),
			))
		}
	}

	// Error / status (non-edit)
	if !m.editing && m.editErr != "" {
		sections = append(sections, tui.ErrorStyle.Render("  "+m.editErr))
	}
	if m.statusMsg != "" {
		sections = append(sections, tui.SuccessStyle.Render("  "+m.statusMsg))
	}

	// Footer
	sections = append(sections, "")
	if m.editing {
		// footer already in popup
	} else {
		sections = append(sections, tui.HelpStyle.Render(
			"[j/k] navigate  [ctrl+d/u] half-page  [Enter] edit/jump  [d] deps  [c] calc  [Tab] next form  [f] flagged  [i] inputs  [s] save  [e] export  [q] quit",
		))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, sections...)
	contentW := tui.ContentWidth(m.width)
	return tui.BorderStyle.Width(contentW).Render(body) + "\n"
}

// viewCalc renders the calculator overlay.
func (m RollforwardView) viewCalc() string {
	contentW := tui.ContentWidth(m.width)

	// Target field info
	var targetInfo string
	visible := m.visibleFields()
	if m.cursor < len(visible) {
		field := visible[m.cursor]
		if field.FieldType == forms.UserInput && !field.IsString {
			targetInfo = tui.PromptStyle.Render(fmt.Sprintf("Result will be applied to: %s", field.Label))
		} else {
			targetInfo = tui.HelpStyle.Render("Result will not be applied (select an editable input field first)")
		}
	}

	title := tui.TitleStyle.Width(contentW).Render("Calculator")

	instructions := tui.HelpStyle.Width(contentW).Render(
		"Type an expression and press Enter to evaluate.\n" +
			"Supports: +, -, *, /  and currency codes (e.g., 1000 EUR, 500 GBP + 200 SEK)")

	var rateStatus string
	if m.calcLoading {
		rateStatus = tui.HelpStyle.Render("Loading exchange rates...")
	} else if m.calcRatesErr != "" {
		rateStatus = tui.ErrorStyle.Render("Rates: " + m.calcRatesErr)
	} else if m.calcRates != nil {
		rateStatus = tui.SuccessStyle.Render("Exchange rates loaded")
	}

	cursor := tui.HighlightStyle.Render("\u25b8 ")
	inputLine := cursor + tui.InputStyle.Render(m.calcInput) +
		tui.HighlightStyle.Render("\u2588")

	var resultBlock string
	if m.calcResult != "" {
		resultStyle := lipgloss.NewStyle().Bold(true).Width(contentW)
		if m.calcHasResult {
			resultStyle = resultStyle.Foreground(lipgloss.Color("#98C379"))
		} else {
			resultStyle = resultStyle.Foreground(lipgloss.Color("#E06C75"))
		}
		resultBlock = resultStyle.Render(m.calcResult)
	}

	var helpText string
	if m.calcHasResult {
		helpText = tui.HelpStyle.Render("Enter: use result  |  Esc: cancel  |  Type: new expression")
	} else {
		helpText = tui.HelpStyle.Render("Enter: calculate  |  Esc: cancel")
	}

	parts := []string{title, "", instructions}
	if rateStatus != "" {
		parts = append(parts, rateStatus)
	}
	if targetInfo != "" {
		parts = append(parts, targetInfo)
	}
	parts = append(parts, "", inputLine)
	if resultBlock != "" {
		parts = append(parts, resultBlock)
	}
	parts = append(parts, "", helpText)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return tui.BorderStyle.Width(contentW).Render(content) + "\n"
}

func (m RollforwardView) renderFieldRow(idx int, field interview.RollforwardField) string {
	cursor := "  "
	if idx == m.cursor {
		cursor = "\u25b8 "
	}

	// Type tag
	typeTag := ""
	switch field.FieldType {
	case forms.UserInput:
		typeTag = "[input]"
	case forms.Computed:
		typeTag = "[comp] "
	case forms.Lookup:
		typeTag = "[look] "
	case forms.FederalRef:
		typeTag = "[fref] "
	case forms.PriorYear:
		typeTag = "[prior]"
	}

	// Label (truncate if needed)
	label := field.Label
	if len(label) > 40 {
		label = label[:37] + "..."
	}

	// Values — format based on field type
	var valueStr, priorStr, deltaStr string
	isCurrency := !field.IsString && !field.IsInteger && !isNonCurrencyField(field)
	if field.IsString {
		valueStr = field.StrValue
		priorStr = field.PriorStr
		// Show human-readable label for enum/option fields
		if len(field.Options) > 0 {
			valueStr = formatOptionLabel(valueStr)
			priorStr = formatOptionLabel(priorStr)
		}
		if len(valueStr) > 14 {
			valueStr = valueStr[:11] + "..."
		}
		if len(priorStr) > 14 {
			priorStr = priorStr[:11] + "..."
		}
		deltaStr = "-"
	} else if !isCurrency {
		// Integer/count/factor/boolean fields — no dollar sign
		valueStr = formatPlainNumber(field.Value)
		priorStr = formatPlainNumber(field.PriorValue)
		delta := field.Value - field.PriorValue
		if delta == 0 {
			deltaStr = "-"
		} else {
			deltaStr = fmt.Sprintf("%+.0f", delta)
		}
	} else {
		// Currency fields
		valueStr = formatDollar(field.Value)
		priorStr = formatDollar(field.PriorValue)
		delta := field.Value - field.PriorValue
		if delta == 0 {
			deltaStr = "-"
		} else {
			deltaStr = fmt.Sprintf("%+.0f", delta)
		}
	}

	row := fmt.Sprintf("%s%-42s %-7s %14s %14s %10s",
		cursor, label, typeTag, valueStr, priorStr, deltaStr)

	// Flag indicator
	flag := ""
	if field.Flagged {
		flag = " !"
	}
	row += flag

	// Apply styling
	if m.editing && idx == m.cursor {
		return rfEditStyle.Render(row)
	}
	if m.flashKeys[field.Key] {
		return rfFlashStyle.Render(row)
	}
	if idx == m.cursor {
		return rfCursorStyle.Render(row)
	}
	if field.Flagged {
		return rfFlaggedStyle.Render(row)
	}
	if field.Changed {
		return rfChangedStyle.Render(row)
	}
	if field.FieldType == forms.UserInput {
		return rfInputStyle.Render(row)
	}
	return rfComputedStyle.Render(row)
}

// isNonCurrencyField detects fields that are counts, factors, booleans, or
// other non-dollar numeric values based on label keywords.
func isNonCurrencyField(field interview.RollforwardField) bool {
	lower := strings.ToLower(field.Label)
	keywords := []string{
		"number of", "factor", "check", "qualifying child",
		"full year", "self employ", "foreign accounts",
		"foreign interest", "accrued or paid",
		"basis reported", "lives abroad",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	// Small whole numbers (0 or 1) that look like booleans/counts
	if field.Value == 0 && field.PriorValue == 0 {
		return false // can't tell, default to currency
	}
	if (field.Value == 0 || field.Value == 1) &&
		(field.PriorValue == 0 || field.PriorValue == 1) {
		return true
	}
	return false
}

// formatPlainNumber formats a number without dollar sign.
func formatPlainNumber(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}
