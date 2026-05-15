package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/example/vial-helper/internal/app"
	"github.com/example/vial-helper/internal/buildinfo"
	"github.com/example/vial-helper/internal/config"

	hid "github.com/sstallion/go-hid"
)

func main() {
	var (
		configPath string
		command    string
		rawDoctor  bool
		showVer    bool
	)

	flag.StringVar(&configPath, "config", "", "Path to config.toml. Default: OS user config dir / vial-helper / config.toml")
	flag.StringVar(&command, "command", "run", "Command: run | refresh-layout | refresh-now | paths | init | doctor | status | version")
	flag.BoolVar(&rawDoctor, "raw", false, "With --command doctor, also print the decoded layout JSON snapshot")
	flag.BoolVar(&showVer, "version", false, "Print build version and exit")
	flag.Parse()

	if showVer || command == "version" {
		fmt.Println(buildinfo.String())
		return
	}

	if err := hid.Init(); err != nil {
		log.Fatalf("hid init: %v", err)
	}
	defer func() { _ = hid.Exit() }()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	cfg, paths, err := config.Resolve(configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	switch command {
	case "run":
		if err := app.New(cfg, paths).Run(); err != nil {
			log.Fatalf("daemon: %v", err)
		}
	case "refresh-layout":
		if err := app.TouchRefreshFlag(paths); err != nil {
			log.Fatalf("refresh flag: %v", err)
		}
		fmt.Println(paths.RefreshFlag)
	case "refresh-now":
		if err := app.New(cfg, paths).RefreshOnce(); err != nil {
			log.Fatalf("refresh layout: %v", err)
		}
	case "paths":
		fmt.Printf("config=%s\n", paths.ConfigFile)
		fmt.Printf("state=%s\n", paths.StateFile)
		fmt.Printf("layout=%s\n", paths.LayoutFile)
		fmt.Printf("refresh_flag=%s\n", paths.RefreshFlag)
	case "init":
		fmt.Println(paths.ConfigFile)
	case "doctor":
		if err := app.Doctor(cfg, paths, app.DoctorOptions{Raw: rawDoctor}, os.Stdout); err != nil {
			os.Exit(1)
		}
	case "status":
		if err := app.Status(paths, os.Stdout); err != nil {
			log.Fatalf("status: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		os.Exit(2)
	}
}
