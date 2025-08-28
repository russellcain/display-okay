package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework CoreGraphics
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <CoreGraphics/CoreGraphics.h>
#import <stdlib.h>

struct Display {
    int id;
    char* name;
    double x, y, width, height;
};

struct Window {
    int id;
    char* title;
    double x, y, width, height;
    AXUIElementRef ref;
};

bool check_accessibility() {
    const void *keys[] = { kAXTrustedCheckOptionPrompt };
    const void *vals[] = { kCFBooleanTrue };
    CFDictionaryRef options = CFDictionaryCreate(kCFAllocatorDefault, keys, vals, 1, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
    return AXIsProcessTrustedWithOptions(options);
}

int get_display_info(struct Display* displays, int max_displays) {
    @autoreleasepool {
        NSArray<NSScreen *> *screens = [NSScreen screens];
        int count = [screens count];
        for (int i = 0; i < count && i < max_displays; i++) {
            NSScreen *screen = screens[i];
            NSDictionary<NSDeviceDescriptionKey, id> *description = [screen deviceDescription];
            NSNumber *screenNumber = description[@"NSScreenNumber"];
            CGDirectDisplayID displayID = [screenNumber unsignedIntValue];
            NSString *displayName = [screen localizedName];
            NSRect frame = [screen frame];

            displays[i].id = displayID;
            displays[i].x = frame.origin.x;
            displays[i].y = frame.origin.y;
            displays[i].width = frame.size.width;
            displays[i].height = frame.size.height;
            displays[i].name = (char*)[displayName UTF8String];
        }
        return count;
    }
}

int get_window_info(struct Window* windows, int max_windows) {
    CFArrayRef windowList = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly, kCGNullWindowID);
    int count = CFArrayGetCount(windowList);
    for (int i = 0; i < count && i < max_windows; i++) {
        CFDictionaryRef window_info = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, i);
        CFNumberRef window_id_ref = (CFNumberRef)CFDictionaryGetValue(window_info, kCGWindowNumber);
        int window_id;
        CFNumberGetValue(window_id_ref, kCFNumberIntType, &window_id);

        pid_t pid;
        CFNumberRef pid_ref = (CFDictionaryGetValue(window_info, kCGWindowOwnerPID));
        CFNumberGetValue(pid_ref, kCFNumberSInt32Type, &pid);

        AXUIElementRef app_ref = AXUIElementCreateApplication(pid);
        CFArrayRef window_array;
        AXUIElementCopyAttributeValue(app_ref, kAXWindowsAttribute, (CFTypeRef *)&window_array);

        windows[i].ref = NULL;
        if (window_array != NULL) {
            int window_count = CFArrayGetCount(window_array);
            for (int j = 0; j < window_count; j++) {
                AXUIElementRef window_ref = (AXUIElementRef)CFArrayGetValueAtIndex(window_array, j);
                if (window_ref != NULL) {
                    windows[i].ref = window_ref;
                }
            }
        }

        CFStringRef title = (CFStringRef)CFDictionaryGetValue(window_info, kCGWindowName);
        if (title != NULL) {
            long len = CFStringGetLength(title);
            char *buffer = (char *)malloc(len + 1);
            CFStringGetCString(title, buffer, len + 1, kCFStringEncodingUTF8);
            windows[i].title = buffer;
        } else {
            windows[i].title = NULL;
        }

        CGRect bounds;
        CFDictionaryRef bounds_dict = (CFDictionaryRef)CFDictionaryGetValue(window_info, kCGWindowBounds);
        CGRectMakeWithDictionaryRepresentation(bounds_dict, &bounds);
        windows[i].x = bounds.origin.x;
        windows[i].y = bounds.origin.y;
        windows[i].width = bounds.size.width;
        windows[i].height = bounds.size.height;
        windows[i].id = window_id;
    }
    CFRelease(windowList);
    return count;
}

void move_window(AXUIElementRef window_ref, int x, int y) {
    CGPoint new_pos = {.x = x, .y = y};
    CFTypeRef pos_ref = AXValueCreate(kAXValueCGPointType, &new_pos);
    AXUIElementSetAttributeValue(window_ref, kAXPositionAttribute, pos_ref);
    CFRelease(pos_ref);
}

void resize_window(AXUIElementRef window_ref, int w, int h) {
    CGSize new_size = {.width = w, .height = h};
    CFTypeRef size_ref = AXValueCreate(kAXValueCGSizeType, &new_size);
    AXUIElementSetAttributeValue(window_ref, kAXSizeAttribute, size_ref);
    CFRelease(size_ref);
}

bool is_window_ref_null(AXUIElementRef ref) {
    return ref == NULL;
}

void free_window_titles(struct Window* windows, int count) {
    for (int i = 0; i < count; i++) {
        if (windows[i].title != NULL) {
            free(windows[i].title);
        }
    }
}

*/
import "C"
import (
    "fmt"
	"os"
    "time"
)

func main() {
	if !C.check_accessibility() {
		fmt.Println("Accessibility permissions are not granted. Please grant them in System Preferences.")
		os.Exit(1)
	}

	fmt.Println("Accessibility permissions are granted.")

    const target_display_name = "C27F390"

    // Note: This program cannot manage fullscreen windows. When a window enters
    // fullscreen mode, it is moved to a new "space" (virtual desktop) and is no
    // longer manageable by this program.

    ticker := time.NewTicker(3 * time.Second)
    for range ticker.C {
        displays := make([]C.struct_Display, 10)
        display_count := C.get_display_info(&displays[0], 10)

        var external_display *C.struct_Display
        for i := 0; i < int(display_count); i++ {
            if C.GoString(displays[i].name) == target_display_name {
                external_display = &displays[i]
                break
            }
        }

        if external_display != nil {
            windows := make([]C.struct_Window, 100)
            window_count := C.get_window_info(&windows[0], 100)

            for i := 0; i < int(window_count); i++ {
                if !C.is_window_ref_null(windows[i].ref) {
                    // Check if the window is on the external display
                    if windows[i].x >= external_display.x && windows[i].x < external_display.x + external_display.width {
                        // Check if the window's right edge extends beyond the 84% mark of the screen
                        if (windows[i].x + windows[i].width) > (external_display.x + (external_display.width * 0.84)) {
                            new_x := external_display.x
                            new_y := windows[i].y // Keep the original y position
                            new_width := external_display.width * 0.84
                            new_height := windows[i].height // Keep the original height

                            C.move_window(windows[i].ref, C.int(new_x), C.int(new_y))
                            C.resize_window(windows[i].ref, C.int(new_width), C.int(new_height))
                        }
                    }
                }
            }

            C.free_window_titles(&windows[0], C.int(window_count))
        }
    }
}