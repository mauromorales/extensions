package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/mod/semver"
	yaml "gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

func (p *Package) PrintInfo() {
	fmt.Printf("# Checking updates for package %s\n", p.Name)
	fmt.Printf("- Github: %s / %s\n", p.Label("github.owner"), p.Label("github.repo"))
	fmt.Printf("- Autobump Strategy: %s\n", p.Label("autobump.strategy"))
	fmt.Printf("- Autobump Prefix: %s\n", p.Label("autobump.prefix"))
	fmt.Printf("- Autobump reverse dependencies: %v\n", p.AutobumpReverseDependencies())
	fmt.Printf("- Prefix trim: %s\n", p.Label("autobump.trim_prefix"))
	fmt.Printf("- String replace: %s\n", p.Label("autobump.string_replace"))
	fmt.Printf("- Skip if contains: %s\n", p.Label("autobump.skip_if_contains"))
	fmt.Printf("- Consider only if version contains: %s\n", p.Label("autobump.version_contains"))
}

func (p *Package) GetGithubTag() (string, error) {
	apiUrl, _ := url.JoinPath("https://api.github.com/repos", p.Label("github.owner"), p.Label("github.repo"), "tags")
	response, err := http.Get(apiUrl)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request returned non-200 status code: %d\n", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Create a variable to store the data
	var data []GithubTag

	// Unmarshal the JSON response into your data structure
	if err := json.Unmarshal(responseBody, &data); err != nil {
		return "", err
	}

	latestTag := ""
	for _, item := range data {
		if p.Label("autobump.version_contains") != "" {
			if item.Name == p.Label("autobump.version_contains") {
				latestTag = item.Name
			}
		} else {
			if semver.Compare(item.Name, latestTag) > 0 {
				latestTag = item.Name
			}
		}
	}
	return latestTag, nil
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

type GithubTag struct {
	Name string `json:"name"`
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
			pkg.PrintInfo()

			latestTag, err := pkg.GetGithubTag()

			fmt.Printf("Latest version found for %s is: %s. Current at %s\n", pkg.Name, latestTag, pkg.Version)

			switch semver.Compare(pkg.Version, latestTag) {
			case -1:
				fmt.Println("Newer version available")
			case 0:
				fmt.Println("Up to date")
			case 1:
				fmt.Println("Newer version installed")
			}
		}

	}
}
