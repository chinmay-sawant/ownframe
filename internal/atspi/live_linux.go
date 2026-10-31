//go:build linux && !android && !ios && !js

package atspi

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/safego"
)

// pendingReply is one outstanding method call. sender is the bus name the
// reply must come from, so a peer cannot answer another process's call by
// guessing its serial.
type pendingReply struct {
	ch     chan *dbusMsg
	sender string
}

type liveBus struct {
	mu         sync.Mutex
	conn       net.Conn
	reader     *bufio.Reader
	writeMu    sync.Mutex
	connectMu  sync.Mutex
	pending    map[uint32]pendingReply
	done       chan struct{}
	connDone   chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
	name       string
	uniqueName string
	addr       string
	claimed    bool
	closed     bool
	reason     string
	serial     uint32
	app        Application
	view       *serverView
	originX    int
	originY    int
	parent     dbusRef
	appID      int32
	lastTry    time.Time
	failAt     time.Time
	embedded   bool
	width      int
	height     int
}

const (
	liveDialTimeout   = 2 * time.Second
	liveCallTimeout   = 2 * time.Second
	liveEmitTimeout   = 500 * time.Millisecond
	liveRetryInterval = 2 * time.Second

	dbusDaemonName   = "org.freedesktop.DBus"
	a11yBusName      = "org.a11y.Bus"
	a11yRegistryName = "org.a11y.atspi.Registry"

	serialHello        = 1
	serialStartService = 2
	serialGetNameOwner = 3
	serialGetAddress   = 4

	authLineLimit = 4096
)

func newLiveBus() Bus {
	return &liveBus{
		pending: make(map[uint32]pendingReply),
		done:    make(chan struct{}),
		reason:  "not connected",
	}
}

// Ready reports whether Publish should attempt to connect or publish.
// A dead bus is retried at most once per liveRetryInterval.
func (b *liveBus) Ready() bool {
	if b == nil || disabledByEnvironment() {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return false
	}
	if !b.failAt.IsZero() && time.Since(b.failAt) < liveRetryInterval {
		return false
	}
	if b.conn != nil && b.claimed {
		return true
	}
	if b.lastTry.IsZero() {
		return true
	}
	return time.Since(b.lastTry) >= liveRetryInterval
}

// refNameLocked is the bus name used in object references. Callers hold b.mu.
func (b *liveBus) refNameLocked() string {
	if b.uniqueName != "" {
		return b.uniqueName
	}
	return b.name
}

func (b *liveBus) ensure() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	if b.conn != nil {
		b.mu.Unlock()
		return nil
	}
	if time.Since(b.lastTry) < liveRetryInterval {
		reason := b.reason
		b.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrLiveUnavailable, reason)
	}
	b.lastTry = time.Now()
	b.mu.Unlock()
	return b.connect()
}

func (b *liveBus) connect() error {
	b.connectMu.Lock()
	defer b.connectMu.Unlock()

	b.mu.Lock()
	if b.conn != nil {
		b.mu.Unlock()
		return nil
	}
	if b.closed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	b.mu.Unlock()

	netw, addr, guid, err := discoverATSPISocket()
	if err != nil {
		b.setReason(err.Error())
		return fmt.Errorf("%w: %w", ErrLiveUnavailable, err)
	}

	conn, err := (&net.Dialer{Timeout: liveDialTimeout}).DialContext(context.Background(), netw, addr)
	if err != nil {
		b.setReason(fmt.Sprintf("connect to %s failed: %v", addr, err))
		return fmt.Errorf("%w: connect %s: %w", ErrLiveUnavailable, addr, err)
	}

	reader, unique, err := greet(conn, guid)
	if err != nil {
		_ = conn.Close()
		b.setReason(err.Error())
		return fmt.Errorf("%w: %w", ErrLiveUnavailable, err)
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		_ = conn.Close()
		return ErrLiveUnavailable
	}
	b.conn = conn
	b.connDone = make(chan struct{})
	b.reader = reader
	b.uniqueName = unique
	b.addr = addr
	b.serial = 1
	b.reason = "connected to " + addr
	b.embedded = false
	b.wg.Add(1)
	b.mu.Unlock()

	safego.Go("atspi reader", func() { b.readLoop(conn, reader) })
	return nil
}

func (b *liveBus) setReason(reason string) {
	b.mu.Lock()
	b.reason = reason
	b.mu.Unlock()
}

