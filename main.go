package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type lineNumber = uint32

type header struct {
	startLine lineNumber
	endLine lineNumber
	size uint32
}

type contextTree struct {
	lines []string
	lineCount uint32

	nodes [][]*sitter.Node
	headers []header

	scopes []map[lineNumber]struct{}
}

func (t *contextTree) walkTree(node *sitter.Node) error {
	startLine := node.StartPoint().Row
	endLine := node.EndPoint().Row

	t.nodes[startLine] = append(t.nodes[startLine], node)

	nodeSize := endLine - startLine 
	if nodeSize > 0 {
		if t.headers[startLine].size == 0 || nodeSize < t.headers[startLine].size{
			t.headers[startLine] = header{
				startLine: startLine,
				endLine: endLine,
				size: nodeSize,
			}
		}
	}

	for i := startLine; i <= endLine; i++ {
		if t.scopes[i] == nil {
			t.scopes[i] = make(map[lineNumber]struct{})
		}

		t.scopes[i][startLine] = struct{}{}
	}

	for i := range node.ChildCount() {
		t.walkTree(node.Child(int(i)))
	}

	return nil
}

func main() {
	data, err := os.ReadFile("./test/test.go")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(data), "\n")

	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, data)
	if err != nil {
		panic(err)
	}

	root := tree.RootNode()

	treeContext := contextTree{
		lines: lines,
		lineCount: uint32(len(lines)),
		nodes: make([][]*sitter.Node, root.EndPoint().Row+1),
		headers: make([]header, root.EndPoint().Row+1),
		scopes: make([]map[lineNumber]struct{}, root.EndPoint().Row+1),
	}

	treeContext.walkTree(root)
	fmt.Printf("treeContext: %v\n", treeContext)
}
