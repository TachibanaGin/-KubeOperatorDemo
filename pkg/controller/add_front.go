package controller

import (
	"src/op-demo-front/pkg/controller/front"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, front.Add)
}
