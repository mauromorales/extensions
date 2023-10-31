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
	"regexp"
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
	fmt.Printf("- Consider only if version contains: %s\n\n", p.Labels["autobump.version_contains"])
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

func replace(p *types.Package, version string) (string, error) {
	var data map[string]string
	stringReplace := p.Labels["autobump.string_replace"]
	if stringReplace == "" {
		return version, nil
	}

	if err := json.Unmarshal([]byte(stringReplace), &data); err != nil {
		return "", err
	}

	for exp, with := range data {
		fmt.Printf("Replacing %s with %s\n", exp, with)
		var re = regexp.MustCompile(exp)
		version = re.ReplaceAllString(version, with)
	}

	return version, nil
}

func skipIfContains(p *types.Package, version string) bool {
	skipIfContains := p.Labels["autobump.skip_if_contains"]
	if skipIfContains == "" {
		return false
	}

	var data []string

	if err := json.Unmarshal([]byte(skipIfContains), &data); err != nil {
		return false
	}

	for _, exp := range data {
		if strings.Contains(version, exp) {
			fmt.Println("Skipping because latest release contains: ", exp)
			return true
		}
	}

	return false
}

func installDependencies() error {
	rootDir := os.Getenv("ROOT_DIR")
	if rootDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		rootDir = wd
	}

	jqRelease := os.Getenv("JQ_RELEASE")
	if jqRelease == "" {
		jqRelease = "1.6"
	}
	qyRelease := os.Getenv("YQ_RELEASE")
	if qyRelease == "" {
		qyRelease = "3.3.4"
	}
	hubRelease := os.Getenv("HUB_RELEASE")
	if hubRelease == "" {
		hubRelease = "2.14.2"
	}

	err := utils.RunSH("", fmt.Sprintf("hash jq 2>/dev/null || {\n    mkdir -p %s/.bin/;\n    wget https://github.com/stedolan/jq/releases/download/jq-%s/jq-linux64 -O %s/.bin/jq\n    chmod +x %s/.bin/jq\n}", rootDir, jqRelease, rootDir, rootDir))
	if err != nil {
		return err
	}
	err = utils.RunSH("", fmt.Sprintf("hash yq 2>/dev/null || {\n    mkdir -p %s/.bin/;\n    wget https://github.com/mikefarah/yq/releases/download/%s/yq_linux_amd64 -O %s/.bin/yq\n    chmod +x %s/.bin/yq\n}", rootDir, qyRelease, rootDir, rootDir))
	if err != nil {
		return err
	}
	err = utils.RunSH("", fmt.Sprintf("hash hub 2>/dev/null || {\n    mkdir -p %s/.bin/;\n    wget https://github.com/github/hub/releases/download/v%s/hub-linux-amd64-%s.tgz -O %s/.bin/hub\n    chmod +x %s/.bin/hub\n}", rootDir, hubRelease, hubRelease, rootDir, rootDir))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	failOnError := false
	if os.Getenv("FAIL_ON_ERROR") == "true" {
		failOnError = true
	}

	// err := installDependencies()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	os.Exit(1)
	// }

	treeDir := os.Getenv("TREE_DIR")
	if treeDir == "" {
		fmt.Println("TREE_DIR is not set")

		if failOnError {
			os.Exit(1)
		}

		return
	}
	cmd := fmt.Sprintf("luet tree pkglist --tree %s -o json", treeDir)
	pkgJson, err := utils.RunSHOUT("", cmd)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	var luetOutput struct {
		Packages types.Packages `json:"packages"`
	}
	if err := json.Unmarshal(pkgJson, &luetOutput); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	for _, entry := range luetOutput.Packages {

		_, err := os.ReadFile(filepath.Join(entry.Path, "collection.yaml"))
		if err == nil {
			fmt.Println("Skipping collection:", entry.Name)
			continue
		}

		definition, err := NewDefinition(entry.Path)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		PrintInfo(definition.Package)
		if definition.Package.Labels["autobump.ignore"] == "1" {
			fmt.Printf("Ignoring package: %s\n", definition.Package.Name)
			continue
		}

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

		latestTag, err = replace(definition.Package, latestTag)
		if err != nil {
			fmt.Println("Error:", err)

			if failOnError {
				os.Exit(1)
			}

			return
		}

		if !updateSrc || skipIfContains(definition.Package, latestTag) {
			continue
		}

		latestVersion := latestTag

		trimPrefix := definition.Package.Labels["autobump.trim_prefix"]
		if trimPrefix != "" {
			latestVersion = strings.TrimPrefix(latestVersion, trimPrefix)
		} else {
			latestVersion = strings.Replace(latestVersion, "v", "", 1)
		}

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
