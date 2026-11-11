package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"log"
	"log/slog"
	"reflect"
	"time"

	"guiforcores/bridge"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/dist/favicon.ico
var icon []byte

//go:embed build/appicon.png
var appIcon []byte

// main is the Wails v3 entry point. On Linux build with -tags gtk3; it is kept
// beside main.go (which carries !v3) so both toolkits can be built from the
// same tree while the migration is in progress.
func main() {
	app := bridge.CreateApp(assets)
	// v2 supplied a context via OnStartup; bridge still uses it for HTTP timeouts.
	app.Ctx = context.Background()

	trayIcon := icon
	if bridge.Env.OS == "linux" {
		trayIcon = appIcon
	}
	trayStart, trayEnd := bridge.CreateTray(app, trayIcon)

	startState := application.WindowStateNormal
	if bridge.Env.WindowStartState == bridge.WindowStartMinimised {
		startState = application.WindowStateMinimised
	}

	uniqueID := func() string {
		if bridge.Config.MultipleInstance {
			return time.Now().String()
		}
		sum := sha256.Sum256([]byte(bridge.Env.DataDir))
		return bridge.Env.AppId + "-" + hex.EncodeToString(sum[:8])
	}()

	var window *application.WebviewWindow
	var v3app *application.App
	notifier := notifications.New()

	v3app = application.New(application.Options{
		Name:        bridge.Env.AppName,
		Description: "GUI.for.Cores",
		Icon:        appIcon,
		Services: []application.Service{
			application.NewService(app),
			application.NewService(notifier),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: bridge.RollingRelease,
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: uniqueID,
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				window.Show()
				// v2 delivered args as one array argument; wrapping keeps the
				// frontend shim's "spread ev.data" rule compatible.
				v3app.Event.EmitEvent(&application.CustomEvent{Name: "onLaunchApp", Data: []any{data.Args}})
			},
		},
		ShouldQuit: func() bool {
			if !bridge.Env.PreventExit {
				return true
			}
			v3app.Event.Emit("onBeforeExitApp")
			return false
		},
		OnShutdown: func() {
			trayEnd()
		},
		LogLevel: slog.LevelInfo,
	})

	window = v3app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:                    bridge.Env.AppName,
		URL:                      "/",
		Width:                    max(bridge.Config.Width, 1024),
		Height:                   max(bridge.Config.Height, 768),
		MinWidth:                 600,
		MinHeight:                400,
		Frameless:                bridge.Env.OS != "darwin" && !bridge.Config.SystemTitleBar,
		StartState:               startState,
		BackgroundColour:         application.NewRGB(255, 255, 255),
		BackgroundType:           backgroundType(),
		ContentProtectionEnabled: bridge.Config.ContentProtection,
		DevToolsEnabled:          true,
		Windows: application.WindowsWindow{
			BackdropType: application.Acrylic,
		},
		Mac: application.MacWindow{
			TitleBar:   application.MacTitleBarHiddenInset,
			Appearance: application.DefaultAppearance,
			Backdrop:   application.MacBackdropTranslucent,
		},
		Linux: application.LinuxWindow{
			Icon:             appIcon,
			WebviewGpuPolicy: application.WebviewGpuPolicy(bridge.Config.WebviewGpuPolicy),
		},
	})

	bridge.SetHost(bridge.Host{
		Emit: emitV3(v3app),
		On: func(name string, handler func(data ...any)) {
			v3app.Event.On(name, func(e *application.CustomEvent) { handler(toArgs(e.Data)...) })
		},
		Off:            v3app.Event.Off,
		WindowShow:     func() { window.Show() },
		WindowHide:     func() { window.Hide() },
		WindowSetTitle: func(title string) { window.SetTitle(title) },
		Quit:           v3app.Quit,
	})

	v3app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		if bridge.Env.OS == "darwin" {
			v3app.Menu.SetApplicationMenu(createMacOSMenus(window, v3app))
		}
		trayStart()
	})

	if err := v3app.Run(); err != nil {
		log.Fatal(err)
	}
}

// macOS 需要显式菜单，否则 Cmd+C/Cmd+V 之类的编辑快捷键不生效。
func createMacOSMenus(win *application.WebviewWindow, app *application.App) *application.Menu {
	appMenu := application.NewMenu()
	appSub := appMenu.AddSubmenu("App")
	appSub.Add("Show").SetAccelerator("cmdorctrl+s").OnClick(func(*application.Context) { win.Show() })
	appSub.Add("Hide").SetAccelerator("cmdorctrl+h").OnClick(func(*application.Context) { win.Hide() })
	appSub.AddSeparator()
	appSub.Add("Quit").SetAccelerator("cmdorctrl+q").OnClick(func(*application.Context) {
		app.Event.Emit("onExitApp")
	})
	appMenu.AddRole(application.EditMenu)
	return appMenu
}

// v2 kept Linux windows opaque and made the Windows/macOS ones translucent.
func backgroundType() application.BackgroundType {
	if bridge.Env.OS == "linux" {
		return application.BackgroundTypeSolid
	}
	return application.BackgroundTypeTranslucent
}

// emitV3 adapts bridge's variadic payloads to v3 event data. v3 keeps the
// payload as-is when there is exactly one argument, so a lone slice has to be
// wrapped in a slice to stay distinguishable from a multi-argument event on
// the frontend side.
func emitV3(app *application.App) func(string, ...any) {
	return func(name string, data ...any) {
		app.Event.Emit(name, normalizePayload(data)...)
	}
}

func normalizePayload(data []any) []any {
	if len(data) == 1 && isSlice(data[0]) {
		return []any{data[0]}
	}
	return data
}

func isSlice(v any) bool {
	if v == nil {
		return false
	}
	k := reflect.TypeOf(v).Kind()
	return k == reflect.Slice || k == reflect.Array
}

// toArgs turns v3's single data field back into v2's variadic arguments.
func toArgs(data any) []any {
	if data == nil {
		return nil
	}
	if args, ok := data.([]any); ok {
		return args
	}
	return []any{data}
}
