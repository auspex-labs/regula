// Copyright 2021 Fugue, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package loader

// StdIn is the path used for stdin.
const StdIn = "<stdin>"

// InputType is a flag that determines which types regula should look for.
type InputType int

const (
	// Auto means that regula will automatically try to determine which input types are
	// in the given paths.
	Auto InputType = iota
	// TfPlan means that regula will only look for Terraform plan JSON files in given
	// directories and it will assume that given files are Terraform plan JSON.
	TfPlan
	// Cfn means that regula will only look for CloudFormation template files in given
	// directories and it will assume that given files are CloudFormation JSON.
	Cfn
)

// InputTypeIDs maps the InputType enums to string values that can be specified in
// CLI options.
var InputTypeIDs = map[InputType][]string{
	Auto:   {"auto"},
	TfPlan: {"tf-plan"},
	Cfn:    {"cfn"},
}

// LoadedConfigurations is a container for IACConfigurations loaded by Regula.
type LoadedConfigurations interface {
	// AddConfiguration adds a configuration entry for the given path
	AddConfiguration(path string, config IACConfiguration)
	// Location resolves a file path and attribute path from the regula output to a
	// location within a file.
	Location(path string, attributePath []string) (*Location, error)
	// AlreadyLoaded indicates whether the given path has already been loaded as part
	// of another IACConfiguration.
	AlreadyLoaded(path string) bool
	// RegulaInput renders the RegulaInput from all of the contained configurations.
	RegulaInput() []RegulaInput
	// Count returns the number of loaded configurations.
	Count() int
}

// RegulaInput is a generic map that can be fed to OPA for regula.
type RegulaInput map[string]interface{}

// IACConfiguration is a loaded IaC Configuration.
type IACConfiguration interface {
	// RegulaInput returns a input for regula.
	RegulaInput() RegulaInput
	// LoadedFiles are all of the files contained within this configuration.
	LoadedFiles() []string
	// Location resolves an attribute path to to a file, line and column.
	Location(attributePath []string) (*Location, error)
}

// Location is a filepath, line and column.
type Location struct {
	Path string
	Line int
	Col  int
}

// DetectOptions are options passed to the configuration detectors.
type DetectOptions struct {
	IgnoreExt bool
}

// ConfigurationDetector implements the visitor part of the visitor pattern for the
// concrete InputPath implementations. A ConfigurationDetector implementation must
// contain functions to visit both directories and files. An empty implementation
// must return nil, nil to indicate that the InputPath has been ignored.
type ConfigurationDetector interface {
	DetectDirectory(i InputDirectory, opts DetectOptions) (IACConfiguration, error)
	DetectFile(i InputFile, opts DetectOptions) (IACConfiguration, error)
}

// InputPath is a generic interface to represent both directories and files that
// can serve as inputs for a ConfigurationDetector.
type InputPath interface {
	DetectType(d ConfigurationDetector, opts DetectOptions) (IACConfiguration, error)
	IsDir() bool
	Path() string
	Name() string
}

type InputDirectory interface {
	InputPath
	Walk(w func(i InputPath) error) error
	Children() []InputPath
}

type InputFile interface {
	InputPath
	Ext() string
	Contents() ([]byte, error)
}
