package bridge

// Host is the slice of the GUI toolkit that bridge needs. The v2 entry point
// (main.go) and the v3 one (main_v3.go) each install their own via SetHost;
// unset callbacks are no-ops, so bridge stays usable without a toolkit.
type Host struct {
	Emit           func(name string, data ...any)
	On             func(name string, handler func(data ...any))
	Off            func(name string)
	WindowShow     func()
	WindowHide     func()
	WindowSetTitle func(title string)
	Quit           func()
}

var host Host

func SetHost(h Host) { host = h }

func emit(name string, data ...any) {
	if host.Emit != nil {
		host.Emit(name, data...)
	}
}

func on(name string, handler func(data ...any)) {
	if host.On != nil {
		host.On(name, handler)
	}
}

func off(name string) {
	if host.Off != nil {
		host.Off(name)
	}
}

func windowShow() {
	if host.WindowShow != nil {
		host.WindowShow()
	}
}

func windowHide() {
	if host.WindowHide != nil {
		host.WindowHide()
	}
}

func windowSetTitle(title string) {
	if host.WindowSetTitle != nil {
		host.WindowSetTitle(title)
	}
}

func quitApp() {
	if host.Quit != nil {
		host.Quit()
	}
}
