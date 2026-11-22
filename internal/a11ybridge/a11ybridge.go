// Package a11ybridge is the hand-off between a page's accessibility tree and
// a host that lives outside Go, such as the Android view that answers
// TalkBack.
//
// The window loop publishes an encoded snapshot whenever the page changes and
// installs a performer that runs actions on the page loop. The host reads the
// latest snapshot and requests actions from its own threads. Neither side
// touches the page directly, because the tree and the actions belong to the
// page's owning loop.
package a11ybridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Action numbers are the host.AccessibilityActionKind values.
const (
	ActionFocus    = int(host.AccessibilityFocus)
	ActionActivate = int(host.AccessibilityActivate)
	ActionSetValue = int(host.AccessibilitySetValue)
)

// performTimeout bounds one action so a stuck page cannot hold a binder
// thread: TalkBack gives up on its own after a few seconds.
const performTimeout = 3 * time.Second

const emptyTree = `{"nodes":[]}`

var (
	errUnknownAction = errors.New("unknown accessibility action")
	errNoPerformer   = errors.New("no page is attached")
)

// Performer runs one request on the page loop. atspi.Actions.Request has this
// shape.
type Performer func(context.Context, host.AccessibilityRequest) error

var (
	latest    atomic.Pointer[string]
	performer atomic.Pointer[Performer]
)

// SetTree publishes the encoded snapshot the host will read.
func SetTree(encoded string) { latest.Store(&encoded) }

// Tree returns the latest published snapshot, or an empty tree before the
// first one, so a host can always parse the result.
func Tree() string {
	if s := latest.Load(); s != nil {
		return *s
	}

	return emptyTree
}

// SetPerformer installs the function that runs actions on the page loop. A
// nil performer detaches the page.
func SetPerformer(p Performer) {
	if p == nil {
		performer.Store(nil)

		return
	}

	performer.Store(&p)
}

// Reset forgets the published tree and detaches the performer.
func Reset() {
	latest.Store(nil)
	performer.Store(nil)
}

// Perform runs one action a host asked for. document is the token from the
// tree the host read, so an action aimed at a node of a page that has since
// navigated is refused instead of landing on a different node. It returns an
// empty string on success and a short message otherwise.
func Perform(document int64, node string, action int, value string) string {
	kind, err := actionKind(action)
	if err != nil {
		return err.Error()
	}

	p := performer.Load()
	if p == nil {
		return errNoPerformer.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), performTimeout)
	defer cancel()

	err = (*p)(ctx, host.AccessibilityRequest{
		Document: uint64(document), //nolint:gosec // reason: the token came from Tree as a non-negative counter.
		Node:     host.AccessibilityNodeID(node),
		Kind:     kind,
		Value:    value,
	})
	if err != nil {
		return err.Error()
	}

	return ""
}

func actionKind(action int) (host.AccessibilityActionKind, error) {
	switch action {
	case ActionFocus:
		return host.AccessibilityFocus, nil
	case ActionActivate:
		return host.AccessibilityActivate, nil
	case ActionSetValue:
		return host.AccessibilitySetValue, nil
	}

	return 0, fmt.Errorf("%w: %d", errUnknownAction, action)
}

type jsonNode struct {
	ID         string   `json:"id"`
	Parent     string   `json:"parent,omitempty"`
	Children   []string `json:"children,omitempty"`
	ElementID  string   `json:"element,omitempty"`
	Role       string   `json:"role"`
	Name       string   `json:"name,omitempty"`
	Desc       string   `json:"desc,omitempty"`
	Value      string   `json:"value,omitempty"`
	X          float64  `json:"x"`
	Y          float64  `json:"y"`
	W          float64  `json:"w"`
	H          float64  `json:"h"`
	HasBounds  bool     `json:"bounds"`
	Disabled   bool     `json:"disabled,omitempty"`
	Focusable  bool     `json:"focusable,omitempty"`
	Focused    bool     `json:"focused,omitempty"`
	Password   bool     `json:"password,omitempty"`
	HasChecked bool     `json:"hasChecked,omitempty"`
	Checked    bool     `json:"checked,omitempty"`
	Actions    []int    `json:"actions,omitempty"`
}

type jsonTree struct {
	Document int64      `json:"doc"`
	Revision int64      `json:"rev"`
	Root     string     `json:"root"`
	Focused  string     `json:"focus,omitempty"`
	Nodes    []jsonNode `json:"nodes"`
}

// Encode returns the snapshot as JSON, in document order, with bounds in CSS
// pixels of the page content. Password values are never included.
func Encode(tree host.AccessibilityTree) string {
	out := jsonTree{
		Document: int64(tree.Document), //nolint:gosec // reason: a document token is a small counter.
		Revision: int64(tree.Revision), //nolint:gosec // reason: a revision is a small counter.
		Root:     string(tree.Root),
		Focused:  string(tree.Focused),
		Nodes:    make([]jsonNode, 0, len(tree.Nodes)),
	}

	for i := range tree.Nodes {
		out.Nodes = append(out.Nodes, encodeNode(&tree.Nodes[i]))
	}

	data, err := json.Marshal(out)
	if err != nil {
		return emptyTree
	}

	return string(data)
}

func encodeNode(n *host.AccessibilityNode) jsonNode {
	node := jsonNode{
		ID:         string(n.ID),
		Parent:     string(n.Parent),
		ElementID:  n.ElementID,
		Role:       string(n.Role),
		Name:       n.Name,
		Desc:       n.Description,
		Value:      n.Value,
		X:          n.Bounds.X,
		Y:          n.Bounds.Y,
		W:          n.Bounds.W,
		H:          n.Bounds.H,
		HasBounds:  n.HasBounds,
		Disabled:   n.Disabled,
		Focusable:  n.Focusable,
		Focused:    n.Focused,
		Password:   n.Password,
		HasChecked: n.HasChecked,
		Checked:    n.Checked,
	}

	if n.Password {
		node.Value = ""
	}

	for _, child := range n.Children {
		node.Children = append(node.Children, string(child))
	}

	for _, action := range n.Actions {
		node.Actions = append(node.Actions, int(action))
	}

	return node
}
