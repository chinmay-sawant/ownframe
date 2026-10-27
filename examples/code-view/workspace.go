package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	tabWidth            = 4
	tabSpaces           = "    "
	maxFileBytes        = 512 << 10
	maxFileLines        = 20000
	maxLineRunes        = 600
	maxWorkspaceFiles   = 10000
	maxIgnoredDirs      = 200
	maxIgnoreCandidates = 5000
	maxTreeDepth        = 12
	maxFileBytesKB      = maxFileBytes >> 10
	byteKB              = 1 << 10
	byteMB              = 1 << 20
)

var errBinaryFile = errors.New("binary file")

// fileMeta is one file the explorer found. The viewer reads the bytes only
// when the file opens.
type fileMeta struct {
	Path     string
	Name     string
	Dir      string
	Size     int64
	Lang     string
	Big      bool
	Ignored  bool
	Lines    int
	Encoding string
	Indent   string
}

// fileNode is one directory or file in the tree.
type fileNode struct {
	Path     string
	Name     string
	Dir      bool
	Ignored  bool
	Meta     *fileMeta
	Children []*fileNode
}

// workspace is a scanned directory, either the embedded sample or a real
// directory passed with -root.
type workspace struct {
	Root      string
	Sample    bool
	Files     []*fileMeta
	ByPath    map[string]*fileMeta
	Tree      []*fileNode
	Bytes     int64
	Truncated bool
	// Limit is the file cap the scan ran under, so the UI can name it.
	Limit int
	// TruncatedNote explains a stopped scan. It is empty when the whole
	// workspace was scanned.
	TruncatedNote string
	// ScanMillis is how long the walk took, in milliseconds.
	ScanMillis int64
}

// loadSampleWorkspace scans the embedded demo project.
func loadSampleWorkspace() (*workspace, error) {
	sub, err := fs.Sub(sampleFiles, "testdata/sample")
	if err != nil {
		return nil, fmt.Errorf("sample workspace: %w", err)
	}

	w := &workspace{Root: "sample", Sample: true, ByPath: map[string]*fileMeta{}}
	started := time.Now()

	err = fs.WalkDir(sub, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == "." {
			return nil
		}

		name := entry.Name()
		if entry.IsDir() {
			if skipDir(name) || strings.Count(path, "/") >= maxTreeDepth {
				return fs.SkipDir
			}

			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return nil //nolint:nilerr // one unreadable entry must not hide the workspace
		}

		w.addFile(path, name, info.Size())

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("sample workspace: %w", err)
	}

	w.ScanMillis = time.Since(started).Milliseconds()
	w.finish()

	return w, nil
}

// loadDiskWorkspace scans a real directory for the -root flag.
func loadDiskWorkspace(root string) (*workspace, error) {
	return loadDiskWorkspaceWithLimit(root, maxWorkspaceFiles)
}

// loadDiskWorkspaceWithLimit scans a real directory and stops before the file
// after limit. A limit of zero or less means maxWorkspaceFiles.
func loadDiskWorkspaceWithLimit(root string, limit int) (*workspace, error) {
	if limit <= 0 {
		limit = maxWorkspaceFiles
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("workspace path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("workspace %s: %w", abs, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("workspace %s is not a directory", abs)
	}

	w := &workspace{Root: abs, ByPath: map[string]*fileMeta{}, Limit: limit}
	dirs := make([]string, 0, maxIgnoreCandidates)
	started := time.Now()

	err = filepath.WalkDir(abs, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}

			return nil //nolint:nilerr // keep scanning around one unreadable entry
		}

		if path == abs {
			return nil
		}

		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return nil //nolint:nilerr // a path outside the root is not listed
		}

		rel = filepath.ToSlash(rel)
		name := entry.Name()

		if entry.IsDir() {
			if len(dirs) < maxIgnoreCandidates {
				dirs = append(dirs, rel)
			}

			if skipDir(name) || strings.Count(rel, "/") >= maxTreeDepth {
				return fs.SkipDir
			}

			return nil
		}

		if strings.HasPrefix(name, ".") && name != ".gitignore" && name != ".env.example" {
			return nil
		}

		if skipExt(name) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return nil //nolint:nilerr // one unreadable entry must not hide the workspace
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Stop before adding the file that trips the limit, so a scan that
		// ends exactly at the limit is not reported as truncated.
		if len(w.Files) >= limit {
			w.Truncated = true
			w.TruncatedNote = truncatedNote(limit)

			return fs.SkipAll
		}

		w.addFile(rel, name, info.Size())

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("workspace %s: %w", abs, err)
	}

	w.ScanMillis = time.Since(started).Milliseconds()
	w.finish()

	ignored := loadIgnoredPaths(abs)
	for path := range loadIgnoredDirs(abs, dirs) {
		if ignored == nil {
			ignored = map[string]bool{}
		}

		ignored[path] = true
	}

	if len(ignored) > 0 {
		w.markIgnored(ignored)
	}

	return w, nil
}

