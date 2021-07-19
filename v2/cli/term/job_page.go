package term

import (
	"context"
	"fmt"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const jobPageName = "job"

// jobPage is a custom UI component that displays Job info and a list of
// associated logs.
type jobPage struct {
	*page
	jobInfo *tview.TextView
	logsBox *tview.TextView
	usage   *tview.TextView
}

// newJobPage returns a custom UI component that displays Job info and a list of
// associated logs.
func newJobPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *jobPage {
	j := &jobPage{
		page:    newPage(apiClient, app, router),
		jobInfo: tview.NewTextView().SetDynamicColors(true),
		logsBox: tview.NewTextView().SetDynamicColors(true),
		usage: tview.NewTextView().SetDynamicColors(true).SetText(
			"[yellow](F5) [white]Reload    [yellow](<-/Del) [white]Back    [yellow](ESC) [white]Home    [yellow](Q) [white]Quit", // nolint: lll
		),
	}
	j.jobInfo.SetBorder(true).SetBorderColor(tcell.ColorYellow)
	j.logsBox.SetChangedFunc(
		func() {
			j.app.Draw()
		},
	)
	j.logsBox.SetBorder(true).SetTitle(" Logs ")
	// Create the layout
	j.page.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(j.jobInfo, 0, 1, false).
		AddItem(j.logsBox, 0, 5, true).
		AddItem(j.usage, 1, 1, false)
	return j
}

// refresh refreshes Job info and repaints the page.
func (j *jobPage) refresh(eventID, jobName string) {
	event, err := j.apiClient.Events().Get(context.TODO(), eventID)
	if err != nil {
		// TODO: Handle this
	}
	job, found := event.Worker.Job(jobName)
	if !found {
		// TODO: Handle this
	}
	j.fillJobInfo(eventID, job)
	j.fillLogs(eventID, job.Name)
	// Set key handlers
	j.logsBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF5: // Reload
			j.router.loadJobPage(eventID, jobName)
		case // Back
			tcell.KeyLeft,
			tcell.KeyDelete,
			tcell.KeyBackspace,
			tcell.KeyBackspace2:
			j.router.loadEventPage(eventID)
		case tcell.KeyEsc: // Home
			j.router.loadProjectsPage()
		case tcell.KeyRune: // Regular key handling:
			switch event.Rune() {
			case 'r', 'R': // Reload
				j.router.loadJobPage(eventID, jobName)
			case 'q', 'Q': // Exit
				j.router.exit()
			}
		}
		return event
	})
}

func (j *jobPage) fillJobInfo(eventID string, job core.Job) {
	j.jobInfo.SetTitle(fmt.Sprintf(" %s: %s ", eventID, job.Name))
	j.jobInfo.SetBorderColor(getColorFromJobPhase(job.Status.Phase))
	j.jobInfo.Clear()
	infoText := fmt.Sprintf(
		`[grey]Primary Image: [white]%s
[grey]Created: [white]%s
[grey]Started: [white]%s
[grey]Ended: [white]%s`,
		job.Spec.PrimaryContainer.Image,
		"TODO", // We need a job created field; it doesn't exist yet
		formatDateTimeToString(job.Status.Started),
		formatDateTimeToString(job.Status.Ended),
	)
	if job.Status.Started != nil && job.Status.Ended != nil {
		infoText = fmt.Sprintf(
			"%s\n[grey]Duration: [white]%v",
			infoText,
			job.Status.Ended.Sub(*job.Status.Started),
		)
	}
	j.jobInfo.SetText(infoText)
}

func (j *jobPage) fillLogs(eventID, jobName string) {
	j.logsBox.Clear()
	go j.streamLog(eventID, jobName)
}

// nolint: lll
func (j *jobPage) streamLog(eventID, jobName string) {
	// // Initialize control channels for the streaming.
	// j.stopStreaming = make(chan struct{})
	// j.canStream = make(chan struct{})

	// // Save the context on goroutine.
	// ss := j.stopStreaming
	// cs := j.canStream
	// l := ctx.Log

	// // Close our reader when finished streaming, ignore if error.
	// defer l.Close()

	// // When finished we are ready to stream again. Only one can stream at a time.
	// defer func() {
	// 	close(cs)
	// 	cs = nil
	// }()

	// // Run a goroutine to check the state of the job on inteval N.
	// // If job finished we could reload everything and stop our streaming.
	// go func() {
	// 	t := time.NewTicker(5 * time.Second)
	// 	defer t.Stop()
	// 	for range t.C {
	// 		// Check if another streaming has been started before finishing this
	// 		// and we need to stop checking this job status.
	// 		select {
	// 		case <-ss:
	// 			return
	// 		default:
	// 		}

	// 		// If not running is time to reload everything.
	// 		if ctx.Job.Phase != core.JobPhaseRunning {
	// 			j.Refresh(projectID, eventID, ctx.Job.Name)
	// 			return
	// 		}
	// 	}
	// }()

	// // Start showing the stream on the textView.
	// // Ignore the copy error.
	// j.copyWithAnsiColors(j.logBox, readerFunc(func(p []byte) (n int, err error) {
	// 	select {
	// 	// if we don't want to continue reading return 0.
	// 	case <-ss:
	// 		return 0, io.EOF
	// 	default: // Fallback to read.
	// 		return l.Read(p)
	// 	}
	// }))
}
