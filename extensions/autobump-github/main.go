package main

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type Package struct {
	Name     string            `yaml:"name"`
	Category string            `yaml:"category"`
	Version  string            `yaml:"version"`
	Labels   map[string]string `yaml:"labels"`
}

func (p *Package) Label(key string) string {
	return p.Labels[key]
}

func (p *Package) AutobumpReverseDependencies() bool {
	revDeps := p.Label("autobump.reverse_dependencies")
	return revDeps == "true" || revDeps == ""
}

func readDefinitionFile(pkg *Package, path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(file, &pkg); err != nil {
		return err
	}

	return nil
}

func readPackage(dir string) (*Package, error) {
	pkg := &Package{}

	err := readDefinitionFile(pkg, filepath.Join(dir, "definition.yaml"))
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

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
			pkg, err := readPackage(filepath.Join(treeDir, entry.Name()))
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			fmt.Printf("# Checking updates for package %s\n", pkg.Name)
			fmt.Printf("- Github: %s / %s\n", pkg.Label("github.owner"), pkg.Label("github.repo"))
			fmt.Printf("- Autobump Strategy: %s\n", pkg.Label("autobump.strategy"))
			fmt.Printf("- Autobump Prefix: %s\n", pkg.Label("autobump.prefix"))
			fmt.Printf("- Autobump reverse dependencies: %s\n", pkg.AutobumpReverseDependencies())
		}
	}
}
