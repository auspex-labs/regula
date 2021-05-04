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
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/configs"
	"github.com/zclconf/go-cty/cty"

	"tf_resource_schemas"
)

type TfDirDetector struct{}

func (t *TfDirDetector) DetectFile(InputFile, DetectOptions) (IACConfiguration, error) {
	return nil, nil
}

func (t *TfDirDetector) DetectDirectory(i InputDirectory, opts DetectOptions) (IACConfiguration, error) {
	// First check that a `.tf` file exists in the directory.
	if matches, err := filepath.Glob(i.Path() + "/*"); err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("Directory does not contain a .tf file: %v", i.Path())
	}

	configuration := new(HclConfiguration)
	parser := configs.NewParser(nil)
	var diags hcl.Diagnostics
	configuration.module, diags = parser.LoadConfigDir(i.Path())
	if diags.HasErrors() {
		return nil, fmt.Errorf(diags.Error())
	}

	configuration.schemas = tf_resource_schemas.LoadResourceSchemas()
	return configuration, nil
}

type HclConfiguration struct {
	module  *configs.Module
	schemas tf_resource_schemas.ResourceSchemas
}

func (c *HclConfiguration) LoadedFiles() []string {
	// TODO
	return []string{}
}

func (c *HclConfiguration) Location(attributePath []string) (*Location, error) {
	return nil, nil
}

func (c *HclConfiguration) RegulaInput() RegulaInput {
	return c.RenderResourceView()
}

func (c *HclConfiguration) RenderResourceView() map[string]interface{} {
	resourceView := make(map[string]interface{})

	resources := make(map[string]interface{})
	resourceView["resources"] = resources

	for resourceId, resource := range c.module.ManagedResources {
		resources[resourceId] = c.RenderResource(resourceId, resource)
	}
	for resourceId, resource := range c.module.DataResources {
		resources[resourceId] = c.RenderResource(resourceId, resource)
	}

	return resourceView
}

func (c *HclConfiguration) RenderResource(
	resourceId string, resource *configs.Resource,
) interface{} {
	schema := c.schemas[resource.Type]
	properties := make(map[string]interface{})
	properties["_type"] = resource.Type
	properties["id"] = resourceId

	body, ok := resource.Config.(*hclsyntax.Body)
	if !ok {
		return properties
	}

	bodyProperties := c.RenderBody(body, schema)
	for k, v := range bodyProperties {
		properties[k] = v
	}

	return properties
}

func (c *HclConfiguration) RenderBody(
	body *hclsyntax.Body, schema *tf_resource_schemas.Schema,
) map[string]interface{} {
	properties := make(map[string]interface{})

	for _, attribute := range body.Attributes {
		properties[attribute.Name] = c.RenderAttribute(
			attribute,
			tf_resource_schemas.GetAttribute(schema, attribute.Name),
		)
	}

	// TODO: Add defaults NOT in body.
	// TODO: Double check defaults
	// TODO: Turn undefined lists into [], other attributes into null?
	tf_resource_schemas.SetDefaultAttributes(schema, properties)

	renderedBlocks := make(map[string][]interface{})
	for _, block := range body.Blocks {
		if _, ok := properties[block.Type]; !ok {
			renderedBlocks[block.Type] = make([]interface{}, 0)
		}
		entry := c.RenderBlock(
			block,
			tf_resource_schemas.GetAttribute(schema, block.Type),
		)
		renderedBlocks[block.Type] = append(renderedBlocks[block.Type], entry)
	}
	for key, renderedBlock := range renderedBlocks {
		properties[key] = renderedBlock
	}

	return properties
}

func (c *HclConfiguration) RenderAttribute(
	attribute *hclsyntax.Attribute, schema *tf_resource_schemas.Schema,
) interface{} {
	if attribute.Expr == nil {
		return nil
	}
	return c.RenderExpr(attribute.Expr, schema)
}

func (c *HclConfiguration) RenderBlock(
	block *hclsyntax.Block, schema *tf_resource_schemas.Schema,
) interface{} {
	if block.Body == nil {
		return nil
	}
	return c.RenderBody(block.Body, schema)
}

