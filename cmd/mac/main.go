package main

import (
	"context"
	"fmt"
	"time"

	"github.com/FergusInLondon/PRList/internal/icons"
	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Empty!")
	systray.SetIcon(icons.CircleRegular)
	systray.SetTooltip("What's a tooltip?")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-mQuit.ClickedCh
		fmt.Println("Requesting quit")
		cancel()
		systray.Quit()
		fmt.Println("Finished quitting")
	}()

	go toggle(ctx)

	// Sets the icon of a menu item. Only available on Mac and Windows.
	mQuit.SetIcon(icons.CircleSolid)
}

func toggle(ctx context.Context) {
	def := true
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(5) * time.Second):
			if def {
				systray.SetTitle("Empty!")
				systray.SetIcon(icons.CircleRegular)
			} else {
				systray.SetTitle("Full!")
				systray.SetIcon(icons.CircleSolid)
			}
		}

		def = !def
	}
}
func onExit() {
	// clean up here
}
