// Wails v2 runtime API re-implemented on top of the Wails v3 runtime.
// The app's bridge modules and views import these names, so keeping the same
// surface keeps every existing call site untouched.
//
// `wails build` (v2) regenerates this directory and would overwrite this file;
// on the wails-v3 branch build the frontend with `pnpm build-only` instead.
import { Browser, Clipboard, Events, Window } from '@wailsio/runtime'

// v3 delivers a single event object: its payload is the bare value when Go sent
// one argument, or the argument slice when it sent several.
const toArgs = (event) => {
  const data = event?.data
  if (data === undefined || data === null) return []
  return Array.isArray(data) ? data : [data]
}

export const EventsOn = (name, callback) => Events.On(name, (event) => callback(...toArgs(event)))

export const EventsOff = (...names) => Events.Off(...names)

export const EventsEmit = (name, ...data) => {
  Events.Emit(name, data.length > 1 ? data : data[0])
}

export const BrowserOpenURL = (url) => {
  Browser.OpenURL(url)
}

export const ClipboardGetText = () => Clipboard.Text()

export const ClipboardSetText = async (text) => {
  try {
    await Clipboard.SetText(text)
    return true
  } catch {
    return false
  }
}

export const Show = () => {
  Window.Show()
}

export const WindowHide = () => {
  Window.Hide()
}

export const WindowMinimise = () => {
  Window.Minimise()
}

export const WindowToggleMaximise = () => {
  Window.ToggleMaximise()
}

export const WindowIsMaximised = () => Window.IsMaximised()

export const WindowIsMinimised = () => Window.IsMinimised()

export const WindowIsFullscreen = () => Window.IsFullscreen()

export const WindowSetSize = (width, height) => {
  Window.SetSize(width, height)
}

export const WindowGetSize = async () => {
  const { width, height } = await Window.Size()
  return { w: width, h: height }
}

export const WindowSetPosition = (x, y) => {
  Window.SetPosition(x, y)
}

export const WindowGetPosition = () => Window.Position()

export const WindowSetAlwaysOnTop = (alwaysOnTop) => {
  Window.SetAlwaysOnTop(alwaysOnTop)
}

export const WindowReloadApp = () => {
  Window.Reload()
}

// v3 dropped the v2 theme setters; the theme is driven by CSS and the
// common:ThemeChanged event, so this is intentionally a no-op.
export const WindowSetSystemDefaultTheme = () => {}

// v3's JS runtime has no notification API, so use the Web Notifications API,
// which the platform webviews surface as native notifications.
export const IsNotificationAvailable = async () => typeof Notification !== 'undefined'

export const RequestNotificationAuthorization = async () => {
  if (typeof Notification === 'undefined') return false
  if (Notification.permission === 'granted') return true
  return (await Notification.requestPermission()) === 'granted'
}

export const SendNotification = async (options) => {
  new Notification(options.title, { body: options.body })
}
