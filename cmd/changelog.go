package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/kr/text"
	"github.com/skaldarnar/gh-terasology/pkg/fun"
	"github.com/spf13/cobra"
)

type ChangelogOptions struct {
	Since  string
	Until  string
	Repo   string
	Pretty bool
}

func NewChangelogCmd() *cobra.Command {
	opts := &ChangelogOptions{
		Since: "",
		Repo:  "",
	}

	example := "changelog --repo MovingBlocks/Terasology\n"
	example += "    Print the changelog since latest release for MovingBlocks/Terasology to the console\n"
	example += "changelog --repo Terasology --since 2021-12-01 --pretty\n"
	example += "    Print markdown-formatted changelog for all repositories under the Terasology org since 1st Dec 2021 to the console\n"

	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "show the changelog of PRs since the last published release",
		Run: func(cmd *cobra.Command, args []string) {
			changelog(opts)
		},
		Example: text.Indent(example, "    "),
	}

	cmd.PersistentFlags().StringVar(&opts.Since, "since", "", "Start the changelog at date `since` (ISO 8601)")
	cmd.PersistentFlags().StringVar(&opts.Until, "until", "", "End the changelog at date `until` (ISO 8601)")
	cmd.PersistentFlags().StringVarP(&opts.Repo, "repo", "R", "", "Select another repository or organization using the OWNER[/REPO] format")
	cmd.PersistentFlags().BoolVar(&opts.Pretty, "pretty", false, "Pretty-print changelog as markdown")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewChangelogCmd())
}

// --------------------------------------------------------------------------------------------------------------------

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

