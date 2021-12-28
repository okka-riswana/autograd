package grader

import (
	"github.com/sirupsen/logrus"

	"github.com/fahmifan/autograd/container"
	"github.com/fahmifan/autograd/model"
)

// Type types of graders
type Type string

// grader types
const (
	TypeCPP              = Type("cpp")
	TypeCPPContainerized = Type("cpp-containerized")
)

// New new grader factory
func New(t Type) model.GraderEngine {
	switch t {
	case TypeCPP:
		return &CPPGrader{}
	case TypeCPPContainerized:
		c, err := container.NewWithBuiltIn()
		if err != nil {
			logrus.Fatal(err)
		}
		return &CPPContainerizedGrader{Container: c}
	default:
		return nil
	}
}
