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
	"time"
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

func GetGitHubRelease(p *types.Package) (string, error) {
	apiUrl, err := url.JoinPath("https://api.github.com/repos", p.Labels["github.owner"], p.Labels["github.repo"], "releases", "latest")
	if err != nil {
		return "", err
	}
	responseBody, err := getGitHubAPI(apiUrl)
	if err != nil {
		return "", err
	}

	var data GithubRelease

	if err := json.Unmarshal(responseBody, &data); err != nil {
		return "", err
	}

	return data.TagName, nil
}

func GetGitHubRefs(p *types.Package) (string, error) {
	apiUrl, err := url.JoinPath("https://api.github.com/repos", p.Labels["github.owner"], p.Labels["github.repo"], "git", "refs", "tags")
	if err != nil {
		return "", err
	}
	responseBody, err := getGitHubAPI(apiUrl)
	if err != nil {
		return "", err
	}

	var data []GitHubRef

	if err := json.Unmarshal(responseBody, &data); err != nil {
		return "", err
	}

	latestTag := ""
	for _, item := range data {
		ref := strings.Replace(item.Ref, "refs/tags/", "", 1)

		if p.Labels["autobump.version_contains"] != "" {
			if ref == p.Labels["autobump.version_contains"] {
				latestTag = ref
			}
		} else {
			if semver.Compare(ref, latestTag) > 0 {
				latestTag = ref
			}
		}
	}
	return latestTag, nil
}

func GetGitHubHeads(p *types.Package) (string, error) {
	branch := "master"
	if p.Labels["github.branch"] != "" {
		branch = p.Labels["github.branch"]
	}
	apiUrl, err := url.JoinPath("https://api.github.com/repos", p.Labels["github.owner"], p.Labels["github.repo"], "git", "refs", "heads", branch)
	if err != nil {
		return "", err
	}
	responseBody, err := getGitHubAPI(apiUrl)
	if err != nil {
		return "", err
	}

	var data GitHubRef

	if err := json.Unmarshal(responseBody, &data); err != nil {
		return "", err
	}

	return data.Object.Sha, nil
}

func GetGitHubReleaseTag(p *types.Package) (string, error) {
	apiUrl, err := url.JoinPath("https://api.github.com/repos", p.Labels["github.owner"], p.Labels["github.repo"], "releases", "tags", p.Labels["github.tag"])
	if err != nil {
		return "", err
	}
	responseBody, err := getGitHubAPI(apiUrl)
	if err != nil {
		return "", err
	}

	var data GithubRelease

	if err := json.Unmarshal(responseBody, &data); err != nil {
		return "", err
	}

	return data.TagName, nil
}

func GetGitHubTag(p *types.Package) (string, error) {
	apiUrl, err := url.JoinPath("https://api.github.com/repos", p.Labels["github.owner"], p.Labels["github.repo"], "tags")
	if err != nil {
		return "", err
	}
	responseBody, err := getGitHubAPI(apiUrl)
	if err != nil {
		return "", err
	}

	var data []GitHubTag

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

func getGitHubAPI(url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	token := os.Getenv("TOKEN")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request returned non-200 status code: %d\n", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return responseBody, nil
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

type GitHubTag struct {
	Name string `json:"name"`
}

type GithubRelease struct {
	TagName string `json:"tag_name"`
}

type GitHubRef struct {
	Ref    string `json:"ref"`
	Object struct {
		Sha string `json:"sha"`
	} `json:"object"`
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

			updateSrc := true
			latestTag := ""
			switch definition.Package.Labels["autobump.strategy"] {
			case "release":
				latestTag, err = GetGitHubRelease(definition.Package)
			case "refs":
				latestTag, err = GetGitHubRefs(definition.Package)
			case "git_hash":
				updateSrc = false
				currentTime := time.Now()
				formattedTime := currentTime.Format("20060102")
				yqReplace(filepath.Join(definition.Path, "definition.yaml"), "version", formattedTime)
				latestTag, err = GetGitHubHeads(definition.Package)
				yqReplace(filepath.Join(definition.Path, "definition.yaml"), "labels.\\\"git.hash\\\"", latestTag)
			case "release_tag":
				latestTag, err = GetGitHubReleaseTag(definition.Package)
			default:
				latestTag, err = GetGitHubTag(definition.Package)
			}

			if !updateSrc {
				return
			}

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
