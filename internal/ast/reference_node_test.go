package ast

import (
	"slices"
	"testing"
)

func Test_RefNodes(t *testing.T) {
	input := `
a="global"

foo() {
  local b="scoped"
  c="global_in_func"
  declare d=123
  typeset e
  echo "$a $b $c"
}

read f

for g in 1 2 3; do
  echo $e
done
`
	fileAst, _ := ParseDocument(input, "", false)
	refNodes := fileAst.RefNodes(true)
	for _, r := range refNodes {
		t.Log(r.Name)
	}

	expected := []struct {
		name      string
		startLine uint
	}{
		{"a", 2},
		{"foo", 4},
		{"b", 5},
		{"c", 6},
		{"d", 7},
		{"e", 8},
		{"echo", 9},
		{"a", 9},
		{"b", 9},
		{"c", 9},
		{"f", 12},
		{"g", 14},
		{"echo", 15},
		{"e", 15},
	}

	if len(refNodes) != len(expected) {
		t.Fatalf("expected %d definition nodes, got %d", len(expected), len(refNodes))
	}

	for i := range refNodes {
		if refNodes[i].Name != expected[i].name {
			t.Errorf("expected '%s', got '%s'", expected[i].name, refNodes[i].Name)
		}
		if refNodes[i].StartLine != expected[i].startLine {
			t.Errorf("expected '%d', got '%d'", expected[i].startLine, refNodes[i].StartLine)
		}
	}
}

func Test_FindRefInFile(t *testing.T) {
	input := `
a="global"
b="global2"

foo() {
  local b="scoped"
  c="global_in_func"
  echo "$a $b $c"
  local a="shadows"
  echo "$a"
}

echo "$a $b $c"
`
	fileAst, _ := ParseDocument(input, "", false)

	tests := []struct {
		cursor     Cursor
		name       string
		startLines []uint
	}{
		{NewCursor(1, 1), "a", []uint{2, 8, 13}},
		{NewCursor(2, 12), "b", []uint{3, 13}},
		{NewCursor(5, 9), "b", []uint{6, 8}},
		{NewCursor(6, 3), "c", []uint{7, 8, 13}},
		{NewCursor(9, 10), "a", []uint{9, 10}},
	}

	for _, e := range tests {
		refNodes := fileAst.FindRefsInFile(e.cursor, true)
		for _, refNode := range refNodes {
			t.Logf("%s - %d", refNode.Name, refNode.StartLine)
			if refNode.Name != e.name {
				t.Errorf("expected '%s', got '%s'", e.name, refNode.Name)
			}
			if !slices.Contains(e.startLines, refNode.StartLine) {
				t.Errorf("expected '%v', got '%d'", e.startLines, refNode.StartLine)
			}
		}
	}
}
