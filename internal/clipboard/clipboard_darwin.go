//go:build darwin && !ios && cgo

package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>
#import <stdlib.h>

static void ownframePut(const void *p, size_t n) {
	@autoreleasepool {
		NSPasteboard *pb = [NSPasteboard generalPasteboard];
		[pb clearContents];
		NSString *s = [[NSString alloc] initWithBytes:p
			length:(NSUInteger)n
			encoding:NSUTF8StringEncoding];
		if (s == nil) {
			[pb setString:@"" forType:NSPasteboardTypeString];
			return;
		}
		[pb setString:s forType:NSPasteboardTypeString];
		[s release];
	}
}

static char *ownframeGet(void) {
	@autoreleasepool {
		NSString *s = [[NSPasteboard generalPasteboard]
			stringForType:NSPasteboardTypeString];
		if (s == nil || [s UTF8String] == NULL) {
			return NULL;
		}
		return strdup([s UTF8String]);
	}
}
*/
import "C"
import "unsafe"

func writeOS(text string) {
	if text == "" {
		C.ownframePut(nil, 0)

		return
	}

	p := unsafe.Pointer(unsafe.StringData(text))
	C.ownframePut(p, C.size_t(len(text)))
}

func readOS() (string, bool) {
	p := C.ownframeGet()
	if p == nil {
		return "", false
	}

	defer C.free(unsafe.Pointer(p))

	return C.GoString(p), true
}

func clearOS() { C.ownframePut(nil, 0) }
