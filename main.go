package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"starora/embedded"
	"starora/internal/app"
)

func main() {
	a := app.New()
	mygo.App.WhenReady(func() {
		a.Win = mygo.NewWindow(mygo.WindowOptions{
			Title:     "Starora Studio - AI Upscaler & Video Super-Resolution",
			Width:     1120,
			Height:    760,
			MinWidth:  900,
			MinHeight: 600,
			StateKey:  "main",
			Content:   ui.View(a.View),
		})
		if len(embedded.IconPNG) > 0 {
			_ = a.Win.SetIcon(embedded.IconPNG)
		}
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
