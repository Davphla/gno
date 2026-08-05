package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyPathError(t *testing.T) {
	t.Parallel()

	cfg := &cfg{
		docsPath: "",
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFn()

	_, err := execLint(cfg, ctx)
	assert.ErrorIs(t, err, errEmptyPath)
}

func TestExtractLinks(t *testing.T) {
	t.Parallel()

	// Create mock file content with random links
	mockFileContent := `# Lorem Ipsum
Lorem ipsum dolor sit amet, 
[consectetur](https://example.org)
adipiscing elit. Vivamus lacinia odio
vitae [vestibulum vestibulum](http://localhost:3000).
Cras [vel ex](http://192.168.1.1) et
turpis egestas luctus. Nullam
[eleifend](https://www.wikipedia.org)
nulla ac [blandit tempus](https://gitlab.org). 
## Valid Links Here are some valid links:
- [Mozilla](https://mozilla.org) 
- [Valid URL](https://valid-url.net) 
- [Another Valid URL](https://another-valid-url.info) 
- [Valid Link](https://valid-link.edu)
`

	// Expected URLs
	expectedUrls := []string{
		"https://example.org",
		"http://192.168.1.1",
		"https://www.wikipedia.org",
		"https://gitlab.org",
		"https://mozilla.org",
		"https://valid-url.net",
		"https://another-valid-url.info",
		"https://valid-link.edu",
	}

	// Extract URLs from each file in the sourceDir
	extractedUrls := extractUrls([]byte(mockFileContent))

	if len(expectedUrls) != len(extractedUrls) {
		t.Fatal("did not extract correct amount of URLs")
	}

	sort.Strings(extractedUrls)
	sort.Strings(expectedUrls)

	for i, u := range expectedUrls {
		require.Equal(t, u, extractedUrls[i])
	}
}

func TestExtractJSX(t *testing.T) {
	t.Parallel()

	// Create mock file content with random JSX tags
	mockFileContent := `
#### Usage

### getFunctionSignatures

Fetches public facing function signatures

#### Parameters

Returns **Promise<FunctionSignature[]>**

# test text from gnodev.md <node-rpc-listener>

#### Usage
### evaluateExpression

Evaluates any expression in readonly mode and returns the results

#### Parameters

Returns **Promise<string>**
`

	// Expected JSX tags
	expectedTags := []string{
		"<FunctionSignature[]>",
		"<string>",
		"<node-rpc-listener>",
	}

	// Extract JSX tags from the mock file content
	extractedTags := extractJSX([]byte(mockFileContent))

	if len(expectedTags) != len(extractedTags) {
		t.Fatal("did not extract the correct amount of JSX tags")
	}

	sort.Strings(extractedTags)
	sort.Strings(expectedTags)

	for i, tag := range expectedTags {
		require.Equal(t, tag, extractedTags[i])
	}
}

func TestExtractLocalLinks(t *testing.T) {
	t.Parallel()

	// Create mock file content with random local links
	mockFileContent := `
Here is some text with a link to a local file: [text](../concepts/file1.md)
Here is another local link: [another](./path/to/file1.md)
Here is another local link: [another](./path/to/file2.md#header-1-2)
Here is another local link without ./ or ../: [another](path/to/another-file.md)
And a link to an external website: [example](https://example.com)
And a websocket link: [websocket](ws://example.com/socket)
Here's an embedmd link: [embedmd]:# (../assets/how-to-guides/simple-library/tapas.gno go)
Here's an embedmd link: [embedmd]:# (../assets/myfile.sol go)
Here's an embedmd link: [embedmd]:# (../assets/myfi()le.gno c)
Here's an embedmd link: [embedmd]:# (../assets/)myfi(le.gno c)
Here's another link: [embedmd]:# (../folder/myfile.gno c
Here's a tcp link: [tcp](tcp://localhost:25567)
`

	// Expected local links
	expectedLinks := []string{
		"../concepts/file1.md",
		"./path/to/file1.md",
		"./path/to/file2.md",
		"path/to/another-file.md",
		"../assets/how-to-guides/simple-library/tapas.gno",
		"../assets/myfile.sol",
		"../assets/myfi()le.gno",
		"../assets/)myfi(le.gno",
		"../folder/myfile.gno",
	}

	// Extract local links tags from the mock file content
	extractedLinks := extractLocalLinks([]byte(mockFileContent))

	if len(expectedLinks) != len(extractedLinks) {
		t.Fatalf("did not extract the correct amount of local links, expected %d, got %d", len(expectedLinks), len(extractedLinks))
	}

	sort.Strings(extractedLinks)
	sort.Strings(expectedLinks)

	for i, tag := range expectedLinks {
		require.Equal(t, tag, extractedLinks[i])
	}
}

func TestFindFilePaths(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp(".", "test")
	require.NoError(t, err)
	t.Cleanup(removeDir(t, tempDir))

	numSourceFiles := 20
	testFiles := make([]string, numSourceFiles)

	for i := 0; i < numSourceFiles; i++ {
		testFiles[i] = "sourceFile" + strconv.Itoa(i) + ".md"
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		require.NoError(t, err)

		_, err = os.Create(filePath)
		require.NoError(t, err)
	}

	results, err := findFilePaths(tempDir)
	require.NoError(t, err)

	expectedResults := make([]string, 0, len(testFiles))

	for _, testFile := range testFiles {
		expectedResults = append(expectedResults, filepath.Join(tempDir, testFile))
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})

	sort.Slice(expectedResults, func(i, j int) bool {
		return expectedResults[i] < expectedResults[j]
	})

	require.Equal(t, len(results), len(expectedResults))

	for i, result := range results {
		if result != expectedResults[i] {
			require.Equal(t, result, expectedResults[i])
		}
	}
}

