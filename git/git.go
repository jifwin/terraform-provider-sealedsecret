package git

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"io"
	"sync"
)

type Git struct {
	url  string
	repo *git.Repository
	fs   billy.Filesystem
	auth *http.BasicAuth
	mu   *sync.Mutex
}

type BasicAuth struct {
	Username, Token string
}

const remoteName = "origin"

type Giter interface {
	Push(ctx context.Context, file []byte, path string) error
	GetFile(filePath string) ([]byte, error)
	DeleteFile(ctx context.Context, filePath string) error
}

func NewGit(ctx context.Context, url string, auth BasicAuth) (*Git, error) {
	ba := &http.BasicAuth{
		Username: auth.Username,
		Password: auth.Token,
	}
	fs := memfs.New()
	r, err := git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		URL:  url,
		Auth: ba,
	})
	if err != nil {
		return nil, err
	}
	return &Git{repo: r, fs: fs, auth: ba, url: url, mu: &sync.Mutex{}}, nil
}

// Push creates the new file and pushes the changes to Git remote.
//
// filePath must specify the path to where the new file should be created
func (g *Git) Push(ctx context.Context, file []byte, filePath string) error {
	newFile, err := g.fs.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to create file: %w", err)
	}

	// when multiple resources are created we need to update the git refs head after push
	g.mu.Lock()
	defer g.mu.Unlock()

	_, err = newFile.Write(file)
	if err != nil {
		return fmt.Errorf("unable to write to file: %w", err)
	}
	err = newFile.Close()
	if err != nil {
		return err
	}

	w, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(filePath)
	if err != nil {
		return err
	}
	_, err = w.Commit(createCommitMsg("created", filePath), &git.CommitOptions{})
	if err != nil {
		return err
	}

	if err := g.repo.PushContext(ctx, &git.PushOptions{RemoteName: remoteName, Auth: g.auth}); err != nil {
		return err
	}

	if err := g.repo.FetchContext(ctx, &git.FetchOptions{RemoteName: remoteName, Auth: g.auth}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	return nil
}

func (g *Git) GetFile(filePath string) ([]byte, error) {
	f, err := g.fs.Open(filePath)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(f)
}

func (g *Git) DeleteFile(ctx context.Context, filePath string) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	// when multiple resources are created we need to update the git refs head after push
	g.mu.Lock()
	defer g.mu.Unlock()
	_, err = w.Remove(filePath)
	if err != nil {
		return err
	}
	_, err = w.Commit(createCommitMsg("deleted", filePath), &git.CommitOptions{Author: &object.Signature{
		Name: "SEALEDSECRET-PROVIDER",
	}})
	if err != nil {
		return err
	}
	if err := g.repo.PushContext(ctx, &git.PushOptions{RemoteName: "origin", Auth: g.auth}); err != nil {
		return err
	}

	if err := g.repo.FetchContext(ctx, &git.FetchOptions{RemoteName: remoteName, Auth: g.auth}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}

func createCommitMsg(action, filePath string) string {
	return fmt.Sprintf("[SEALEDSECRET-PROVIDER] %s --> %s", action, filePath)

}