func handshake(conn net.Conn, expectedGUID string) (*bufio.Reader, error) {
	_ = conn.SetDeadline(time.Now().Add(liveCallTimeout))
	defer func() { _ = conn.SetDeadline(time.Time{}) }()

	if _, err := conn.Write([]byte{0}); err != nil {
		return nil, fmt.Errorf("write null byte: %w", err)
	}

	uidStr := strconv.Itoa(os.Getuid())
	hexUID := hex.EncodeToString([]byte(uidStr))
	authMsg := fmt.Sprintf("AUTH EXTERNAL %s\r\n", hexUID)
	if _, err := conn.Write([]byte(authMsg)); err != nil {
		return nil, fmt.Errorf("write auth: %w", err)
	}

	reader := bufio.NewReaderSize(conn, authLineLimit)
	line, err := reader.ReadSlice('\n')
	if err != nil {
		if errors.Is(err, bufio.ErrBufferFull) {
			return nil, fmt.Errorf("auth response line exceeds %d bytes", authLineLimit)
		}
		return nil, fmt.Errorf("read auth response: %w", err)
	}
	authLine := strings.TrimRight(string(line), "\r\n")
	fields := strings.Fields(authLine)
	if len(fields) == 0 || fields[0] != "OK" {
		return nil, fmt.Errorf("auth rejected: %s", authLine)
	}
	authGUID := ""
	if len(fields) > 1 {
		authGUID = fields[1]
	}
	if expectedGUID != "" && authGUID != "" && !strings.EqualFold(authGUID, expectedGUID) {
		return nil, fmt.Errorf("auth guid mismatch: server sent %s, address expects %s", authGUID, expectedGUID)
	}

	if _, err := conn.Write([]byte("BEGIN\r\n")); err != nil {
		return nil, fmt.Errorf("write begin: %w", err)
	}
	return reader, nil
}

// greet performs the auth handshake and Hello, returning the buffered
// reader and the unique bus name. It reads only the matching Hello
// reply and leaves any queued signal for the read loop. guid, when set,
// must match the GUID the server reports in its auth response.
func greet(conn net.Conn, guid string) (*bufio.Reader, string, error) {
	reader, err := handshake(conn, guid)
	if err != nil {
		return nil, "", err
	}
	hello := &dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  serialHello,
		path:    "/org/freedesktop/DBus",
		dest:    dbusDaemonName,
		iface:   dbusDaemonName,
		member:  "Hello",
	}
	reply, err := syncCall(conn, reader, hello, dbusDaemonName)
	if err != nil {
		return nil, "", fmt.Errorf("hello failed: %w", err)
	}
	name, err := decodeStringBody(reply.body)
	if err != nil {
		return nil, "", fmt.Errorf("decode hello: %w", err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "", errors.New("hello returned an empty unique name")
	}
	return reader, name, nil
}

// syncCall sends call and waits for the reply from sender. Replies from
// any other peer, including a forged reply with the same serial, are
// ignored.
func syncCall(conn net.Conn, reader *bufio.Reader, call *dbusMsg, sender string) (*dbusMsg, error) {
	_ = conn.SetDeadline(time.Now().Add(liveCallTimeout))
	defer func() { _ = conn.SetDeadline(time.Time{}) }()
	if _, err := conn.Write(encodeMessage(call)); err != nil {
		return nil, err
	}
	for {
		reply, err := readMessage(reader)
		if err != nil {
			return nil, err
		}
		if reply.msgType == dbusTypeSignal {
			continue
		}
		if reply.replySerial != call.serial || reply.sender != sender {
			continue
		}
		if reply.msgType == dbusTypeError {
			return nil, fmt.Errorf("%s: %s", reply.errorName, decodeStringOrEmpty(reply.body))
		}
		if reply.msgType != dbusTypeMethodReturn {
			return nil, fmt.Errorf("unexpected reply type %d", reply.msgType)
		}
		return reply, nil
	}
}

func decodeStringOrEmpty(body []byte) string {
	s, _ := decodeStringBody(body)
	return s
}

