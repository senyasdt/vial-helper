package hidclient

import (
	"errors"
	"fmt"
	"time"

	"github.com/example/vial-helper/internal/config"
	hid "github.com/sstallion/go-hid"
)

const (
	ReportPayloadSize = 32
	WriteReportSize   = 33 // report id + payload
)

type Client struct {
	dev     *hid.Device
	timeout time.Duration
}

func Open(cfg config.Config, timeout time.Duration, probe func(*Client) error) (*Client, error) {
	var candidates []hid.DeviceInfo
	err := hid.Enumerate(cfg.Device.VID, cfg.Device.PID, func(info *hid.DeviceInfo) error {
		candidates = append(candidates, *info)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("enumerate HID devices: %w", err)
	}
	if len(candidates) == 0 {
		return nil, errors.New("no matching HID devices")
	}

	ordered := make([]hid.DeviceInfo, 0, len(candidates))
	for _, info := range candidates {
		if info.UsagePage == cfg.Device.UsagePage && info.Usage == cfg.Device.Usage {
			ordered = append(ordered, info)
		}
	}
	for _, info := range candidates {
		if info.UsagePage != cfg.Device.UsagePage || info.Usage != cfg.Device.Usage {
			ordered = append(ordered, info)
		}
	}

	var lastErr error
	for _, info := range ordered {
		dev, err := hid.OpenPath(info.Path)
		if err != nil {
			lastErr = err
			continue
		}
		client := &Client{dev: dev, timeout: timeout}
		if probe == nil || probe(client) == nil {
			return client, nil
		}
		lastErr = errors.New("helper API probe failed")
		_ = dev.Close()
	}
	if lastErr == nil {
		lastErr = errors.New("no usable HID device")
	}
	return nil, lastErr
}

func (c *Client) Close() error {
	if c == nil || c.dev == nil {
		return nil
	}
	return c.dev.Close()
}

func (c *Client) Send(payload []byte, expectedCmd byte) ([]byte, error) {
	if len(payload) > ReportPayloadSize {
		return nil, fmt.Errorf("payload too large: %d", len(payload))
	}
	writeBuf := make([]byte, WriteReportSize)
	writeBuf[0] = 0x00
	copy(writeBuf[1:], payload)
	if _, err := c.dev.Write(writeBuf); err != nil {
		return nil, fmt.Errorf("HID write: %w", err)
	}

	readBuf := make([]byte, WriteReportSize)
	n, err := c.dev.ReadWithTimeout(readBuf, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("HID read: %w", err)
	}
	if n == 0 {
		return nil, errors.New("empty HID response")
	}
	resp := normalize(readBuf[:n])
	if len(resp) == 0 {
		return nil, errors.New("normalized HID response is empty")
	}
	if resp[0] != expectedCmd {
		return nil, fmt.Errorf("unexpected response command: expected 0x%02X, got 0x%02X", expectedCmd, resp[0])
	}
	return resp, nil
}

func normalize(data []byte) []byte {
	if len(data) == WriteReportSize && data[0] == 0x00 {
		return data[1:]
	}
	if len(data) > ReportPayloadSize {
		return data[:ReportPayloadSize]
	}
	return data
}
