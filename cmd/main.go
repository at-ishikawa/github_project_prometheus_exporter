package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/at-ishikawa/github_project_prometheus_exporter/internal/github"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	otelsdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"golang.org/x/sync/errgroup"
)

const (
	port           = 11111
	exitCodeOK int = 0
	meterName      = "github.com/at-ishikawa/github_project_prometheus_exporter"
)

func main() {
	exitCode, err := runMain()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(exitCode)
}

// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go#L15C1-L94C2
func newPrometheusProvider() (*otelsdkmetric.MeterProvider, error) {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("prometheus.New: %w", err)
	}

	return otelsdkmetric.NewMeterProvider(otelsdkmetric.WithReader(exporter)), nil
}

func startMetricsServer(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		fmt.Printf("serving metrics at localhost:%d/metrics\n", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
			fmt.Printf("error serving http: %v", err)
			return
		}
	}()
}

type App struct {
	meter         otelmetric.Meter
	meterProvider *otelsdkmetric.MeterProvider
}

func NewApp() (*App, error) {
	provider, err := newPrometheusProvider()
	if err != nil {
		return nil, fmt.Errorf("setupPrometheus: %w", err)
	}

	return &App{
		meterProvider: provider,
		meter:         provider.Meter(meterName),
	}, nil
}

func (app App) Run(ctx context.Context, f func(ctx context.Context) error) error {
	if err := f(ctx); err != nil {
		return err
	}

	ctx, _ = signal.NotifyContext(ctx,
		os.Interrupt,
		os.Kill,
	)
	<-ctx.Done()

	app.meterProvider.Shutdown(ctx)
	return nil
}

func runMain() (int, error) {
	exitCode := 1

	rootCommand := cobra.Command{
		Use:  "exporter [userId]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// TODO: Support both of env vars and arguments
			userId := args[0]

			githubToken := os.Getenv("GITHUB_TOKEN")
			if githubToken == "" {
				return fmt.Errorf("GITHUB_TOKEN environment variable is required")
			}

			client, err := github.NewClient(githubToken)
			if err != nil {
				return fmt.Errorf("github.NewClient: %w", err)
			}

			app, err := NewApp()
			if err != nil {
				return fmt.Errorf("NewApp: %w", err)
			}
			return app.Run(ctx, func(ctx context.Context) error {
				// https://pkg.go.dev/go.opentelemetry.io/otel/metric#section-readme
				defaultAttributes := []attribute.KeyValue{
					attribute.Key("user").String(userId),
				}
				if _, err := app.meter.Int64ObservableGauge("github_project_items_count",
					otelmetric.WithDescription("The number of items in a GitHub project"),
					otelmetric.WithInt64Callback(func(ctx context.Context, observer otelmetric.Int64Observer) error {
						projects, err := client.FetchUserProjects(ctx, userId)
						if err != nil {
							return fmt.Errorf("client.FetchUserProject: %w", err)
						}

						eg, childCtx := errgroup.WithContext(ctx)
						for _, project := range projects {
							project := project
							eg.Go(func() error {
								stats, err := client.FetchProjectStats(childCtx, project.ID)
								if err != nil {
									return fmt.Errorf("client.FetchProjectStats: %w", err)
								}

								for fieldName, statValues := range stats {
									for fieldValue, fieldCount := range statValues {
										observer.Observe(int64(fieldCount), otelmetric.WithAttributes(
											attribute.Key("project").String(project.Title),
											attribute.Key(strings.ToLower(fieldName)).String(fieldValue),
										), otelmetric.WithAttributes(
											defaultAttributes...,
										))
									}
								}
								return nil
							})
						}

						return eg.Wait()
					}),
				); err != nil {
					return fmt.Errorf("meter.Int64ObservableGauge: %w", err)
				}

				startMetricsServer(port)
				return nil
			})
		},
	}

	if err := rootCommand.Execute(); err != nil {
		return exitCode, err
	}
	return exitCodeOK, nil
}
