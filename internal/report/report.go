package report

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/Qwental/port-scanner-alert-system/internal/diff"
	"github.com/Qwental/port-scanner-alert-system/internal/model"
)

func BuildScanReport(results []model.ScanResult) string {
	if len(results) == 0 {
		return "Scan completed. No open ports found."
	}

	var b strings.Builder

	grouped := make(map[string][]model.ScanResult)
	var hosts []string

	for _, r := range results {
		if _, exists := grouped[r.IP]; !exists {
			hosts = append(hosts, r.IP)
		}
		grouped[r.IP] = append(grouped[r.IP], r)
	}

	sort.Strings(hosts)

	b.WriteString("\n SCAN REPORT \n")
	b.WriteString(fmt.Sprintf("Hosts: %d | Open ports: %d\n", len(hosts), len(results)))

	for _, ip := range hosts {
		ports := grouped[ip]
		sort.Slice(ports, func(i, j int) bool {
			return ports[i].Port < ports[j].Port
		})

		var portList []string
		for _, p := range ports {
			portList = append(portList, fmt.Sprintf("%d/%s", p.Port, p.Proto))
		}

		b.WriteString(fmt.Sprintf("%-18s  %s\n", ip, strings.Join(portList, ", ")))
	}

	b.WriteString("\n END\n")
	return b.String()
}

func BuildDiffReport(d diff.DiffResult) string {
	if len(d.New) == 0 && len(d.Changed) == 0 && len(d.Closed) == 0 {
		return "\nDIFF: No changes detected\n"
	}

	var b strings.Builder

	b.WriteString("\nDIFF REPORT\n\n")

	if len(d.New) > 0 {
		b.WriteString(fmt.Sprintf("[NEW] (%d):\n", len(d.New)))
		for _, r := range d.New {
			if r.Banner != "" {
				b.WriteString(fmt.Sprintf("   + %s:%d/%s  [%s]\n", r.IP, r.Port, r.Proto, r.Banner))
			} else {
				b.WriteString(fmt.Sprintf("   + %s:%d/%s\n", r.IP, r.Port, r.Proto))
			}
		}
		b.WriteString("\n")
	}

	if len(d.Changed) > 0 {
		b.WriteString(fmt.Sprintf("[CHANGED] (%d):\n", len(d.Changed)))
		for _, r := range d.Changed {
			b.WriteString(fmt.Sprintf("   ~ %s:%d/%s  [%s]\n", r.IP, r.Port, r.Proto, r.Banner))
		}
		b.WriteString("\n")
	}

	if len(d.Closed) > 0 {
		b.WriteString(fmt.Sprintf("[CLOSED] (%d):\n", len(d.Closed)))
		for _, r := range d.Closed {
			b.WriteString(fmt.Sprintf("   - %s:%d/%s\n", r.IP, r.Port, r.Proto))
		}
		b.WriteString("\n")
	}

	b.WriteString("END DIFF\n")
	return b.String()
}

func BuildDiffHTML(d diff.DiffResult) string {
	if len(d.New) == 0 && len(d.Changed) == 0 && len(d.Closed) == 0 {
		return "<h3>No changes detected</h3>"
	}

	var b strings.Builder

	b.WriteString("<h2>Scan Diff Report</h2>")

	if len(d.New) > 0 {
		b.WriteString(fmt.Sprintf("<h3>New ports (%d)</h3><ul>", len(d.New)))
		for _, r := range d.New {
			b.WriteString(fmt.Sprintf("<li><b>%s:%d/%s</b> %s</li>",
				html.EscapeString(r.IP), r.Port, r.Proto,
				html.EscapeString(r.Banner)))
		}
		b.WriteString("</ul>")
	}

	if len(d.Changed) > 0 {
		b.WriteString(fmt.Sprintf("<h3>Changed (%d)</h3><ul>", len(d.Changed)))
		for _, r := range d.Changed {
			b.WriteString(fmt.Sprintf("<li><b>%s:%d/%s</b> — %s</li>",
				html.EscapeString(r.IP), r.Port, r.Proto,
				html.EscapeString(r.Banner)))
		}
		b.WriteString("</ul>")
	}

	if len(d.Closed) > 0 {
		b.WriteString(fmt.Sprintf("<h3>Closed (%d)</h3><ul>", len(d.Closed)))
		for _, r := range d.Closed {
			b.WriteString(fmt.Sprintf("<li><b>%s:%d/%s</b></li>",
				html.EscapeString(r.IP), r.Port, r.Proto))
		}
		b.WriteString("</ul>")
	}

	return b.String()
}
