package term

import (
	"context"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/util/duration"
)

const projectsPageName = "projects"

// projectsPage is a custom UI component that displays the list of all
// Projects.
type projectsPage struct {
	*page
	projectsTable *tview.Table
	usage         *tview.TextView
}

// newProjectsPage returns a custom UI component that displays the list of all
// Projects.
func newProjectsPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *projectsPage {
	p := &projectsPage{
		page:          newPage(apiClient, app, router),
		projectsTable: tview.NewTable().SetSelectable(true, false),
		usage: tview.NewTextView().SetDynamicColors(true).SetText(
			"[yellow](F5) [white]Reload    [yellow](Q) [white]Quit",
		),
	}
	p.projectsTable.SetBorder(true).SetTitle("Projects")
	// Create the layout
	p.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(p.projectsTable, 0, 1, true).
		AddItem(p.usage, 1, 1, false)
	return p
}

// refresh refreshes the list of all Projects and repaints the page.
func (p *projectsPage) refresh() {
	projects, err := p.apiClient.Projects().List(context.TODO(), nil, nil)
	if err != nil {
		// TODO: Handle this
	}
	mostRecentEventByProject := map[string]core.Event{}
	for _, project := range projects.Items {
		events, err := p.apiClient.Events().List(
			context.TODO(),
			&core.EventsSelector{
				ProjectID: project.ID,
			},
			&meta.ListOptions{
				Limit: 1,
			},
		)
		if err != nil {
			// TODO: Handle this
		}
		if len(events.Items) > 0 {
			mostRecentEventByProject[project.ID] = events.Items[0]
		}
	}
	p.fillProjectsTable(projects, mostRecentEventByProject)
	// Key handling...
	p.projectsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF5: // Reload
			p.router.loadProjectsPage()
		case tcell.KeyRune: // Regular key handling
			switch event.Rune() {
			case 'r', 'R': // Reload
				p.router.loadProjectsPage()
			case 'q', 'Q': // Exit
				p.router.exit()
			}
		}
		return event
	})
}

func (p *projectsPage) fillProjectsTable(
	projects core.ProjectList,
	mostRecentEventByProject map[string]core.Event,
) {
	const (
		statusCol int = iota
		idCol
		descriptionCol
		lastEventTimeCol
	)
	p.projectsTable.Clear()
	p.projectsTable.SetCell(
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
		descriptionCol,
		&tview.TableCell{
			Text:  "Description",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		lastEventTimeCol,
		&tview.TableCell{
			Text:  "Last Event",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	)
	for r, project := range projects.Items {
		row := r + 1
		var since time.Duration
		color := unknownColor
		icon := unknownIcon
		lastEvent, found := mostRecentEventByProject[project.ID]
		if found {
			color = getColorFromWorkerPhase(lastEvent.Worker.Status.Phase)
			icon = getIconFromWorkerPhase(lastEvent.Worker.Status.Phase)
			since = time.Since(*lastEvent.Worker.Status.Started).Truncate(time.Second)
		}
		p.projectsTable.SetCell(
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
				Text:  project.ID,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			descriptionCol,
			&tview.TableCell{
				Text:  project.Description,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			lastEventTimeCol,
			&tview.TableCell{
				Text:  duration.ShortHumanDuration(since),
				Align: tview.AlignLeft,
				Color: color,
			},
		)
	}
	p.projectsTable.SetSelectedFunc(func(row, _ int) {
		if row > 0 { // Header row cells aren't selectable
			projectID := p.projectsTable.GetCell(row, idCol).Text
			p.router.loadProjectPage(projectID)
		}
	})
}
