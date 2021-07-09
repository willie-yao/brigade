package term

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/util/duration"
)

const eventPageName = "event"

// eventPage is a custom UI component that displays Event info and a list of
// associated Jobs.
type eventPage struct {
	*page
	eventInfo *tview.TextView
	jobsTable *tview.Table
	usage     *tview.TextView
}

// newEventPage returns a custom UI component that displays Event info and a
// list of associated Jobs.
func newEventPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *eventPage {
	e := &eventPage{
		page:      newPage(apiClient, app, router),
		eventInfo: tview.NewTextView().SetDynamicColors(true),
		jobsTable: tview.NewTable().SetSelectable(true, false),
		usage: tview.NewTextView().SetDynamicColors(true).SetText(
			"[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit", // nolint: lll
		),
	}
	e.eventInfo.SetBorder(true).SetBorderColor(tcell.ColorYellow)
	e.jobsTable.SetBorder(true).SetTitle("Jobs")
	// Create the layout
	e.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.eventInfo, 0, 3, false).
		AddItem(e.jobsTable, 0, 6, true).
		AddItem(e.usage, 1, 1, false)
	return e
}

// refresh refreshes Event info and associated Jobs and repaints the page.
func (e *eventPage) refresh(eventID string) {
	event, err := e.apiClient.Events().Get(context.TODO(), eventID)
	if err != nil {
		// TODO: Handle this
	}
	e.fillEventInfo(event)
	e.fillJobsTable(event)
	// Set key handlers
	e.jobsTable.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		switch evt.Key() {
		case tcell.KeyF5: // Reload
			e.router.loadEventPage(eventID)
		case // Back
			tcell.KeyLeft,
			tcell.KeyDelete,
			tcell.KeyBackspace,
			tcell.KeyBackspace2:
			e.router.loadProjectPage(event.ProjectID)
		case tcell.KeyEsc: // Home
			e.router.loadProjectsPage()
		case tcell.KeyRune: // Regular key handling
			switch evt.Rune() {
			case 'r', 'R': // Reload
				e.router.loadEventPage(eventID)
			case 'q', 'Q': // Exit
				e.router.exit()
			}
		}
		return evt
	})

}

func (e *eventPage) fillEventInfo(event core.Event) {
	e.eventInfo.Clear()
	e.eventInfo.SetText(
		fmt.Sprintf(
			"[yellow]Event: [white]%s\n"+
				"[yellow]Source: [white]%s\n"+
				"[yellow]Type: [white]%s\n"+
				"[yellow]Time Created: [white]%s",
			event.ID,
			event.Source,
			event.Type,
			event.Created.Local().Format("2006-01-02 15:04:05"),
		),
	)

	e.eventTimeInfo.Clear()
	e.eventTimeInfo.SetText(
		fmt.Sprintf(
			"[yellow]Time Created: [white]%s\n"+
				"[yellow]Worker Started: [white]%s\n"+
				"[yellow]Worker Ended: [white]%s\n",
			event.Created.Local().Format("2006-01-02 15:04:05"),
			event.Worker.Status.Started.Local().Format("2006-01-02 15:04:05"),
			event.Worker.Status.Ended.Local().Format("2006-01-02 15:04:05"),
		),
	)
}

func (e *eventPage) fillJobsTable(event core.Event) {
	const (
		statusCol int = iota
		nameCol
		imageCol
		ageCol
		startedCol
		endedCol
		durationCol
	)
	e.jobsTable.Clear()
	e.jobsTable.SetCell(
		0,
		statusCol,
		&tview.TableCell{
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0,
		nameCol,
		&tview.TableCell{
			Text:  "Name",
			Align: tview.AlignCenter,
			Color: tcell.ColorYellow,
		},
	).SetCell(
		0, imageCol,
		&tview.TableCell{
			Text:  "Image",
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
	for r, job := range event.Worker.Jobs {
		row := r + 1
		icon := getIconFromJobPhase(job.Status.Phase)
		color := getColorFromJobPhase(job.Status.Phase)
		e.jobsTable.SetCell(
			row,
			statusCol,
			&tview.TableCell{
				Text:  icon,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			nameCol,
			&tview.TableCell{
				Text:  job.Name,
				Align: tview.AlignLeft,
				Color: color,
			},
		).SetCell(
			row,
			imageCol,
			&tview.TableCell{
				Text:  job.Spec.PrimaryContainer.Image,
				Align: tview.AlignLeft,
				Color: color,
			},
		)
		// TODO: Add age-- needs Job to track create time
		if job.Status.Started != nil {
			started := time.Since(*job.Status.Started).Truncate(time.Second)
			e.jobsTable.SetCell(
				row,
				startedCol,
				&tview.TableCell{
					Text:  duration.ShortHumanDuration(started),
					Align: tview.AlignLeft,
					Color: color,
				},
			)
		}
		if job.Status.Ended != nil {
			ended := time.Since(*job.Status.Ended).Truncate(time.Second)
			e.jobsTable.SetCell(
				row,
				endedCol,
				&tview.TableCell{
					Text:  duration.ShortHumanDuration(ended),
					Align: tview.AlignLeft,
					Color: color,
				},
			)
		}
		if job.Status.Started != nil && job.Status.Ended != nil {
			duration :=
				job.Status.Ended.Sub(*job.Status.Started).Truncate(time.Second)
			e.jobsTable.SetCell(
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
	e.jobsTable.SetSelectedFunc(func(row, _ int) {
		if row > 0 { // Header row cells aren't selectable
			jobName := e.jobsTable.GetCell(row, nameCol).Text
			e.router.loadJobPage(event.ID, jobName)
		}
	})
}
