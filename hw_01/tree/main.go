package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type dirNode struct {
	File     os.DirEntry
	Children []dirNode
}

func (d dirNode) String() string {
	fileInfo, err := d.File.Info()
	if err != nil {
		return ""
	}
	if fileInfo.IsDir() {
		return d.File.Name()
	}

	if fileInfo.Size() != 0 {
		return fmt.Sprintf("%s (%db)", d.File.Name(), fileInfo.Size())
	}

	return fmt.Sprintf("%s (empty)", d.File.Name())
}

func getDirNodes(rootPath string, printFiles bool) ([]dirNode, error) {
	files, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	var dirNodes []dirNode

	for _, file := range files {
		if !file.IsDir() && !printFiles {
			continue
		}

		node := dirNode{file, nil}

		if file.IsDir() {
			children, err := getDirNodes(filepath.Join(rootPath, file.Name()), printFiles)
			if err != nil {
				return nil, err
			}

			node.Children = children
		}

		dirNodes = append(dirNodes, node)
	}

	return dirNodes, nil
}

func printNodes(out io.Writer, dirNodes []dirNode, parentPrefix string) {
	var (
		currentPrefix = "├───"
		childPrefix   = "│\t"
	)

	for i, node := range dirNodes {
		if i == len(dirNodes)-1 {
			currentPrefix = "└───"
			childPrefix = "\t"
		}

		formattedName := parentPrefix + currentPrefix + node.String() + "\n"

		_, err := out.Write([]byte(formattedName))
		if err != nil {
			return
		}

		if node.File.IsDir() {
			printNodes(out, node.Children, parentPrefix+childPrefix)
		}
	}
}

func dirTree(out io.Writer, rootPath string, printFiles bool) error {
	nodes, err := getDirNodes(rootPath, printFiles)
	if err != nil {
		return err
	}

	printNodes(out, nodes, "")

	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
