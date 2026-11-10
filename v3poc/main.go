// Command v3poc is a Wails v3 migration probe.
// It is deliberately standalone: it does not import the existing bridge package,
// because that package still talks to the Wails v2 runtime.
package main

import (
	"embed"
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend
var assets embed.FS

// PocService exercises the two things we care about: bound method calls
// (with arguments and a non-string return) and event delivery.
type PocService struct{}

func (p *PocService) Echo(text string) string {
	log.Printf("Echo(%q)", text)
	return "echo: " + text
}

func (p *PocService) Add(a, b int) int {
	return a + b
}

func main() {
	app := application.New(application.Options{
		Name:        "GUI.for.SingBox v3 PoC",
		Description: "Wails v3 migration probe",
		Services: []application.Service{
			application.NewService(&PocService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Wails v3 PoC",
		Width:            700,
		Height:           420,
		BackgroundColour: application.NewRGB(24, 24, 27),
		URL:              "/",
	})

	go func() {
		for i := 0; ; i++ {
			app.Event.Emit("poc:tick", map[string]any{"count": i, "at": time.Now().Format("15:04:05")})
			time.Sleep(time.Second)
		}
	}()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
