package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type areaStats struct {
	Area    string  `json:"area"`
	Files   int     `json:"files"`
	Covered int     `json:"covered_statements"`
	Total   int     `json:"total_statements"`
	Percent float64 `json:"percent"`
}

type report struct {
	GeneratedAt  string      `json:"generated_at"`
	Coverprofile string      `json:"coverprofile"`
	Areas        []areaStats `json:"areas"`
	Total        areaStats   `json:"total"`
}

func main() {
	var coverprofile string
	var outJSON string
	var outMD string

	flag.StringVar(&coverprofile, "coverprofile", "", "path to coverprofile (e.g. artifacts/coverage/internal.out)")
	flag.StringVar(&outJSON, "out-json", "", "write JSON report to path (optional)")
	flag.StringVar(&outMD, "out-md", "", "write Markdown report to path (optional)")
	flag.Parse()

	if strings.TrimSpace(coverprofile) == "" {
		fatalf("missing -coverprofile")
	}

	areaByFile := map[string]string{}
	filesByArea := map[string]map[string]struct{}{}
	coveredByArea := map[string]int{}
	totalByArea := map[string]int{}

	coveredTotal := 0
	totalTotal := 0
	filesTotal := map[string]struct{}{}

	f, err := os.Open(coverprofile)
	if err != nil {
		fatalf("open coverprofile: %v", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "mode:") {
			continue
		}

		// Format: <file>:<range> <numStatements> <count>
		parts := strings.Fields(line)
		if len(parts) != 3 {
			fatalf("invalid coverprofile line %d: %q", lineNo, line)
		}

		fileField := parts[0]
		colonIdx := strings.LastIndex(fileField, ":")
		if colonIdx <= 0 {
			fatalf("invalid file field line %d: %q", lineNo, fileField)
		}
		filePath := fileField[:colonIdx]

		numStatements, err := strconv.Atoi(parts[1])
		if err != nil {
			fatalf("invalid numStatements line %d: %v", lineNo, err)
		}
		count, err := strconv.Atoi(parts[2])
		if err != nil {
			fatalf("invalid count line %d: %v", lineNo, err)
		}

		area := areaForFile(filePath)
		areaByFile[filePath] = area

		if _, ok := filesByArea[area]; !ok {
			filesByArea[area] = map[string]struct{}{}
		}
		filesByArea[area][filePath] = struct{}{}
		filesTotal[filePath] = struct{}{}

		totalByArea[area] += numStatements
		totalTotal += numStatements

		if count > 0 {
			coveredByArea[area] += numStatements
			coveredTotal += numStatements
		}
	}
	if err := sc.Err(); err != nil {
		fatalf("scan coverprofile: %v", err)
	}
	_ = areaByFile

	areas := make([]areaStats, 0, len(totalByArea))
	for area, total := range totalByArea {
		covered := coveredByArea[area]
		files := len(filesByArea[area])
		areas = append(areas, areaStats{
			Area:    area,
			Files:   files,
			Covered: covered,
			Total:   total,
			Percent: percent(covered, total),
		})
	}
	sort.Slice(areas, func(i, j int) bool { return areas[i].Area < areas[j].Area })

	totalStats := areaStats{
		Area:    "total",
		Files:   len(filesTotal),
		Covered: coveredTotal,
		Total:   totalTotal,
		Percent: percent(coveredTotal, totalTotal),
	}

	rep := report{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Coverprofile: coverprofile,
		Areas:        areas,
		Total:        totalStats,
	}

	if strings.TrimSpace(outJSON) != "" {
		writeJSON(outJSON, rep)
	}
	if strings.TrimSpace(outMD) != "" {
		writeMD(outMD, rep)
	}

	// Always print minimal summary so script users can see result.
	fmt.Printf("coverage_total=%.1f%% statements=%d files=%d\n", rep.Total.Percent, rep.Total.Total, rep.Total.Files)
}

func areaForFile(rawPath string) string {
	p := strings.TrimSpace(rawPath)
	p = strings.TrimPrefix(p, "./")

	// Convert module import paths into repo-relative paths when possible.
	if idx := strings.Index(p, "/internal/"); idx >= 0 {
		p = p[idx+1:] // keep "internal/..."
	} else if idx := strings.Index(p, "/cmd/"); idx >= 0 {
		p = p[idx+1:] // keep "cmd/..."
	}

	p = filepath.ToSlash(p)

	switch {
	case strings.HasPrefix(p, "internal/modules/"):
		parts := strings.Split(p, "/")
		if len(parts) >= 3 && strings.TrimSpace(parts[2]) != "" {
			return "modules/" + parts[2]
		}
		return "modules"
	case strings.HasPrefix(p, "internal/platform/"):
		parts := strings.Split(p, "/")
		if len(parts) >= 3 && strings.TrimSpace(parts[2]) != "" {
			return "platform/" + parts[2]
		}
		return "platform"
	case strings.HasPrefix(p, "internal/app/"):
		parts := strings.Split(p, "/")
		if len(parts) >= 3 && strings.TrimSpace(parts[2]) != "" {
			return "app/" + parts[2]
		}
		return "app"
	case strings.HasPrefix(p, "internal/shared/"):
		parts := strings.Split(p, "/")
		if len(parts) >= 3 && strings.TrimSpace(parts[2]) != "" {
			return "shared/" + parts[2]
		}
		return "shared"
	case strings.HasPrefix(p, "internal/"):
		return "internal/other"
	case strings.HasPrefix(p, "cmd/"):
		parts := strings.Split(p, "/")
		if len(parts) >= 2 && strings.TrimSpace(parts[1]) != "" {
			return "cmd/" + parts[1]
		}
		return "cmd"
	default:
		return "other"
	}
}

func percent(covered, total int) float64 {
	if total <= 0 {
		return 0
	}
	return (float64(covered) / float64(total)) * 100
}

func writeJSON(path string, payload any) {
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fatalf("marshal json: %v", err)
	}
	encoded = append(encoded, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		fatalf("write json: %v", err)
	}
}

func writeMD(path string, rep report) {
	var b strings.Builder
	b.WriteString("# Coverage Report (By Area)\n\n")
	b.WriteString(fmt.Sprintf("- generated_at: `%s`\n", rep.GeneratedAt))
	b.WriteString(fmt.Sprintf("- coverprofile: `%s`\n\n", rep.Coverprofile))

	b.WriteString("| area | covered | total | percent | files |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, a := range rep.Areas {
		b.WriteString(fmt.Sprintf("| %s | %d | %d | %.1f%% | %d |\n", a.Area, a.Covered, a.Total, a.Percent, a.Files))
	}
	b.WriteString(fmt.Sprintf("| **%s** | **%d** | **%d** | **%.1f%%** | **%d** |\n", rep.Total.Area, rep.Total.Covered, rep.Total.Total, rep.Total.Percent, rep.Total.Files))

	b.WriteString("\nNotes:\n")
	b.WriteString("- statements aggregated from coverprofile blocks; covered statements counted when block count > 0.\n")
	b.WriteString("- area grouping based on file path prefixes under `internal/`.\n")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		fatalf("write md: %v", err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
