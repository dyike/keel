//go:build darwin && !ios && cgo
#import <AppKit/AppKit.h>
extern void keel_application_menu_command(uint64_t generation,const char *id);

static void keel_edit(NSString *action) {
    NSDictionary *keys=@{@"copy":@"c",@"cut":@"x",@"paste":@"v",@"select-all":@"a",@"undo":@"z",@"redo":@"z"};
    NSString *key=keys[action];
    NSWindow *window=NSApp.keyWindow;
    id responder=window.firstResponder;
    if(!key || ![responder respondsToSelector:@selector(keyDown:)]) return;
    NSEventModifierFlags flags=NSEventModifierFlagCommand;
    if([action isEqualToString:@"redo"]) flags|=NSEventModifierFlagShift;
    NSEvent *event=[NSEvent keyEventWithType:NSEventTypeKeyDown location:NSZeroPoint modifierFlags:flags timestamp:NSProcessInfo.processInfo.systemUptime windowNumber:window.windowNumber context:nil characters:key charactersIgnoringModifiers:key isARepeat:NO keyCode:0];
    [responder keyDown:event];
}
void keel_application_menu_edit(const char *action) {
    NSString *name=[[NSString alloc]initWithUTF8String:action];
    dispatch_async(dispatch_get_main_queue(),^{keel_edit(name);});
    [name release];
}

@interface KeelMenuActions : NSObject
@property(nonatomic) uint64_t generation;
-(void)run:(NSMenuItem *)item;
-(void)edit:(NSMenuItem *)item;
@end
@implementation KeelMenuActions
-(void)run:(NSMenuItem *)item {keel_application_menu_command(self.generation,[item.representedObject UTF8String]);}
-(void)edit:(NSMenuItem *)item {keel_edit(item.representedObject);}
@end

static NSString *keel_key(NSString *name) {
    NSDictionary *keys=@{@"⏎":@"\r",@"⌤":@"\r",@"⎋":@"\033",@"space":@" ",@"tab":@"\t",@"⌫":@"\177",@"⌦":@"\uF728",@"↑":@"\uF700",@"↓":@"\uF701",@"←":@"\uF702",@"→":@"\uF703",@"⇱":@"\uF729",@"⇲":@"\uF72B",@"⇞":@"\uF72C",@"⇟":@"\uF72D"};
    if(keys[name]) return keys[name];
    if([name hasPrefix:@"f"] && name.length>1) {
        NSInteger number=[[name substringFromIndex:1] integerValue];
        if(number>=1 && number<=35) {unichar key=(unichar)(NSF1FunctionKey+number-1);return [NSString stringWithCharacters:&key length:1];}
    }
    return name;
}
static NSMenu *keel_build_menu(NSString *title,NSArray *items,KeelMenuActions *actions) {
    NSMenu *menu=[[NSMenu alloc]initWithTitle:title];menu.autoenablesItems=NO;
    for(NSDictionary *entry in items) {
        if([entry[@"separator"] boolValue]) {[menu addItem:NSMenuItem.separatorItem];continue;}
        NSString *action=entry[@"action"],*key=keel_key(entry[@"key"]);
        NSMenuItem *item=[menu addItemWithTitle:entry[@"title"] action:action.length?@selector(edit:):@selector(run:) keyEquivalent:key];
        item.target=actions;item.representedObject=action.length?action:entry[@"id"];
        NSUInteger mods=[entry[@"modifiers"] unsignedIntegerValue];
        NSEventModifierFlags flags=0;
        if(mods&1) flags|=NSEventModifierFlagCommand;
        if(mods&2) flags|=NSEventModifierFlagControl;
        if(mods&4) flags|=NSEventModifierFlagOption;
        if(mods&8) flags|=NSEventModifierFlagShift;
        item.keyEquivalentModifierMask=flags;
        item.enabled=![entry[@"disabled"] boolValue];item.state=[entry[@"checked"] boolValue]?NSControlStateValueOn:NSControlStateValueOff;
        NSArray *children=entry[@"children"];
        if(children.count) {
            NSMenu *sub=keel_build_menu(entry[@"title"],children,actions);item.submenu=sub;
            NSString *role=entry[@"role"];
            if([role isEqualToString:@"window"]) NSApp.windowsMenu=sub;
            if([role isEqualToString:@"help"]) NSApp.helpMenu=sub;
            if([role isEqualToString:@"services"]) NSApp.servicesMenu=sub;
            [sub release];
        }
    }
    return menu;
}
void keel_install_application_menu(const void *json,size_t len) {
    NSData *data=[[NSData alloc]initWithBytes:json length:len];
    dispatch_async(dispatch_get_main_queue(),^{
        NSDictionary *model=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
        if(!model) return;
        static KeelMenuActions *current;
        KeelMenuActions *next=[[KeelMenuActions alloc]init];next.generation=[model[@"generation"] unsignedLongLongValue];
        NSApp.windowsMenu=nil;NSApp.helpMenu=nil;NSApp.servicesMenu=nil;
        NSMenu *bar=keel_build_menu(@"",model[@"items"],next);
        NSApp.mainMenu=bar;
        [bar release];[current release];current=next;
    });
    [data release];
}
