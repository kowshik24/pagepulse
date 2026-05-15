package app

import (
	"flag"
	"fmt"
	"time"
)

type Config struct {
	Host           string
	Port           int
	SampleInterval time.Duration
}

func ParseConfig(args []string) (Config, error) {
	cfg := Config{}
	fs := flag.NewFlagSet("pagepulse", flag.ContinueOnError)

	public := fs.Bool("public", false, "listen on all interfaces (same as --host 0.0.0.0)")
	fs.StringVar(&cfg.Host, "host", "127.0.0.1", "host interface to bind")
	fs.IntVar(&cfg.Port, "port", 8421, "port to bind")
	fs.DurationVar(&cfg.SampleInterval, "sample-interval", time.Second, "sampling interval (e.g. 1s, 2s)")

	if err := fs.Parse(args); err != nil {
		return cfg, err
	}

	if *public {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return cfg, fmt.Errorf("port must be between 1 and 65535")
	}
	if cfg.SampleInterval < 250*time.Millisecond {
		return cfg, fmt.Errorf("sample-interval must be at least 250ms")
	}
	return cfg, nil
}
