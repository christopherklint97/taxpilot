package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

	case tea.KeyMsg:
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
		if len(visible) > 0 && m.cursor <= maxIdx {
			field := visible[m.cursor]
			if field.FieldType == forms.UserInput {
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
			}
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

	// Error / status
	if m.editErr != "" {
		sections = append(sections, tui.ErrorStyle.Render("  "+m.editErr))
	}
	if m.statusMsg != "" {
		sections = append(sections, tui.SuccessStyle.Render("  "+m.statusMsg))
	}

	// Footer
	sections = append(sections, "")
	if m.editing {
		sections = append(sections, tui.HelpStyle.Render(
			"[Enter] confirm  [Esc] cancel  [Backspace] delete",
		))
	} else {
		sections = append(sections, tui.HelpStyle.Render(
			"[j/k] navigate  [ctrl+d/u] half-page  [Enter] edit  [Tab] next form  [f] flagged  [i] inputs  [s] save  [e] export  [q] quit",
		))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, sections...)
	contentW := tui.ContentWidth(m.width)
	return tui.BorderStyle.Width(contentW).Render(body) + "\n"
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

	// Handle edit mode
	if m.editing && idx == m.cursor {
		// Show edit buffer with cursor
		editDisplay := m.editBuffer
		if m.editCursor < len(editDisplay) {
			editDisplay = editDisplay[:m.editCursor] + "\u2588" + editDisplay[m.editCursor:]
		} else {
			editDisplay += "\u2588"
		}
		valueStr = editDisplay
		if len(valueStr) > 14 {
			valueStr = valueStr[:14]
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