func (b *liveBus) readLoop(conn net.Conn, reader *bufio.Reader) {
	defer b.wg.Done()
	// A panic in the reader must not leave the bus claimed without a
	// reader: disconnect releases the name and unblocks pending calls.
	defer b.disconnect(conn, "bus reader stopped")
	for {
		msg, err := readMessage(reader)
		if err != nil {
			b.disconnect(conn, fmt.Sprintf("bus read failed: %v", err))
			return
		}
		switch msg.msgType {
		case dbusTypeMethodReturn, dbusTypeError:
			b.deliver(msg)
		case dbusTypeMethodCall:
			b.handleCall(msg)
		}
	}
}

func (b *liveBus) deliver(msg *dbusMsg) {
	b.mu.Lock()
	pending, ok := b.pending[msg.replySerial]
	if !ok || pending.sender != msg.sender {
		// Leave the real call pending so a forged reply cannot consume it.
		b.mu.Unlock()
		return
	}
	delete(b.pending, msg.replySerial)
	b.mu.Unlock()
	select {
	case pending.ch <- msg:
	default:
	}
}

func (b *liveBus) handleCall(msg *dbusMsg) {
	// A caller that set NO_REPLY_EXPECTED does not want the result, so
	// skip the read handlers and any reply. The registry's one-way
	// Application.Id write is the exception handled below.
	if msg.flags&dbusFlagNoReplyExpected != 0 {
		b.handleNoReplyCall(msg)
		return
	}
	b.mu.Lock()
	view := b.view
	b.mu.Unlock()
	var res callResult
	if view == nil {
		res = errResult("org.freedesktop.DBus.Error.UnknownObject", "application has no published tree")
	} else {
		res = view.handle(msg)
	}
	b.reply(msg, res)
}

// handleNoReplyCall performs only the cheap Application.Id write that
// the registry sends without waiting for a reply. It applies the same
// sender guard as serve.go; every other one-way call is dropped before
// it can touch a handler or encode a result.
func (b *liveBus) handleNoReplyCall(msg *dbusMsg) {
	if msg.iface != dbusProperties || msg.member != "Set" {
		return
	}
	b.mu.Lock()
	view := b.view
	b.mu.Unlock()
	if view == nil || view.appID == nil || msg.path != view.app.RootPath {
		return
	}
	if msg.sender == "" || msg.sender != view.parent.Name {
		return
	}
	d := dbusDecoder{body: msg.body}
	iface, err := d.stringVal()
	if err != nil || iface != atspiApplication {
		return
	}
	name, err := d.stringVal()
	if err != nil || name != "Id" {
		return
	}
	sig, val, err := d.variant()
	if err != nil || sig != "i" {
		return
	}
	id, err := val.int32Val()
	if err != nil {
		return
	}
	b.mu.Lock()
	if b.view == view {
		*view.appID = id
	}
	b.mu.Unlock()
}

func (b *liveBus) reply(call *dbusMsg, res callResult) {
	msg := &dbusMsg{
		msgType:     dbusTypeMethodReturn,
		replySerial: call.serial,
		dest:        call.sender,
		path:        call.path,
		sig:         res.sig,
		body:        res.body,
	}
	if res.errName != "" {
		msg.msgType = dbusTypeError
		msg.errorName = res.errName
		msg.body = encodeErrorBody(res.errText)
		msg.sig = "s"
	}
	if err := b.send(msg, liveCallTimeout); err != nil {
		b.disconnect(b.connSnapshot(), "reply failed: "+err.Error())
	}
}

func (b *liveBus) call(msg *dbusMsg, sender string, timeout time.Duration) (*dbusMsg, error) {
	b.mu.Lock()
	if b.conn == nil || b.closed {
		b.mu.Unlock()
		return nil, ErrLiveUnavailable
	}
	b.serial++
	msg.serial = b.serial
	ch := make(chan *dbusMsg, 1)
	b.pending[msg.serial] = pendingReply{ch: ch, sender: sender}
	connDone := b.connDone
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.pending, msg.serial)
		b.mu.Unlock()
	}()

	if err := b.writeMessage(msg, timeout); err != nil {
		return nil, err
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case reply := <-ch:
		if reply.msgType == dbusTypeError {
			return nil, fmt.Errorf("%s: %s", reply.errorName, decodeStringOrEmpty(reply.body))
		}
		return reply, nil
	case <-timer.C:
		return nil, fmt.Errorf("timed out waiting for %s", msg.member)
	case <-connDone:
		return nil, ErrLiveUnavailable
	case <-b.done:
		return nil, ErrLiveUnavailable
	}
}

