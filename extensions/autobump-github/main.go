package main

import (
	"encoding/json"
	"fmt"
	"github.com/mudler/luet/pkg/api/client/utils"
	"github.com/mudler/luet/pkg/api/core/types"
	"golang.org/x/mod/semver"
	yaml "gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Definition struct {
	Package *types.Package
	Path    string
}

func AutobumpReverseDependencies(p *types.Package) bool {
	revDeps := p.Labels["autobump.reverse_dependencies"]
	return revDeps == "true" || revDeps == ""
}

func PrintInfo(p *types.Package) {
	fmt.Printf("# Checking updates for package %s\n", p.Name)
	fmt.Printf("- Github: %s / %s\n", p.Labels["github.owner"], p.Labels["github.repo"])
	fmt.Printf("- Autobump Strategy: %s\n", p.Labels["autobump.strategy"])
	fmt.Printf("- Autobump Prefix: %s\n", p.Labels["autobump.prefix"])
	fmt.Printf("- Autobump reverse dependencies: %v\n", AutobumpReverseDependencies(p))
	fmt.Printf("- Prefix trim: %s\n", p.Labels["autobump.trim_prefix"])
	fmt.Printf("- String replace: %s\n", p.Labels["autobump.string_replace"])
	fmt.Printf("- Skip if contains: %s\n", p.Labels["autobump.skip_if_contains"])
	fmt.Printf("- Consider only if version contains: %s\n", p.Labels["autobump.version_contains"])
}

func GetGithubTag(p *types.Package) (string, error) {
	apiUrl, _ := url.JoinPath("https://api.github.com/repos", p.Labels["github.owner"], p.Labels["github.repo"], "tags")
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
		if p.Labels["autobump.version_contains"] != "" {
			if item.Name == p.Labels["autobump.version_contains"] {
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

func NewDefinition(path string) (*Definition, error) {
	definition := &Definition{
		Package: &types.Package{},
		Path:    path,
	}

	err := readDefinitionFile(definition.Package, filepath.Join(path, "definition.yaml"))
	if err != nil {
		return definition, err
	}

	return definition, nil
}

func readDefinitionFile(pkg *types.Package, path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(file, &pkg); err != nil {
		return err
	}

	return nil
}

type GithubTag struct {
	Name string `json:"name"`
}

func yqReplace(file, key, value string) {
	utils.RunSH("", fmt.Sprintf("yq w -i %s %s %s --style double", file, key, value))
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
			definition, err := NewDefinition(filepath.Join(treeDir, entry.Name()))
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			PrintInfo(definition.Package)

			latestTag, err := GetGithubTag(definition.Package)
			latestVersion := strings.Replace(latestTag, "v", "", 1)

			fmt.Printf("Latest version found for %s is: %s. Current at %s\n", definition.Package.Name, latestVersion, definition.Package.Version)

			switch semver.Compare(definition.Package.Version, latestTag) {
			case -1:
				fmt.Printf("Bumping %s/%s to %s\n", definition.Package.Category, definition.Package.Name, latestVersion)
				if definition.Package.Labels["autobump.strategy"] == "github_tag" {
					yqReplace(filepath.Join(definition.Path, "definition.yaml"), "labels.\\\"github.tag\\\"", latestTag)
				}
				yqReplace(filepath.Join(definition.Path, "definition.yaml"), "version", latestVersion)
			case 0:
				fmt.Println("Up to date")
			case 1:
				fmt.Println("Newer version installed")
			}
		}

	}
}
