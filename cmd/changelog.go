package cmd

import (
	"fmt"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/spf13/cobra"
)

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

func mergedPrsSince(client api.GQLClient, owner string, name string, publishedAt string) ([]pr, error) {

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
		"searchQuery": graphql.String(fmt.Sprintf(`repo:%s/%s is:merged merged:>=%s`, graphql.String(owner), graphql.String(name), graphql.String(publishedAt))),
	}

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

func changelog() {
	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	response := struct{ Login string }{}
	err = client.Get("user", &response)
	if err != nil {
		fmt.Println(err)
		return
	}

	repo, err := gh.CurrentRepository()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("★ running as @%s in %s/%s\n", response.Login, repo.Owner(), repo.Name())

	publishedAt, err := dateOfLastRelease(client, repo.Owner(), repo.Name())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("★ latest release published at %s\n", publishedAt)

	gql, err := gh.GQLClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	prs, err := mergedPrsSince(gql, repo.Owner(), repo.Name(), publishedAt)
	for _, pr := range prs {
		line := fmt.Sprintf("#%d %s (@%s)", pr.PullRequest.Number, pr.PullRequest.Title, pr.PullRequest.Author.User.Login)
		fmt.Println(line)
	}
}

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "show the changelog of PRs since the last published release",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UseLine())
		changelog()
	},
}

func init() {
	rootCmd.AddCommand(changelogCmd)
}
