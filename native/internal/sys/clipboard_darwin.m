//go:build darwin && !ios && cgo

#import <AppKit/AppKit.h>
#include <stdint.h>

extern void keel_clipboard_read_done(uintptr_t token, const char *json, int code);

void keel_clipboard_read(uintptr_t token) {
    dispatch_async(dispatch_get_main_queue(), ^{
        @autoreleasepool {
            NSPasteboard *board = NSPasteboard.generalPasteboard;
            NSInteger version = board.changeCount;
            NSArray<NSPasteboardItem *> *items = board.pasteboardItems;
            if (items.count > 128) { keel_clipboard_read_done(token, "", 7); return; }
            NSString *text = [board stringForType:NSPasteboardTypeString] ?: @"";
            NSUInteger size = [text lengthOfBytesUsingEncoding:NSUTF8StringEncoding];
            NSMutableArray *images = [NSMutableArray array];
            NSMutableArray *files = [NSMutableArray array];
            for (NSPasteboardItem *item in items) {
                NSString *urlText = [item stringForType:NSPasteboardTypeFileURL];
                NSURL *url = urlText ? [NSURL URLWithString:urlText] : nil;
                if (url.isFileURL && url.path) {
                    size += [url.path lengthOfBytesUsingEncoding:NSUTF8StringEncoding];
                    [files addObject:url.path];
                    // Finder may attach an image preview; preserve the file identity.
                    continue;
                }
                NSData *data = [item dataForType:NSPasteboardTypePNG];
                NSString *mime = @"image/png";
                if (!data) { data = [item dataForType:NSPasteboardTypeTIFF]; mime = @"image/tiff"; }
                if (data) {
                    size += data.length;
                    if (size > 16*1024*1024) { keel_clipboard_read_done(token, "", 7); return; }
                    [images addObject:@{@"mime":mime, @"data":[data base64EncodedStringWithOptions:0]}];
                }
            }
            if (size > 16*1024*1024 || version != board.changeCount) {
                keel_clipboard_read_done(token, "", 7); return;
            }
            NSData *json = [NSJSONSerialization dataWithJSONObject:@{@"text":text,@"images":images,@"files":files} options:0 error:nil];
            NSString *encoded = json ? [[NSString alloc] initWithData:json encoding:NSUTF8StringEncoding] : nil;
            keel_clipboard_read_done(token, encoded ? encoded.UTF8String : "", encoded ? 0 : 7);
        }
    });
}
