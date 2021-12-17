package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/spf13/cobra"
)

type ChangelogOptions struct {
	Since string
	Repo  string
}

type Repo struct {
	Owner string
	Name  string
}

func (r Repo) SearchString() string {
	if r.Name != "" {
		return fmt.Sprintf(`repo:%s/%s`, r.Owner, r.Name)
	}
	return fmt.Sprintf(`org:%s`, r.Owner)
}

type pr struct {
	PullRequest struct {
		Title  string
		Number int
		Author struct {
			User struct {
				Login string
			} `graphql:"... on User"`
		}
		Repository struct {
			Name          string
			NameWithOwner string
		}
	} `graphql:"... on PullRequest"`
}

// Make a GET request to retrieve the publish date of the latest release, if present.
//
// owner	- the Github owner (user or organization), must not be empty
//
// name 	- the Github repository
func dateOfLastRelease(client api.RESTClient, owner string, name string) (string, error) {
	response := struct {
		PublishedAt string `json:"published_at"`
		Name        string
	}{}
	err := client.Get(fmt.Sprintf("repos/%s/%s/releases/latest", owner, name), &response)
	if err != nil {
		return "", err
	}
	return response.PublishedAt, nil
}

func mergedPrsSince(client api.GQLClient, repo *Repo, publishedAt string) ([]pr, error) {

	var query struct {
		Search struct {
			Nodes    []pr
			PageInfo struct {
				HasNextPage bool
				EndCursor   graphql.String
			}
		} `graphql:"search(query:$searchQuery, type: ISSUE, first: 20, after: $cursor)"`
	}

	variables := map[string]interface{}{
		"cursor":      (*graphql.String)(nil),
		"searchQuery": graphql.String(fmt.Sprintf(`%s is:merged merged:>=%s`, graphql.String(repo.SearchString()), graphql.String(publishedAt))),
	}

	fmt.Println(variables["searchQuery"])

	var allPrs []pr
	for {
		err := client.Query("Changelog", &query, variables)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		allPrs = append(allPrs, query.Search.Nodes...)
		if !query.Search.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = graphql.NewString(query.Search.PageInfo.EndCursor)
	}

	return allPrs, nil
}

func repository(repoOpts string) (*Repo, error) {

	if repoOpts != "" {
		// the user selected a different repository or owner
		regex := regexp.MustCompile(`^(?P<Owner>[\w\d\-]+)(?:\/(?P<Repo>[\w\d\-]+))?$`)
		match := regex.FindStringSubmatch(repoOpts)

		if len(match) == 3 {
			fmt.Printf("owner: %s\n", match[1])
			fmt.Printf("repo: %s\n", match[2])
			return &Repo{
				Owner: match[1],
				Name:  match[2],
			}, nil
		}

		return nil, errors.New("Invalid repository selector. Must be of format `OWNER[/REPO]`")
	} else {
		// the user did not provide a different selection, let's use the current repository
		repo, err := gh.CurrentRepository()
		if err != nil {
			return nil, err
		}
		return &Repo{
			Owner: repo.Owner(),
			Name:  repo.Name(),
		}, nil
	}
}

func since(client api.RESTClient, sinceUserInput string, repo *Repo) (string, error) {
	if sinceUserInput != "" {
		return sinceUserInput, nil
	} else if repo.Name != "" {
		publishedAt, err := dateOfLastRelease(client, repo.Owner, repo.Name)
		if err != nil {
			return "", err
		}
		return publishedAt, nil
	} else {
		return "", errors.New("Cannot determine start date. Either provide `--since` or select a single repository.")
	}
}

func changelog(opts *ChangelogOptions) {
	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	repo, err := repository(opts.Repo)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("★ target: \t %s\n", repo.SearchString())

	gql, err := gh.GQLClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	since, err := since(client, opts.Since, repo)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("★ since : \t %s\n", since)

	prs, err := mergedPrsSince(gql, repo, since)
	for _, pr := range prs {
		repoString := ""
		if repo.Name == "" {
			repoString = pr.PullRequest.Repository.NameWithOwner
		}
		line := fmt.Sprintf("%s#%d %s (@%s)", repoString, pr.PullRequest.Number, pr.PullRequest.Title, pr.PullRequest.Author.User.Login)
		fmt.Println(line)
	}
}

func NewChangelogCmd() *cobra.Command {
	opts := &ChangelogOptions{
		Since: "",
		Repo:  "",
	}

	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "show the changelog of PRs since the last published release",
		Run: func(cmd *cobra.Command, args []string) {
			changelog(opts)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Since, "since", "", "Start changelog at date `since`")
	cmd.PersistentFlags().StringVarP(&opts.Repo, "repo", "R", "", "Select another repository or organization using the OWNER[/REPO] format")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewChangelogCmd())
}