// loadIgnoredDirs asks git which scanned directories match .gitignore. One
// check-ignore call covers skipped folders, empty folders, and folders whose
// files git does not list because they have no content.
func loadIgnoredDirs(root string, dirs []string) map[string]bool {
	if len(dirs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	input := strings.Join(dirs, "\x00") + "\x00"
	raw, err := gitBytesStdin(ctx, root, input, "check-ignore", "--no-index", "--stdin", "-z")
	if err != nil {
		return nil
	}

	ignored := map[string]bool{}
	for _, path := range strings.Split(string(raw), "\x00") {
		path = strings.TrimSpace(path)
		path = strings.TrimSuffix(path, "/")
		if path != "" {
			ignored[path] = true
		}
	}

	return ignored
}

// loadIgnoredPaths reads the paths git says are ignored. A directory entry is
// returned without its trailing slash, so a parent match covers its children.
// A non-repository or a failed git call yields nil and the tree stays as it is.
func loadIgnoredPaths(root string) map[string]bool {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	raw, err := gitBytes(ctx, root, "ls-files", "--others", "--ignored", "--exclude-standard", "--directory", "-z")
	if err != nil {
		return nil
	}

	ignored := map[string]bool{}

	for _, path := range strings.Split(string(raw), "\x00") {
		path = strings.TrimSpace(path)
		path = strings.TrimSuffix(path, "/")
		if path != "" {
			ignored[path] = true
		}
	}

	return ignored
}

// markIgnored sets the ignored flag on file metadata and every tree node whose
// path is listed or sits under a listed directory. Skipped directories that
// git ignores are added as dim, empty rows so they stay visible.
func (w *workspace) markIgnored(ignored map[string]bool) {
	for _, meta := range w.Files {
		meta.Ignored = pathIgnored(ignored, meta.Path)
	}

	markTreeIgnored(w.Tree, ignored)
	w.addIgnoredDirs(ignored)
}

func markTreeIgnored(nodes []*fileNode, ignored map[string]bool) {
	for _, node := range nodes {
		node.Ignored = pathIgnored(ignored, node.Path)
		markTreeIgnored(node.Children, ignored)
	}
}

// addIgnoredDirs adds an empty row for an ignored directory the scanner
// skipped, such as node_modules. The row carries no children, so the scan
// budget is not spent on a tree the person did not open.
func (w *workspace) addIgnoredDirs(ignored map[string]bool) {
	if w.Sample || len(ignored) == 0 {
		return
	}

	paths := make([]string, 0, len(ignored))

	for path := range ignored {
		paths = append(paths, path)
	}

	sort.Slice(paths, func(i, j int) bool {
		di := strings.Count(paths[i], "/")
		dj := strings.Count(paths[j], "/")

		if di != dj {
			return di < dj
		}

		return paths[i] < paths[j]
	})

	existing := treeNodePaths(w.Tree)
	root := &fileNode{Dir: true, Children: w.Tree}
	added := 0

	for _, path := range paths {
		if added >= maxIgnoredDirs {
			break
		}

		if existing[path] {
			continue
		}

		info, err := os.Stat(filepath.Join(w.Root, filepath.FromSlash(path)))
		if err != nil || !info.IsDir() {
			continue
		}

		parent := ensureTreeNode(root, dirName(path), existing, ignored)
		if parent == nil {
			continue
		}

		node := &fileNode{Path: path, Name: baseName(path), Dir: true, Ignored: true}
		parent.Children = append(parent.Children, node)
		existing[path] = true
		added++
	}

	if added > 0 {
		sortTree(root.Children)
		w.Tree = root.Children
	}
}

// ensureTreeNode finds or creates a directory below root.
func ensureTreeNode(root *fileNode, path string, existing map[string]bool, ignored map[string]bool) *fileNode {
	if path == "" || path == "." {
		return root
	}

	if node := findTreeNode(root.Children, path); node != nil {
		return node
	}

	parent := ensureTreeNode(root, dirName(path), existing, ignored)
	if parent == nil {
		return nil
	}

	node := &fileNode{Path: path, Name: baseName(path), Dir: true, Ignored: ignored[path]}
	parent.Children = append(parent.Children, node)
	existing[path] = true

	return node
}

func findTreeNode(nodes []*fileNode, path string) *fileNode {
	for _, node := range nodes {
		if node.Path == path {
			return node
		}

		if node.Dir && strings.HasPrefix(path, node.Path+"/") {
			if found := findTreeNode(node.Children, path); found != nil {
				return found
			}
		}
	}

	return nil
}

func treeNodePaths(nodes []*fileNode) map[string]bool {
	out := map[string]bool{}
	var walk func([]*fileNode)

	walk = func(list []*fileNode) {
		for _, node := range list {
			out[node.Path] = true
			walk(node.Children)
		}
	}

	walk(nodes)

	return out
}

// pathIgnored walks the parent chain so a listed ignored directory covers all
// of its children without a prefix scan over every pattern.
func pathIgnored(ignored map[string]bool, path string) bool {
	if ignored == nil || path == "" {
		return false
	}

	for {
		if ignored[path] {
			return true
		}

		i := strings.LastIndexByte(path, '/')
		if i < 0 {
			return false
		}

		path = path[:i]
	}
}

// truncatedNote is the explanation shown when the scan stopped at its file
// limit. The wording is stable so the UI and tests read the same sentence.
func truncatedNote(limit int) string {
	return fmt.Sprintf(
		"Scan stopped at the limit of %d files. The explorer, quick open, and search only show the first %d files; raise the limit to scan the whole repository.",
		limit,
		limit,
	)
}

// ScanBudget is the short budget sentence for the explorer, or "" when the
// scan saw the whole workspace.
func (w *workspace) ScanBudget() string {
	if !w.Truncated {
		return ""
	}

	limit := w.Limit
	if limit <= 0 {
		limit = maxWorkspaceFiles
	}

	return fmt.Sprintf("Showing the first %d files...", limit)
}

// addFile appends one scanned file.
func (w *workspace) addFile(path, name string, size int64) {
	dir := ""
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		dir = path[:i]
	}

	meta := &fileMeta{
		Path: path,
		Name: name,
		Dir:  dir,
		Size: size,
		Lang: languageOf(path),
		Big:  size > maxFileBytes,
	}

	w.Files = append(w.Files, meta)
	w.ByPath[path] = meta
	w.Bytes += size
}

