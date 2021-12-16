package main

import (
	"fmt"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
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

func mergedPrsSince(client api.GQLClient) {

	var query struct {
		Search struct {
			Nodes []struct {
				PullRequest struct {
					Title string
				} `graphql:"... on PullRequest"`
			}
		} `graphql:"search(query:\"repo:MovingBlocks/Terasology is:merged merged:>=2021-10-01\", type: ISSUE, first: 20)"`
	}

	variables := map[string]interface{}{}

	err := client.Query("Changelog", &query, variables)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(query)
}

func main() {
	fmt.Println("hi world, this is the gh-terasology extension!")
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
	fmt.Printf("running as %s\n", response.Login)

	publishedAt, err := dateOfLastRelease(client, "MovingBlocks", "Terasology")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("latest release published at %s\n", publishedAt)

	gql, err := gh.GQLClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	mergedPrsSince(gql)
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
