package diff

import (
	"github.com/Qwental/port-scanner-alert-system/internal/model"
)

type DiffResult struct {
	New     []model.ScanResult 
	Changed []model.ScanResult 
	Closed  []model.ScanResult 
}

func Compare(current []model.ScanResult, previous map[string]model.ScanResult) DiffResult {
	var result DiffResult

	currentMap := make(map[string]model.ScanResult, len(current))

	for _, r := range current {
		key := r.Key()
		currentMap[key] = r

		prev, exists := previous[key]
		if !exists {
			result.New = append(result.New, r)
		} else if r.Banner != "" && r.Banner != prev.Banner {
			result.Changed = append(result.Changed, r)
		}
	}

	for key, prev := range previous {
		if _, exists := currentMap[key]; !exists {
			result.Closed = append(result.Closed, prev)
		}
	}

	return result
}