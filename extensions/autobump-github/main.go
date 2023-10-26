package main

import (
	"fmt"
	"os"
)

func main() {
	treeDir := os.Getenv("TREE_DIR")
	if treeDir == "" {
		fmt.Println("TREE_DIR is not set")
		return
	}

	entries, err := os.ReadDir(treeDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("# Checking updates for package %s\n", entry.Name())
		}
	}
}
