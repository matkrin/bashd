package ast

import (
	"testing"
)

func TestDefNodes_GlobalVariables(t *testing.T) {
	input := `
a="test"
b=123
`
	fileAst, _ := ParseDocument(input, "")
	defNodes := fileAst.DefNodes()

	if len(defNodes) != 2 {
		t.Fatalf("expected 2 definition nodes, got %d", len(defNodes))
	}

	if defNodes[0].Name != "a" {
		t.Errorf("expected def node 'a', got '%s'", defNodes[0].Name)
	}
	if defNodes[0].IsScoped {
		t.Error("expected 'a' to not be scoped")
	}

	if defNodes[1].Name != "b" {
		t.Errorf("expected def node 'b', got '%s'", defNodes[1].Name)
	}
	if defNodes[1].IsScoped {
		t.Error("expected 'b' to not be scoped")
	}
}

func TestDefNodes_FunctionDeclaration(t *testing.T) {
	input := `
foo() {
  echo "bar"
}
`
	fileAst, _ := ParseDocument(input, "")
	defNodes := fileAst.DefNodes()

	if len(defNodes) != 1 {
		t.Fatalf("expected 1 definition node, got %d", len(defNodes))
	}

	if defNodes[0].Name != "foo" {
		t.Errorf("expected def node 'foo', got '%s'", defNodes[0].Name)
	}
	if defNodes[0].IsScoped {
		t.Error("expected 'foo' to not be scoped")
	}
}

func TestDefNodes_ScopedVariables(t *testing.T) {
	input := `
foo() {
  local a="test"
  declare b=123
  typeset c
}
`
	fileAst, _ := ParseDocument(input, "")
	defNodes := fileAst.DefNodes()

	// Includes the function declaration
	if len(defNodes) != 4 {
		t.Fatalf("expected 4 definition nodes, got %d", len(defNodes))
	}

	// The function
	if defNodes[0].Name != "foo" {
		t.Errorf("expected def node 'foo', got '%s'", defNodes[0].Name)
	}

	// The scoped variables
	for _, tt := range []struct {
		name string
	}{
		{"a"},
		{"b"},
		{"c"},
	} {
		found := false
		for _, defNode := range defNodes {
			if defNode.Name == tt.name {
				found = true
				if !defNode.IsScoped {
					t.Errorf("expected '%s' to be scoped", tt.name)
				}
			}
		}
		if !found {
			t.Errorf("expected to find definition for '%s'", tt.name)
		}
	}
}

func TestDefNodes_ReadStatement(t *testing.T) {
	input := `read a b`
	fileAst, _ := ParseDocument(input, "")
	defNodes := fileAst.DefNodes()

	if len(defNodes) != 2 {
		t.Fatalf("expected 2 definition nodes, got %d", len(defNodes))
	}

	if defNodes[0].Name != "a" {
		t.Errorf("expected def node 'a', got '%s'", defNodes[0].Name)
	}
	if defNodes[1].Name != "b" {
		t.Errorf("expected def node 'b', got '%s'", defNodes[1].Name)
	}
}

func TestDefNodes_ForLoop(t *testing.T) {
	input := `for i in 1 2 3; do echo $i; done`
	fileAst, _ := ParseDocument(input, "")
	defNodes := fileAst.DefNodes()

	if len(defNodes) != 1 {
		t.Fatalf("expected 1 definition node, got %d", len(defNodes))
	}

	if defNodes[0].Name != "i" {
		t.Errorf("expected def node 'i', got '%s'", defNodes[0].Name)
	}
}

func TestDefNodes_Complex(t *testing.T) {
	input := `
a="global"

foo() {
  local b="scoped"
  c="global_in_func"
  echo "$a $b $c"
}

read d

for e in 1 2 3; do
  echo $e
done
`
	fileAst, _ := ParseDocument(input, "")
	defNodes := fileAst.DefNodes()

	expected := map[string]bool{
		"a":    false,
		"foo":  false,
		"b":    true,
		"c":    false,
		"d":    false,
		"e":    false,
	}

	if len(defNodes) != len(expected) {
		t.Fatalf("expected %d definition nodes, got %d", len(expected), len(defNodes))
	}

	for _, defNode := range defNodes {
		isScoped, ok := expected[defNode.Name]
		if !ok {
			t.Errorf("unexpected definition node '%s'", defNode.Name)
			continue
		}
		if defNode.IsScoped != isScoped {
			t.Errorf("expected '%s' to have IsScoped=%t, got %t", defNode.Name, isScoped, defNode.IsScoped)
		}
		delete(expected, defNode.Name)
	}

	if len(expected) > 0 {
		for name := range expected {
			t.Errorf("missing expected definition for '%s'", name)
		}
	}
}