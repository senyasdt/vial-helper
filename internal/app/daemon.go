package app

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/example/vial-helper/internal/config"
	"github.com/example/vial-helper/internal/hidclient"
	"github.com/example/vial-helper/internal/model"
	"github.com/example/vial-helper/internal/protocol"
	"github.com/example/vial-helper/internal/storage"
)

type Daemon struct {
	Cfg   config.Config
	Paths config.Paths
}

func New(cfg config.Config, paths config.Paths) *Daemon {
	return &Daemon{Cfg: cfg, Paths: paths}
}

func (d *Daemon) Run() error {
	if err := storage.WriteJSONAtomic(d.Paths.StateFile, OfflineState("INIT")); err != nil {
		log.Printf("write initial state: %v", err)
	}

	var lastState *model.State
	connected := false

	for {
		client, err := hidclient.Open(
			d.Cfg,
			time.Duration(d.Cfg.Polling.RequestTimeoutMS)*time.Millisecond,
			protocol.Probe,
		)
		if err != nil {
			state := OfflineState("OFF")
			if lastState == nil || !reflect.DeepEqual(*lastState, state) {
				if writeErr := storage.WriteJSONAtomic(d.Paths.StateFile, state); writeErr != nil {
					log.Printf("write OFF state: %v", writeErr)
				}
				lastState = &state
			}
			if connected || lastState == nil {
				log.Printf("device unavailable: %v", err)
			}
			connected = false
			time.Sleep(time.Duration(d.Cfg.Polling.ReconnectMS) * time.Millisecond)
			continue
		}

		reader := protocol.NewReader(client, d.Cfg)
		connected = true

		if err := d.refreshLayout(reader); err != nil {
			log.Printf("initial layout refresh failed: %v", err)
			_ = client.Close()
			connected = false
			time.Sleep(time.Duration(d.Cfg.Polling.ReconnectMS) * time.Millisecond)
			continue
		}

		for {
			if d.refreshRequested() {
				if err := d.refreshLayout(reader); err != nil {
					log.Printf("layout refresh failed: %v", err)
					_ = client.Close()
					connected = false
					break
				}
				d.clearRefreshFlag()
			}

			state, err := reader.LayerState(d.Cfg.Layers)
			if err != nil {
				log.Printf("layer read failed: %v", err)
				_ = client.Close()
				connected = false
				break
			}

			if lastState == nil || !reflect.DeepEqual(*lastState, state) {
				if err := storage.WriteJSONAtomic(d.Paths.StateFile, state); err != nil {
					log.Printf("write state: %v", err)
				}
				lastState = &state
			}

			time.Sleep(time.Duration(d.Cfg.Polling.LayerPollMS) * time.Millisecond)
		}
	}
}

func (d *Daemon) RefreshOnce() error {
	client, err := hidclient.Open(
		d.Cfg,
		time.Duration(d.Cfg.Polling.RequestTimeoutMS)*time.Millisecond,
		protocol.Probe,
	)
	if err != nil {
		return err
	}
	defer client.Close()
	reader := protocol.NewReader(client, d.Cfg)
	return d.refreshLayout(reader)
}

func (d *Daemon) refreshLayout(reader *protocol.Reader) error {
	layout, err := reader.Layout()
	if err != nil {
		return err
	}
	if err := storage.WriteJSONAtomic(d.Paths.LayoutFile, layout); err != nil {
		return err
	}
	return nil
}

func (d *Daemon) refreshRequested() bool {
	_, err := os.Stat(d.Paths.RefreshFlag)
	return err == nil
}

func (d *Daemon) clearRefreshFlag() {
	_ = os.Remove(d.Paths.RefreshFlag)
}

func OfflineState(label string) model.State {
	return model.State{
		Label:     label,
		Top:       -1,
		Name:      "Keyboard disconnected",
		Effective: "—",
		Temp:      "—",
		Default:   "—",
		Tooltip:   "Vial helper device is not connected",
	}
}

func TouchRefreshFlag(paths config.Paths) error {
	f, err := os.Create(paths.RefreshFlag)
	if err != nil {
		return fmt.Errorf("create refresh flag: %w", err)
	}
	return f.Close()
}