func (b *liveBus) send(msg *dbusMsg, timeout time.Duration) error {
	b.mu.Lock()
	if b.conn == nil || b.closed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	b.serial++
	msg.serial = b.serial
	b.mu.Unlock()
	return b.writeMessage(msg, timeout)
}

func (b *liveBus) writeMessage(msg *dbusMsg, timeout time.Duration) error {
	b.mu.Lock()
	conn := b.conn
	b.mu.Unlock()
	if conn == nil {
		return ErrLiveUnavailable
	}
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err := conn.Write(encodeMessage(msg))
	_ = conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}
	return nil
}

func (b *liveBus) connSnapshot() net.Conn {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.conn
}

func (b *liveBus) disconnect(conn net.Conn, reason string) {
	b.mu.Lock()
	if b.closed || b.conn != conn {
		b.mu.Unlock()
		return
	}
	done := b.connDone
	b.conn = nil
	b.reader = nil
	b.uniqueName = ""
	b.claimed = false
	b.embedded = false
	b.parent = dbusRef{}
	b.reason = reason
	b.connDone = nil
	b.mu.Unlock()
	if done != nil {
		close(done)
	}
	if conn != nil {
		_ = conn.Close()
	}
}

func (b *liveBus) Live() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.conn != nil && !b.closed && b.claimed
}

