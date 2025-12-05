package main

import (
	"fmt"
	"os"

	"waypasta/cmd"
)

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
	fmt.Println("WayPasta - Wayland clipboard manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  waypasta start          Start the daemon")
	fmt.Println("  waypasta store          Store clipboard data from stdin")
	fmt.Println("  waypasta list           List all clipboard items")
	fmt.Println("  waypasta get <index>    Get clipboard item by index")
	fmt.Println("  waypasta wipe           Clear all clipboard items")
	fmt.Println("  waypasta stop           Stop the daemon")
	fmt.Println("  waypasta help           Show this help message")
	fmt.Println()
	fmt.Println("Integration with wl-paste:")
	fmt.Println("  wl-paste --watch waypasta store")
	fmt.Println()
	fmt.Println("Select and copy clipboard item:")
	fmt.Println("  waypasta list | rofi -dmenu | cut -f1 | waypasta get | wl-copy")
}