type Pr interface {
	Title() string
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

func (pr pr) Title() string {
	return pr.PullRequest.Title
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

func timespanSearchQuery(since string, until string) string {
	if since != "" && until != "" {
		return fmt.Sprintf(`%s..%s`, since, until)
	} else if since != "" {
		return ">=" + since
	} else {
		return "<" + until
	}
}

func mergedPrsSince(client api.GQLClient, repo *Repo, since string, until string) ([]pr, error) {

	var query struct {
		Search struct {
			Nodes    []pr
			PageInfo struct {
				HasNextPage bool
				EndCursor   graphql.String
			}
		} `graphql:"search(query:$searchQuery, type: ISSUE, first: 100, after: $cursor)"`
	}

	variables := map[string]interface{}{
		"cursor":      (*graphql.String)(nil),
		"searchQuery": graphql.String(fmt.Sprintf(`%s is:merged merged:%s`, graphql.String(repo.SearchString()), graphql.String(timespanSearchQuery(since, until)))),
	}

	//DEBUG: fmt.Printf("★ search: \t %s\n", variables["searchQuery"])

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

func getPrCategory(pr Pr) PrCategory {
	title := strings.ToLower(pr.Title())
	// features
	if strings.HasPrefix(title, "feat") {
		// TODO: or has label 'Type: Improvement'
		return FEATURES
	}
	// bug fixes
	if strings.HasPrefix(title, "bug") || strings.HasPrefix(title, "fix") {
		// TODO: or has label 'Type: Bug'
		return BUG_FIXES
	}
	// maintenance
	if strings.HasPrefix(title, "chore") || strings.HasPrefix(title, "refactor") {
		// TODO: or has label 'Topic: Stabilization' or 'Type: Chore' or 'Type: Refactoring'
		return MAINTENANCE
	}
	// documentation
	if strings.HasPrefix(title, "doc") {
		// TODO: or has label 'Category: Doc'
		// TODO: or is targeting a tutorial repository?
		return DOCUMENTATION
	}
	// logistics
	if strings.HasPrefix(title, "build") || strings.HasPrefix(title, "ci") {
		// TODO: or has label 'Category: Build/CI'
		return LOGISTICS
	}
	// performance
	if strings.HasPrefix(title, "perf") {
		// TODO: or has label 'Category: Performance'
		return PERFORMANCE
	}
	// tests
	if strings.HasPrefix(title, "test") {
		// TODO: or has label 'Category: Test/QA'
		return TESTS
	}
	return GENERAL
}

func changelog(opts *ChangelogOptions) {
	// create clients to talk to GitHub's REST (v3) or GraphQL (v4) API
	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	gql, err := gh.GQLClient(nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// which repository or organization are we targeting?
	repo, err := repository(opts.Repo)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//DEBUG: fmt.Printf("★ target: \t %s\n", repo.SearchString())

	// what is the starting point of our changelog?
	since, err := since(client, opts.Since, repo)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//DEBUG: fmt.Printf("★ since : \t %s\n", since)

	// let's get all the PRs in some order defined by github (probabyl chronologically sorted by merge date)
	prs, err := mergedPrsSince(gql, repo, since, opts.Until)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if opts.Pretty {
		fmt.Println(prettyPrint(prs))
	} else {
		// group by repository (in case we are targeting all repos of an organization)
		prsByRepo := fun.GroupBy(prs, func(p pr) string { return p.PullRequest.Repository.NameWithOwner })

		keys := make([]string, 0, len(prsByRepo))
		for k := range prsByRepo {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// TODO: generate a "pretty printed" output, sorting changes into buckets such as "Features" depending on the PR title prefixes

		shouldPrintRepoName := len(keys) > 1
		for _, r := range keys {
			repoPrefix := ""
			if shouldPrintRepoName {
				repoPrefix = r
			}
			for _, pr := range prsByRepo[r] {
				line := fmt.Sprintf("%s#%d %s (@%s)", repoPrefix, pr.PullRequest.Number, pr.PullRequest.Title, pr.PullRequest.Author.User.Login)
				fmt.Println(line)
			}
		}
	}
}

func addSection(sb *strings.Builder, category PrCategory, prs []pr, usePrefix bool) {
	sb.WriteString("## " + category.Pretty() + "\n")
	sb.WriteString("\n")
	addLines(sb, prs, usePrefix)
	sb.WriteString("\n")
}

func addLines(sb *strings.Builder, prs []pr, usePrefix bool) {
	for _, pr := range prs {
		repoPrefix := ""
		if usePrefix {
			repoPrefix = pr.PullRequest.Repository.NameWithOwner
		}
		line := fmt.Sprintf("- %s#%d %s (@%s)\n", repoPrefix, pr.PullRequest.Number, pr.PullRequest.Title, pr.PullRequest.Author.User.Login)
		sb.WriteString(line)
	}
}

func prettyPrint(prs []pr) string {

	var changelog = make(map[PrCategory][]pr)
	// the set of repositories that are present in the changelog (we don't have a set data structure, therefore using a map)
	var repos = make(map[string]struct{})

	for _, pr := range prs {
		category := getPrCategory(pr)
		changelog[category] = append(changelog[category], pr)
		repos[pr.PullRequest.Repository.NameWithOwner] = struct{}{}
	}

	// sort by repsoitory name so that changes within the same repo appear together
	for _, prs := range changelog {
		sort.SliceStable(prs, func(i, j int) bool {
			return prs[i].PullRequest.Repository.NameWithOwner < prs[j].PullRequest.Repository.NameWithOwner
		})
	}

	shouldPrintRepoName := len(repos) > 1

	var sb strings.Builder
	sb.WriteString("# Changelog\n\n")
	addSection(&sb, FEATURES, changelog[FEATURES], shouldPrintRepoName)
	addSection(&sb, BUG_FIXES, changelog[BUG_FIXES], shouldPrintRepoName)
	addSection(&sb, MAINTENANCE, changelog[MAINTENANCE], shouldPrintRepoName)
	addSection(&sb, TESTS, changelog[TESTS], shouldPrintRepoName)
	addSection(&sb, PERFORMANCE, changelog[PERFORMANCE], shouldPrintRepoName)
	addSection(&sb, DOCUMENTATION, changelog[DOCUMENTATION], shouldPrintRepoName)
	addSection(&sb, LOGISTICS, changelog[LOGISTICS], shouldPrintRepoName)
	addSection(&sb, GENERAL, changelog[GENERAL], shouldPrintRepoName)

	return sb.String()
}
