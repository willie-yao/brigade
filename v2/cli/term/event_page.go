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
	eventInfo  *tview.TextView
	workerInfo *tview.TextView
	jobsTable  *tview.Table
	workerLogs *tview.TextView
	logModal   tview.Primitive
	usage      *tview.TextView
}

// newEventPage returns a custom UI component that displays Event info and a
// list of associated Jobs.
func newEventPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *eventPage {
	e := &eventPage{
		page:       newPage(apiClient, app, router),
		eventInfo:  tview.NewTextView().SetDynamicColors(true),
		workerInfo: tview.NewTextView().SetDynamicColors(true),
		jobsTable:  tview.NewTable().SetSelectable(true, false),
		workerLogs: tview.NewTextView().SetDynamicColors(true),
		usage: tview.NewTextView().SetDynamicColors(true).SetText(
			"[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](L) [white]Logs    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit", // nolint: lll
		),
	}
	e.eventInfo.SetBorder(true).SetBorderColor(tcell.ColorWhite)
	e.workerInfo.SetBorder(true).SetTitle("Worker")
	e.jobsTable.SetBorder(true).SetTitle("Jobs")
	e.workerLogs.SetBorder(true).SetTitle("Logs (<-/Del) Quit")

	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	e.logModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(e.workerLogs, 25, 1, false).
			AddItem(nil, 0, 1, false), 85, 1, false).
		AddItem(nil, 0, 1, false)

	// Create the layout
	e.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().
				AddItem(e.eventInfo, 0, 1, false).
				AddItem(e.workerInfo, 0, 1, false),
			0,
			1,
			false,
		).
		AddItem(e.jobsTable, 0, 3, true).
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
	e.fillWorkerInfo(event)
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
			case 'l', 'L':
				e.workerLogs.SetText("Placeholder logs")
				e.router.ShowPage("Event Logs")
				e.router.app.SetFocus(e.workerLogs)
			case 'q', 'Q': // Exit
				e.router.exit()
			}
		}
		return evt
	})

	e.workerLogs.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		switch evt.Key() {
		case // Back
			tcell.KeyLeft,
			tcell.KeyDelete,
			tcell.KeyBackspace,
			tcell.KeyBackspace2:
			e.router.HidePage("Event Logs")
			e.router.app.SetFocus(e.page)
		}
		return evt
	})

}

func (e *eventPage) fillEventInfo(event core.Event) {
	e.eventInfo.Clear()
	e.eventInfo.SetTitle(fmt.Sprintf(" %s ", event.ID))
	infoText := fmt.Sprintf(
		`[grey]Project: [white]%s
[grey]Source: [white]%s
[grey]Type: [white]%s`,
		event.ProjectID,
		event.Source,
		event.Type,
	)
	if len(event.Qualifiers) > 0 {
		infoText = fmt.Sprintf("%s\n[grey]Qualifiers:", infoText)
		for k, v := range event.Qualifiers {
			infoText = fmt.Sprintf("%s\n  [grey]%s: [white]%s", infoText, k, v)
		}
	}
	if len(event.Labels) > 0 {
		infoText = fmt.Sprintf("%s\n[grey]Labels:", infoText)
		for k, v := range event.Labels {
			infoText = fmt.Sprintf("%s\n  [grey]%s: [white]%s", infoText, k, v)
		}
	}
	if event.Git != nil {
		infoText = fmt.Sprintf("%s\n[grey]Git:", infoText)
		if event.Git.CloneURL != "" {
			infoText = fmt.Sprintf(
				"%s\n  [grey]Clone URL: [white]%s",
				infoText,
				event.Git.CloneURL,
			)
		}
		if event.Git.Commit != "" {
			infoText = fmt.Sprintf(
				"%s\n  [grey]Commit: [white]%s",
				infoText,
				event.Git.Commit,
			)
		}
		if event.Git.Ref != "" {
			infoText = fmt.Sprintf(
				"%s\n  [grey]Ref: [white]%s",
				infoText,
				event.Git.Ref,
			)
		}
	}
	infoText = fmt.Sprintf(
		"%s\n[grey]Created: [white]%s",
		infoText,
		formatDateTimeToString(event.Created),
	)
	e.eventInfo.SetText(infoText)
}

func (e *eventPage) fillWorkerInfo(event core.Event) {
	e.workerInfo.Clear()
	workerPhaseColor := getColorFromWorkerPhase(event.Worker.Status.Phase)
	e.workerInfo.SetBorderColor(workerPhaseColor).SetTitleColor(workerPhaseColor)
	image := "DEFAULT"
	if event.Worker.Spec.Container != nil {
		image = event.Worker.Spec.Container.Image
	}
	infoText := fmt.Sprintf(
		`[grey]Image: [white]%s
[grey]Started: [white]%s
[grey]Ended: [white]%s`,
		image,
		formatDateTimeToString(event.Worker.Status.Started),
		formatDateTimeToString(event.Worker.Status.Ended),
	)
	if event.Worker.Status.Started != nil && event.Worker.Status.Ended != nil {
		infoText = fmt.Sprintf(
			"%s\n[grey]Duration: [white]%s",
			infoText,
			event.Worker.Status.Ended.Sub(*event.Worker.Status.Started),
		)
	}
	infoText = fmt.Sprintf(
		"%s\n[grey]Phase: %s%s",
		infoText,
		getTextColorFromWorkerPhase(event.Worker.Status.Phase),
		event.Worker.Status.Phase,
	)
	e.workerInfo.SetText(infoText)
}

// nolint: lll
func (e *eventPage) streamEventLog(eventID string) {
	logEntryCh, errCh, err := e.apiClient.Events().Logs().Stream(
		context.Background(),
		eventID,
		&core.LogsSelector{},
		&core.LogStreamOptions{},
	)
	if errCh != nil || err != nil {
		// TODO: Handle this
	}

	logText := ""
	e.workerLogs.SetText(logText)

	for {
		select {
		case logEntry, ok := <-logEntryCh:
			if ok {
				logText = fmt.Sprintf("%s\n%s", logText, logEntry.Message)
				e.workerLogs.SetText(logText)
			} else {
				// logEntryCh was closed, but want to keep looping through this select
				// in case there are pending errors on the errCh still. nil channels are
				// never readable, so we'll just nil out logEntryCh and move on.
				logEntryCh = nil
			}
		case err, ok := <-errCh:
			if ok {
				// TODO: Remove and handle this
				fmt.Println(err)
			}
			// errCh was closed, but want to keep looping through this select in case
			// there are pending messages on the logEntryCh still. nil channels are
			// never readable, so we'll just nil out errCh and move on.
			errCh = nil
		case <-context.Background().Done():
			return
		}
		// If BOTH logEntryCh and errCh were closed, we're done.
		if logEntryCh == nil && errCh == nil {
			// TODO: Handle this
		}
	}
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
			Text:  "Primary Image",
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
		color := getColorFromJobPhase(job.Status.Phase)
		e.jobsTable.SetCell(
			row,
			statusCol,
			&tview.TableCell{
				Text:  getIconFromJobPhase(job.Status.Phase),
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
