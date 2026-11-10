import { Events } from '/wails/runtime.js'
import { PocService } from './bindings/guiforcores/v3poc/index.js'

const log = document.getElementById('log')
const lines = []

function add(line) {
  lines.push(line)
  log.textContent = lines.join('\n')
}

// 1. Bound Go methods, with args and a non-string return.
add(`Echo("hello") -> ${await PocService.Echo('hello')}`)
add(`Add(2, 40)      -> ${await PocService.Add(2, 40)}`)

// 2. Event from Go: v3 delivers a single event object, payload in ev.data.
Events.On('poc:tick', (ev) => {
  add(`poc:tick #${ev.data.count} at ${ev.data.at}`)
  // round-trip back to Go: proves the event was really received here
  PocService.Echo(`tick ${ev.data.count}`)
})
