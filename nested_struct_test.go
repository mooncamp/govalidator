package govalidator

import (
	"context"
	"testing"
)

func TestMooncampNestedStruct(t *testing.T) {
	type Node struct {
		Name  string
		Child *Node `valid:"childless"`
	}

	input := Node{
		Name:  "root",
		Child: nil,
	}

	vd := New()
	vd.AddCustomTypeTagFn("childless", func(ctx context.Context, in interface{}, o interface{}) (bool, error) {
		node := in.(*Node)
		return node.Child == nil, nil
	})

	ok, err := vd.ValidateStruct(input)
	if err != nil {
		t.Fatalf("validate with nested struct: %v", err)
	}

	if true != ok {
		t.Errorf("expected %v, got %v", false, ok)
	}
}

// Testing mooncamp use-case
func TestMooncampNestedStructWithError(t *testing.T) {
	type Node struct {
		Name  string
		Child *Node `valid:"childless"`
	}

	input := Node{
		Name: "root",
		Child: &Node{
			Name: "child",
			Child: &Node{
				Name: "child second gen",
			},
		},
	}

	vd := New()
	vd.AddCustomTypeTagFn("childless", func(ctx context.Context, in interface{}, o interface{}) (bool, error) {
		node := in.(*Node)
		return node.Child == nil, nil
	})

	ok, err := vd.ValidateStruct(input)
	if err == nil {
		t.Fatal("didn't validate nested struct correctly")
	}

	if false != ok {
		t.Errorf("expected %v, got %v", false, ok)
	}
}
