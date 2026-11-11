# Wails v3 迁移（`wails-v3` 分支）

主线 `main` 保持 v2。所有 v3 工作只在这个分支上。目标是先让 v3 能构建、能起窗口、
`bridge` 层不依赖具体 GUI 工具包，再逐个补齐前端与平台集成。

## 构建

```bash
# v2（与 main 一致，任何时候都应能构建）
~/go/bin/wails build -tags webkit2_41

# v3（同一棵树，用 build tag 切换 main 包）
go build -tags "v3 gtk3" -o /tmp/gfs-v3 .
```

`main.go` 带 `//go:build !v3`，`main_v3.go` 带 `//go:build v3`，两者共享 `bridge` 包。
v3 不需要 `Taskfile.yml`，`go build` 直接出可执行文件；正式打包阶段再加。

Linux 构建标签：`gtk3`（webkit2gtk-4.1，本机 2.54.2）或 `gtk4`（webkitgtk-6.0）。
v2 的 `-tags webkit2_41` 在 v3 下无效。

## 已完成

### 第 1 步：骨架探针（提交 60716a3）

验证了 v3.1.0 在 Arch/WebKitGTK 2.54.2 上：窗口 + `AssetFileServerFS(embed.FS)` 资产、
`wails generate bindings` 生成的绑定调用、`app.Event.Emit` → 前端 `Events.On` 事件往返，
端到端可用。

### 第 2 步：`bridge` 与 GUI 工具包解耦（当前）

- 新增 `bridge/host.go`：`bridge.Host` 结构体 + `SetHost()`，把 bridge 需要的工具包能力收敛成
  7 个回调（`Emit` / `On` / `Off` / `WindowShow` / `WindowHide` / `WindowSetTitle` / `Quit`）。
  未设置的字段是空操作，bridge 在无工具包环境（测试）也能跑。
- `bridge/` 里 22 处 `runtime.*` 调用全部改成 `emit` / `on` / `off` / `windowShow` /
  `windowHide` / `windowSetTitle` / `quitApp`，**不再 import `wails/v2/pkg/runtime`**。
- `main.go` 用 v2 的 `runtime.*` 实现 Host；`main_v3.go` 用 v3 的 `application.App` 实现。
- `main_v3.go`：`application.New(Options{Name, Icon, Services, Assets, SingleInstance,
  ShouldQuit, OnShutdown, LogLevel})` + `Window.NewWithOptions(...)`，服务是
  `application.NewService(app)`——`wails generate bindings` 扫出 **1 Service / 45 Methods**，
  与 v2 `Bind: []any{app}` 完全一致。
- 事件 payload 规范化（`normalizePayload` / `toArgs`，带测试 `main_v3_test.go`）：
  v3 的 `Emit(name, data...)` 在只有一个参数时把 `Data` 设成该值本身、多参数时设成 `[]any`，
  与 v2 的可变参语义不完全一致。单参数的 slice 会被包一层 `[]any{...}`，
  前端 shim 只要按「`ev.data` 是数组就展开、否则单参」处理就能对上 v2。
  `onLaunchApp` 用 `EmitEvent(&CustomEvent{Data: []any{args}})` 保留 v2 的「一个数组参数」语义。
- 托盘仍用 `energye/systray`（只依赖 Host 的两个回调，无需改），`ApplicationStarted` 事件里启动，
  `OnShutdown` 里 `trayEnd()`。
- 验证：v2 构建通过（6.6s）；v3 构建通过；v3 起窗口、资产正常服务；`bridge` 单测 + 主包单测通过。

## v2 → v3 对照

| v2 | v3 |
| --- | --- |
| `wails.Run(&options.App{...})` | `application.New(application.Options{...})` + `app.Run()` |
| `Bind: []any{app}` | `Services: []application.Service{application.NewService(app)}` |
| `OnStartup: func(ctx)` | `app.Event.OnApplicationEvent(events.Common.ApplicationStarted, ...)` |
| `OnBeforeClose: func(ctx) bool` | `ShouldQuit: func() bool`（返回 true 才退出） |
| `OnShutdown: func(ctx)` | `OnShutdown: func()` |
| `options.SingleInstanceLock{UniqueId, OnSecondInstanceLaunch}` | `application.SingleInstanceOptions{UniqueID, OnSecondInstanceLaunch}` |
| `runtime.EventsEmit(ctx, n, d...)` | `app.Event.Emit(n, d...)`（见上面的 payload 差异） |
| `runtime.EventsOn(ctx, n, func(d ...any))` | `app.Event.On(n, func(e *CustomEvent))`，payload 在 `e.Data` |
| `runtime.EventsOff(ctx, n)` | `app.Event.Off(n)` |
| `runtime.WindowShow/Hide/SetTitle(ctx, ...)` | 窗口句柄 `window.Show()/Hide()/SetTitle()` |
| `runtime.Quit(ctx)` | `app.Quit()` |
| `runtime.InitializeNotifications/CleanupNotifications` | `pkg/services/notifications`（未做） |
| `options.App.Menu`（`*menu.Menu`） | `application.Menu` / `LinuxWindow.Menu`（未做） |
| `-tags webkit2_41` | `-tags gtk3` / `gtk4` |
| `wails.json` | `Taskfile.yml` + `build/config.yml`（未做） |

## 还没做（第 3-6 步）

1. **前端 bridge shim**（下一个阻塞点）：`frontend/src/bridge/index.ts` 现在 re-export
   `@wails/runtime/runtime`，需要改成基于 `/wails/runtime.js` 的同名实现
   （`EventsOn(name, cb)` → `Events.On(name, ev => cb(...args))`、`Window*`、`Clipboard*`、
   `Browser*` 等），视图里 88 处调用不动。
   `frontend/index.html` 的 `if (!window.WailsInvoke)` 检测也要换（v3 没有 `WailsInvoke`），
   否则会显示 WebView2 错误页——**这正是当前 v3 启动后看到那个白底错误页的原因**。
2. v3 主题 API 缺失：`WindowSetSystemDefaultTheme` / `SetLightTheme` / `SetDarkTheme`
   在 v3 JS runtime 里没有，需要改成 `common:ThemeChanged` 事件 + CSS。
3. 原生托盘/菜单：`app.SystemTray`、`application.Menu` 替换 `energye/systray` 和 `macOS` 菜单。
4. 通知：`pkg/services/notifications`。
5. 构建与 CI：`Taskfile.yml`、`build/config.yml`，`.github/workflows/release.yml` 三处 `wails build`。
6. `bridge/bridge.go` 里 `createMacOSMenus` 仍用 v2 的 `menu`/`keys` 包——只在 darwin 生效，
   v3 构建里是死代码，第 5 步一起换掉。

## 已知坑

- `go.mod` 同时有 `wails/v2` 和 `wails/v3`，`go mod tidy` 不会改动既有版本，但最终迁移完要删 v2 依赖。
- v3 进程对 SIGTERM 不退出（`PreventExit` 默认 true，`ShouldQuit` 返回 false），
  手动测试要 `kill -9`。这与 v2「关窗即隐藏到托盘」的行为一致，不是新 bug。
- `frontend/dist` 是 `go:embed` 进来的，v3 构建前要先 `pnpm build-only`（与 v2 相同）。
