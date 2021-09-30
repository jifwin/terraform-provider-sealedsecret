package git

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
)

const (
	testGitUrlKey   = "TEST_GIT_URL"
	testGitToken    = "TEST_GIT_TOKEN"
	testGitUsername = "TEST_GIT_USERNAME"
)

func getEnv(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("cannot run test when env var %s is empty", key)
	}
	return value
}

func TestCreateBranch(t *testing.T) {
	g := newGit(t)
	defer cleanupBranch(t, g)

	branches, err := g.repo.Branches()
	assert.Nil(t, err)

	branchNameFound := false
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		if strings.Contains(ref.Name().String(), providerBranchName) {
			branchNameFound = true
		}
		return nil
	})
	assert.Nil(t, err)

	assert.True(t, branchNameFound, "expected to find a branch by name "+providerBranchName)
}

func TestGit_Push(t *testing.T) {
	testFile, testPath := []byte("my awesome test file"), "testpath/test.txt"
	g := newGit(t)
	defer cleanupBranch(t, g)
	err := g.Push(context.Background(), testFile, testPath)
	assert.Nil(t, err)

	// updating the git instance to ensure that the file actually was pushed
	g = newGit(t)
	g.fs.Open(testPath)
	wt, err := g.repo.Worktree()
	assert.Nil(t, err)
	assert.Nil(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(providerBranchName + "0"),
	}))
	assert.True(t, findFile(t, testPath, wt))
}

func cleanupBranch(t *testing.T, g *Git) {
	t.Fatal(g.repo.DeleteBranch(providerBranchName + "0"))
}

func findFile(t *testing.T, filePath string, wt *git.Worktree) bool {
	status, err := wt.Status()
	assert.Nil(t, err)
	for filePathActual := range status {
		log.Println("FILEPATH ACTUAL =====================>", filePathActual)
		if filePathActual == filePath {
			return true
		}
	}
	return false
}

func newGit(t *testing.T) *Git {
	g, err := NewGit(context.Background(), getEnv(t, testGitUrlKey), BasicAuth{
		Username: getEnv(t, testGitUsername),
		Token:    getEnv(t, testGitToken),
	})
	assert.Nil(t, err)

	return g
}
