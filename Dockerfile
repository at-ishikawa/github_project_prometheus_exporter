FROM golang:1.22 AS builder
WORKDIR /app
COPY . /app
RUN go build -o main ./cmd

FROM scratch
COPY --from=builder /app/main /github_project_prometheus_exporter
CMD ["/github_project_prometheus_exporter"]
