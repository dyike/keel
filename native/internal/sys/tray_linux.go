//go:build linux && !android

package sys

import (
	"context"
	"fmt"
	"time"

	"github.com/dyike/keel/native"
	"github.com/godbus/dbus/v5"
)

func TrayMenuCapabilities() TrayMenuFeatures { return TrayMenuFeatures{Icons: true} }

func TrayOpen(options TrayOptions) (TrayBackend, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("%w: tray session bus: %v", native.ErrUnsupported, err)
	}
	t := &dbusTray{conn: conn, options: options, ident: fmt.Sprintf("keel-tray-%d", dbusTrayID.Add(1)), revision: 1, signals: make(chan *dbus.Signal, 32)}
	t.pixmaps, err = trayPixmaps(options.PNG)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err = t.export(); err != nil {
		conn.Close()
		return nil, trayDBusFailure(err)
	}
	conn.Signal(t.signals)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err = conn.AddMatchSignalContext(ctx, dbus.WithMatchSender("org.freedesktop.DBus"), dbus.WithMatchInterface("org.freedesktop.DBus"), dbus.WithMatchMember("NameOwnerChanged"))
	cancel()
	if err != nil {
		conn.Close()
		return nil, trayDBusFailure(err)
	}
	for _, watcher := range trayWatchers {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = conn.AddMatchSignalContext(ctx, dbus.WithMatchInterface(watcher), dbus.WithMatchObjectPath("/StatusNotifierWatcher"), dbus.WithMatchMember("StatusNotifierHostRegistered"))
		cancel()
		if err != nil {
			conn.Close()
			return nil, trayDBusFailure(err)
		}
	}
	if err = t.register(); err != nil {
		conn.Close()
		return nil, err
	}
	go t.watch()
	return t, nil
}

func TrayMain() { select {} }
