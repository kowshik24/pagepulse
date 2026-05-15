package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"pagepulse/internal/buildinfo"
	"pagepulse/internal/metrics"
	"pagepulse/internal/web"
)

type Instance struct {
	cfg       Config
	collector *metrics.Collector
	httpSrv   *http.Server
}

func New(cfg Config) (*Instance, error) {
	collector, err := metrics.NewCollector(cfg.SampleInterval)
	if err != nil {
		return nil, fmt.Errorf("collector init failed: %w", err)
	}
	server, err := web.NewServer(collector, buildinfo.Current())
	if err != nil {
		return nil, fmt.Errorf("web server init failed: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Instance{cfg: cfg, collector: collector, httpSrv: httpSrv}, nil
}

func (i *Instance) Run(ctx context.Context) error {
	go i.collector.Run(ctx)

	printStartupInfo(i.cfg)
	log.Printf("pagepulse listening on http://%s", i.httpSrv.Addr)
	return web.RunHTTP(ctx, i.httpSrv)
}

func printStartupInfo(cfg Config) {
	urlHost := cfg.Host
	if urlHost == "0.0.0.0" {
		urlHost = "localhost"
	}
	fmt.Printf("PagePulse started\n")
	fmt.Printf("Dashboard URL: http://%s:%d\n", urlHost, cfg.Port)
	fmt.Printf("JSON API: http://%s:%d/api/v1/summary\n", urlHost, cfg.Port)
	fmt.Printf("SSE stream: http://%s:%d/api/v1/stream\n", urlHost, cfg.Port)
	fmt.Printf("\nIf running remotely, tunnel with:\n")
	fmt.Printf("ssh -L %d:127.0.0.1:%d <user>@<server>\n", cfg.Port, cfg.Port)
}