// finish sorts the files, builds the tree, and counts lines only for files
// small enough to open.
func (w *workspace) finish() {
	sort.Slice(w.Files, func(i, j int) bool {
		return strings.ToLower(w.Files[i].Path) < strings.ToLower(w.Files[j].Path)
	})

	root := &fileNode{Dir: true}

	for _, meta := range w.Files {
		parts := strings.Split(meta.Path, "/")
		cur := root

		for i, part := range parts {
			child := findChild(cur.Children, part)
			if child == nil {
				path := strings.Join(parts[:i+1], "/")
				child = &fileNode{Path: path, Name: part, Dir: i < len(parts)-1}
				cur.Children = append(cur.Children, child)
			}

			if i == len(parts)-1 {
				child.Meta = meta
			}

			cur = child
		}
	}

	sortTree(root.Children)
	w.Tree = root.Children
}

// findChild returns the child with a name.
func findChild(children []*fileNode, name string) *fileNode {
	for _, child := range children {
		if child.Name == name {
			return child
		}
	}

	return nil
}

// sortTree puts directories before files and sorts each group by name.
func sortTree(nodes []*fileNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Dir != nodes[j].Dir {
			return nodes[i].Dir
		}

		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})

	for _, node := range nodes {
		sortTree(node.Children)
	}
}

// skipDir reports whether a directory is outside the code view.
func skipDir(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", ".cache", ".idea", ".next", ".nuxt", ".venv",
		"__pycache__", "build", "coverage", "dist", "node_modules", "target",
		"temp", "tmp", "vendor":
		return true
	}

	return strings.HasPrefix(name, ".") && name != ".github"
}

// skipExt reports whether a file is binary or an archive.
func skipExt(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".7z", ".a", ".apk", ".avi", ".bmp", ".bz2", ".class", ".db", ".dll",
		".dylib", ".exe", ".gif", ".gz", ".ico", ".jar", ".jpeg", ".jpg",
		".m4a", ".mkv", ".mov", ".mp3", ".mp4", ".o", ".ogg", ".otf", ".pdf",
		".png", ".so", ".sqlite", ".tar", ".tgz", ".ttf", ".wav", ".webm",
		".webp", ".woff", ".woff2", ".xz", ".zip":
		return true
	}

	return false
}

