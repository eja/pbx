// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func MarkdownToText(md []byte) (string, error) {
	doc := goldmark.DefaultParser().Parse(text.NewReader(md))
	buf := &bytes.Buffer{}

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch v := n.(type) {
			case *ast.Text:
				buf.Write(v.Text(md))
			case *ast.String:
				buf.Write(v.Value)
			case *ast.CodeBlock, *ast.FencedCodeBlock:
				lines := n.Lines()
				for i := 0; i < lines.Len(); i++ {
					seg := lines.At(i)
					buf.Write(seg.Value(md))
				}
			}
		} else {
			if n.Type() == ast.TypeBlock {
				buf.WriteByte('\n')
			}
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

func FileCopy(sourcePath, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func FileReadJson(filename string) (map[string]any, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var data map[string]any

	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return data, nil
}
