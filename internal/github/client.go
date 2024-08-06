package github

//go:generate go run github.com/Khan/genqlient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

type Client struct {
	graphQLClient graphql.Client
}

func NewClient(githubToken string) (*Client, error) {
	httpClient := http.Client{
		Transport: &authedTransport{
			key:     githubToken,
			wrapped: http.DefaultTransport,
		},
	}

	return &Client{
		graphQLClient: graphql.NewClient("https://api.github.com/graphql", &httpClient),
	}, nil
}

// See https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects#finding-information-about-projects
func (client *Client) FetchUserProject(ctx context.Context, userId string, projectNumber int) (string, error) {
	response, err := FetchUserProject(ctx, client.graphQLClient, userId, projectNumber)
	if err != nil {
		return "", fmt.Errorf("FetchUserProject: %w", err)
	}

	return response.User.ProjectV2.Id, nil
}