// readLines reads one file as display lines. Tabs become spaces, a trailing
// CR is stripped, and an overlong line is clipped with an ellipsis.
func (w *workspace) readLines(path string) ([]string, bool, error) {
	meta := w.ByPath[path]
	if meta == nil {
		return nil, false, fmt.Errorf("no file %s", path)
	}

	if meta.Big {
		return nil, false, fmt.Errorf("%s is larger than %d KB", path, maxFileBytesKB)
	}

	raw, err := w.readFile(path)
	if err != nil {
		return nil, false, err
	}

	if isBinary(raw) {
		return nil, false, errBinaryFile
	}

	text := string(raw)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")

	switch {
	case strings.HasPrefix(text, "\ufeff"):
		meta.Encoding = "UTF-8 BOM"
		text = strings.TrimPrefix(text, "\ufeff")
	default:
		meta.Encoding = "UTF-8"
	}

	var lines []string
	if text != "" {
		lines = strings.Split(text, "\n")
	}

	meta.Indent = detectIndent(lines)
	truncated := false

	if len(lines) > maxFileLines {
		lines = lines[:maxFileLines]
		truncated = true
	}

	for i, line := range lines {
		lines[i] = clipLine(expandTabs(line))
	}

	meta.Lines = len(lines)

	return lines, truncated, nil
}

// readFile returns the bytes of one scanned file.
func (w *workspace) readFile(path string) ([]byte, error) {
	if w.Sample {
		sub, err := fs.Sub(sampleFiles, "testdata/sample")
		if err != nil {
			return nil, err
		}

		return fs.ReadFile(sub, path)
	}

	clean := filepath.Clean(filepath.FromSlash(path))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path escapes workspace: %s", path)
	}

	return os.ReadFile(filepath.Join(w.Root, clean)) //nolint:gosec // path is checked against the scanned root above
}

// detectIndent reports the indentation the file actually uses.
func detectIndent(lines []string) string {
	tabs := false
	smallest := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "\t") {
			tabs = true
		}

		n := len(line) - len(strings.TrimLeft(line, " "))
		if n > 0 && (smallest == 0 || n < smallest) {
			smallest = n
		}
	}

	if tabs {
		return "Tabs"
	}

	if smallest > 0 {
		return fmt.Sprintf("Spaces: %d", smallest)
	}

	return "None"
}

// isBinary reports whether the first block contains a NUL byte.
func isBinary(raw []byte) bool {
	const probe = 8192

	if len(raw) > probe {
		raw = raw[:probe]
	}

	for _, b := range raw {
		if b == 0 {
			return true
		}
	}

	return false
}

// clipLine caps one display line so a minified file cannot dominate layout.
func clipLine(line string) string {
	if len(line) <= maxLineRunes {
		return line
	}

	r := []rune(line)
	if len(r) <= maxLineRunes {
		return line
	}

	return string(r[:maxLineRunes]) + "…"
}

// fileDoc is one opened file: its lines, tokens, and outline.
type fileDoc struct {
	Path        string
	Lang        string
	ByteSize    int64
	Lines       []string
	Tokens      [][]segment
	Outline     []symbolRow
	Preview     []previewRow
	PreviewLine []int
	Truncated   bool
	Binary      bool
	Err         string
	Encoding    string
	Indent      string
	Wrap        *wrapPlan
}

// openDoc reads and tokenizes one file.
func (w *workspace) openDoc(path string) *fileDoc {
	meta := w.ByPath[path]
	if meta == nil {
		return &fileDoc{Path: path, Err: "This file is not in the workspace."}
	}

	if !w.Sample {
		if info, statErr := os.Stat(filepath.Join(w.Root, filepath.FromSlash(path))); statErr == nil {
			meta.Size = info.Size()
			meta.Big = meta.Size > maxFileBytes
		}
	}

	doc := &fileDoc{Path: path, Lang: meta.Lang, ByteSize: meta.Size}

	lines, truncated, err := w.readLines(path)
	if err != nil {
		doc.Err = err.Error()
		if errors.Is(err, errBinaryFile) {
			doc.Binary = true
		}

		return doc
	}

	doc.Lines = lines
	doc.Truncated = truncated
	doc.Encoding = meta.Encoding
	doc.Indent = meta.Indent
	doc.Tokens = tokenize(doc.Lang, lines)
	doc.Outline = outline(doc.Lang, lines)

	return doc
}

// baseName returns the last path element.
func baseName(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}

	return path
}

// dirName returns everything before the last slash.
func dirName(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[:i]
	}

	return ""
}

// humanSize formats a byte count for the details panel.
func humanSize(size int64) string {
	switch {
	case size >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(size)/byteMB)
	case size >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(size)/byteKB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
