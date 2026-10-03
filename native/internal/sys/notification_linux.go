//go:build linux && !android

package sys

import (
	"context"
	"fmt"
	"github.com/dyike/keel/native"
	"github.com/godbus/dbus/v5"
	"sync"
	"time"
)

var linuxNotifications struct {
	sync.Mutex
	conn    *dbus.Conn
	service notificationDBus
}

func linuxNotificationConnection() error {
	if linuxNotifications.conn != nil && linuxNotifications.conn.Connected() {
		return nil
	}
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("%w: session bus: %v", native.ErrUnsupported, err)
	}
	ch := make(chan *dbus.Signal, 32)
	conn.Signal(ch)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err = conn.AddMatchSignalContext(ctx, dbus.WithMatchInterface(notificationService), dbus.WithMatchMember("NotificationClosed"), dbus.WithMatchObjectPath("/org/freedesktop/Notifications"))
	cancel()
	if err != nil {
		conn.RemoveSignal(ch)
		conn.Close()
		return notificationDBusError(err)
	}
	linuxNotifications.conn = conn
	linuxNotifications.service = notificationDBus{call: func(destination, method string, args ...any) *dbus.Call {
		path := dbus.ObjectPath("/org/freedesktop/Notifications")
		if destination == "org.freedesktop.DBus" {
			path = "/org/freedesktop/DBus"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return conn.Object(destination, path).CallWithContext(ctx, method, 0, args...)
	}}
	go func() {
		for signal := range ch {
			if signal == nil || signal.Name != notificationService+".NotificationClosed" || len(signal.Body) < 1 {
				continue
			}
			id, ok := signal.Body[0].(uint32)
			if !ok {
				continue
			}
			linuxNotifications.Lock()
			if linuxNotifications.conn == conn {
				linuxNotifications.service.closed(signal.Sender, id)
			}
			linuxNotifications.Unlock()
		}
	}()
	return nil
}
func NotificationAvailable() bool {
	linuxNotifications.Lock()
	defer linuxNotifications.Unlock()
	if linuxNotificationConnection() != nil {
		return false
	}
	return linuxNotifications.service.refresh() == nil
}
func NotificationPermission(done func(error)) {
	go func() {
		linuxNotifications.Lock()
		err := linuxNotificationConnection()
		if err == nil {
			err = linuxNotifications.service.refresh()
		}
		linuxNotifications.Unlock()
		done(err)
	}()
}
func NotificationPost(id, title, body string, done func(error)) {
	go func() {
		linuxNotifications.Lock()
		err := linuxNotificationConnection()
		if err == nil {
			err = linuxNotifications.service.post(id, title, body)
		}
		linuxNotifications.Unlock()
		done(err)
	}()
}
func NotificationRemove(id string, done func(error)) {
	go func() {
		linuxNotifications.Lock()
		err := linuxNotificationConnection()
		if err == nil {
			err = linuxNotifications.service.remove(id)
		}
		linuxNotifications.Unlock()
		done(err)
	}()
}
