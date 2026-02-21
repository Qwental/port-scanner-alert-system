package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Qwental/port-scanner-alert-system/internal/config"
	"github.com/Qwental/port-scanner-alert-system/internal/model"
	"go.uber.org/zap"
)

type MasscanWrapper struct {
	cfg config.MasscanConfig
	log *zap.SugaredLogger
}

func NewMasscanWrapper(cfg config.MasscanConfig, log *zap.SugaredLogger) *MasscanWrapper {
	return &MasscanWrapper{cfg: cfg, log: log}
}

func (m *MasscanWrapper) Run(ctx context.Context, targets []string) ([]model.ScanResult, error) {

	args := []string{"-oJ", "-", "--banners", "--rate", m.cfg.Rate, "-p", m.cfg.Ports}

	if m.cfg.Interface != "" {
		args = append(args, "-e", m.cfg.Interface)
	}

	args = append(args, targets...)

	m.log.Infof("Masscan args: %s", strings.Join(args, " "))

	cmd := exec.Command("masscan", args...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start masscan: %w", err)
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
		case <-done:
		}
	}()

	var results []model.ScanResult
	decoder := json.NewDecoder(stdout)

	if _, err := decoder.Token(); err != nil {
		close(done)
		_ = cmd.Wait()
		return results, nil
	}

	for decoder.More() {
		if ctx.Err() != nil {
			break
		}

		var raw model.MasscanOutput
		if err := decoder.Decode(&raw); err != nil {
			if ctx.Err() != nil {
				break
			}
			continue
		}

		for _, p := range raw.Ports {
			if p.Status == "open" {
				results = append(results, model.ScanResult{
					IP:     raw.IP,
					Port:   p.Port,
					Proto:  p.Proto,
					Banner: p.Service.Banner,
				})
			}
		}
	}

	close(done)
	_ = cmd.Wait()

	return results, nil
}
