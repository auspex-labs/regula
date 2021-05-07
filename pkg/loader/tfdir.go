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
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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
	configuration.path = i.Path()

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
	path    string
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
	return RegulaInput{
		"filepath": c.path,
		"content":  c.RenderResourceView(),
	}
}

func (c *HclConfiguration) RenderResourceView() map[string]interface{} {
	resourceView := make(map[string]interface{})
	resourceView["hcl_resource_view_version"] = "0.0.1"

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
	properties["_provider"] = resource.Provider.ForDisplay()
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
		if _, ok := renderedBlocks[block.Type]; !ok {
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

func (c *HclConfiguration) GetResource(id string) (*configs.Resource, bool) {
	if r, ok := c.module.ManagedResources[id]; ok {
		return r, true
	}
	if r, ok := c.module.DataResources[id]; ok {
		return r, true
	}
	return nil, false
}

func (c *HclConfiguration) ResolveResourceReference(traversal hcl.Traversal) interface{} {
	parts := c.RenderTraversal(traversal)
	if len(parts) < 1 {
		return nil
	}
	idx := 2
	if parts[0] == "data" {
		idx = 3
	}

	resourceId := strings.Join(parts[:idx], ".")
	if resource, ok := c.GetResource(resourceId); ok {
		resourceNode := TfNode{Object: resource.Config, Range: resource.DeclRange}
		if node, err := resourceNode.GetDescendant(parts[idx:]); err == nil {
			// TODO: non-attribute cases.
			if node.Attribute != nil {
				expr := node.Attribute.Expr
				if e, ok := expr.(hclsyntax.Expression); ok {
					return c.RenderExpr(e, nil)
				}
			}
		}

		return resourceId
	}
	return nil
}

// This returns a string or array of references.
// TODO: limit this to resource references?
func (c *HclConfiguration) ExpressionReferences(expr hclsyntax.Expression) interface{} {
	references := make([]interface{}, 0)
	for _, traversal := range expr.Variables() {
		resolved := c.ResolveResourceReference(traversal)
		if resolved != nil {
			references = append(references, resolved)
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

// Auxiliary function to determine if the expression should be ignored from
// sets, lists, etc.
func voidExpression(expr hclsyntax.Expression) bool {
	switch e := expr.(type) {
	case *hclsyntax.TemplateExpr:
		return len(e.Parts) == 0
	}
	return false
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
		if len(e.Parts) == 1 {
			return c.RenderExpr(e.Parts[0], schema)
		}

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
			if !voidExpression(elem) {
				arr = append(arr, c.RenderExpr(elem, elemSchema))
			}
		}
		return arr
	case *hclsyntax.ObjectConsExpr:
		object := make(map[string]interface{})
		for _, item := range e.Items {
			key := c.RenderExpr(item.KeyExpr, nil)   // Or pass string schema?
			val := c.RenderExpr(item.ValueExpr, nil) // Or get elem schema?
			if str, ok := key.(string); ok {
				object[str] = val
			} else {
				fmt.Fprintf(os.Stderr, "warning: non-string key: %s\n", reflect.TypeOf(key).String())
			}
		}
		return object
	case *hclsyntax.ObjectConsKeyExpr:
		// Keywords are interpreted as keys.
		if key := hcl.ExprAsKeyword(e); key != "" {
			return key
		} else {
			return c.RenderExpr(e.Wrapped, schema)
		}
	}

	fmt.Fprintf(os.Stderr, "warning: unhandled expression type %s\n", reflect.TypeOf(expr).String())

	// Fall back to normal eval.
	return c.EvaluateExpr(expr, schema)
}

func (c *HclConfiguration) EvaluateExpr(
	expr hcl.Expression, schema *tf_resource_schemas.Schema,
) interface{} {
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
		b := val.AsBigFloat()
		if b.IsInt() {
			i, _ := b.Int64()
			return i
		} else {
			f, _ := b.Float64()
			return f
		}
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

	fmt.Fprintf(os.Stderr, "Unknown type: %v\n", val.Type().GoString())
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// utilities for traversing to a path in a HCL tree somewhat generically

// A `TfNode` represents a syntax tree in the HCL config.
type TfNode struct {
	// Exactly one of the next three fields will be set.
	Object    hcl.Body
	Array     hcl.Blocks
	Attribute *hcl.Attribute

	// This will always be set.
	Range hcl.Range
}

func (node *TfNode) GetChild(key string) (*TfNode, error) {
	child := TfNode{}

	if node.Object != nil {
		bodyContent, _, diags := node.Object.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name:     key,
					Required: false,
				},
			},
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type: key,
				},
			},
		})
		if diags.HasErrors() {
			return nil, fmt.Errorf(diags.Error())
		}

		blocks := bodyContent.Blocks.OfType(key)
		if len(blocks) > 0 {
			child.Array = blocks
			child.Range = blocks[0].DefRange
		}

		if attribute, ok := bodyContent.Attributes[key]; ok {
			child.Attribute = attribute
			child.Range = attribute.Range
		}
	} else if node.Array != nil {
		index, err := strconv.Atoi(key)
		if err != nil {
			return nil, err
		} else {
			if index < 0 || index >= len(node.Array) {
				return nil, fmt.Errorf("TfNode.Get: out of bounds: %d", index)
			}

			child.Object = node.Array[index].Body
			child.Range = node.Array[index].DefRange
		}
	}

	return &child, nil
}

func (node *TfNode) GetDescendant(path []string) (*TfNode, error) {
	if len(path) == 0 {
		return node, nil
	}

	child, err := node.GetChild(path[0])
	if err != nil {
		return nil, err
	}

	return child.GetDescendant(path[1:])
}

func (node *TfNode) Location() string {
	return fmt.Sprintf(
		"%s:%d:%d",
		node.Range.Filename,
		node.Range.Start.Line,
		node.Range.Start.Column,
	)
}
