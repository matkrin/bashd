package ast

import (
	"testing"
)

func Test_DefNodes(t *testing.T) {
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
	fileAst, _ := ParseDocument(input, "")
	defNodes := fileAst.DefNodes()

	expected := []struct {
		name      string
		isScoped  bool
		startLine uint
	}{
		{"a", false, 2},
		{"foo", false, 4},
		{"b", true, 5},
		{"c", false, 6},
		{"d", true, 7},
		{"e", true, 8},
		{"f", false, 12},
		{"g", false, 14},
	}

	if len(defNodes) != len(expected) {
		t.Fatalf("expected %d definition nodes, got %d", len(expected), len(defNodes))
	}

	for i := range defNodes {
		if defNodes[i].Name != expected[i].name {
			t.Errorf("expected '%s', got '%s'", expected[i].name, defNodes[i].Name)
		}
		if defNodes[i].IsScoped != expected[i].isScoped {
			t.Errorf("expected '%t', got '%t'", expected[i].isScoped, defNodes[i].IsScoped)
		}
		if defNodes[i].StartLine != expected[i].startLine {
			t.Errorf("expected '%d', got '%d'", expected[i].startLine, defNodes[i].StartLine)
		}
	}
}

func Test_FindDefInFile(t *testing.T) {
	input := `
a="global"
b="global2"

foo() {
  local b="scoped"
  c="global_in_func"
  echo "$a $b $c"
}

echo "$b"
`
	fileAst, _ := ParseDocument(input, "")

	expected := []struct {
		cursor    Cursor
		name      string
		startLine uint
	}{
		{NewCursor(7, 9), "a", 2},
		{NewCursor(7, 12), "b", 6},
		{NewCursor(7, 14), "c", 7},
		{NewCursor(10, 7), "b", 3},
	}

	for _, e := range expected {
		defNode := fileAst.FindDefInFile(e.cursor)
		if defNode.Name != e.name {
			t.Errorf("expected '%s', got '%s'", e.name, defNode.Name)
		}
		if defNode.StartLine != e.startLine {
			t.Errorf("expected '%d', got '%d'", e.startLine, defNode.StartLine)
		}
	}
}
