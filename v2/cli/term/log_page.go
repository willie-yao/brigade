package term

import (
	"context"
	"fmt"
	"log"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const logPageName = "log"

type logPage struct {
	*page
	logText   *tview.TextView
	logString string
}

func newLogPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *logPage {
	l := &logPage{
		page:    newPage(apiClient, app, router),
		logText: tview.NewTextView().SetDynamicColors(true),
	}

	l.logText.SetBorder(true).SetTitle("Logs (<-/Del) Quit")
	l.logString = "Waiting for logs..."

	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	l.page.Flex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(l.logText, 25, 1, false).
			AddItem(nil, 0, 1, false), 85, 1, false).
		AddItem(nil, 0, 1, false)

	return l
}

// refresh refreshes Event info and associated Jobs and repaints the page.
func (l *logPage) refresh(page page, eventID string, jobID string) {

	// TODO: Implement log streaming and uncomment
	// go l.streamLogs(eventID, jobID)

	l.logText.SetText(l.logString)

	l.logText.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		switch evt.Key() {
		case // Back
			tcell.KeyLeft,
			tcell.KeyDelete,
			tcell.KeyBackspace,
			tcell.KeyBackspace2:
			l.router.HidePage(logPageName)
			l.router.app.SetFocus(page)
		}
		return evt
	})

}

// nolint: lll
func (l *logPage) streamLogs(eventID string, jobID string, quit chan bool) {
	var logsSelector core.LogsSelector
	if jobID == "" {
		logsSelector = core.LogsSelector{}
	} else {
		logsSelector = core.LogsSelector{Job: jobID}
	}
	logEntryCh, errCh, err := l.apiClient.Events().Logs().Stream(
		context.Background(),
		eventID,
		&logsSelector,
		&core.LogStreamOptions{Follow: true},
	)
	if err != nil {
		// TODO: Handle this
		log.Fatal(err)
	}

	// l.logText.SetText(logText)

	for {
		select {
		case logEntry, ok := <-logEntryCh:
			l.logString = fmt.Sprintf("%s\n%s", l.logString, logEntry.Message)
			if ok {
				l.logText.SetText(l.logString)
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
			l.logText.SetText(l.logString)
			quit <- true
			// return
		}
	}
}