// mockTransport returns a fixed HTTP status code for every request.
type mockTransport struct{ statusCode int }

func (m *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       http.NoBody,
	}, nil
}

func TestFlow(t *testing.T) {
	t.Parallel()

	// Replace the package-level HTTP client so checkUrl never hits the network.
	origClient := httpClient
	httpClient = &http.Client{Transport: &mockTransport{statusCode: http.StatusNotFound}}
	t.Cleanup(func() { httpClient = origClient })

	const brokenURL = "https://example.com/non-existent-page"

	tempDir, err := os.MkdirTemp(".", "test")
	require.NoError(t, err)
	t.Cleanup(removeDir(t, tempDir))

	contents := `This is a [broken link](` + brokenURL + `).
Here's an embedmd link that links to a non-existing file: [embedmd]:# (../assets/myfile.sol go)
and here is some JSX tags <string\> <random-unescaped-text-tag>
and [this is a link to a non-existent](../myfolder/myfile.md) file.`

	expectedItems := []string{
		brokenURL,
		"../assets/myfile.sol",
		"<random-unescaped-text-tag>",
		"../myfolder/myfile.md",
	}

	filePath := filepath.Join(tempDir, "examplefile.md")

	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	require.NoError(t, err)

	f, err := os.Create(filePath)
	require.NoError(t, err)

	_, err = f.WriteString(contents)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	res, err := execLint(&cfg{
		docsPath: tempDir,
	},
		context.Background(),
	)

	assert.ErrorIs(t, err, errFoundLintItems)

	for _, item := range expectedItems {
		assert.True(t, strings.Contains(res, item))
	}
}

