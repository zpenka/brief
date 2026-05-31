package brief

import (
	"strconv"
	"strings"
)

// TODOItem is a marker (TODO/FIXME/HACK/NOTE/XXX) found on an added line in the diff.
type TODOItem struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Kind string `json:"kind"`
	Text string `json:"text"`
}

var todoMarkers = []string{"TODO", "FIXME", "HACK", "NOTE", "XXX"}

// todosInDiff scans unified diff output for markers on added lines.
func todosInDiff(diff string) []TODOItem {
	var (
		items   []TODOItem
		file    string
		lineNum int
	)

	for _, raw := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(raw, "diff --git "):
			parts := strings.Fields(raw)
			if len(parts) >= 4 {
				file = strings.TrimPrefix(parts[3], "b/")
			}
			lineNum = 0

		case strings.HasPrefix(raw, "@@ "):
			lineNum = parseHunkStart(raw)

		case strings.HasPrefix(raw, "+") && !strings.HasPrefix(raw, "+++"):
			if lineNum > 0 {
				content := raw[1:]
				if kind, text, ok := findTODO(content); ok {
					items = append(items, TODOItem{File: file, Line: lineNum, Kind: kind, Text: text})
				}
				lineNum++
			}

		case strings.HasPrefix(raw, "-") && !strings.HasPrefix(raw, "---"):
			// removed line: don't advance new-file line counter

		case !strings.HasPrefix(raw, "\\") && !strings.HasPrefix(raw, "diff") &&
			!strings.HasPrefix(raw, "index") && !strings.HasPrefix(raw, "---") &&
			!strings.HasPrefix(raw, "+++") && raw != "":
			// context line
			if lineNum > 0 {
				lineNum++
			}
		}
	}

	return items
}

// parseHunkStart extracts the new-file starting line number from "@@ -a,b +c,d @@".
func parseHunkStart(line string) int {
	i := strings.Index(line, " +")
	if i < 0 {
		return 0
	}
	s := line[i+2:]
	end := strings.IndexAny(s, ", @")
	if end < 0 {
		end = len(s)
	}
	n, _ := strconv.Atoi(s[:end])
	return n
}

// findTODO checks a line for a TODO/FIXME/HACK/NOTE/XXX marker.
func findTODO(line string) (kind, text string, ok bool) {
	upper := strings.ToUpper(line)
	for _, marker := range todoMarkers {
		idx := strings.Index(upper, marker)
		if idx < 0 {
			continue
		}
		rest := line[idx+len(marker):]
		rest = strings.TrimLeft(rest, ":( \t")
		return marker, strings.TrimSpace(rest), true
	}
	return "", "", false
}
