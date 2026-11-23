//go:build linux && !android && !ios && !js && cgo

package window

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <pthread.h>
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
// Xlib's default error handler exits the process. A window that closes while
// the tree is walked or queried raises BadWindow, so each Xlib call runs under
// a scoped handler that ignores errors for the active Display connection.
// Because XSetErrorHandler is process-wide in Xlib, ownframe_x_mu serializes
// ownframe_x_begin/ownframe_x_end and restores the previous handler after XSync
// flushes pending errors on the connection.
static pthread_mutex_t ownframe_x_mu = PTHREAD_MUTEX_INITIALIZER;
static Display *ownframe_active_display = NULL;
static XErrorHandler ownframe_prev_handler = NULL;
static int ownframe_ignore_x_error(Display *d, XErrorEvent *e) {
 if (d == ownframe_active_display || ownframe_active_display == NULL) {
  return 0;
 }
 if (ownframe_prev_handler != NULL) {
  return ownframe_prev_handler(d, e);
 }
 return 0;
}
static XErrorHandler ownframe_x_begin(Display *d) {
 pthread_mutex_lock(&ownframe_x_mu);
 ownframe_active_display = d;
 ownframe_prev_handler = XSetErrorHandler(ownframe_ignore_x_error);
 return ownframe_prev_handler;
}
static void ownframe_x_end(Display *d, XErrorHandler prev) {
 if (d) XSync(d, False);
 XSetErrorHandler(prev);
 ownframe_prev_handler = NULL;
 ownframe_active_display = NULL;
 pthread_mutex_unlock(&ownframe_x_mu);
}
static Window ownframe_client_in(Display *d, Window start) {
 XErrorHandler prev = ownframe_x_begin(d);
 Window found = ownframe_find_window(d,start,XInternAtom(d,"_NET_WM_PID",False),2);
 ownframe_x_end(d, prev);
 return found;
}
static Window ownframe_client(Display *d, Window root) {
 return ownframe_client_in(d,root);
}
static int ownframe_geometry(Display *d, Window w, Window root, int *x, int *y, int *width, int *height) {
 XErrorHandler prev = ownframe_x_begin(d);
 Window child; XWindowAttributes attrs;
 int ok = XTranslateCoordinates(d,w,root,0,0,x,y,&child) && XGetWindowAttributes(d,w,&attrs);
 ownframe_x_end(d, prev);
 if (!ok) return 0;
 *width=attrs.width; *height=attrs.height;
 return 1;
}
static char *ownframe_title(Display *d,Window w) {
 XErrorHandler prev = ownframe_x_begin(d);
 Atom actual;int format;unsigned long count,after;unsigned char *data=NULL;
 Atom name=XInternAtom(d,"_NET_WM_NAME",False);
 Atom utf8=XInternAtom(d,"UTF8_STRING",False);
 char *res = NULL;
 if(XGetWindowProperty(d,w,name,0,4096,False,utf8,&actual,&format,&count,&after,&data)==Success && format==8 && count>0) {
  res = (char*)data;
 } else if(data) {
  XFree(data);
 }
 ownframe_x_end(d, prev);
 return res;
}
static void ownframe_free_name(char *name) {XFree(name);}
static int ownframe_frame_geometry(Display *d,Window w,Window root,int *x,int *y,int *width,int *height) {
 XErrorHandler prev = ownframe_x_begin(d);
 Window parent,r,*children=NULL;unsigned int n;
 int q = XQueryTree(d,w,&r,&parent,&children,&n);
 if(children) XFree(children);
 ownframe_x_end(d, prev);
 if(!q) return 0;
 return ownframe_geometry(d,parent==root?w:parent,root,x,y,width,height);
}
static Window ownframe_root(Display *d) { return DefaultRootWindow(d); }

