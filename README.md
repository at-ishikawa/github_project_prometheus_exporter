# github_project_prometheus_exporter

Export following metrics from a GitHub project.

| Name                       | Type  | Labels                                                                                                                                                       |
| -------------------------- | ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| github_project_items_count | Gauge | project, user, [single select fields](https://docs.github.com/en/issues/planning-and-tracking-with-projects/understanding-fields/about-single-select-fields) |

## Container

Container is hosted in [DockerHub](https://hub.docker.com/r/atishikawa/github_project_prometheus_exporter).
To run a container using a github token from GitHub CLI,

```shell
docker run --env GITHUB_TOKEN=$(gh auth token) -p 11111:11111 atishikawa/github_project_prometheus_exporter:0.1.0 $GITHUB_USER
```

## Development

- Use [genqlient](https://github.com/Khan/genqlient/tree/main) for GraphQL client
- Use [GitHub Public GraphQL Schema](https://docs.github.com/en/graphql/overview/public-schema)
