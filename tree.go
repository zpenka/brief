package brief

import (
	"fmt"
	"sort"
	"strings"
)

type treeNode struct {
	children map[string]*treeNode
}

func newTreeNode() *treeNode {
	return &treeNode{children: make(map[string]*treeNode)}
}

// buildTree returns indented lines representing the file tree for the given
// slash-separated paths, collapsed at maxDepth directory levels.
func buildTree(paths []string, maxDepth int) []string {
	if len(paths) == 0 {
		return nil
	}

	root := newTreeNode()
	for _, p := range paths {
		node := root
		for _, part := range strings.Split(p, "/") {
			if part == "" {
				continue
			}
			if _, ok := node.children[part]; !ok {
				node.children[part] = newTreeNode()
			}
			node = node.children[part]
		}
	}

	var lines []string
	formatNode(root, 0, maxDepth, &lines)
	return lines
}

func formatNode(node *treeNode, depth, maxDepth int, lines *[]string) {
	var dirs, files []string
	for k, child := range node.children {
		if len(child.children) > 0 {
			dirs = append(dirs, k)
		} else {
			files = append(files, k)
		}
	}
	sort.Strings(dirs)
	sort.Strings(files)

	indent := strings.Repeat("  ", depth)

	for _, k := range dirs {
		child := node.children[k]
		if depth+1 >= maxDepth {
			n := countFiles(child)
			*lines = append(*lines, fmt.Sprintf("%s%s/ (%d files)", indent, k, n))
		} else {
			*lines = append(*lines, indent+k+"/")
			formatNode(child, depth+1, maxDepth, lines)
		}
	}
	for _, k := range files {
		*lines = append(*lines, indent+k)
	}
}

func countFiles(node *treeNode) int {
	if len(node.children) == 0 {
		return 1
	}
	n := 0
	for _, child := range node.children {
		n += countFiles(child)
	}
	return n
}

// repoFiles returns tracked and untracked (non-ignored) file paths relative
// to dir, sorted alphabetically.
func repoFiles(dir string) ([]string, error) {
	tracked, err := gitCmd(dir, "ls-files")
	if err != nil {
		return nil, err
	}
	untracked, _ := gitCmd(dir, "ls-files", "--others", "--exclude-standard")

	seen := make(map[string]bool)
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(tracked), "\n") {
		if line != "" && !seen[line] {
			seen[line] = true
			files = append(files, line)
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(untracked), "\n") {
		if line != "" && !seen[line] {
			seen[line] = true
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files, nil
}
