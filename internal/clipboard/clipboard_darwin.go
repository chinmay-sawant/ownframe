//go:build darwin && !ios && cgo

package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>
#import <stdlib.h>
#include <stdbool.h>

static bool ownframePut(const void *p, size_t n) {
	@autoreleasepool {
		NSPasteboard *pb = [NSPasteboard generalPasteboard];
		[pb clearContents];
		NSString *s = [[NSString alloc] initWithBytes:p
			length:(NSUInteger)n
			encoding:NSUTF8StringEncoding];
		if (s == nil) {
			return [pb setString:@"" forType:NSPasteboardTypeString];
		}
		bool ok = [pb setString:s forType:NSPasteboardTypeString];
		[s release];
		return ok;
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
import (
	"context"
	"unsafe"
)

func writeOS(_ context.Context, text string) bool {
	if text == "" {
		return bool(C.ownframePut(nil, 0))
	}

	p := unsafe.Pointer(unsafe.StringData(text))
	return bool(C.ownframePut(p, C.size_t(len(text))))
}

func readOS(context.Context) (string, bool) {
	p := C.ownframeGet()
	if p == nil {
		return "", false
	}

	defer C.free(unsafe.Pointer(p))

	return C.GoString(p), true
}

func clearOS() { C.ownframePut(nil, 0) }

func diagnoseOS() Status { return Status{Backend: "darwin", Available: true} }
