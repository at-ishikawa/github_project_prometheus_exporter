FROM golang:1.22 AS builder
COPY . /app
WORKDIR /app
# https://stackoverflow.com/questions/55106186/no-such-file-or-directory-with-docker-scratch-image
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main ./cmd

FROM scratch
COPY --from=builder /app/main /github_project_prometheus_exporter
# https://stackoverflow.com/questions/75696690/how-to-resolve-tls-failed-to-verify-certificate-x509-certificate-signed-by-un
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/github_project_prometheus_exporter"]
CMD ["exporter"]