func (b *liveBus) Status() Status {
	if b == nil {
		return Status{Platform: runtime.GOOS, Live: false, Reason: "no bus"}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	live := b.conn != nil && !b.closed && b.claimed
	reason := b.reason
	if b.closed {
		live = false
	}
	return Status{Platform: runtime.GOOS, Live: live, Reason: reason}
}

func (b *liveBus) Claim(name string) error {
	if err := CheckName(name); err != nil {
		return err
	}
	if err := b.ensure(); err != nil {
		return err
	}
	b.mu.Lock()
	if b.closed || b.conn == nil {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	if b.claimed && b.name == name {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()
	return b.requestName(name)
}

func (b *liveBus) requestName(name string) error {
	code, err := b.requestNameCode(name)
	if err != nil {
		b.setReason("claim failed: " + err.Error())
		return fmt.Errorf("claim failed: %w", err)
	}
	if code == 1 || code == 4 {
		b.claimedAs(name)
		b.mu.Lock()
		b.failAt = time.Time{}
		b.mu.Unlock()
		return nil
	}
	if name == DefaultName {
		fallback, ferr := randomFallbackName()
		if ferr != nil {
			b.mu.Lock()
			b.failAt = time.Now()
			b.mu.Unlock()
			b.setReason("claim failed: " + ferr.Error())
			return fmt.Errorf("claim failed: %w", ferr)
		}
		if fcode, ferr := b.requestNameCode(fallback); ferr == nil && (fcode == 1 || fcode == 4) {
			b.claimedAs(fallback)
			b.mu.Lock()
			b.failAt = time.Time{}
			b.mu.Unlock()
			return nil
		}
	}
	b.mu.Lock()
	b.failAt = time.Now()
	b.mu.Unlock()
	b.setReason(fmt.Sprintf("claim failed: name %s returned code %d", name, code))
	return fmt.Errorf("claim failed: name %s returned code %d", name, code)
}

func (b *liveBus) requestNameCode(name string) (uint32, error) {
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		path:    "/org/freedesktop/DBus",
		dest:    dbusDaemonName,
		iface:   dbusDaemonName,
		member:  "RequestName",
		sig:     "su",
		body:    encodeRequestNameBody(name, 0),
	}
	reply, err := b.call(msg, dbusDaemonName, liveCallTimeout)
	if err != nil {
		return 0, err
	}
	return decodeUint32Body(reply.body)
}

func startServiceCall(name string) *dbusMsg {
	var e dbusEncoder
	e.stringVal(name)
	e.uint32Val(0)
	return &dbusMsg{
		msgType: dbusTypeMethodCall,
		path:    "/org/freedesktop/DBus",
		dest:    dbusDaemonName,
		iface:   dbusDaemonName,
		member:  "StartServiceByName",
		sig:     "su",
		body:    e.bytes(),
	}
}

func getNameOwnerCall(name string) *dbusMsg {
	return &dbusMsg{
		msgType: dbusTypeMethodCall,
		path:    "/org/freedesktop/DBus",
		dest:    dbusDaemonName,
		iface:   dbusDaemonName,
		member:  "GetNameOwner",
		sig:     "s",
		body:    encodeStringBody(name),
	}
}

// nameOwner resolves the unique name that owns name on the connected
// bus. Only org.freedesktop.DBus may answer the lookup itself.
func (b *liveBus) nameOwner(name string) (string, error) {
	// Match the activation a direct call to the well-known name would
	// have caused. A failed activation is fine when the name already
	// has an owner; GetNameOwner below decides that.
	_, _ = b.call(startServiceCall(name), dbusDaemonName, liveCallTimeout)
	msg := getNameOwnerCall(name)
	reply, err := b.call(msg, dbusDaemonName, liveCallTimeout)
	if err != nil {
		return "", err
	}
	owner, err := decodeStringBody(reply.body)
	if err != nil {
		return "", fmt.Errorf("decode name owner: %w", err)
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return "", fmt.Errorf("name %s has no owner", name)
	}
	return owner, nil
}

// randomFallbackName returns a name a squatter cannot predict from the
// process id. The suffix is 16 hex characters appended to DefaultName,
// keeping the single-element shape CheckName accepts.
func randomFallbackName() (string, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("random fallback name: %w", err)
	}
	name := DefaultName + hex.EncodeToString(raw[:])
	if err := CheckName(name); err != nil {
		return "", err
	}
	return name, nil
}

func (b *liveBus) claimedAs(name string) {
	b.mu.Lock()
	b.name = name
	b.claimed = true
	b.reason = "connected to " + b.addr + " as " + name
	b.mu.Unlock()
}

func (b *liveBus) ensureClaimed() error {
	if err := b.ensure(); err != nil {
		return err
	}
	b.mu.Lock()
	claimed := b.claimed
	name := b.name
	b.mu.Unlock()
	if claimed && name != "" {
		return nil
	}
	return b.Claim(DefaultName)
}

// SetOrigin records the window content origin used for screen coordinates.
func (b *liveBus) SetOrigin(x, y int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed || (b.originX == x && b.originY == y) {
		return
	}
	b.originX, b.originY = x, y
	if b.app.RootPath != "" {
		b.view = newServerView(b.app, b.refNameLocked(), b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	}
}

// SetSize records the window content size used for the root extents.
func (b *liveBus) SetSize(width, height int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed || (b.width == width && b.height == height) {
		return
	}
	b.width, b.height = width, height
	if b.app.RootPath != "" {
		b.view = newServerView(b.app, b.refNameLocked(), b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	}
}

func (b *liveBus) Publish(app Application) error {
	if err := b.ensureClaimed(); err != nil {
		return err
	}
	b.mu.Lock()
	if b.conn == nil || b.closed || !b.claimed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	refName := b.refNameLocked()
	b.app = app
	b.view = newServerView(app, refName, b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	embedded := b.embedded && b.parent.Path != ""
	b.mu.Unlock()
	if embedded {
		return nil
	}

	owner, err := b.nameOwner(a11yRegistryName)
	if err != nil {
		b.mu.Lock()
		b.failAt = time.Now()
		b.mu.Unlock()
		b.setReason("embed failed: " + err.Error())
		return fmt.Errorf("embed failed: %w", err)
	}

	// The registry writes Application.Id before it sends the Embed
	// reply. Seed the parent name so serve.go's sender guard admits
	// that write; the reply installs the full parent below.
	b.mu.Lock()
	if b.conn == nil || b.closed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	b.parent = dbusRef{Name: owner}
	b.view = newServerView(b.app, b.refNameLocked(), b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	b.mu.Unlock()

	plug := encodeStructSO(refName, app.RootPath)
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		path:    accessibleRootPath,
		dest:    a11yRegistryName,
		iface:   "org.a11y.atspi.Socket",
		member:  "Embed",
		sig:     "(so)",
		body:    plug,
	}
	reply, err := b.call(msg, owner, liveCallTimeout)
	if err != nil {
		b.clearPendingParent(owner)
		b.mu.Lock()
		b.failAt = time.Now()
		b.mu.Unlock()
		b.setReason("embed failed: " + err.Error())
		return fmt.Errorf("embed failed: %w", err)
	}
	name, path, err := decodeSO(reply.body)
	if err != nil {
		b.clearPendingParent(owner)
		b.setReason("embed reply: " + err.Error())
		return fmt.Errorf("embed reply: %w", err)
	}
	b.mu.Lock()
	b.parent = dbusRef{Name: name, Path: path}
	b.embedded = true
	b.failAt = time.Time{}
	b.view = newServerView(b.app, refName, b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	b.reason = "connected to " + b.addr + " as " + b.name
	b.mu.Unlock()
	return nil
}

// clearPendingParent drops the owner name seeded before Embed when the
// handshake fails. It leaves a fully installed parent untouched.
func (b *liveBus) clearPendingParent(owner string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.parent.Name != owner || b.parent.Path != "" {
		return
	}
	b.parent = dbusRef{}
	b.view = newServerView(b.app, b.refNameLocked(), b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
}

func (b *liveBus) Emit(ev Event) error {
	b.mu.Lock()
	live := b.conn != nil && !b.closed && b.claimed
	name := b.refNameLocked()
	root := b.app.RootPath
	b.mu.Unlock()
	if !live {
		return ErrLiveUnavailable
	}
	sig, body, member, err := encodeEvent(ev, name, root)
	if err != nil {
		return err
	}
	path := ev.Path
	if path == "" {
		path = root
	}
	if path == "" {
		path = accessibleRootPath
	}
	msg := &dbusMsg{
		msgType: dbusTypeSignal,
		path:    path,
		iface:   "org.a11y.atspi.Event.Object",
		member:  member,
		sig:     sig,
		body:    body,
	}
	return b.send(msg, liveEmitTimeout)
}

func (b *liveBus) Close() error {
	b.connectMu.Lock()
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		b.connectMu.Unlock()
		b.wg.Wait()
		return nil
	}
	b.closed = true
	b.claimed = false
	b.embedded = false
	b.reason = "bus closed"
	conn := b.conn
	b.conn = nil
	b.reader = nil
	b.uniqueName = ""
	b.parent = dbusRef{}
	connDone := b.connDone
	b.connDone = nil
	b.closeOnce.Do(func() { close(b.done) })
	if connDone != nil {
		close(connDone)
	}
	b.mu.Unlock()
	b.connectMu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
	b.wg.Wait()
	return nil
}

func discoverATSPISocket() (string, string, string, error) {
	if disabledByEnvironment() {
		return "", "", "", errors.New("at-spi bus disabled by environment")
	}
	addr := os.Getenv("AT_SPI_BUS_ADDRESS")
	if addr != "" {
		return parseDBusAddress(addr)
	}
	if netw, sock, guid, err := sessionA11yAddress(); err == nil && sock != "" {
		return netw, sock, guid, nil
	}
	uid := os.Getuid()
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		matches, _ := filepath.Glob(filepath.Join(runtimeDir, "at-spi", "bus*"))
		if len(matches) > 0 {
			return parseDBusAddress("unix:path=" + matches[0])
		}
	}
	matches, _ := filepath.Glob(fmt.Sprintf("/run/user/%d/at-spi/bus*", uid))
	if len(matches) > 0 {
		return parseDBusAddress("unix:path=" + matches[0])
	}
	return "", "", "", errors.New(liveReason())
}

// disabledByEnvironment reports whether AT-SPI has been switched off.
// NO_AT_BRIDGE is the GTK convention; AT_SPI_BUS_ADDRESS also accepts
// the disabled and none sentinels.
func disabledByEnvironment() bool {
	switch strings.TrimSpace(os.Getenv("AT_SPI_BUS_ADDRESS")) {
	case "disabled", "none":
		return true
	}
	bridge := strings.TrimSpace(os.Getenv("NO_AT_BRIDGE"))
	return bridge == "1" || strings.EqualFold(bridge, "true")
}

// sessionA11yAddress asks the session bus for the accessibility bus
// address. The two buses are separate; DBUS_SESSION_BUS_ADDRESS alone
// is not an AT-SPI bus. A discovered address is trusted only when it is
// a Unix socket inside the user's runtime directory, is not the session
// bus itself, and its GUID matches the one the socket reports.
func sessionA11yAddress() (string, string, string, error) {
	session := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	if session == "" {
		return "", "", "", errors.New("no session bus address")
	}
	sessionNetw, sessionAddr, sessionGUID, err := parseDBusAddress(session)
	if err != nil {
		return "", "", "", err
	}
	conn, err := (&net.Dialer{Timeout: liveDialTimeout}).DialContext(context.Background(), sessionNetw, sessionAddr)
	if err != nil {
		return "", "", "", err
	}
	defer func() { _ = conn.Close() }()

	reader, _, err := greet(conn, sessionGUID)
	if err != nil {
		return "", "", "", err
	}

	// Start the accessibility bus the way a direct GetAddress call
	// would have. Activation may fail when the service is not
	// activatable; GetNameOwner below then says whether it is running.
	startCall := startServiceCall(a11yBusName)
	startCall.serial = serialStartService
	_, _ = syncCall(conn, reader, startCall, dbusDaemonName)

	ownerCall := getNameOwnerCall(a11yBusName)
	ownerCall.serial = serialGetNameOwner
	ownerReply, err := syncCall(conn, reader, ownerCall, dbusDaemonName)
	if err != nil {
		return "", "", "", err
	}
	owner, err := decodeStringBody(ownerReply.body)
	if err != nil {
		return "", "", "", fmt.Errorf("decode %s owner: %w", a11yBusName, err)
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return "", "", "", fmt.Errorf("name %s has no owner", a11yBusName)
	}

	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  serialGetAddress,
		path:    "/org/a11y/bus",
		dest:    a11yBusName,
		iface:   a11yBusName,
		member:  "GetAddress",
	}
	reply, err := syncCall(conn, reader, msg, owner)
	if err != nil {
		return "", "", "", err
	}
	raw, err := decodeStringBody(reply.body)
	if err != nil {
		return "", "", "", err
	}
	netw, sock, guid, err := parseDBusAddress(raw)
	if err != nil {
		return "", "", "", err
	}
	if err := checkDiscoveredAddress(netw, sock, sessionNetw, sessionAddr); err != nil {
		return "", "", "", err
	}
	return netw, sock, guid, nil
}

// checkDiscoveredAddress refuses a GetAddress result that points at the
// session bus or outside the user's runtime directory. Abstract sockets
// are refused because the runtime-directory policy cannot bound them.
func checkDiscoveredAddress(netw, addr, sessionNetw, sessionAddr string) error {
	if netw != "unix" {
		return fmt.Errorf("discovered bus uses unsupported transport %q", netw)
	}
	if strings.HasPrefix(addr, "@") {
		return errors.New("discovered bus uses an abstract socket")
	}
	if netw == sessionNetw && sameUnixSocket(addr, sessionAddr) {
		return errors.New("discovered bus address matches the session bus")
	}
	if !insideRuntimeDir(addr) {
		return fmt.Errorf("discovered bus socket outside the runtime directory: %s", addr)
	}
	return nil
}

func sameUnixSocket(a, b string) bool {
	if strings.HasPrefix(a, "@") || strings.HasPrefix(b, "@") {
		return a == b
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func insideRuntimeDir(path string) bool {
	if !filepath.IsAbs(path) {
		return false
	}
	clean := filepath.Clean(path)
	dirs := []string{fmt.Sprintf("/run/user/%d", os.Getuid())}
	if runtimeDir := strings.TrimSpace(os.Getenv("XDG_RUNTIME_DIR")); runtimeDir != "" {
		dirs = append(dirs, runtimeDir)
	}
	for _, dir := range dirs {
		if !filepath.IsAbs(dir) {
			continue
		}
		root := filepath.Clean(dir)
		if strings.HasPrefix(clean, root+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func parseDBusAddress(addr string) (string, string, string, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", "", "", errors.New("empty bus address")
	}
	if strings.HasPrefix(addr, "unix:") {
		rest := addr[len("unix:"):]
		parts := strings.Split(rest, ",")
		kind, sock, guid := "", "", ""
		for _, part := range parts {
			switch {
			case strings.HasPrefix(part, "guid="):
				guid = strings.TrimSpace(part[len("guid="):])
			case kind == "" && strings.HasPrefix(part, "path="):
				kind, sock = "path", part[len("path="):]
			case kind == "" && strings.HasPrefix(part, "abstract="):
				kind, sock = "abstract", "@"+part[len("abstract="):]
			}
		}
		if kind == "" {
			return "", "", "", fmt.Errorf("no path or abstract in unix address: %s", addr)
		}
		if sock == "" || sock == "@" {
			return "", "", "", fmt.Errorf("empty %s in unix address: %s", kind, addr)
		}
		return "unix", sock, guid, nil
	}
	if strings.HasPrefix(addr, "/") {
		return "unix", addr, "", nil
	}
	return "", "", "", fmt.Errorf("unsupported bus address: %s", addr)
}
