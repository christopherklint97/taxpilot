package tui

import tea "github.com/charmbracelet/bubbletea"

// ViewName identifies which view is currently active.
type ViewName int

const (
	ViewWelcome   ViewName = iota
	ViewInterview
	ViewSummary
	ViewExport
	ViewEFile
)

// ViewFactory creates tea.Model instances for view transitions.
// The App calls these when it receives transition messages.
type ViewFactory struct {
	// MakeInterview creates the interview view.
	// Called when StartInterviewMsg is received.
	MakeInterview func(msg StartInterviewMsg) (tea.Model, error)

	// MakeSummary creates the summary view.
	// Called when ShowSummaryMsg is received.
	MakeSummary func(msg ShowSummaryMsg) tea.Model

	// ImportPriorYear handles importing a prior-year PDF.
	// Called when ImportPriorYearMsg is received.
	ImportPriorYear func(msg ImportPriorYearMsg) tea.Msg

	// Explain triggers a RAG-powered explanation for a form field.
	// Called when RequestExplanationMsg is received. May be nil if no LLM is configured.
	Explain func(msg RequestExplanationMsg) tea.Msg
}

// App is the top-level Bubble Tea model that routes between views.
type App struct {
	currentView ViewName
	width       int
	height      int
	// The active sub-model. Each view implements tea.Model.
	active  tea.Model
	factory ViewFactory
	// err holds any error from view transitions
	err string
}

// NewApp creates a new App with the given initial sub-model and view factory.
func NewApp(initial tea.Model, factory ViewFactory) *App {
	return &App{
		currentView: ViewWelcome,
		active:      initial,
		factory:     factory,
	}
}

// Init satisfies tea.Model.
func (a *App) Init() tea.Cmd {
	if a.active != nil {
		return a.active.Init()
	}
	return nil
}

// Update satisfies tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case StartInterviewMsg:
		if a.factory.MakeInterview != nil {
			view, err := a.factory.MakeInterview(msg)
			if err != nil {
				a.err = err.Error()
				return a, nil
			}
			a.active = view
			a.currentView = ViewInterview
			return a, a.active.Init()
		}

	case ShowSummaryMsg:
		if a.factory.MakeSummary != nil {
			a.active = a.factory.MakeSummary(msg)
			a.currentView = ViewSummary
			return a, a.active.Init()
		}

	case ImportPriorYearMsg:
		if a.factory.ImportPriorYear != nil {
			fn := a.factory.ImportPriorYear
			return a, func() tea.Msg {
				return fn(msg)
			}
		}

	case RequestExplanationMsg:
		if a.factory.Explain != nil {
			fn := a.factory.Explain
			return a, func() tea.Msg {
				return fn(msg)
			}
		}
		// No LLM configured — return a message indicating that
		return a, func() tea.Msg {
			return ExplanationResponseMsg{
				Explanation: "AI explanations are not available. Set OPENROUTER_API_KEY to enable them.",
			}
		}
	}

	if a.active != nil {
		var cmd tea.Cmd
		a.active, cmd = a.active.Update(msg)
		return a, cmd
	}
	return a, nil
}

// View satisfies tea.Model.
func (a *App) View() string {
	if a.err != "" {
		return ErrorStyle.Render("Error: "+a.err) + "\n"
	}
	if a.active != nil {
		return a.active.View()
	}
	return ""
}
