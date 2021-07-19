package term

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/util/duration"
)

const projectPageName = "project"

// projectPage is a custom UI component that displays Project info and a list of
// associated Events.
type projectPage struct {
	*page
	currentProjectID     string
	projectInfo          *tview.TextView
	eventsContinueValues []string // Stack of "continue" values to aid paging
	eventsTable          *tview.Table
	usage                *tview.TextView
}

// newProjectPage returns a custom UI component that displays Project info and a
// list of associated Events.
func newProjectPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *projectPage {
	p := &projectPage{
		page:                 newPage(apiClient, app, router),
		projectInfo:          tview.NewTextView().SetDynamicColors(true),
		eventsContinueValues: []string{""}, // "" == continue value for first page
		eventsTable:          tview.NewTable().SetSelectable(true, false),
		usage:                tview.NewTextView().SetDynamicColors(true),
	}
	p.projectInfo.SetBorder(true).SetBorderColor(tcell.ColorWhite)
	p.eventsTable.SetBorder(true).SetTitle(" Events ")
	// Create the layout
	p.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(p.projectInfo, 0, 2, false).
		AddItem(p.eventsTable, 0, 9, true).
		AddItem(p.usage, 1, 1, false)
	return p
}

// refresh refreshes Projects info and associated Events and repaints the page.
func (p *projectPage) refresh(projectID string) {
	if projectID != p.currentProjectID {
		p.currentProjectID = projectID
		// "" == continue value for first page
		p.eventsContinueValues = []string{""}
	}

	project, err := p.apiClient.Projects().Get(context.TODO(), projectID)
	if err != nil {
		// TODO: Handle this
	}
	events, err := p.apiClient.Events().List(
		context.TODO(),
		&core.EventsSelector{
			ProjectID: projectID,
		},
		&meta.ListOptions{
			Continue: p.eventsContinueValues[len(p.eventsContinueValues)-1],
			Limit:    20,
		},
	)
	if err != nil {
		// TODO: Handle this
	}
	p.fillProjectInfo(project)
	p.fillEventsTable(events)
	p.fillUsage(events)
	// Set key handlers
	p.eventsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF5: // Reload
			p.router.loadProjectPage(projectID)
		case // Back
			tcell.KeyLeft,
			tcell.KeyDelete,
			tcell.KeyEsc,
			tcell.KeyBackspace,
			tcell.KeyBackspace2:
			p.router.loadProjectsPage()
		case tcell.KeyRune: // Regular key handling:
			switch event.Rune() {
			case 'r', 'R': // Reload
				p.router.loadProjectPage(projectID)
			case 'p', 'P': // Previous page
				// Pop the current page continue value from the stack, but never pop it
				// if it's the only continue value. i.e. Never pop it if it's the empty
				// continue value that gets you the first page.
				if len(p.eventsContinueValues) > 1 {
					p.eventsContinueValues =
						p.eventsContinueValues[:len(p.eventsContinueValues)-1]
					p.router.loadProjectPage(projectID)
				}
			case 'n', 'N': // Next page
				if events.Continue != "" {
					// Push the continue value for the next page onto the stack
					p.eventsContinueValues =
						append(p.eventsContinueValues, events.Continue)
					p.router.loadProjectPage(projectID)
				}
			case 'q', 'Q': // Exit
				p.router.exit()
			}
		}
		return event
	})

}

func (p *projectPage) fillProjectInfo(project core.Project) {
	p.projectInfo.Clear()
	p.projectInfo.SetTitle(fmt.Sprintf(" %s ", project.ID))
	infoText := fmt.Sprintf("[grey]Description: [white]%s", project.Description)
	if project.Spec.WorkerTemplate.Git != nil {
		infoText = fmt.Sprintf("%s\n[grey]Git:", infoText)
		if project.Spec.WorkerTemplate.Git.CloneURL != "" {
			infoText = fmt.Sprintf(
				"%s\n  [grey]Clone URL: [white]%s",
				infoText,
				project.Spec.WorkerTemplate.Git.CloneURL,
			)
		}
	}
	infoText = fmt.Sprintf(
		"%s\n[grey]Created: %s",
		infoText,
		formatDateTimeToString(project.Created),
	)
	p.projectInfo.SetText(infoText)
}

