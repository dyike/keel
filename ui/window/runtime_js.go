//go:build js

package window

func serveAutomation()                    {}
func restoreWindow(*Window)               {}
func registerDevelopmentWindow(*Window)   {}
func unregisterDevelopmentWindow(*Window) {}
func (*Window) saveSizeState()            {}
