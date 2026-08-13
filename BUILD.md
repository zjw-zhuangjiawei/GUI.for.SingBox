# Build & Install

## Prerequisites

- **Node.js** — [nodejs.org](https://nodejs.org/)
- **pnpm** — `npm i -g pnpm`
- **Go** — [go.dev](https://go.dev/)
- **Wails** — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Build

```bash
# Install frontend dependencies and build
cd frontend
pnpm install --frozen-lockfile && pnpm build
cd ..

# Build the Go binary (webkit2_41 tag for older WebKitGTK compatibility)
wails build -tags webkit2_41
```

The binary will be at `build/bin/GUI.for.SingBox`.

## Install

```bash
mkdir -p ~/.local/opt/
cp build/bin/GUI.for.SingBox ~/.local/opt/
```

## Optional: add to PATH

```bash
ln -sf ~/.local/opt/GUI.for.SingBox ~/.local/bin/gfs
```

## Optional: register .desktop entry

Create `~/.local/share/applications/gui-for-singbox.desktop`:

```ini
[Desktop Entry]
Name=GUI.for.SingBox
Comment=GUI for SingBox
Exec=/home/yourusername/.local/opt/GUI.for.SingBox
Icon=network-server
Terminal=false
Type=Application
Categories=Network;
```

## Update

```bash
git pull origin main
cd frontend && pnpm install --frozen-lockfile && pnpm build && cd ..
wails build -tags webkit2_41
cp build/bin/GUI.for.SingBox ~/.local/opt/
```

## Uninstall

```bash
rm ~/.local/opt/GUI.for.SingBox
rm ~/.local/share/applications/gui-for-singbox.desktop   # if created
rm ~/.local/bin/gfs                                      # if linked
```