static int ownframe_move(Display *d, Window root, Window w) {
 XErrorHandler prev = ownframe_x_begin(d);
 Window rr,child; int rx,ry,wx,wy; unsigned int mask;
 if (!XQueryPointer(d,root,&rr,&child,&rx,&ry,&wx,&wy,&mask)) {
  ownframe_x_end(d, prev);
  return 0;
 }
 XEvent event = {0};
 event.xclient.type=ClientMessage;
 event.xclient.window=w;
 event.xclient.message_type=XInternAtom(d,"_NET_WM_MOVERESIZE",False);
 event.xclient.format=32;
 event.xclient.data.l[0]=rx; event.xclient.data.l[1]=ry;
 event.xclient.data.l[2]=8; event.xclient.data.l[3]=1; event.xclient.data.l[4]=1;
 int sent=XSendEvent(d,root,False,SubstructureRedirectMask|SubstructureNotifyMask,&event);
 ownframe_x_end(d, prev);
 return sent;
}
static void ownframe_set_rect(Display *d, Window w, int x, int y, unsigned int width, unsigned int height) {
 XErrorHandler prev = ownframe_x_begin(d);
 XResizeWindow(d, w, width, height);
 XMoveWindow(d, w, x, y);
 ownframe_x_end(d, prev);
}
static int ownframe_query_pointer(Display *d, Window w, int *x, int *y) {
 XErrorHandler prev = ownframe_x_begin(d);
 Window root, child; int rx, ry; unsigned int mask;
 int ok = XQueryPointer(d, w, &root, &child, &rx, &ry, x, y, &mask);
 ownframe_x_end(d, prev);
 return ok;
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
	// XResizeWindow with a zero dimension raises a fatal BadValue.
	if w <= 0 || h <= 0 {
		return false
	}
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
			C.ownframe_set_rect(d.display, d.window, C.int(x+m.x), C.int(y+m.y), C.uint(w), C.uint(h))
			return true
		}
	}
	return false
}

// nativeClientFrom walks the window tree below start for this process's
// client window. An invalid start window yields 0 instead of ending the
// process. It exists so the error-handler guard can be exercised directly.
func nativeClientFrom(start uint64) (uint64, bool) {
	display := C.XOpenDisplay(nil)
	if display == nil {
		return 0, false
	}
	defer C.XCloseDisplay(display)

	return uint64(C.ownframe_client_in(display, C.Window(start))), true
}

func nativeWindowGeometry() (workarea, bool) {
	if !isWSLg() {
		return workarea{}, false
	}
	var d nativeWindowDrag
	d.begin()
	defer d.end()
	if d.display == nil {
		return workarea{}, false
	}
	var x, y, width, height C.int
	if C.ownframe_geometry(d.display, d.window, d.root, &x, &y, &width, &height) == 0 {
		return workarea{}, false
	}
	layout, ok := xrandrLayout()
	if !ok {
		return workarea{}, false
	}
	for _, m := range layout {
		if m.name == "XWAYLAND0" {
			return workarea{x: int(x) - m.x, y: int(y) - m.y, w: int(width), h: int(height)}, true
		}
	}
	return workarea{}, false
}

func nativeWindowPosition() (int, int, bool) { r, ok := nativeWindowGeometry(); return r.x, r.y, ok }

func (d *nativeWindowDrag) cursor() (int, int, bool) {
	if d.display == nil {
		d.begin()
	}
	if d.display == nil {
		return 0, 0, false
	}
	var x, y C.int
	ok := C.ownframe_query_pointer(d.display, d.window, &x, &y)
	return int(x), int(y), ok != 0
}

func nativeWindowTarget() (hostWindowTarget, bool) {
	var d nativeWindowDrag
	d.begin()
	defer d.end()
	if d.display == nil {
		return hostWindowTarget{}, false
	}
	client, ok := nativeWindowGeometry()
	if !ok {
		return hostWindowTarget{}, false
	}
	name := C.ownframe_title(d.display, d.window)
	if name == nil {
		return hostWindowTarget{}, false
	}
	title := C.GoString(name)
	C.ownframe_free_name(name)
	var x, y, w, h, cx, cy, cw, ch C.int
	if C.ownframe_frame_geometry(d.display, d.window, d.root, &x, &y, &w, &h) == 0 ||
		C.ownframe_geometry(d.display, d.window, d.root, &cx, &cy, &cw, &ch) == 0 {
		return hostWindowTarget{}, false
	}
	return hostWindowTarget{Title: title, Client: desktopRect{X: client.x, Y: client.y, W: client.w, H: client.h}, Frame: desktopRect{X: client.x + int(x-cx), Y: client.y + int(y-cy), W: int(w), H: int(h)}}, true
}
