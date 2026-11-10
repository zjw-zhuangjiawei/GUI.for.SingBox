# Wails v3 迁移探针（PoC）

独立的最小 v3 应用，用来验证迁移可行性。**不导入 `bridge/` 包**——那个包仍然对接 v2 runtime。
主线 `main` 保持 v2，本目录只在 `wails-v3` 分支存在。

## 跑起来

```bash
go build -tags gtk3 -o /tmp/gfs-v3poc ./v3poc
/tmp/gfs-v3poc
```

不需要 `Taskfile.yml`，也不需要 `wails3 build`——直接 `go build` 即可（`gtk3` 标签对应本机
`webkit2gtk-4.1`；要 GTK4 就用 `gtk4` 标签 + `webkitgtk-6.0`）。

## 已验证（2026-11-11，Wails v3.1.0，Arch/WebKitGTK 2.54.2）

1. **窗口 + 资产**：`application.New` + `app.Window.NewWithOptions` + `AssetFileServerFS(embed.FS)`
   正常，页面加载。
2. **绑定生成**：`wails generate bindings -b -d v3poc/frontend/bindings ./v3poc`
   生成 `bindings/guiforcores/v3poc/pocservice.js`，2 个方法。
   `-b`（bundled runtime）让生成代码 import `/wails/runtime.js`，**不需要 npm 装 `@wailsio/runtime`**，
   纯 HTML + ES module 就能跑。
3. **方法调用**：`PocService.Echo("hello")`、`Add(2,40)` 前后端往返成功（Go 日志可见）。
4. **事件**：Go 端 `app.Event.Emit("poc:tick", map[string]any{...})` 每秒发一次，
   前端 `Events.On("poc:tick", (ev) => ev.data.count)` 收到后回调 Go，日志连续 7 次无丢失。
5. **v2 与 v3 共存**：`go.mod` 同时存在 `wails/v2` 与 `wails/v3`，v2 构建（`~/go/bin/wails build
   -tags webkit2_41`）仍正常，`go mod tidy` 只新增 4 行依赖、未改动已有版本。

## 关键差异（本 PoC 实测 + 源码核对）

| v2 | v3 |
| --- | --- |
| `wails.Run(&options.App{...})` | `application.New(application.Options{...})` + `app.Run()` |
| `Bind: []any{app}` | `Services: []application.Service{application.NewService(app)}` |
| 前端 `EventsOn(n, (...args)=>…)` | `Events.On(n, (ev)=>ev.data)`，**单对象**，参数在 `ev.data` |
| `runtime.EventsEmit(ctx, n, data...)` | `app.Event.Emit(n, data...)`，多参数打包成 `[]any` |
| `runtime.WindowShow(ctx)` | 窗口句柄 `window.Show()` |
| `-tags webkit2_41` | `-tags gtk3`（webkit2gtk-4.1）或 `gtk4`（webkitgtk-6.0） |
| `wails.json` 驱动构建 | `Taskfile.yml` + `build/config.yml`（PoC 阶段可绕过） |

## 下一步（尚未做）

1. 把 `bridge.App` 的 45 个方法挂到 v3 service 上，替换 `runtime.*` 33 处调用（14 `EventsEmit` / 4 `EventsOn`）
2. 事件 payload 规范化：`bridge/server.go:182` 一次发 6 个参数，需要定成对象还是数组
3. 前端 `frontend/src/bridge/index.ts` 用 v3 函数写同名 shim，视图 88 处调用不动
4. 托盘：原生 `app.SystemTray` 替换 `energye/systray`，或保留 systray 只改 2 处
5. 菜单 `bridge/bridge.go:242-254`、通知 `pkg/services/notifications`
6. 构建/CI：`Taskfile.yml` + `release.yml` 三平台 job
