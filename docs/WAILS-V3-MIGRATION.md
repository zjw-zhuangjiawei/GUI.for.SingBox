# Wails v3 迁移（`wails-v3` 分支）

主线 `main` 保持 v2。**这个分支已经完全切到 v3**：`wails/v2` 依赖、v2 入口、v2 CLI 配置都删掉了，
`main.go` 就是 v3 入口。目标是先让 v3 能构建、能起窗口、`bridge` 层不依赖具体 GUI 工具包，
再逐个补齐平台集成。

## 构建

```bash
# 前端（dist 是 go:embed 进来的，必须先建）
cd frontend && pnpm install && pnpm build-only && cd ..

# Linux
go build -tags gtk3 -o GUI.for.SingBox .
```

v3 不需要 `Taskfile.yml`，`go build` 直接出可执行文件；正式打包阶段再加。

Linux 构建标签：`gtk3`（webkit2gtk-4.1，本机 2.54.2）或 `gtk4`（webkitgtk-6.0）。
v2 的 `-tags webkit2_41` 在 v3 下无效。

## 已完成

### 第 1 步：骨架探针（提交 60716a3）

验证了 v3.1.0 在 Arch/WebKitGTK 2.54.2 上：窗口 + `AssetFileServerFS(embed.FS)` 资产、
`wails generate bindings` 生成的绑定调用、`app.Event.Emit` → 前端 `Events.On` 事件往返，
端到端可用。

### 第 2 步：`bridge` 与 GUI 工具包解耦（提交 a83e9e5）

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

### 第 3 步：前端 v2 runtime shim

前端只动 `frontend/src/bridge/wailsjs/` 这一个目录，`src/bridge/*.ts` 和视图里 88 处调用一行没改：

- `runtime/runtime.js` 用 `@wailsio/runtime`（npm，3.1.0）实现 v2 的 runtime 同名函数，
  只实现实际用到的 24 个。`EventsOn(name, cb)` 把 v3 的单个事件对象展开成 v2 的可变参；
  `WindowGetSize` 把 `{width,height}` 映射回 `{w,h}`；`ClipboardSetText` 保留成功布尔返回。
- `runtime/runtime.d.ts` 只声明上面这些函数，未实现的 v2 API 不再声明——上游新增用法会在
  `pnpm type-check` 阶段报错，而不是运行时静默失效。
- `go/bridge/App.js` 改为 `export * from '../../guiforcores/bridge/app.js'`，保留
  `@wails/go/bridge/App` 这个 import 路径；删除 v2 生成的 `App.d.ts` 和 `go/models.ts`。
- 生成绑定（用 **v3 CLI** `/usr/bin/wails`）：`wails generate bindings -f "-tags gtk3" -clean=false -d frontend/src/bridge/wailsjs .`
  （**不要加 `-b`**，加了会 import `/wails/runtime.js`，vite 解析不到）。
  生成器输出 JSDoc 类型的 `.js` 而不是 `.d.ts`，所以 `tsconfig.app.json` 需要 `allowJs: true`。
- `frontend/index.html` 的 IPC 检测从 `window.WailsInvoke` 改成 `window._wails`。
- 通知：`@wailsio/runtime` 没有通知 API，三个函数用 Web Notification API 实现。
- 主题：`WindowSetSystemDefaultTheme` 在 v3 没有对应物，暂时是空函数（第 5 步换成
  `common:ThemeChanged` + CSS）。
- 验证：`pnpm type-check` 0 错误、`pnpm build-only` 通过；v3 二进制启动后日志出现
  `GetEnv`、`ReadFile data/*.yaml`、`GetSystemProxyBypass`、`UpdateTrayAndMenus`，
  说明 Vue 界面在 v3 下完整跑起来，绑定与事件都通。

### 第 4 步：通知走 v3 service，托盘/主题核对

- 通知：v3 有 `pkg/services/notifications`（Linux 走 D-Bus，macOS 走 UNUserNotificationCenter，
  Windows 走 wintoast）。`main_v3.go` 里 `application.NewService(notifications.New())` 注册成
  第二个 service，`wails generate bindings` 后前端 shim 的
  `IsNotificationAvailable` / `RequestNotificationAuthorization` / `SendNotification`
  改为调用生成的 `NotificationService`，不再用 Web Notification API。
  实测：启动 6 秒后发一条通知，`SendNotification` 返回 `nil`。
- 托盘：`energye/systray` 在 v3 的 GTK 主循环下**照常工作**，无需改成 `app.SystemTray`。
  实测 v3 进程注册了 `org.kde.StatusNotifierItem-<pid>-1`，且出现在
  `org.kde.StatusNotifierWatcher.RegisteredStatusNotifierItems` 里。菜单回调本来就走
  `bridge.App` 的方法，与工具包无关。
- 主题：前端的主题切换是 `document.body.setAttribute('theme-mode', ...)` 的纯 CSS 方案，
  v2 的 `WindowSetSystemDefaultTheme()` 只是通知原生控件跟随系统；v3 没有对应 API，
  shim 里的空函数就是正解，不需要 `common:ThemeChanged`。

