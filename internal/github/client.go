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

type Project struct {
	ID    string
	Title string
}

// See https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects#finding-information-about-projects
func (client *Client) FetchUserProjects(ctx context.Context, userId string) ([]Project, error) {
	response, err := FetchUserProjects(ctx, client.graphQLClient, userId)
	if err != nil {
		return nil, fmt.Errorf("FetchUserProjects: %w", err)
	}

	projects := make([]Project, len(response.User.ProjectsV2.GetNodes()))
	for i, node := range response.User.ProjectsV2.GetNodes() {
		projects[i] = Project{
			ID:    node.GetId(),
			Title: node.GetTitle(),
		}
	}
	return projects, nil
}

func (client *Client) FetchProjectStats(ctx context.Context, projectId string) (map[string]map[string]int, error) {
	cursor := ""
	stats := make(map[string]map[string]int)

	for {
		response, err := PaginateProjectItems(ctx, client.graphQLClient, projectId, cursor)
		if err != nil {
			return nil, fmt.Errorf("FetchProjectItems: %w", err)
		}

		projectNode := response.GetNode()
		if projectNode.GetTypename() != "ProjectV2" {
			return nil, fmt.Errorf("unexpected typename: %s", response.GetNode().GetTypename())
		}

		projectV2 := projectNode.(*PaginateProjectItemsNodeProjectV2)
		for _, itemNode := range projectV2.Items.GetNodes() {
			for _, fieldValueNode := range itemNode.FieldValues.GetNodes() {
				if fieldValueNode.GetTypename() != "ProjectV2ItemFieldSingleSelectValue" {
					continue
				}
				// TDODO: Replace this with https://github.com/shurcooL/githubv4
				fieldValue := fieldValueNode.(*PaginateProjectItemsNodeProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemFieldValuesProjectV2ItemFieldValueConnectionNodesProjectV2ItemFieldSingleSelectValue)
				// support only a single select
				if fieldValue.Field.GetTypename() != "ProjectV2SingleSelectField" {
					continue
				}

				field := fieldValue.Field.(*PaginateProjectItemsNodeProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemFieldValuesProjectV2ItemFieldValueConnectionNodesProjectV2ItemFieldSingleSelectValueFieldProjectV2SingleSelectField)
				fieldName := field.GetName()
				if _, ok := stats[fieldName]; !ok {
					stats[fieldName] = make(map[string]int)
				}
				valueName := fieldValue.GetName()
				stats[fieldName][valueName]++
			}
		}
		pageInfo := projectV2.Items.GetPageInfo()
		if !pageInfo.GetHasNextPage() {
			break
		}
		cursor = pageInfo.GetEndCursor()
	}

	return stats, nil
}
