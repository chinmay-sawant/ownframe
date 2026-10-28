# Code Review example map

This page maps the `examples/code-view` app only. It shows how a window opens, how the shell stays fixed, and where Explorer, the editor, search, diff, and session state meet.

```mermaid
flowchart TD
    subgraph Startup["startup - main.go / app.go / build.go"]
        Flags["flags: root -files -dev -web"] --> NewLimit["NewWithLimit: disk workspace or sample"]
        NewLimit --> OwnNew["ownframe.New: buildHTML + Config Title Size DevTools"]
        OwnNew --> Handlers["Handle onClick Press Key Wheel Type + SetScrollWindow + SetTick tickWatch"]
        Handlers --> Initial["openInitial README + refresh: buildView + SetData"]
        Initial --> SessEn["EnableSession + reloadChanges + SetDevTools + SetPerf"]
        SessEn --> RunMode{"Run desktop or Serve -web?"}
        RunMode -- desktop --> RunD["ownframe.Run"]
        RunMode -- "-web preview" --> ServeW["ServeWithOptions"]
    end

    subgraph Shell["shell - titlebar / activity / statusbar"]
        RunD --> Title["titlebar: brand mode-explorer + cmdbar palette-open"]
        ServeW --> Title
        RunD --> Act["activity: explorer changes search details settings devtools"]
        ServeW --> Act
        RunD --> Status["statusbar: branch + counts + devtools"]
        ServeW --> Status
        Title --> Mode["onClick mode-explorer / mode-changes to setMode"]
        Act --> Mode
        Status --> Mode
        Mode --> Refresh["refresh: buildView Mode Tree Tabs Diff History + SetData + Redraw"]
        Refresh --> ScrollWin["onScroll slice windowed rows + SetScrollWindow"]
        ScrollWin --> FocusRoute["focus: filter palette find branch-filter workspace-search"]
    end

    subgraph Explorer["explorer + source control - workspace.go / rescan.go / sidebar"]
        Scan["loadDiskWorkspace WalkDir skipDir skipExt limit"] --> AddFile["addFile + finish sort Tree"]
        Scan --> Ign["loadIgnoredPaths ls-files + check-ignore"]
        Ign --> Dim["markIgnored + addIgnoredDirs placeholder"]
        AddFile --> ViewB["buildView + SetData"]
        Dim --> ViewB
        ViewB --> TreeR["treeRows walkTree filteredRows + statuses"]
        ViewB --> DiffL["fillDiff sidebarChangeFiles selectDiff"]
        ViewB --> HistR["fillHistory buildCommitGraph fillBranchList"]
        TreeR --> SideUI["sidebar.html trow ignored filter"]
        DiffL --> SideUI
        HistR --> SideUI
        SideUI --> ClickT["click open: / dir: / collapse-all / filter-clear"]
        SideUI --> ClickD["click diff-file: history-working branch-changes"]
        SideUI --> ClickC["click commit: toggleCommit branch-picker pick"]
        Watch["fsnotify watchLoop debounce"] --> Rescan["rescanDiskWorkspace diffWorkspace"]
        Rescan --> GitLive["loadLiveChanges diff + branches + commit log"]
        GitLive --> ApplyW["applyWatchUpdate installRescan refresh Redraw"]
        ClickC --> CommitLog["loadCommitLogLimit parseCommitLog patch cache"]
        CommitLog --> ApplyW
    end

    subgraph Editor["editor - pane.go / wrap.go / markdown"]
        OpenF["openFile / openFileIn"] --> Doc["ws.openDoc + fileDoc Lines Tokens Outline"]
        Doc --> TabsG["tabs + active per group + secondView"]
        TabsG --> FillC["fillCode windowSliceForGroup"]
        FillC --> WrapQ{"wrap?"}
        WrapQ -- yes --> WrapPlan["doc.wrapped cols wrap.go"]
        WrapQ -- no --> RawTok["doc.Tokens slice"]
        WrapPlan --> SliceR["split_scroll slice + find/edit decor"]
        RawTok --> SliceR
        SliceR --> RowsV["View.Rows to editor.html track"]
        FillC --> PrevQ{"preview markdown?"}
        PrevQ -- yes --> MDPrev["ensurePreview previewRows"]
        MDPrev --> MerQ{"mermaid fence?"}
        MerQ -- text --> MerTxt["appendFlowchart / Sequence text"]
        MerImg --> PrevRows["previewRow to preview-track"]
        MerTxt --> PrevRows
        RowsV --> EditOp["toggleEdit commit save + outline jump"]
    end

    subgraph Search["search + palette + settings - palette.go / search_ops.go"]
        CmdB["cmdbar palette-open / Ctrl+P"] --> OpenP["openPalette + Focus palette"]
        CtrlF["Ctrl+Shift+F"] --> OpenW["openWorkspaceSearch + Focus workspace-search"]
        OpenP --> QPal["onChange palette prefix then contains max 8"]
        QPal --> SubmitP["Enter openFile"]
        OpenW --> RunS["runWorkspaceSearch go searchWorkspace limit 200"]
        RunS --> ChanS["searchReply chan + WantFrameIn"]
        ChanS --> TickS["tickWatch receiveSearch"]
        TickS --> WinS["searchViewState 8-row window"]
        WinS --> HitS["search-hit:N / Enter openSearchHit scrollToSourceLine"]
        SetOpen["settings-open"] --> ThemeCh["theme:ID validTheme + Page.SetTheme"]
        ThemeCh --> Refresh
        OpenP --> EscAll["Esc close"]
        OpenW --> EscAll
        SetOpen --> EscAll
    end

    subgraph DiffSession["diff + session + tick - diff.go / session.go / watch.go"]
        GitDiff["git diff loadLiveChanges"] --> ParseD["parseDiff files + hunks"]
        ParseD --> Emph["applyEmphasis del/add pairs"]
        Emph --> RowsU["diffRowsUnified"]
        Emph --> RowsS["diffRowsSplit"]
        RowsU --> Gaps["identifyGaps"]
        RowsS --> Gaps
        Gaps --> Rebuild["rebuildDiffRows select file + split"]
        Rebuild --> WrapD["wrapDiffRows clipDiffRows"]
        WrapD --> FillD["fillDiffFile windowSlice + markCurrentDiffChange"]
        FillD --> ViewB
        ExpandG["expandDiffGap loadContextLines"] --> Rebuild
        NavM["modifiedNav Next Prev Select"] --> OpenCh["openChangedPath selectDiff"]
        OpenCh --> Rebuild
        NavC["fileNavChangeRows inlineChangeLines"] --> ScrollC["scrollTo scrollGroupTo"]
        ScrollC --> ViewB
        InlineM["inlineChanges map"] --> DecorR["decorateCodeRows showInlineCard"]
        DecorR --> ViewB
        SessLoad["EnableSession loadSession sanitize restoreState"] --> SessRef["refresh persistSession saveSession"]
        SessRef --> ViewB
        Watch --> TickLoop["tickWatch pollPatches receiveSearch"]
        TickLoop --> ApplyW
    end

    Refresh --> OpenF
    Refresh --> TreeR
    Refresh --> QPal
    Refresh --> Rebuild
    FocusRoute --> QPal
    FocusRoute --> RunS
    ClickT --> OpenF
    ClickD --> Rebuild
    HitS --> OpenF
    ScrollWin --> SliceR
    ScrollWin --> FillD
    ScrollWin --> WinS
```