### 第 5 步：bridge 彻底不碰 v2，CI 换成 go build

- `bridge/` 里最后两个 v2 类型清掉：`App.AppMenu *menu.Menu` 删掉，macOS 菜单改由各入口点自己建；
  `Env.WindowStartState` 从 `options.WindowStartState` 改成 `bridge.WindowStartNormal/Minimised`
  两个字符串常量，由 `main.go` / `main_v3.go` 各自映射。
  现在 `grep -r wails/v2 bridge/` 为空——**bridge 包完全不依赖 GUI 工具包**。
- macOS 菜单：`main.go` 用 v2 的 `menu`/`keys` 建（`createMacOSMenus(app)`），
  `main_v3.go` 用 `application.NewMenu()` + `AddSubmenu` + `SetAccelerator("cmdorctrl+s")` +
  `AddRole(application.EditMenu)` 建同构菜单，在 `ApplicationStarted` 里
  `v3app.Menu.SetApplicationMenu(...)`。只在 `Env.OS == "darwin"` 时执行。
- CI（`.github/workflows/release.yml`）：三个 job 的 `go install wails v2` 和 `wails build`
  全部删掉，改成 `go build`：
  - Windows：`go build -trimpath -ldflags="-s -w -H windowsgui" -o build/bin/APP.exe .`
  - Linux：`go build -tags gtk3 -trimpath -ldflags="-s -w" -o build/bin/APP .`
  - macOS：手动拼 `.app` 目录 + `build/darwin/Info.static.plist`（新增的静态 plist，
    `APP_NAME` / `APP_VERSION` 由 CI sed 替换），然后 `go build`。
  - 顺手删掉 macOS job 里对 v2 `AppDelegate.m` 的 `go mod vendor` + sed 补丁。
- **没有加 `Taskfile.yml` / `build/config.yml`**：v3 的 `go build` 已经够用，Taskfile 只是
  `wails3` CLI 打包（`.icns`、`.syso`、安装包）才需要，等真要做安装包再加。
- 本地能验证的：Windows 交叉编译通过（`GOOS=windows go build`）、Linux 构建通过、
  v3 二进制启动无回归。**macOS 那段 CI 没法在本地验证**（需要 macOS SDK 和 cgo），
  首次真跑 release 时要盯着。

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

### 第 6 步：删掉 v2（提交见下）

- `git rm main.go`（v2 入口），`main_v3.go` → `main.go`、`main_v3_test.go` → `main_test.go`，
  去掉 `//go:build v3` 标签；CI 的三个 `-tags v3` 相应改成无标签（Windows/macOS）和 `-tags gtk3`（Linux）。
- `go mod tidy` 后 `go.mod` 只剩 `github.com/wailsapp/wails/v3 v3.1.0`，`wails/v2` 及其依赖全部消失。
- 删 `wails.json`（v2 CLI 的配置，v3 的 `go build` 不读它；前端构建由 pnpm 直接跑）。
- **这条分支现在没有 v2 兜底了**。要回到 v2 就用 `main` 分支。

## 还没做

1. 安装包：`.icns` / `.syso` / `.desktop` / winget 需要 `Taskfile.yml` + `wails3 package` 才有。
2. 前端 Windows/macOS 行为（透明窗口、Mica/Acrylic、macOS 隐藏标题栏）需要上机验证。
3. 托盘菜单点击、通知点击回调只在 Linux 验证过。
4. macOS 的 `.app` 没有图标（`CFBundleIconFile` 从静态 plist 里去掉了，仓库里没有 `.icns`）。
5. `build/darwin/Info.plist`（Go template）与 `build/windows/{info.json,wails.exe.manifest}` 是 v2 打包
   残留，现在没人读，留给以后做安装包时参考。

## 已知坑

- 第 6 步之前 `go.mod` 同时有 `wails/v2` 和 `wails/v3`；现在只剩 v3。若上游 `main` 又改了 v2 入口，合并时会冲突在 `main.go`（见下）。
- v3 进程对 SIGTERM 不退出（`PreventExit` 默认 true，`ShouldQuit` 返回 false），
  手动测试要 `kill -9`。这与 v2「关窗即隐藏到托盘」的行为一致，不是新 bug。
- `frontend/dist` 是 `go:embed` 进来的，v3 构建前要先 `pnpm build-only`（与 v2 相同）。
- **这个分支上不要再跑 `~/go/bin/wails build`**：v2 CLI 会用 v2 的 Go 代码重新生成
  `frontend/src/bridge/wailsjs/`（`runtime/*`、`go/bridge/App.js`、`go/models.ts`），
  把 v3 的前端 shim 覆盖掉，前端随即 type-check 失败。v3 的构建路径是
  `pnpm build-only` + `go build -tags gtk3`。
- 本分支的前端只服务 v3（`window._wails` 检测）；v2 编译出来的包 UI 会显示「WebView2 未安装」兜底页。
- 会话锁屏时截图拿不到窗口内容（KWin 不渲染 webview 表面），只能看到白底；验证 UI 要看日志或解锁后看屏幕。
