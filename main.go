package main

import (
	"fmt"
	"os"

	"noodles/cmd"
)

const version = "0.3"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		cmd.RunDaemon()
	case "store":
		cmd.RunStore()
	case "list":
		cmd.RunList()
	case "get":
		// Pass remaining args (index) to get command
		cmd.RunGet(os.Args[2:])
	case "wipe":
		cmd.RunWipe()
	case "stop":
		cmd.RunStop()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("Noodles v%s - Wayland clipboard manager\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  noodles start          Start the daemon")
	fmt.Println("  noodles store          Store clipboard data from stdin")
	fmt.Println("  noodles list           List all clipboard items")
	fmt.Println("  noodles get <index>    Get clipboard item by index")
	fmt.Println("  noodles wipe           Clear all clipboard items")
	fmt.Println("  noodles stop           Stop the daemon")
	fmt.Println("  noodles help           Show this help message")
	fmt.Println()
	fmt.Println("Integration with wl-paste:")
	fmt.Println("  wl-paste --watch noodles store")
	fmt.Println()
	fmt.Println("Select and copy clipboard item:")
	fmt.Println("  noodles list | rofi -dmenu | cut -f1 | noodles get | wl-copy")
}