## How to read it

- Startup builds one HTML shell from `components/*.html` and `components/*.css`, then opens the initial README and starts either desktop `Run` or `-web` `Serve`.
- The shell never scrolls. Title bar, activity bar, and status bar dispatch `mode-explorer`, `mode-changes`, `palette-open`, `search-open`, `details-toggle`, `settings-open`, and `devtools`.
- Explorer scans disk with `WalkDir`, marks git-ignored paths dimmed, and renders tree, changes, and history rows from `buildView`. Clicks route through `actions.go`.
- The editor opens docs per group, slices windowed 20px rows, and keeps wrap, find, edit, Markdown preview, text diagrams, and outline jumps on the same row pipeline.
- Search runs bounded async scans with an 8-row result window. Settings changes theme through `Page.SetTheme` and returns focus to the settings button.
- Diff pairs deleted and added lines, rebuilds unified or split rows, and shares the page scroll window with the editor. Session save and restore plus the fsnotify watcher feed back into `refresh`.

## File map

- Shell: `main.go`, `app.go`, `build.go`, `view.go`, `actions.go`, `pane.go`, `components/titlebar.html`, `components/activity.html`, `components/statusbar.html`.
- Explorer: `workspace.go`, `rescan.go`, `watch.go`, `branches.go`, `commits.go`, `history_load.go`, `git.go`, `components/sidebar.html`.
- Editor: `editor_group.go`, `split_scroll.go`, `wrap.go`, `find.go`, `text_edit.go`, `edit_ops.go`, `markdown_preview.go`, `markdown_mermaid.go`, `outline.go`, `components/editor.html`, `components/details.html`.
- Search: `palette.go`, `search_ops.go`, `workspace_search.go`, `themes.go`, `components/palette.html`, `components/settings.html`.
- Diff and session: `diff.go`, `diff_wrap.go`, `diff_expand.go`, `change_summary.go`, `file_nav_ops.go`, `modified_nav.go`, `inline_changes.go`, `inline_ops.go`, `session.go`, `app_session.go`.