func (p *projectPage) fillUsage(events core.EventList) {
	usageText := "[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home" // nolint: lll
	if len(p.eventsContinueValues) > 1 {
		usageText = fmt.Sprintf("%s    [yellow](P) [white]Previous Page", usageText)
	}
	if events.Continue != "" {
		usageText = fmt.Sprintf("%s    [yellow](N) [white]Next Page", usageText)
	}
	usageText = fmt.Sprintf("%s    [yellow](Q) [white]Quit", usageText)
	p.usage.SetText(usageText)
}

func (p *projectPage) fillEventsTable(events core.EventList) {
	const (
		statusCol int = iota
		idCol
		sourceCol
		typeCol
		ageCol
		startedCol
		endedCol
		durationCol
	)
	p.eventsTable.Clear()
	p.eventsTable.SetCell(
		0,
		statusCol,
		&tview.TableCell{
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		idCol,
		&tview.TableCell{
			Text:  "ID",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		sourceCol,
		&tview.TableCell{
			Text:  "Source",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		typeCol,
		&tview.TableCell{
			Text:  "Type",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		ageCol,
		&tview.TableCell{
			Text:  "Age",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		startedCol,
		&tview.TableCell{
			Text:  "Started",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		endedCol,
		&tview.TableCell{
			Text:  "Ended",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		durationCol,
		&tview.TableCell{
			Text:  "Duration",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	)
	for r, event := range events.Items {
		row := r + 1
		icon := getIconFromWorkerPhase(event.Worker.Status.Phase)
		color := getColorFromWorkerPhase(event.Worker.Status.Phase)
		p.eventsTable.SetCell(
			row,
			statusCol,
			&tview.TableCell{
				Text:  icon,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			idCol,
			&tview.TableCell{
				Text:  event.ID,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			sourceCol,
			&tview.TableCell{
				Text:  event.Source,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			typeCol,
			&tview.TableCell{
				Text:  event.Type,
				Align: tview.AlignLeft,
				Color: color,
			},
		)
		age := time.Since(*event.Created).Truncate(time.Second)
		p.eventsTable.SetCell(
			row,
			ageCol,
			&tview.TableCell{
				Text:  duration.ShortHumanDuration(age),
				Align: tview.AlignLeft,
				Color: color,
			},
		)
		if event.Worker.Status.Started != nil {
			started := time.Since(*event.Worker.Status.Started).Truncate(time.Second)
			p.eventsTable.SetCell(
				row,
				startedCol,
				&tview.TableCell{
					Text:  duration.ShortHumanDuration(started),
					Align: tview.AlignLeft,
					Color: color,
				},
			)
		}
		if event.Worker.Status.Ended != nil {
			ended := time.Since(*event.Worker.Status.Ended).Truncate(time.Second)
			p.eventsTable.SetCell(
				row,
				endedCol,
				&tview.TableCell{
					Text:  duration.ShortHumanDuration(ended),
					Align: tview.AlignLeft,
					Color: color,
				},
			)
		}
		if event.Worker.Status.Started != nil && event.Worker.Status.Ended != nil {
			duration := event.Worker.Status.Ended.Sub(
				*event.Worker.Status.Started,
			).Truncate(time.Second)
			p.eventsTable.SetCell(
				row,
				durationCol,
				&tview.TableCell{
					Text:  fmt.Sprintf("%v", duration),
					Align: tview.AlignLeft,
					Color: color,
				},
			)
		}
	}
	p.eventsTable.SetSelectedFunc(func(row, _ int) {
		if row > 0 { // Header row cells aren't selectable
			eventID := p.eventsTable.GetCell(row, idCol).Text
			p.router.loadEventPage(eventID)
		}
	})
}
