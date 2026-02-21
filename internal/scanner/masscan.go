package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

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
	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets specified")
	}

	// single target — no need for goroutines
	if len(targets) == 1 {
		return m.scanTarget(ctx, targets[0])
	}

	var (
		mu         sync.Mutex
		wg         sync.WaitGroup
		allResults []model.ScanResult
		errs       []error
	)

	for _, target := range targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()

			m.log.Infof("Starting scan for target: %s", t)
			results, err := m.scanTarget(ctx, t)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errs = append(errs, fmt.Errorf("target %s: %w", t, err))
				return
			}
			allResults = append(allResults, results...)
		}(target)
	}

	wg.Wait()

	if len(errs) > 0 {
		for _, e := range errs {
			m.log.Errorf("Scan error: %v", e)
		}
	}

	m.log.Infof("Scan complete: %d results from %d targets", len(allResults), len(targets))
	return allResults, nil
}

func (m *MasscanWrapper) scanTarget(ctx context.Context, target string) ([]model.ScanResult, error) {
	args := []string{
		"-oJ", "-",
		"--banners",
		"--rate", m.cfg.Rate,
		"-p", m.cfg.Ports,
	}

	if m.cfg.Interface != "" {
		args = append(args, "-e", m.cfg.Interface)
	}

	args = append(args, target)

	m.log.Infof("Masscan args: %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "masscan", args...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start masscan: %w", err)
	}

	var results []model.ScanResult
	decoder := json.NewDecoder(stdout)

	if _, err := decoder.Token(); err != nil {
		m.log.Warnf("No scan output for %s: %v", target, err)
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

	_ = cmd.Wait()
	return results, nil
}
