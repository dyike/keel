//go:build darwin && !ios && cgo

#import <AppKit/AppKit.h>

// keel_set_app_icon copies the PNG and sets the Dock icon on the main
// queue, once the app object exists.
void keel_set_app_icon(const void *png, size_t len) {
    NSData *data = [NSData dataWithBytes:png length:len];
    dispatch_async(dispatch_get_main_queue(), ^{
        NSImage *image = [[NSImage alloc] initWithData:data];
        if (image) {
            [NSApplication sharedApplication].applicationIconImage = image;
        }
    });
}