// mockRepo writes a docs folder holding a page, a go.mod marking the repository
// root above it, and an examples folder sitting outside the docs root
func mockRepo(t *testing.T, page string) (repoRoot, docsRoot string) {
	t.Helper()

	repoRoot, err := os.MkdirTemp(".", "repo")
	require.NoError(t, err)
	t.Cleanup(removeDir(t, repoRoot))

	repoRoot, err = filepath.Abs(repoRoot)
	require.NoError(t, err)

	docsRoot = filepath.Join(repoRoot, "docs")

	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module mock\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "examples", "pkg"), os.ModePerm))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "examples", "pkg", "README.md"), []byte("# pkg\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(docsRoot, "resources"), os.ModePerm))
	require.NoError(t, os.WriteFile(filepath.Join(docsRoot, "resources", "inside.md"), []byte("# inside\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(docsRoot, "resources", "page.md"), []byte(page), 0o644))

	return repoRoot, docsRoot
}

// mockPage carries one link of every kind the outside-the-docs rules recognise
const mockPage = "A link that stays inside: [inside](./inside.md)\n" +
	"A link that leaves: [pkg](../../examples/pkg/README.md#usage)\n" +
	"A folder that leaves: [examples](../../examples/pkg)\n" +
	"An embedmd directive: [embedmd]:# (../../examples/pkg/README.md md)\n" +
	"```\n" +
	"A fenced sample: [pkg](../../examples/pkg/README.md)\n" +
	"```\n"

func TestFindRepoRoot(t *testing.T) {
	t.Parallel()

	repoRoot, docsRoot := mockRepo(t, "# page\n")

	assert.Equal(t, repoRoot, findRepoRoot(docsRoot))
	assert.Equal(t, "", findRepoRoot(string(filepath.Separator)))
}

func TestLintLinksOutsideDocs(t *testing.T) {
	t.Parallel()

	_, docsRoot := mockRepo(t, mockPage)

	res, err := execLint(&cfg{docsPath: docsRoot, treatUrlsAsErr: false}, context.Background())
	assert.ErrorIs(t, err, errFoundLintItems)

	// Both links that leave the docs root are reported with the URL to use
	assert.Contains(t, res, "../../examples/pkg/README.md#usage -> "+
		repoURL+"/blob/"+repoRef+"/examples/pkg/README.md#usage")
	assert.Contains(t, res, "../../examples/pkg -> "+repoURL+"/tree/"+repoRef+"/examples/pkg")

	// The link that stays inside the docs root is not reported, and neither the
	// embedmd directive nor the fenced sample is
	assert.NotContains(t, res, "./inside.md")
	assert.Equal(t, 2, strings.Count(res, " -> "))
}

func TestFixLinksOutsideDocs(t *testing.T) {
	t.Parallel()

	_, docsRoot := mockRepo(t, mockPage)
	pagePath := filepath.Join(docsRoot, "resources", "page.md")

	res, err := execLint(&cfg{docsPath: docsRoot, fix: true}, context.Background())
	require.NoError(t, err)
	assert.Contains(t, res, "Rewrote 2 link(s)")

	expected := "A link that stays inside: [inside](./inside.md)\n" +
		"A link that leaves: [pkg](" + repoURL + "/blob/" + repoRef + "/examples/pkg/README.md#usage)\n" +
		"A folder that leaves: [examples](" + repoURL + "/tree/" + repoRef + "/examples/pkg)\n" +
		"An embedmd directive: [embedmd]:# (../../examples/pkg/README.md md)\n" +
		"```\n" +
		"A fenced sample: [pkg](../../examples/pkg/README.md)\n" +
		"```\n"

	fixed, err := os.ReadFile(pagePath)
	require.NoError(t, err)
	assert.Equal(t, expected, string(fixed))

	// Fixing is idempotent: a second run rewrites nothing
	res, err = execLint(&cfg{docsPath: docsRoot, fix: true}, context.Background())
	require.NoError(t, err)
	assert.Equal(t, "", res)

	refixed, err := os.ReadFile(pagePath)
	require.NoError(t, err)
	assert.Equal(t, expected, string(refixed))
}

func TestFixLeavesMissingTargetsAlone(t *testing.T) {
	t.Parallel()

	const page = "A link to nothing: [gone](../../examples/gone/README.md)\n"

	_, docsRoot := mockRepo(t, page)
	pagePath := filepath.Join(docsRoot, "resources", "page.md")

	res, err := execLint(&cfg{docsPath: docsRoot, fix: true}, context.Background())
	require.NoError(t, err)
	assert.Equal(t, "", res)

	unchanged, err := os.ReadFile(pagePath)
	require.NoError(t, err)
	assert.Equal(t, page, string(unchanged))

	// A target that is not in the tree is a missing file, not a link to rewrite
	res, err = execLint(&cfg{docsPath: docsRoot, treatUrlsAsErr: false}, context.Background())
	assert.ErrorIs(t, err, errFoundLintItems)
	assert.Contains(t, res, "Could not find files with the following paths:")
	assert.NotContains(t, res, " -> ")
}

func removeDir(t *testing.T, dirPath string) func() {
	return func() {
		require.NoError(t, os.RemoveAll(dirPath))
	}
}
