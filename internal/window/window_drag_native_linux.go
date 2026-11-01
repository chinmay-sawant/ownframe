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
static int ownframe_geometry(Display *d, Window w, Window root, int *x, int *y, int *width, int *height) {
 Window child; XWindowAttributes attrs;
 if (!XTranslateCoordinates(d,w,root,0,0,x,y,&child) || !XGetWindowAttributes(d,w,&attrs)) return 0;
 *width=attrs.width; *height=attrs.height;
 return 1;
}
static char *ownframe_title(Display *d,Window w) {
 Atom actual;int format;unsigned long count,after;unsigned char *data=NULL;
 Atom name=XInternAtom(d,"_NET_WM_NAME",False);
 Atom utf8=XInternAtom(d,"UTF8_STRING",False);
 if(XGetWindowProperty(d,w,name,0,4096,False,utf8,&actual,&format,&count,&after,&data)==Success && format==8 && count>0) return (char*)data;
 if(data) XFree(data);
 return NULL;
}
static void ownframe_free_name(char *name) {XFree(name);}
static int ownframe_frame_geometry(Display *d,Window w,Window root,int *x,int *y,int *width,int *height) {
 Window parent,r,*children=NULL;unsigned int n;
 if(!XQueryTree(d,w,&r,&parent,&children,&n)) return 0;
 if(children) XFree(children);
 return ownframe_geometry(d,parent==root?w:parent,root,x,y,width,height);
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
			C.XResizeWindow(d.display, d.window, C.uint(w), C.uint(h))
			C.XMoveWindow(d.display, d.window, C.int(x+m.x), C.int(y+m.y))
			C.XSync(d.display, C.False)
			return true
		}
	}
	return false
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
	var root, child C.Window
	var rx, ry, x, y C.int
	var mask C.uint
	ok := C.XQueryPointer(d.display, d.window, &root, &child, &rx, &ry, &x, &y, &mask)
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
	if C.ownframe_frame_geometry(d.display, d.window, d.root, &x, &y, &w, &h) == 0 || C.ownframe_geometry(d.display, d.window, d.root, &cx, &cy, &cw, &ch) == 0 {
		return hostWindowTarget{}, false
	}
	return hostWindowTarget{Title: title, Client: desktopRect{X: client.x, Y: client.y, W: client.w, H: client.h}, Frame: desktopRect{X: client.x + int(x-cx), Y: client.y + int(y-cy), W: int(w), H: int(h)}}, true
}
