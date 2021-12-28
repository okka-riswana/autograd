package grader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/fahmifan/autograd/container"
	"github.com/fahmifan/autograd/model"
)

// make sure implement interface
var _ model.GraderEngine = (*CPPContainerizedGrader)(nil)

// CPPContainerizedGrader implements model.Grader
type CPPContainerizedGrader struct {
	Container container.Container

	_ struct{}
}

// Grade ..
func (c *CPPContainerizedGrader) Grade(arg *model.GradingArg) (*model.GradingResult, error) {
	binPath, err := c.Compile(arg.SourceCodePath)
	if err != nil {
		logrus.WithField("source", arg.SourceCodePath).Error(err)
		return nil, err
	}

	defer func() {
		if err := c.Remove(binPath); err != nil {
			logrus.WithField("path", binPath).Error(err)
		}
	}()

	if len(arg.Inputs) != len(arg.Expecteds) {
		return nil, fmt.Errorf("expecteds & inputs not match in length")
	}

	result := &model.GradingResult{}
	for i := range arg.Inputs {
		input := arg.Inputs[i]
		expected := arg.Expecteds[i]

		out, err := c.Run(binPath, input)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		result.Outputs = append(result.Outputs, out)
		result.Corrects = append(result.Corrects, out == expected)
	}

	return result, nil
}

// Compile compile programs
func (c *CPPContainerizedGrader) Compile(inputPath string) (outPath string, err error) {
	contInPath, err := c.copyToContainer(inputPath)
	if err != nil {
		return "", err
	}

	baseExecFn := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	outPath = path.Join(c.Container.HostSideSharedDir(), baseExecFn)
	contOutPath := path.Join(c.Container.ContainerSideSharedDir(), baseExecFn)

	const cxx = "c++"
	args := []string{
		"-Wall",
		"-Wpedantic",
		"-std=c++17",
		"-o",
		contOutPath,
		contInPath,
	}

	cmd := c.Container.Command(cxx, args...)
	bt, err := cmd.CombinedOutput()
	if err != nil {
		logrus.WithFields(
			logrus.Fields{
				"output":  string(bt),
				"args":    args,
				"outPath": outPath,
			},
		).Error(err)
	}

	return
}

// Run the binary with input as arguments and return the output
func (c *CPPContainerizedGrader) Run(source, input string) (out string, err error) {
	inputs := strings.Split(input, " ")
	input = strings.Join(inputs, "\n")

	baseExecFn := strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	contExecPath := path.Join(c.Container.ContainerSideSharedDir(), baseExecFn)

	cmd := c.Container.Command(contExecPath, inputs...)

	var buffOut bytes.Buffer
	var buffErr bytes.Buffer

	cmd.Stdin = bytes.NewBuffer([]byte(input))
	cmd.Stdout = &buffOut
	cmd.Stderr = &buffErr

	err = cmd.Run()
	if err != nil {
		return
	}

	out = strings.TrimSpace(buffOut.String())
	return
}

// Remove ..
func (c *CPPContainerizedGrader) Remove(source string) error {
	source = filepath.Join(c.Container.HostSideSharedDir(), path.Base(source))
	err := os.Remove(source)
	if err != nil {
		logrus.Error(err)
	}

	return err
}

func (c *CPPContainerizedGrader) copyToContainer(filename string) (string, error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	dst := filepath.Join(c.Container.HostSideSharedDir(), path.Base(filename))

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return "", err
	}

	return filepath.Join(c.Container.ContainerSideSharedDir(), path.Base(filename)), nil
}
