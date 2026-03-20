package main

import (
	"fmt"
	"os"
)

// Version is set by goreleaser at build time.
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("𓂀 Sirsi Anubis Agent %s\n", version)
		os.Exit(0)
	}

	fmt.Println("𓂀 Sirsi Anubis Agent")
	fmt.Println("  Lightweight agent for fleet deployment")
	fmt.Println()
	fmt.Println("This binary is deployed to remote targets by the anubis controller.")
	fmt.Println("It should not be run directly unless for testing.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  anubis-agent serve    Start agent listener (gRPC)")
	fmt.Println("  anubis-agent scan     Run local scan and output JSON")
	fmt.Println("  anubis-agent version  Show version")
}
