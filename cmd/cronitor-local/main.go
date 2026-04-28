package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cronitorio/cronitor-local/internal/config"
)

const defaultConfigPath = "cronitor.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to cronitor YAML config file")
	validate := flag.Bool("validate", false, "validate config file and exit")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[cronitor-local] ")

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *validate {
		fmt.Printf("Config OK — %d job(s) defined.\n", len(cfg.Jobs))
		for _, job := range cfg.Jobs {
			fmt.Printf("  • %-20s  schedule=%s\n", job.Name, job.Schedule)
		}
		os.Exit(0)
	}

	log.Printf("Starting cronitor-local with %d job(s) from %s", len(cfg.Jobs), *configPath)

	// Placeholder: scheduler will be wired in a subsequent commit.
	log.Println("Scheduler not yet implemented — exiting.")
}
