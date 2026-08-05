package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Valid start to an embedmd link
const embedmd = `[embedmd]:# `

// Repository the generated links point at, and the branch they pin
const (
	repoURL   = "https://github.com/gnolang/gno"
	repoRef   = "master"
	fenceMark = "```"
)

// Regular expression to match markdown links
var regex = regexp.MustCompile(`]\(([^)]+)\)`)

// localLink is one link to a local file, as written in a .md file
type localLink struct {
	raw      string // link as written, fragment included
	path     string // link with any #fragment removed
	fragment string // #fragment, empty when there is none
	embedmd  bool   // embedmd directive: embeds a file from disk, never rendered as a link
	fenced   bool   // sits inside a fenced code block: sample text, not a link
}

// extractLocalLinks extracts links to local files from the given file content
func extractLocalLinks(fileContent []byte) []string {
	refs := extractLocalLinkRefs(fileContent)
	links := make([]string, 0, len(refs))

	for _, ref := range refs {
		links = append(links, ref.path)
	}

	return links
}

// extractLocalLinkRefs extracts links to local files, keeping what the fixer
// needs to rewrite them: the text as written, and the context it sits in
func extractLocalLinkRefs(fileContent []byte) []localLink {
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	links := make([]localLink, 0)
	fenced := false

	// Scan file line by line
	for scanner.Scan() {
		line := scanner.Text()

		// Track fenced code blocks
		if strings.HasPrefix(strings.TrimSpace(line), fenceMark) {
			fenced = !fenced
		}

		// Check for embedmd links
		if embedmdPos := strings.Index(line, embedmd); embedmdPos != -1 {
			link := line[embedmdPos+len(embedmd)+1:]

			// Find closing parentheses
			if closePar := strings.LastIndex(link, ")"); closePar != -1 {
				link = link[:closePar]
			}

			// Remove space
			if pos := strings.Index(link, " "); pos != -1 {
				link = link[:pos]
			}

			// Add link to be checked
			links = append(links, localLink{raw: link, path: link, embedmd: true, fenced: fenced})
			continue
		}

		// Find all matches
		matches := regex.FindAllString(line, -1)

		// Extract and print the local file links
		for _, match := range matches {
			// Remove ]( from the beginning and ) from end of link
			match = match[2 : len(match)-1]

			// Ignore http, https, tcp, ws links
			if shouldIgnoreLink(match) {
				continue
			}

			ref := localLink{raw: match, path: match, fenced: fenced}

			// Split off markdown headers in links
			if pos := strings.Index(match, "#"); pos != -1 {
				ref.path, ref.fragment = match[:pos], match[pos:]
			}

			links = append(links, ref)
		}
	}

	return links
}

func lintLocalLinks(filepathToLinks map[string][]localLink, docsRoot, repoRoot string) (string, error) {
	var (
		missing bytes.Buffer
		outside bytes.Buffer
	)

	for filePath, links := range filepathToLinks {
		for _, link := range links {
			path := filepath.Join(filepath.Dir(filePath), link.path)

			info, err := os.Stat(path)
			if err != nil {
				if missing.Len() == 0 {
					missing.WriteString("Could not find files with the following paths:\n")
				}

				absSourcePath, _ := filepath.Abs(filePath)
				missing.WriteString(
					fmt.Sprintf(">>> %s (found in file: file://%s)\n", link.path, absSourcePath),
				)

				continue
			}

			// A link that leaves the docs root cannot resolve on the published
			// site: the generator emits it verbatim, so it ships as a path that
			// does not exist there. It has to be an absolute GitHub URL.
			if !link.rewritable(docsRoot, path) {
				continue
			}

			if outside.Len() == 0 {
				outside.WriteString("Local links that leave the docs folder, and the URLs to replace them with:\n")
			}

			absSourcePath, _ := filepath.Abs(filePath)
			outside.WriteString(fmt.Sprintf(
				">>> %s -> %s (found in file: file://%s)\n",
				link.raw, link.githubURL(repoRoot, path, info.IsDir()), absSourcePath,
			))
		}
	}

	switch {
	case missing.Len() != 0 && outside.Len() != 0:
		return missing.String() + outside.String(), errFoundUnreachableLocalLinks
	case missing.Len() != 0:
		return missing.String(), errFoundUnreachableLocalLinks
	case outside.Len() != 0:
		return outside.String(), errFoundLinksOutsideDocs
	}

	return "", nil
}

// fixLocalLinks rewrites every link that leaves the docs root into an absolute
// GitHub URL, in place. It reports the paths of the files it changed.
func fixLocalLinks(mdFiles []string, docsRoot, repoRoot string) (string, error) {
	var output bytes.Buffer

	for _, filePath := range mdFiles {
		fileContents, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		fixed, replacements := fixContent(fileContents, filePath, docsRoot, repoRoot)
		if len(replacements) == 0 {
			continue
		}

		info, err := os.Stat(filePath)
		if err != nil {
			return "", err
		}

		if err := os.WriteFile(filePath, fixed, info.Mode().Perm()); err != nil {
			return "", err
		}

		absSourcePath, _ := filepath.Abs(filePath)
		output.WriteString(fmt.Sprintf("Rewrote %d link(s) in file://%s\n", len(replacements), absSourcePath))

		for _, replacement := range replacements {
			output.WriteString(fmt.Sprintf(">>> %s -> %s\n", replacement[0], replacement[1]))
		}
	}

	return output.String(), nil
}

// fixContent returns the file content with every rewritable link replaced, and
// the {before, after} pair of each replacement it made
func fixContent(fileContent []byte, filePath, docsRoot, repoRoot string) ([]byte, [][2]string) {
	replacements := make([][2]string, 0)

	for _, link := range extractLocalLinkRefs(fileContent) {
		path := filepath.Join(filepath.Dir(filePath), link.path)

		info, err := os.Stat(path)
		if err != nil || !link.rewritable(docsRoot, path) {
			continue
		}

		url := link.githubURL(repoRoot, path, info.IsDir())
		fileContent = bytes.ReplaceAll(fileContent, []byte("]("+link.raw+")"), []byte("]("+url+")"))
		replacements = append(replacements, [2]string{link.raw, url})
	}

	return fileContent, replacements
}

// rewritable reports whether the link is one the site cannot resolve and the
// tool can replace: a plain markdown link, resolving outside the docs root
func (l localLink) rewritable(docsRoot, target string) bool {
	if l.embedmd || l.fenced || l.path == "" || docsRoot == "" {
		return false
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(docsRoot, absTarget)

	return err == nil && strings.HasPrefix(rel, "..")
}

// githubURL builds the URL for a target given as a path inside the repository
func (l localLink) githubURL(repoRoot, target string, isDir bool) string {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "<no GitHub URL: " + err.Error() + ">"
	}

	rel, err := filepath.Rel(repoRoot, absTarget)
	if repoRoot == "" || err != nil || strings.HasPrefix(rel, "..") {
		return "<no GitHub URL: target is outside the repository>"
	}

	kind := "blob"
	if isDir {
		kind = "tree"
	}

	return fmt.Sprintf("%s/%s/%s/%s%s", repoURL, kind, repoRef, filepath.ToSlash(rel), l.fragment)
}

// findRepoRoot walks up from the given path to the nearest ancestor holding a
// go.mod, which for the docs folder is the repository root
func findRepoRoot(path string) string {
	for {
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			return path
		}

		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}

		path = parent
	}
}

func shouldIgnoreLink(m string) bool {
	return strings.HasPrefix(m, "http") || strings.HasPrefix(m, "https") || strings.HasPrefix(m, "ws") || strings.HasPrefix(m, "tcp")
}
