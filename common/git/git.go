package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
)

const (
	EmptyTreeCommitId = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
)

func CloneRepository(url *string, branch *string, depth int) (*git.Repository, string, error) {
	urlVal := *url
	branchVal := *branch
	dir, err := ioutil.TempDir("", "secretscanner")
	if err != nil {
		return nil, "", err
	}
	repository, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:           urlVal,
		Depth:         depth,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchVal)),
		SingleBranch:  true,
		Tags:          git.NoTags,
	})
	if err != nil {
		return nil, dir, err
	}
	return repository, dir, nil
}

func GetRepositoryHistory(repository *git.Repository) ([]*object.Commit, error) {
	var commits []*object.Commit
	ref, err := repository.Head()
	if err != nil {
		return nil, err
	}
	cIter, err := repository.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}
	cIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	return commits, nil
}

func GetChanges(commit *object.Commit, repo *git.Repository) (object.Changes, error) {
	parentCommit, err := GetParentCommit(commit, repo)
	if err != nil {
		return nil, err
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	parentCommitTree, err := parentCommit.Tree()
	if err != nil {
		return nil, err
	}

	changes, err := object.DiffTree(parentCommitTree, commitTree)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

func GetParentCommit(commit *object.Commit, repo *git.Repository) (*object.Commit, error) {
	if commit.NumParents() == 0 {
		parentCommit, err := repo.CommitObject(plumbing.NewHash(EmptyTreeCommitId))
		if err != nil {
			return nil, err
		}
		return parentCommit, nil
	}
	parentCommit, err := commit.Parents().Next()
	if err != nil {
		return nil, err
	}
	return parentCommit, nil
}

func GetChangeAction(change *object.Change) string {
	action, err := change.Action()
	if err != nil {
		return "Unknown"
	}
	switch action {
	case merkletrie.Insert:
		return "Insert"
	case merkletrie.Modify:
		return "Modify"
	case merkletrie.Delete:
		return "Delete"
	default:
		return "Unknown"
	}
}

func GetChangePath(change *object.Change) string {
	action, err := change.Action()
	if err != nil {
		return change.To.Name
	}

	if action == merkletrie.Delete {
		return change.From.Name
	} else {
		return change.To.Name
	}
}

func GetPatch(change *object.Change) (*object.Patch, error) {
  patch, err := change.Patch()
  if err != nil {
    return nil, err
  }
  return patch, err
}

// GetLatestCommitHash runs a git cmd to return latest commit hash
func GetLatestCommitHash(dir string) (string, error) {
	os.Chdir(dir)
	gitcmd := "git"
	task := "rev-parse"
	op1 := "--verify"
	op2 := "HEAD"
	out, err := exec.Command(gitcmd, task, op1, op2).CombinedOutput()
	if err != nil {
		return "", err
	}
	commitHash := fmt.Sprintf("%s", string(out))
	return commitHash, nil
}

func GatherPaths(dir, branch string, targets []string) ([]string, error) {
	os.Chdir(dir)
	gitcmd := "git"
	listTree := "ls-tree"
	op1 := "-r"
	op2 := "--name-only"
	var paths []string

	if len(targets) == 0 {
		out, err := exec.Command(gitcmd, listTree, op1, branch, op2).CombinedOutput()
		if err != nil {
			return nil, err
		}
		cmdout := fmt.Sprintf("%s", string(out))
		paths = append(paths, strings.Split(cmdout, "\n")...)
	}

	for _, t := range targets {
		out, err := exec.Command(gitcmd, listTree, op1, fmt.Sprintf("%s:%s", branch, t), op2).CombinedOutput()
		if err != nil {
			return nil, err
		}
		cmdout := fmt.Sprintf("%s", string(out))
		currentPaths := strings.Split(cmdout, "\n")
		for i, p := range currentPaths {
			currentPaths[i] = path.Join(t, p)
		}
		paths = append(paths, currentPaths...)
	}
	return paths, nil
}
