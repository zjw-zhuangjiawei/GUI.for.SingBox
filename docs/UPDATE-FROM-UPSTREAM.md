# 从上游更新（fork 同步流程）

本仓库是 `GUI-for-Cores/GUI.for.SingBox` 的 fork。日常更新走 **merge**，不要 rebase，
这样上游历史保持完整，冲突只需解决一次。

| remote | 指向 |
| --- | --- |
| `origin` | 自己的 fork（`git@github.com:zjw-zhuangjiawei/GUI.for.SingBox.git`） |
| `upstream` | 上游（`git@github.com:GUI-for-Cores/GUI.for.SingBox.git`） |

## 一次性配置

```bash
git remote add upstream git@github.com:GUI-for-Cores/GUI.for.SingBox.git
git config pull.rebase false      # 合并而非变基
```

## 更新流程

```bash
# 1. 确保工作区干净，切到 main
git switch main
git status --short --branch

# 2. 拉取上游（带 --prune 清掉已删除的分支，顺带取 tag）
git fetch upstream --prune

# 3. 先看要合进来什么，再决定合不合
git log --oneline main..upstream/main
git rev-list --count main..upstream/main

# 4. 合并
git merge upstream/main --no-edit

# 5. 构建（wails 会自动跑 pnpm install + 前端构建）
wails build -tags webkit2_41

# 6. 安装并重启
mkdir -p ~/.local/opt/gui-for-singbox
systemctl --user stop 'app-gui\x2dfor\x2dsingbox@autostart.service'
cp build/bin/GUI.for.SingBox ~/.local/opt/gui-for-singbox/GUI.for.SingBox
systemctl --user start 'app-gui\x2dfor\x2dsingbox@autostart.service'

# 7. 推到自己的 fork
git push origin main
```

构建产物在 `build/bin/GUI.for.SingBox`，Linux 上需要 `-tags webkit2_41`
（旧版 WebKitGTK 兼容），见 `BUILD.md`。

## 注意事项

- **构建前不必手动装依赖**：`wails build` 会执行 `wails.json` 里的
  `frontend:install`（`pnpm install`）和 `frontend:build`（`pnpm run build`）。
  想锁定依赖版本时才手动跑 `cd frontend && pnpm install --frozen-lockfile && pnpm build`。
- **`go.mod` / `go.sum` / `frontend/pnpm-lock.yaml` 变化**说明上游升了依赖，正常现象，
  不要回滚这些文件，交给构建步骤重算。
- **冲突高发区**：`frontend/src/bridge/`（wails 自动生成的绑定）与
  `frontend/src/lang/locale/`。绑定文件冲突时直接接受上游版本，重新构建会再次生成。
- **不要修改 `BUILD.md`**，它是上游文件，改动会在下次合并时反复冲突。
  fork 专属说明放本文件。
- **构建耗时长就丢给 pueue**（约 20 秒，通常不必）：

  ```bash
  pueue add --label gfs-build \
    'cd ~/projects/GUI.for.SingBox && wails build -tags webkit2_41 > /tmp/gfs-build.log 2>&1; echo EXIT=$? >> /tmp/gfs-build.log'
  ```
- **只同步某几个上游提交**（跳过其余）：

  ```bash
  git cherry-pick <sha>        # 单个
  git cherry-pick -x <a>^..<b> # 一段
  ```

## 查看上游最新发布

```bash
git tag --sort=-v:refname | head -5
git log --oneline upstream/main | head -10
```
