package provisioner

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/alecthomas/errors"
	"github.com/google/go-github/v72/github"
)

type GitRepo string
type GitOrg string
type GitCommit string
type GitBranch string
type GithubPR int

type GitAuthor struct {
	Name  string
	Email string
}

// GitopsChange is a change in a gitops repo. It is used to track the status of a change.
type GitopsChange struct {
	repo GitRepo
	org  GitOrg
	pr   GithubPR

	client *github.Client
}

// GithubFileTemplater uses a template to generate a gitops resource in GitHub,
// and then listens for status updates on the commit / PR.
type GithubFileTemplater[T any] struct {
	content *template.Template
	path    *template.Template
}

func NewGithubFileTemplater[T any](
	path *template.Template,
	content *template.Template,
) *GithubFileTemplater[T] {
	return &GithubFileTemplater[T]{
		path:    path,
		content: content,
	}
}

// GithubChangeTemplater is a templater that generates a change in a gitops repo.
// It clones the repo, creates a branch, applies changes, and then opens a PR.
type GithubChangeTemplater[T any] struct {
	files         []*GithubFileTemplater[T]
	commitMessage *template.Template
	prSubject     *template.Template

	client *github.Client
}

func NewGithubChangeTemplater[T any](
	files []*GithubFileTemplater[T],
	commitMessage *template.Template,
	prSubject *template.Template,
	client *github.Client,
) *GithubChangeTemplater[T] {
	return &GithubChangeTemplater[T]{
		files:         files,
		commitMessage: commitMessage,
		prSubject:     prSubject,
		client:        client,
	}
}

func (t *GithubChangeTemplater[T]) CreateChange(ctx context.Context, org GitOrg, repo GitRepo, branch GitBranch, author GitAuthor, model T) (GitopsChange, error) {
	ref, err := t.getOrCreateRef(ctx, org, repo, branch, branch)
	if err != nil {
		return GitopsChange{}, errors.Wrap(err, "failed to get or create ref")
	}
	tree, err := t.createTree(ctx, ref, org, repo, model)
	if err != nil {
		return GitopsChange{}, errors.Wrap(err, "failed to create tree")
	}
	var message bytes.Buffer
	if err := t.commitMessage.Execute(&message, model); err != nil {
		return GitopsChange{}, errors.Wrap(err, "failed to execute commit message template")
	}
	err = t.pushCommit(ctx, org, repo, ref, tree, message.String(), author)
	if err != nil {
		return GitopsChange{}, errors.Wrap(err, "failed to push commit")
	}
	var subject bytes.Buffer
	if err := t.prSubject.Execute(&subject, model); err != nil {
		return GitopsChange{}, errors.Wrap(err, "failed to execute pr subject template")
	}
	pr, err := t.createPullRequest(ctx, org, repo, branch, subject.String(), message.String())
	if err != nil {
		return GitopsChange{}, errors.Wrap(err, "failed to create pull request")
	}

	return GitopsChange{
		org:    org,
		repo:   repo,
		pr:     GithubPR(pr),
		client: t.client,
	}, nil
}

func (t *GithubChangeTemplater[T]) createTree(ctx context.Context, ref *github.Reference, org GitOrg, repo GitRepo, model T) (*github.Tree, error) {
	treeEntries := []*github.TreeEntry{}

	for _, file := range t.files {
		var content bytes.Buffer
		if err := file.content.Execute(&content, model); err != nil {
			return nil, errors.Wrap(err, "failed to execute content template")
		}
		var path bytes.Buffer
		if err := file.path.Execute(&path, model); err != nil {
			return nil, errors.Wrap(err, "failed to execute path template")
		}
		treeEntries = append(treeEntries, &github.TreeEntry{
			Path:    github.Ptr(path.String()),
			Type:    github.Ptr("blob"),
			Content: github.Ptr(content.String()),
			Mode:    github.Ptr("100644"),
		})
	}

	tree, _, err := t.client.Git.CreateTree(ctx, string(org), string(repo), *ref.Object.SHA, treeEntries)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tree")
	}

	return tree, nil
}

func (t *GithubChangeTemplater[T]) getOrCreateRef(ctx context.Context, owner GitOrg, repo GitRepo, branch, base GitBranch) (*github.Reference, error) {
	if ref, _, err := t.client.Git.GetRef(ctx, string(owner), string(repo), "refs/heads/"+string(branch)); err == nil {
		return ref, nil
	}
	var baseRef *github.Reference
	var err error
	if baseRef, _, err = t.client.Git.GetRef(ctx, string(owner), string(repo), "refs/heads/"+string(base)); err != nil {
		return nil, errors.Wrap(err, "failed to get base ref")
	}
	newRef := &github.Reference{Ref: github.Ptr("refs/heads/" + string(branch)), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err := t.client.Git.CreateRef(ctx, string(owner), string(repo), newRef)
	return ref, errors.Wrap(err, "failed to create ref")
}

func (t *GithubChangeTemplater[T]) pushCommit(ctx context.Context, owner GitOrg, repo GitRepo, ref *github.Reference, tree *github.Tree, message string, author GitAuthor) (err error) {
	parent, _, err := t.client.Repositories.GetCommit(ctx, string(owner), string(repo), *ref.Object.SHA, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get parent commit")
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	commitAuthor := &github.CommitAuthor{Date: &github.Timestamp{Time: date}, Name: &author.Name, Email: &author.Email}
	commit := &github.Commit{Author: commitAuthor, Message: &message, Tree: tree, Parents: []*github.Commit{parent.Commit}}
	opts := github.CreateCommitOptions{}

	newCommit, _, err := t.client.Git.CreateCommit(ctx, string(owner), string(repo), commit, &opts)
	if err != nil {
		return errors.Wrap(err, "failed to create commit")
	}

	// Attach the commit to the base branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = t.client.Git.UpdateRef(ctx, string(owner), string(repo), ref, false)
	return errors.Wrap(err, "failed to update ref")
}

func (t *GithubChangeTemplater[T]) createPullRequest(ctx context.Context, owner GitOrg, repo GitRepo, branch GitBranch, subject string, description string) (int, error) {
	pr, _, err := t.client.PullRequests.Create(ctx, string(owner), string(repo), &github.NewPullRequest{
		Title:               github.Ptr(subject),
		Head:                github.Ptr(string(branch)),
		HeadRepo:            github.Ptr(string(repo)),
		Base:                github.Ptr(string(branch)),
		Body:                github.Ptr(description),
		MaintainerCanModify: github.Ptr(true),
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to create pull request")
	}

	return *pr.Number, nil
}

func (c GitopsChange) IsMerged(ctx context.Context) (bool, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, string(c.org), string(c.repo), int(c.pr))
	if err != nil {
		return false, errors.Wrap(err, "failed to get pull request")
	}
	return pr.MergeCommitSHA != nil, nil
}
