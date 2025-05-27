package provisioner

import (
	"testing"
	"text/template"

	"github.com/alecthomas/assert/v2"
	"github.com/google/go-github/v72/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

type TestData struct {
	Name    string
	Version string
}

func TestGithubFileTemplater(t *testing.T) {
	client := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.EndpointPattern{Pattern: "/repos/fooorg/foorepo/git/ref/heads/change-branch", Method: "GET"},
			github.Reference{
				Ref:    github.Ptr("refs/heads/change-branch"),
				Object: &github.GitObject{SHA: github.Ptr("1")},
			},
		),
		mock.WithRequestMatch(
			mock.EndpointPattern{Pattern: "/repos/fooorg/foorepo/git/trees", Method: "POST"},
			github.Tree{SHA: github.Ptr("1")},
		),
		mock.WithRequestMatch(
			mock.EndpointPattern{Pattern: "/repos/fooorg/foorepo/git/commits", Method: "POST"},
			github.Commit{SHA: github.Ptr("1")},
		),
		mock.WithRequestMatch(
			mock.EndpointPattern{Pattern: "/repos/fooorg/foorepo/commits/1", Method: "GET"},
			github.RepositoryCommit{
				SHA:    github.Ptr("1"),
				Commit: &github.Commit{Message: github.Ptr("commit message")},
			},
		),
		mock.WithRequestMatch(
			mock.EndpointPattern{Pattern: "/repos/fooorg/foorepo/git/refs/heads/change-branch", Method: "PATCH"},
			github.Reference{
				Ref:    github.Ptr("refs/heads/change-branch"),
				Object: &github.GitObject{SHA: github.Ptr("2")},
			},
		),
		mock.WithRequestMatch(
			mock.EndpointPattern{Pattern: "/repos/fooorg/foorepo/pulls", Method: "POST"},
			github.PullRequest{Number: github.Ptr(1)},
		),
	))

	t.Run("opens a PR", func(t *testing.T) {
		author := GitAuthor{
			Name:  "foobar",
			Email: "foobar@example.com",
		}
		data := TestData{
			Name:    "test",
			Version: "1.0.0",
		}
		templater := NewGithubChangeTemplater(
			[]*GithubFileTemplater[TestData]{
				NewGithubFileTemplater[TestData](
					template.Must(template.New("path1/{{.Name}}.yaml").Parse("{{.Name}}")),
					template.Must(template.New("content").Parse("{{.Version}}")),
				),
				NewGithubFileTemplater[TestData](
					template.Must(template.New("path2/{{.Name}}.yaml").Parse("{{.Name}}")),
					template.Must(template.New("content").Parse("{{.Version}}")),
				),
			},
			template.Must(template.New("commitMessage").Parse("commit message")),
			template.Must(template.New("prSubject").Parse("pr subject")),
			client,
		)
		change, err := templater.CreateChange(t.Context(), "fooorg", "foorepo", "change-branch", author, data)
		assert.NoError(t, err)
		assert.Equal(t, "fooorg", change.org)
		assert.Equal(t, "foorepo", change.repo)
		assert.Equal(t, 1, change.pr)
	})
}
