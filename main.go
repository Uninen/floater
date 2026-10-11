package main

import (
	"log"

	"github.com/egoist/mygo"
)

func main() {
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "Floater",
			Width:     1024,
			Height:    768,
			URL:       "https://www.di.fm",
			Frameless: true,
			Page: mygo.PageOptions{
				PreloadScript: `
					document.addEventListener("DOMContentLoaded", () => {
						const style = document.createElement("style");
						style.textContent = "* { scrollbar-width: none !important; } *::-webkit-scrollbar { display: none !important; }";
						document.head.append(style);
					}, { once: true });
				`,
			},
		})
		enableOptionDrag(win)
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
