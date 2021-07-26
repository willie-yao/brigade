package term

import (
	"context"
	"sync"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/rivo/tview"
)

// pageRouter is a custom UI component composed of tview.Pages which can be
// refreshed and brought into focus on command.
type pageRouter struct {
	*tview.Pages
	projectsPage      *projectsPage
	projectPage       *projectPage
	eventPage         *eventPage
	jobPage           *jobPage
	logPage           *logPage
	app               *tview.Application
	cancelRefreshFn   func()
	cancelRefreshFnMu sync.Mutex
}

// NewPageRouter returns a custom UI component composed of tview.Pages which
// can be refreshed and brought into focus on command.
func NewPageRouter(
	apiClient core.APIClient,
	app *tview.Application,
) tview.Primitive {
	r := &pageRouter{
		Pages: tview.NewPages(),
		app:   app,
	}
	r.projectsPage = newProjectsPage(apiClient, app, r)
	r.AddPage(projectsPageName, r.projectsPage, true, false)
	r.projectPage = newProjectPage(apiClient, app, r)
	r.AddPage(projectPageName, r.projectPage, true, false)
	r.eventPage = newEventPage(apiClient, app, r)
	r.AddPage(eventPageName, r.eventPage, true, false)
	r.jobPage = newJobPage(apiClient, app, r)
	r.AddPage(jobPageName, r.jobPage, true, false)
	r.logPage = newLogPage(apiClient, app, r)
	r.AddPage(logPageName, r.logPage, true, false)
	r.loadProjectsPage()
	return r
}

// loadProjectsPage refreshes the projects page and brings it into focus.
func (r *pageRouter) loadProjectsPage() {
	r.loadPage(projectsPageName, func() {
		r.projectsPage.refresh()
	})
}

// loadProjectPage refreshes the project page and brings it into focus.
func (r *pageRouter) loadProjectPage(projectID string) {
	r.loadPage(projectPageName, func() {
		r.projectPage.refresh(projectID)
	})
}

// loadEventPage refreshes the event page and brings it into focus.
func (r *pageRouter) loadEventPage(eventID string) {
	r.loadPage(eventPageName, func() {
		r.eventPage.refresh(eventID)
	})
}

// loadJobPage refreshes the job page and brings it into focus.
func (r *pageRouter) loadJobPage(eventID, jobID string) {
	r.loadPage(jobPageName, func() {
		r.jobPage.refresh(eventID, jobID)
	})
}

// loadLogPage loads a floating window that displays logs and brings it into
// focus.
func (r *pageRouter) loadLogPage(page *page, eventID, jobID string) {
	// go r.logPage.streamLogs(eventID, jobID, quit)
	r.loadPage(logPageName, func() {
		r.logPage.refresh(*page, eventID, jobID)
	}, r.logPage.logText)

	r.logPage.logText.Clear()

	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				r.logPage.streamLogs(eventID, jobID, quit)
			}
		}
	}()
}

// loadPage can refresh any page and bring it into focus, given the name of the
// page and a refresh function.
func (r *pageRouter) loadPage(
	pageName string,
	fn func(),
	focusPage ...tview.Primitive,
) {
	// This is a critical section of code. We only want one page auto-refreshing
	// at a time.
	r.cancelRefreshFnMu.Lock()
	defer r.cancelRefreshFnMu.Unlock()
	// If any page is already auto-refreshing, stop it
	if r.cancelRefreshFn != nil {
		r.cancelRefreshFn()
	}
	// Build a new context for the auto-refresh goroutine to use
	var ctx context.Context
	ctx, r.cancelRefreshFn = context.WithCancel(context.Background())
	if focusPage == nil {
		r.SwitchToPage(pageName) // Focus page and hide background
	} else {
		r.ShowPage(pageName) // Focus page and keep background
		r.app.SetFocus(focusPage[0])
	}
	fn()        // Synchronously refresh the page once
	go func() { // Start auto-refreshing
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				fn()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// exit stops the associated tview.Application.
func (r *pageRouter) exit() {
	r.app.Stop()
}
