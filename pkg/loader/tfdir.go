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

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs"
)

type TfDirDetector struct{}

func (t *TfDirDetector) DetectFile(InputFile, DetectOptions) (IACConfiguration, error) {
	return nil, nil
}

type HclModule struct {
	module *configs.Module
	// schemas resource_schemas.ResourceSchemas
}

func (t *TfDirDetector) DetectDirectory(i InputDirectory, opts DetectOptions) (IACConfiguration, error) {
	// First check that a `.tf` file exists in the directory.
	if matches, err := filepath.Glob(i.Path() + "/*"); err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("Directory does not contain a .tf file: %v", i.Path())
	}

	module := new(HclModule)
	parser := configs.NewParser(nil)
	var diags hcl.Diagnostics
	module.module, diags = parser.LoadConfigDir(i.Path())
	if diags.HasErrors() {
		return nil, fmt.Errorf(diags.Error())
	}

	// module.schemas = resource_schemas.LoadResourceSchemas()
	// return module, nil

	fmt.Println("Oh what to do")
	return nil, nil
}
