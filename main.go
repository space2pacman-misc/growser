package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"growser/css"
	"growser/dom"
	"growser/layout"
	"growser/net"
	"growser/render"
	"growser/ui"
)

func main() {
	useUI := flag.Bool("ui", false, "open interactive window")
	debug := flag.Bool("debug", false, "draw layout boxes")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("usage: growser [-ui] [-debug] <url>")
		os.Exit(1)
	}
	url := args[0]

	if *useUI {
		app, err := ui.NewApp(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "init error: %v\n", err)
			os.Exit(1)
		}
		ebiten.SetWindowSize(ui.WindowWidth, ui.WindowHeight)
		ebiten.SetWindowTitle("Growser — " + url)
		if err := ebiten.RunGame(app); err != nil {
			fmt.Fprintf(os.Stderr, "ui error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// PNG-режим
	body, err := net.Fetch(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch error: %v\n", err)
		os.Exit(1)
	}
	root, err := dom.Parse(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	if title := root.Find("title"); title != nil {
		fmt.Printf("Title: %s\n", title.TextContent())
	}

	var cssText strings.Builder
	var collect func(n *dom.Node)
	collect = func(n *dom.Node) {
		if n.Tag == "style" {
			cssText.WriteString(n.TextContent())
			cssText.WriteString("\n")
		}
		for _, c := range n.Children {
			collect(c)
		}
	}
	collect(root)
	rules := css.ParseStylesheet(cssText.String())
	fmt.Printf("CSS rules: %d\n", len(rules))

	box := layout.Layout(root, rules)
	fmt.Printf("Page height: %.0fpx\n", pageHeight(box))

	render.Debug = *debug
	if err := render.Render(box, "output.png"); err != nil {
		fmt.Fprintf(os.Stderr, "render error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Wrote output.png")
}

func pageHeight(b *layout.Box) float64 {
	maxY := b.Y + b.Height + b.PaddingBottom + b.MarginBottom
	for _, c := range b.Children {
		if h := pageHeight(c); h > maxY {
			maxY = h
		}
	}
	return maxY
}