func (c *HclConfiguration) IsResource(id string) bool {
	if _, ok := c.module.ManagedResources[id]; ok {
		return true
	}
	if _, ok := c.module.DataResources[id]; ok {
		return true
	}
	return false
}

func (c *HclConfiguration) ResolveResourceReference(traversal hcl.Traversal) *string {
	parts := c.RenderTraversal(traversal)
	if len(parts) < 1 {
		return nil
	}
	idx := 2
	if parts[0] == "data" {
		idx = 3
	}

	resourceId := strings.Join(parts[:idx], ".")
	if c.IsResource(resourceId) {
		return &resourceId
	}
	return nil
}

// This returns a string or array of references.
// TODO: limit this to resource references?
func (c *HclConfiguration) ExpressionReferences(expr hclsyntax.Expression) interface{} {
	references := make([]string, 0)
	for _, traversal := range expr.Variables() {
		resolved := c.ResolveResourceReference(traversal)
		if resolved != nil {
			references = append(references, *resolved)
		}
	}
	if len(references) == 0 {
		return nil
	} else if len(references) == 1 {
		return references[0]
	} else {
		return references
	}
}

func (c *HclConfiguration) RenderExpr(
	expr hclsyntax.Expression, schema *tf_resource_schemas.Schema,
) interface{} {
	switch e := expr.(type) {
	case *hclsyntax.TemplateWrapExpr:
		return c.RenderExpr(e.Wrapped, schema)
	case *hclsyntax.ScopeTraversalExpr:
		ref := c.ResolveResourceReference(e.Traversal)
		if ref != nil {
			return ref
		} else {
			// Is this useful?  This should just map to variables?
			return strings.Join(c.RenderTraversal(e.Traversal), ".")
		}
	case *hclsyntax.TemplateExpr:
		// This is commonly used to refer to resources, so we pick out the
		// references.
		refs := c.ExpressionReferences(e)
		if refs != nil {
			return refs
		}

		str := ""
		for _, part := range e.Parts {
			val := c.RenderExpr(part, schema)
			if s, ok := val.(string); ok {
				str += s
			}
		}
		return str
	case *hclsyntax.LiteralValueExpr:
		return c.RenderValue(e.Val, schema)
	case *hclsyntax.TupleConsExpr:
		arr := make([]interface{}, 0)
		elemSchema := tf_resource_schemas.GetElem(schema)
		for _, elem := range e.Exprs {
			arr = append(arr, c.RenderExpr(elem, elemSchema))
		}
		return arr
	}

	fmt.Printf("warning: unhandled expression type %s\n", reflect.TypeOf(expr).String())

	// Fall back to normal eval.
	ctx := hcl.EvalContext{}
	val, _ := expr.Value(&ctx)
	return c.RenderValue(val, schema)
}

func (c *HclConfiguration) RenderTraversal(traversal hcl.Traversal) []string {
	parts := make([]string, 0)

	for _, traverser := range traversal {
		switch t := traverser.(type) {
		case hcl.TraverseRoot:
			parts = append(parts, t.Name)
		case hcl.TraverseAttr:
			parts = append(parts, t.Name)
		case hcl.TraverseIndex:
			// TODO
			_ = t.Key
		}
	}

	return parts
}

func (c *HclConfiguration) RenderValue(
	val cty.Value, schema *tf_resource_schemas.Schema,
) interface{} {
	if val.Type() == cty.Bool {
		return val.True()
	} else if val.Type() == cty.Number {
		return val.AsBigFloat()
	} else if val.Type() == cty.String {
		return val.AsString()
	} else if val.Type().IsTupleType() || val.Type().IsSetType() || val.Type().IsListType() {
		childSchema := tf_resource_schemas.GetElem(schema)
		array := make([]interface{}, 0)
		for _, elem := range val.AsValueSlice() {
			array = append(array, c.RenderValue(elem, childSchema))
		}
		return array
	} else if val.Type().IsMapType() || val.Type().IsObjectType() {
		object := make(map[string]interface{}, 0)
		for key, attr := range val.AsValueMap() {
			attrSchema := tf_resource_schemas.GetAttribute(schema, key)
			object[key] = c.RenderValue(attr, attrSchema)
		}
		return object
	}

	fmt.Printf("Unknown type: %v\n", val.Type().GoString())
	return nil
}
