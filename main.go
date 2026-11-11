//go:build !v3

package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"guiforcores/bridge"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/dist/favicon.ico
var icon []byte

//go:embed build/appicon.png
var appIcon []byte

func main() {
	app := bridge.CreateApp(assets)

	bridge.SetHost(bridge.Host{
		Emit: func(name string, data ...any) { runtime.EventsEmit(app.Ctx, name, data...) },
		On: func(name string, handler func(data ...any)) {
			runtime.EventsOn(app.Ctx, name, handler)
		},
		Off:            func(name string) { runtime.EventsOff(app.Ctx, name) },
		WindowShow:     func() { runtime.WindowShow(app.Ctx) },
		WindowHide:     func() { runtime.WindowHide(app.Ctx) },
		WindowSetTitle: func(title string) { runtime.WindowSetTitle(app.Ctx, title) },
		Quit:           func() { runtime.Quit(app.Ctx) },
	})

	trayIcon := icon
	if bridge.Env.OS == "linux" {
		trayIcon = appIcon
	}
	trayStart, trayEnd := bridge.CreateTray(app, trayIcon)

	appMenu := menu.NewMenu()
	if bridge.Env.OS == "darwin" {
		// macOS 需要显式菜单，否则 Cmd+C/Cmd+V 之类的编辑快捷键不生效。
		appMenu = createMacOSMenus(app)
	}

	// Create application with options
	err := wails.Run(&options.App{
		MinWidth:         600,
		MinHeight:        400,
		DisableResize:    false,
		Menu:             appMenu,
		Title:            bridge.Env.AppName,
		Frameless:        bridge.Env.OS != "darwin" && !bridge.Config.SystemTitleBar,
		Width:            bridge.Config.Width,
		Height:           bridge.Config.Height,
		StartHidden:      bridge.Env.WindowStartState == bridge.WindowStartMinimised,
		WindowStartState: windowStartState(),
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			ContentProtection:    bridge.Config.ContentProtection,
			BackdropType:         windows.Acrylic,
			WebviewBrowserPath:   bridge.Env.WebviewPath,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           mac.DefaultAppearance,
			ContentProtection:    bridge.Config.ContentProtection,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   bridge.Env.AppName,
				Message: "© 2026 GUI.for.Cores",
				Icon:    icon,
			},
		},
		Linux: &linux.Options{
			Icon:                appIcon,
			WindowIsTranslucent: false,
			ProgramName:         bridge.Env.AppName,
			WebviewGpuPolicy:    linux.WebviewGpuPolicy(bridge.Config.WebviewGpuPolicy),
		},
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: bridge.RollingRelease,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: func() string {
				if bridge.Config.MultipleInstance {
					return time.Now().String()
				}
				sum := sha256.Sum256([]byte(bridge.Env.DataDir))
				return bridge.Env.AppId + "-" + hex.EncodeToString(sum[:8])
			}(),
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				runtime.Show(app.Ctx)
				runtime.EventsEmit(app.Ctx, "onLaunchApp", data.Args)
			},
		},
		OnStartup: func(ctx context.Context) {
			app.Ctx = ctx
			runtime.InitializeNotifications(ctx)
			trayStart()
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			if !bridge.Env.PreventExit {
				trayEnd()
				runtime.CleanupNotifications(ctx)
				return false
			}
			runtime.EventsEmit(ctx, "onBeforeExitApp")
			return true
		},
		Bind: []any{
			app,
		},
		LogLevel: logger.INFO,
		Debug: options.Debug{
			OpenInspectorOnStartup: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func windowStartState() options.WindowStartState {
	if bridge.Env.WindowStartState == bridge.WindowStartMinimised {
		return options.Minimised
	}
	return options.Normal
}

func createMacOSMenus(app *bridge.App) *menu.Menu {
	appMenu := menu.NewMenu()
	appSub := appMenu.AddSubmenu("App")
	appSub.AddText("Show", keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) { runtime.WindowShow(app.Ctx) })
	appSub.AddText("Hide", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) { runtime.WindowHide(app.Ctx) })
	appSub.AddSeparator()
	appSub.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.Ctx, "onExitApp")
	})
	appMenu.Append(menu.EditMenu())
	return appMenu
}
