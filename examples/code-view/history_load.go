package main

import "context"

// historyLoadReply is one async commit page.
type historyLoadReply struct {
	Token int
	Limit int
	Log   commitLog
	Err   error
}

// startHistoryLoad reads a larger commit page off the page loop.
func (a *App) startHistoryLoad(ctx context.Context) {
	if a.ws.Sample || a.commits.Unavailable || a.historyScroll.Loading || a.historyDone {
		return
	}

	total := len(a.commits.Commits)
	visible := a.historyVisibleRows()
	if total == 0 || !a.historyScroll.ShouldLoad(total, visible) {
		return
	}

	next := NextHistoryLimit(total, maxHistoryCommits)
	if next <= total {
		a.historyDone = true

		return
	}

	a.historyToken++
	token := a.historyToken
	root := a.ws.Root
	ref := a.historyRef
	a.historyScroll.StartLoad()

	go func() {
		log, err := loadCommitLogLimit(ctx, root, ref, next)

		select {
		case a.historyLoad <- historyLoadReply{Token: token, Limit: next, Log: log, Err: err}:
		default:
		}

		a.page.WantFrameIn(0)
	}()
}

// receiveHistoryLoad drains one finished commit page on the page loop.
func (a *App) receiveHistoryLoad(ctx context.Context) bool {
	select {
	case reply := <-a.historyLoad:
		if reply.Token != a.historyToken {
			return false
		}

		a.historyScroll.DoneLoad()

		if reply.Err != nil {
			return false
		}

		a.commits = reply.Log
		a.historyLimit = reply.Limit
		a.historyScroll.Limit = len(reply.Log.Commits)
		a.historyDone = len(reply.Log.Commits) < reply.Limit || len(reply.Log.Commits) >= maxHistoryCommits

		if a.historyRef != "" {
			a.branchLogs.Put(a.historyRef, reply.Log)
		}

		if a.commitOpen != "" && a.findCommit(a.commitOpen) == nil {
			a.commitOpen = ""
			a.commitFile = ""
			a.commitRows = nil
			a.commitPreview = nil

			if a.diffSource == "commit" {
				a.diffSource = "working"
				a.rebuildDiffRows()
			}
		}

		a.refresh()
		_ = a.page.Redraw(ctx)

		return true
	default:
		return false
	}
}
