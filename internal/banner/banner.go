package banner // or package main depending on your structure

import (
	"github.com/fatih/color"
)

func PrintBanner() {
	// Define color styles using fatih/color
	cyan := color.New(color.FgCyan, color.Bold)
	dimCyan := color.New(color.FgHiBlack)
	yellow := color.New(color.FgYellow)
	white := color.New(color.FgWhite, color.Bold)

	// Print ASCII art
	println()
	cyan.Println(`  ███████╗ ██████╗  ██████╗ ████████╗██████╗ ██████╗ ██╗███╗   ██╗████████╗███████╗`)
	cyan.Println(`  ██╔════╝██╔═══██╗██╔═══██╗╚══██╔══╝██╔══██╗██╔══██╗██║████╗  ██║╚══██╔══╝██╔════╝`)
	cyan.Println(`  █████╗  ██║   ██║██║   ██║   ██║   ██████╔╝██████╔╝██║██╔██╗ ██║   ██║   ███████╗`)
	cyan.Println(`  ██╔══╝  ██║   ██║██║   ██║   ██║   ██╔═══╝ ██╔══██╗██║██║╚██╗██║   ██║   ╚════██║`)
	cyan.Println(`  ██║     ╚██████╔╝╚██████╔╝   ██║   ██║     ██║  ██║██║██║ ╚████║   ██║   ███████║`)
	cyan.Println(`  ╚═╝      ╚═════╝  ╚═════╝    ╚═╝   ╚═╝     ╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝   ╚═╝   ╚══════╝`)

	// Tagline with an icon
	yellow.Print("                        🐾 ")
	dimCyan.Print("a local, privacy-first Chrome history graph visualizer\n\n")

	white.Printf("  Author: ")
	dimCyan.Printf("Mahendra Dani (https://mahendradani.vercel.app)\n")
	white.Printf("  Usage:  ")
	cyan.Printf("footprints [command] --help\n\n")
}
