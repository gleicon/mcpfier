package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/gleicon/mcpfier/internal/config"
)

func main() {
	var (
		generate = flag.Bool("generate", false, "Generate a new random API key")
		hash     = flag.String("hash", "", "Hash an existing API key")
		length   = flag.Int("length", 32, "Length of generated key in bytes")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "MCPFier API Key Utility\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  Generate new API key:  %s -generate [-length 32]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Hash existing key:     %s -hash <api-key>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *generate {
		generateAPIKey(*length)
	} else if *hash != "" {
		hashAPIKey(*hash)
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func generateAPIKey(length int) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating random bytes: %v\n", err)
		os.Exit(1)
	}

	// Encode as URL-safe base64
	key := base64.URLEncoding.EncodeToString(bytes)

	// Remove padding
	key = key[:len(key)-2]

	fmt.Println("Generated API Key:")
	fmt.Println("==================")
	fmt.Printf("Plain text: %s\n\n", key)

	// Also show hash
	hash := config.HashAPIKey(key)
	fmt.Println("Hashed version (SHA-256):")
	fmt.Printf("hash:%s\n\n", hash)

	fmt.Println("Configuration examples:")
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("1. Plain text in config.yaml:")
	fmt.Println("   api_keys:")
	fmt.Printf("     - key: \"%s\"\n", key)
	fmt.Println("       name: \"client-name\"")
	fmt.Println("       permissions: [\"*\"]")
	fmt.Println()
	fmt.Println("2. Hashed in config.yaml:")
	fmt.Println("   api_keys:")
	fmt.Printf("     - key: \"hash:%s\"\n", hash)
	fmt.Println("       name: \"client-name\"")
	fmt.Println("       permissions: [\"*\"]")
	fmt.Println()
	fmt.Println("3. Environment variable in config.yaml:")
	fmt.Println("   api_keys:")
	fmt.Println("     - key: \"${MCPFIER_API_KEY}\"")
	fmt.Println("       name: \"client-name\"")
	fmt.Println("       permissions: [\"*\"]")
	fmt.Println()
	fmt.Printf("   Then set: export MCPFIER_API_KEY=\"%s\"\n", key)
}

func hashAPIKey(key string) {
	hash := config.HashAPIKey(key)

	fmt.Println("API Key Hash:")
	fmt.Println("=============")
	fmt.Printf("Input:  %s\n", key)
	fmt.Printf("SHA-256: %s\n\n", hash)

	fmt.Println("Configuration:")
	fmt.Println("==============")
	fmt.Println("api_keys:")
	fmt.Printf("  - key: \"hash:%s\"\n", hash)
	fmt.Println("    name: \"client-name\"")
	fmt.Println("    permissions: [\"*\"]")
	fmt.Println()
	fmt.Println("Note: Store this hash in your config.yaml")
	fmt.Println("      Clients will use the original key for authentication")
}
