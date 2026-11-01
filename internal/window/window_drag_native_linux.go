//go:build linux && cgo

package window

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <stdlib.h>
#include <unistd.h>

static Window ownframe_find_window(Display *d, Window w, Atom pid_atom, int depth) {
 Atom type; int format; unsigned long count, after; unsigned char *data = NULL;
 if (XGetWindowProperty(d,w,pid_atom,0,1,False,XA_CARDINAL,
       &type,&format,&count,&after,&data)==Success && data) {
  int match = format==32 && count==1 && *(unsigned long*)data==(unsigned long)getpid();
  XFree(data);
  if (match) return w;
 }
 if (depth==0) return 0;
 Window root,parent,*children=NULL; unsigned int n=0;
 if (!XQueryTree(d,w,&root,&parent,&children,&n)) return 0;
 Window found=0;
 for (unsigned int i=0;i<n && !found;i++) found=ownframe_find_window(d,children[i],pid_atom,depth-1);
 if (children) XFree(children);
 return found;
}
static Window ownframe_client(Display *d, Window root) {
 return ownframe_find_window(d,root,XInternAtom(d,"_NET_WM_PID",False),2);
}
static Window ownframe_root(Display *d) { return DefaultRootWindow(d); }
static int ownframe_move(Display *d, Window root, Window w) {
 Window rr,child; int rx,ry,wx,wy; unsigned int mask;
 if (!XQueryPointer(d,root,&rr,&child,&rx,&ry,&wx,&wy,&mask)) return 0;
 XEvent event = {0};
 event.xclient.type=ClientMessage;
 event.xclient.window=w;
 event.xclient.message_type=XInternAtom(d,"_NET_WM_MOVERESIZE",False);
 event.xclient.format=32;
 event.xclient.data.l[0]=rx; event.xclient.data.l[1]=ry;
 event.xclient.data.l[2]=8; event.xclient.data.l[3]=1; event.xclient.data.l[4]=1;
 int sent=XSendEvent(d,root,False,SubstructureRedirectMask|SubstructureNotifyMask,&event);
 XFlush(d);
 return sent;
}
*/
import "C"

import "errors"

type nativeWindowDrag struct {
	display      *C.Display
	root, window C.Window
	started      bool
}

func (d *nativeWindowDrag) end() {
	if d.display != nil {
		C.XCloseDisplay(d.display)
	}
	*d = nativeWindowDrag{}
}

func (d *nativeWindowDrag) begin() {
	d.end()
	if !isWSLg() {
		return
	}
	display := C.XOpenDisplay(nil)
	if display == nil {
		return
	}
	root := C.ownframe_root(display)
	window := C.ownframe_client(display, root)
	if window == 0 {
		C.XCloseDisplay(display)
		return
	}
	d.display, d.root, d.window = display, root, window
}

func (d *nativeWindowDrag) managed() bool { return d.display != nil }

func (d *nativeWindowDrag) move() (bool, error) {
	if d.display == nil {
		return false, nil
	}
	if d.started {
		return true, nil
	}
	if C.ownframe_move(d.display, d.root, d.window) == 0 {
		return true, errors.New("start WSLg window-manager move failed")
	}
	d.started = true
	return true, nil
}

func setNativeWindowRect(x, y, w, h int) bool {
	if !isWSLg() {
		return false
	}
	var d nativeWindowDrag
	d.begin()
	defer d.end()
	if d.display == nil {
		return false
	}
	layout, ok := xrandrLayout()
	if !ok {
		return false
	}
	for _, m := range layout {
		if m.name == "XWAYLAND0" {
			C.XMoveResizeWindow(d.display, d.window, C.int(x+m.x), C.int(y+m.y), C.uint(w), C.uint(h))
			C.XSync(d.display, C.False)
			return true
		}
	}
	return false
}

func nativeWindowPosition() (int, int, bool) {
	if !isWSLg() {
		return 0, 0, false
	}
	var d nativeWindowDrag
	d.begin()
	defer d.end()
	if d.display == nil {
		return 0, 0, false
	}
	var x, y C.int
	var child C.Window
	if C.XTranslateCoordinates(d.display, d.window, d.root, 0, 0, &x, &y, &child) == 0 {
		return 0, 0, false
	}
	layout, ok := xrandrLayout()
	if !ok {
		return 0, 0, false
	}
	for _, m := range layout {
		if m.name == "XWAYLAND0" {
			return int(x) - m.x, int(y) - m.y, true
		}
	}
	return 0, 0, false
}
