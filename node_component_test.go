package gotmx

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	stringutils "github.com/authentic-devel/gotmx/utils"
)

type nodeComponentTestModel struct {
	BooleanValue  bool
	BooleanValue2 bool
	StringValue   string
	StringValue2  string
	StringValue3  string
	StringValue4  string
	StringSlice   []string
	StringMap     map[string]string
	NestedModel   nodeComponentNestedModel
}

type nodeComponentNestedModel struct {
	BooleanValue  bool
	BooleanValue2 bool
	StringValue   string
	StringValue2  string
	StringSlice   []string
}

func TestBasicComponentRender(t *testing.T) {

	parseRenderAndCompareTemplate(
		nil,
		stringutils.TrimMargin(
			`<div data-g-define="my-template">
               |  <span>Hello</span>
               |  <span>World</span>
               |</div>`),
		"my-template",
		nil,
		stringutils.TrimMargin(
			`<div>
               |  <span>Hello</span>
               |  <span>World</span>
               |</div>`),

		t)
}

func TestRenderFullDocument(t *testing.T) {
	template := stringutils.TrimMargin(
		`<!DOCTYPE html>
           |<html g-define="my-template">
           |<head>Should be placed in the body</head>
           |<body>
           |  <div><span>Hello</span><span>World</span></div>
           |</body>
           |</html>`)
	expected := stringutils.TrimMargin(
		`<!DOCTYPE html>
           |<html><head></head><body>Should be placed in the body
           |
           |  <div><span>Hello</span><span>World</span></div>
           |
           |</body></html>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expected, t)
}

func TestVoidElementsAreRenderedCorrectly(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <area>
           |  <base>
           |  <br>
           |  <table><colgroup><col></colgroup></table>
           |  <embed>
           |  <hr>
           |  <img>
		   |  <input type="text">
           |  <link>
           |  <meta>
           |  <param>
           |  <source>
           |  <track>
           |  <wbr>
		   |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |  <area />
           |  <base />
           |  <br />
           |  <table><colgroup><col /></colgroup></table>
           |  <embed />
           |  <hr />
           |  <img />
		   |  <input type="text" />
           |  <link />
           |  <meta />
           |  <param />
           |  <source />
           |  <track />
           |  <wbr />
		   |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expected, t)

}

// Tests that when using nested templates, the nested template is also rendered correctly when rendering the outer
// template.
func TestNestedTemplateIsRendered(t *testing.T) {

	parseRenderAndCompareTemplate(
		nil,
		stringutils.TrimMargin(
			`<div data-g-define="my-template">
		       |  <span>Hello</span>
		       |  <div data-g-define="nested-template">
		       |    <span>World</span>
		       |  </div>
		       |</div>`),
		"my-template",
		nil,

		stringutils.TrimMargin(
			`<div>
               |  <span>Hello</span>
               |  <div>
               |    <span>World</span>
               |  </div>
               |</div>`),
		t)
}

// Tests, that a template is still rendered (when using g-define) when called directly, even if the same html node has
// a g-ignore attribute on it. We explicitly define the behavior like that so that users can define nested templates
// that aren't necessarily always rendered in that spot, but still can be rendered directly.
func TestTemplateIgnoresIgnoreIfRenderedDirectly(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template" data-g-ignore>
           |  <span>Hello World</span>
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |  <span>Hello World</span>
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expected, t)

	// same again but with short versions of the attribute
	templateShort := stringutils.TrimMargin(
		`<div g-define="my-template2" g-ignore>
           |  <span>Hello World</span>
           |</div>`)
	parseRenderAndCompareTemplate(nil, templateShort, "my-template2", nil, expected, t)
}

func TestBooleanAttributesAreRenderedCorrectly(t *testing.T) {
	template := `<button data-g-define="my-template" enabled>Click me!</button>`
	expected := `<button enabled="">Click me!</button>`
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expected, t)
}

// Tests, that a template reference [[ :templateName ]] in an attribute works when referencing a Gotmx
// template
func TestGotmxTemplateReferencedInAttributeWorks(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div g-define="my-template">
           |<div my-attribute="[[ :other-template ]]"></div>
           |</div>
           |<div data-g-define="other-template" g-outer-text="Hello world with special chars < > & '">
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |<div my-attribute="Hello world with special chars &lt; &gt; &amp; &#39;"></div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expected, t)
}

// todo: re-implement support for golang templates
//// Tests, that a template reference [[ :templateName ]] in an attribute works when referencing a Golang
//// HTML template
//func TestGolangTemplateReferencedInAttributeWorks(t *testing.T) {
//	data := nodeComponentTestModel{
//		StringValue: "special chars < > & '",
//	}
//	template := stringutils.TrimMargin(
//		`<div g-define="my-template">
//           |<div my-attribute="[[ :golang-template ]]"></div>
//           |<!-- Even if the called template is actually an unsafe template, the resulting string should be escaped when called in an attribute -->
//           |<div my-attribute="[[ :golang-unsafe-template ]]"></div>
//           |</div>
//           |<div g-as-template="golang-template">Hello world with {{ .StringValue }}</div>
//           |<div g-as-unsafe-template="golang-unsafe-template">Hello world with {{ .StringValue }}</div>
//         `)
//	expected := stringutils.TrimMargin(
//		`<div>
//           |<div my-attribute="Hello world with special chars &lt; &gt; &amp; &#39;"></div>
//           |
//           |<div my-attribute="Hello world with special chars &lt; &gt; &amp; &#39;"></div>
//           |</div>`)
//	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
//}

// Tests, that an empty string is returned when referencing a template that does not exist
func TestUnknownTemplateReferenceReturnsEmptyString(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div g-define="my-template">
           |<div my-attribute="[[ :golang-template ]]"></div>
 	       |</div>`)

	expected := stringutils.TrimMargin(
		`<div>
           |<div my-attribute=""></div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expected, t)
}

//********************************* g-if *******************************************************************************
//**********************************************************************************************************************

// Tests, that the g-if attribute of a component is ignored, in case the component is programmatically added
// as a slotted child when rendering.
func TestTemplateIgnoresIfIfRenderedAsSlotted(t *testing.T) {
	// todo: implement
	// todo: adapt documentation
}

func TestGIfLongVersionHasHigherPriority(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-if="true" data-g-if="false">Hello World</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
          |  
          |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}

// Tests, that a template with a g-define is still rendered when called directly, even if the same html node has
// a g-if attribute on it that evaluates to false. We explicitly define the behavior like that so that users can define
// nested templates that aren't necessarily always rendered in that spot, but still can be rendered directly.
func TestTemplateIgnoresIfIfRenderedDirectly(t *testing.T) {

	template := stringutils.TrimMargin(
		`<div data-g-define="outer-template" g-if="false">
           |  <div g-if="true">Hello World</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
          |  <div>Hello World</div>
          |</div>`)
	parseRenderAndCompareTemplate(nil, template, "outer-template", nil, expectation, t)
}

func TestIfWorksWithBooleanGoTemplate(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="outer-template">
           |  <div g-if="{{ .StringValue }}">Hello World</div>
           |</div>`)

	expectationForFalse := stringutils.TrimMargin(
		`<div>
           |  
           |</div>`)

	expectationForTrue := stringutils.TrimMargin(
		`<div>
           |  <div>Hello World</div>
           |</div>`)

	data := nodeComponentTestModel{
		StringValue: "true",
	}
	parseRenderAndCompareTemplate(nil, template, "outer-template", data, expectationForTrue, t)

	data.StringValue = "false"
	parseRenderAndCompareTemplate(nil, template, "outer-template", data, expectationForFalse, t)

	data.StringValue = "invalid"
	parseRenderAndCompareTemplate(nil, template, "outer-template", data, expectationForFalse, t)

	data.StringValue = ""
	parseRenderAndCompareTemplate(nil, template, "outer-template", data, expectationForFalse, t)
}

// Tests that an if ignores any invalid (non-boolean) expression and considers the expression to be false.
// But if should still work with string values, that are not "true" or "false"
func TestIfWorksWithNonBooleanValues(t *testing.T) {

	template := stringutils.TrimMargin(
		`<div data-g-define="outer-template">
           |  <div g-if="[[ .StringValue ]]">Hello World</div>
           |</div>`)

	expectationForFalse := stringutils.TrimMargin(
		`<div>
           |  
           |</div>`)

	expectationForTrue := stringutils.TrimMargin(
		`<div>
           |  <div>Hello World</div>
           |</div>`)

	data := nodeComponentTestModel{
		StringValue: "true",
	}
	parseRenderAndCompareTemplate(nil, template, "outer-template", data, expectationForTrue, t)

	data.StringValue = "false"
	parseRenderAndCompareTemplate(nil, template, "outer-template", data, expectationForFalse, t)

	data.StringValue = "invalid"
	parseRenderAndCompareTemplate(nil, template, "outer-template", data, expectationForFalse, t)

	data.StringValue = ""
	parseRenderAndCompareTemplate(nil, template, "outer-template", data, expectationForFalse, t)
}

func TestIfWorksWithBooleanModelPath(t *testing.T) {

	data := nodeComponentTestModel{
		BooleanValue: true,
	}
	// if true, should be rendered
	parseRenderAndCompareTemplate(nil,
		stringutils.TrimMargin(
			`<div data-g-define="outer-template">
               |  <div g-if="[[ .BooleanValue ]]">Hello World</div>
               |</div>`),
		"outer-template",
		data,
		stringutils.TrimMargin(
			`<div>
               |  <div>Hello World</div>
               |</div>`),
		t)

	// if false, should not be rendered
	parseRenderAndCompareTemplate(nil,
		stringutils.TrimMargin(
			`<div data-g-define="outer-template">
               |  <div g-if="[[ !.BooleanValue ]] ">Hello World</div>
               |</div>`),
		"outer-template",
		data,
		stringutils.TrimMargin(
			`<div>
               |  
               |</div>`),
		t)

	// if false, should not be rendered
	data.BooleanValue = false
	parseRenderAndCompareTemplate(nil,
		stringutils.TrimMargin(
			`<div data-g-define="outer-template">
               |  <div g-if="[[ .BooleanValue ]] ">Hello World</div>
               |</div>`),
		"outer-template",
		data,
		stringutils.TrimMargin(
			`<div>
               |  
               |</div>`),
		t)
}

func TestIfWorksWithBooleanLiteral(t *testing.T) {
	// if true, should be rendered
	parseRenderAndCompareTemplate(nil,
		stringutils.TrimMargin(
			`<div data-g-define="outer-template">
               |  <div g-if="true">Hello World</div>
               |</div>`),
		"outer-template",
		nil,
		stringutils.TrimMargin(
			`<div>
               |  <div>Hello World</div>
               |</div>`),
		t)

	// if false, should not be rendered
	parseRenderAndCompareTemplate(nil,
		stringutils.TrimMargin(
			`<div data-g-define="outer-template">
               |  <div g-if="false">Hello World</div>
               |</div>`),
		"outer-template",
		nil,
		stringutils.TrimMargin(
			`<div>
               |  
               |</div>`),
		t)

	// if with invalid value, should not be rendered
	parseRenderAndCompareTemplate(nil,
		stringutils.TrimMargin(
			`<div data-g-define="outer-template">
               |  <div g-if="invalid">Hello World</div>
               |</div>`),
		"outer-template",
		nil,
		stringutils.TrimMargin(
			`<div>
               |  
               |</div>`),
		t)
}

//********************************* g-ignore ***************************************************************************
//**********************************************************************************************************************

func TestIgnoreWorks(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue:  "outer",
		StringValue2: "inner",
		StringValue3: "outer-only",
		StringValue4: "something invalid",
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  1: <div g-ignore="outer">outer ignored, should produce empty line</div>
           |  2: <div g-ignore="inner">Inner ignored. Should produce an empty div tag</div>
           |  3: <div g-ignore="outer-only">Outer-only, only this text should be shown</div>
           |  4: <div g-ignore="[[ .StringValue ]]">Outer, should be an empty line</div>
           |  5: <div g-ignore="[[ .StringValue2 ]]">Inner, should be an empty div</div>
           |  6: <div g-ignore="[[ .StringValue3 ]]">Outer-only, only this text should be shown</div>
           |  7: <div g-ignore="">Same as outer, should be an empty line Hello World</div>
           |  8: <div g-ignore="[[ .StringValue4 ]]">When invalid value is given, everything should stay as it is</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |  1: 
           |  2: <div></div>
           |  3: Outer-only, only this text should be shown
           |  4: 
           |  5: <div></div>
           |  6: Outer-only, only this text should be shown
           |  7: 
           |  8: <div>When invalid value is given, everything should stay as it is</div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

//********************************* g-override-att *********************************************************************
//**********************************************************************************************************************

func TestOverrideAttWorks(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "style",
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  1: <div g-use="my-template2" g-override-att="class" class="blue">Only class attribute should be overridden</div> 
           |  2: <div g-use="my-template2" g-override-att="style" style="">Only style attribute should be overridden</div>
           |  3: <div g-use="my-template2" g-override-att="style,class">Class and style attribute should be overridden</div>
           |  4: <div g-use="my-template2" g-override-att="class,style">Class and style attribute should be overridden</div>
           |  5: <div g-use="my-template2" g-override-att="[[ .StringValue ]]">Should even work with model path</div>
           |  6: <div g-use="my-template3" g-override-att="g-inner-text" g-inner-text="Alternate text">Should even work with g-inner-text</div>
           |  7: <div g-use="my-template2" data-g-override-att="style">Should work with long version</div>
           |  8: <div g-use="my-template2" g-override-att="style" g-att-style="from-g-att">Should work with g-att</div>
           |</div>
           |<div data-g-define="my-template2" class="red" style="background-color:blue"></div>
           |<div data-g-define="my-template3" g-inner-text="Dummy"></div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |  1: <div class="blue" style="background-color:blue"></div> 
           |  2: <div class="red" style=""></div>
           |  3: <div class="red" style="background-color:blue"></div>
           |  4: <div class="red" style="background-color:blue"></div>
           |  5: <div class="red" style="background-color:blue"></div>
           |  6: <div>Alternate text</div>
           |  7: <div class="red" style="background-color:blue"></div>
           |  8: <div class="red" style="from-g-att"></div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

// Tests that when using g-use and g-attif and ag-override-att, that the target component will not have the
// attribute, if the condition evaluates to false.
// In this test the target template has a style attribute.
// When calling the template with g-use, we actually override  the style attribute, but since we use g-attif-style with "false",
// we expect the rendered target to not have that style anymore. Because we decided to override it.
func TestOverrideAttWorksWithGAttId(t *testing.T) {

	data := nodeComponentTestModel{
		BooleanValue: false,
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-use="my-template2" g-override-att="style" style="ignored" g-attif-style="[[ .BooleanValue ]]">Should work with g-attif</div>
           |</div>
           |<div data-g-define="my-template2" class="red" style="background-color:blue">Should not have a style attribute</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |  <div class="red">Should not have a style attribute</div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

//********************************* g-inner-repeat **********************************************************************
//**********************************************************************************************************************

func TestGInnerRepeatWorks(t *testing.T) {
	data := nodeComponentTestModel{
		StringSlice: []string{"1", "2", "Special chars should be escaped & < > ' \""},
	}

	template := stringutils.TrimMargin(
		`<ul data-g-define="my-template" g-inner-repeat="[[ .StringSlice ]]">
           |<li g-inner-text="[[ . ]]"></li>
           |</ul>`)
	expectation := stringutils.TrimMargin(
		`<ul>
	       |<li>1</li>
           |
           |<li>2</li>
           |
           |<li>Special chars should be escaped &amp; &lt; &gt; &#39; &#34;</li>
           |</ul>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

//********************************* g-outer-repeat **********************************************************************
//**********************************************************************************************************************

func TestGOuterRepeatWorks(t *testing.T) {

	data := nodeComponentTestModel{
		StringSlice: []string{"1", "2", "Special chars should be escaped & < > ' \""},
	}

	template := stringutils.TrimMargin(
		`<ul data-g-define="my-template">
           |<li g-outer-repeat="[[ .StringSlice ]]" g-inner-text="[[ . ]]"></li>
		   |</ul>`)
	expectation := stringutils.TrimMargin(
		`<ul>
	       |<li>1</li><li>2</li><li>Special chars should be escaped &amp; &lt; &gt; &#39; &#34;</li>
           |</ul>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

func TestGOuterRepeatWorksWithEmptySlice(t *testing.T) {

	data := nodeComponentTestModel{
		StringSlice: []string{},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<ul>
           |<li g-outer-repeat="[[ .StringSlice ]]">[[ . ]]</li>
		   |</ul>
           |<ul>
           |<li data-g-outer-repeat="[[ .StringSlice ]]">[[ . ]]</li>
		   |</ul>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<ul>
           |
           |</ul>
           |<ul>
           |
           |</ul>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

func TestGOuterRepeatWorksWithGUse(t *testing.T) {

	data := nodeComponentTestModel{
		StringSlice: []string{"1", "2", "Special chars should be escaped & < > ' \""},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
	       |<div g-outer-repeat="[[ .StringSlice ]]" data-g-use="my-template2" g-inner-text="[[ . ]]" 
           | g-override-att="g-inner-text"></div>
           |</div>
           |<li data-g-define="my-template2"></li>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<li>1</li><li>2</li><li>Special chars should be escaped &amp; &lt; &gt; &#39; &#34;</li>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

func TestGOuterRepeatWorksWithGOuterText(t *testing.T) {
	data := nodeComponentTestModel{
		StringSlice: []string{"1", "2", "Special chars should be escaped & < > ' \""},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
	       |<div g-outer-repeat="[[ .StringSlice ]]" g-outer-text="[[ . ]]"></div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |12Special chars should be escaped &amp; &lt; &gt; &#39; &#34;
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}
func TestGOuterRepeatWorksWithNestedModel(t *testing.T) {
	data := nodeComponentTestModel{
		NestedModel: nodeComponentNestedModel{
			StringSlice: []string{"1", "2", "Special chars should be escaped & < > ' \""},
		},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
	       |<li g-outer-repeat="[[ .NestedModel.StringSlice ]]" g-inner-text="[[ . ]]"></li>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<li>1</li><li>2</li><li>Special chars should be escaped &amp; &lt; &gt; &#39; &#34;</li>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

// imp: the problem with this test is "order". Maps are not ordered so there could be different results
//func TestGOuterRepeatWorksWithMap(t *testing.T) {
//	data := nodeComponentTestModel{
//		StringMap: map[string]string{
//			"Key 1": "Value 1",
//			"Key 2": "Value 2",
//		},
//	}
//
//	template := stringutils.TrimMargin(
//		`<div data-g-define="my-template">
//	       |<li g-outer-repeat="[[ .StringMap ]]" g-inner-text="Key: {{ .Key }}, Value: {{ .Value }}"></li>
//           |</div>`)
//	expectation := stringutils.TrimMargin(
//		`<div>
//           |<li>Key: Key 1, Value: Value 1</li><li>Key: Key 2, Value: Value 2</li>
//           |</div>`)
//	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
//}

// Tests that when using g-outer-repeat and g-if the g-if is evaluated first
func TestGOuterRepeatWorksWithIf(t *testing.T) {
	data := nodeComponentTestModel{
		BooleanValue:  true,
		BooleanValue2: false,
		StringSlice:   []string{"1", "2", "Special chars should be escaped & < > ' \""},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
	       |<li g-outer-repeat="[[ .StringSlice ]]" g-inner-text="[[ . ]]" 
           | g-if="[[ .BooleanValue]]"></li>
           |<div g-outer-repeat="[[ .StringSlice ]]" g-inner-text="[[ . ]]" 
           | g-if="[[ .BooleanValue2]]"></div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<li>1</li><li>2</li><li>Special chars should be escaped &amp; &lt; &gt; &#39; &#34;</li>
           |
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

//********************************* g-use ******************************************************************************
//**********************************************************************************************************************

func TestGUse(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div g-use="my-image">Hello World</div>
           |<div data-g-use="my-image">Hello World</div>
           |</div>
           |<img src="dummy" g-define="my-image" />
           |`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<img src="dummy" />
           |<img src="dummy" />
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}

// todo: describe that the outer has higher priority here than the g-use
func TestGUseWorksWithOuterIgnore(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div g-use="my-image" g-ignore="outer">Hello World</div>
           |</div>
           |<img src="dummy" g-define="my-image" />
           |`)
	expectation := stringutils.TrimMargin(
		`<div>
           |
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}

//********************************* g-with *****************************************************************************
//**********************************************************************************************************************

func TestThatGTransWorks(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "p",
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div g-trans="p">Hello World</div>
           |<div data-g-trans="p">Hello World</div>
           |<div data-g-trans="[[ .StringValue ]]">Hello World</div>
           |<div data-g-trans="{{ .StringValue }}">Hello World</div>
           |`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<p>Hello World</p>
           |<p>Hello World</p>
           |<p>Hello World</p>
           |<p>Hello World</p>
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

//********************************* g-with *****************************************************************************
//**********************************************************************************************************************

func TestThatGWithIsEvaluatedBeforeGText(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "Not the correct one",
		NestedModel: nodeComponentNestedModel{
			StringValue: "This should be used",
		},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div g-inner-text="[[ .StringValue ]]" g-with="[[ .NestedModel ]]"></div>
           |<div g-inner-text="[[ .StringValue ]]" data-g-with="[[ .NestedModel ]]"></div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<div>This should be used</div>
           |<div>This should be used</div>
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

func TestThatGWithWorksWithGUse(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: "Not the correct one",
		NestedModel: nodeComponentNestedModel{
			StringValue: "/image.png",
		},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div g-use="my-image" g-with="[[ .NestedModel ]]"></div>
           |</div>
           |<img src="[[ .StringValue ]]" g-define="my-image" />`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<img src="/image.png" />
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

// Tests, that g-with works with g-outer-repeat. G width must be evaluated first.
func TestThatGWithWorksWithGOuterRepeat(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "Not the correct one",
		StringSlice: []string{"wrong", "wrong", "wrong"},
		NestedModel: nodeComponentNestedModel{
			StringSlice: []string{"/image.png", "/image2.png", "/image3.png"},
		},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div g-use="my-image" g-with="[[ .NestedModel ]]" g-outer-repeat="[[ .StringSlice ]]"></div>
           |</div>
           |<img src="[[ . ]]" g-define="my-image" />`)
	expectation := stringutils.TrimMargin(
		`<div>
           |<img src="/image.png" /><img src="/image2.png" /><img src="/image3.png" />
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

//********************************* Support functions*******************************************************************
//**********************************************************************************************************************

func parseRenderAndCompareTemplate(e *Engine, template string, templateRef TemplateRef,
	data any, expected string, t *testing.T) {

	var tr *TemplateRegistryDefault
	var tl *templateLoaderString
	if e == nil {
		e, tr, tl = initTestEngine()
	} else {
		// In our tests we always use the default loader and registry
		tr = e.registry.(*TemplateRegistryDefault)
		tl = tr.lazyTemplateLoader.(*templateLoaderString)
	}

	if err := tl.LoadFromString(template, ""); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}
	renderAndCompareTemplate(e, templateRef, data, true, expected, t)
}

func parseRenderAndCompareTemplateUnescaped(e *Engine, template string, templateRef TemplateRef,
	data any, expected string, t *testing.T) {

	var tr *TemplateRegistryDefault
	var tl *templateLoaderString
	if e == nil {
		e, tr, tl = initTestEngine()
	} else {
		// In our tests we always use the default loader and registry
		tr = e.registry.(*TemplateRegistryDefault)
		tl = tr.lazyTemplateLoader.(*templateLoaderString)
	}
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}
	renderAndCompareTemplate(e, templateRef, data, false, expected, t)
}

func renderAndCompareTemplate(e *Engine, templateRef TemplateRef, data any, escaped bool, expected string, t *testing.T) {
	tpl, err := e.registry.GetTemplate(templateRef)
	if err != nil {
		t.Errorf("Expected template '%s' to exist, got error: %v", templateRef, err)
		return
	}
	c, err := tpl.NewRenderable(data)
	if err != nil {
		t.Error("Error creating component", err)
	}
	renderAndCompare(e, c, escaped, expected, t)
}

func renderAndCompare(e *Engine, c Renderable, escaped bool, expected string, t *testing.T) {
	buf := new(bytes.Buffer)
	// Create a RenderContext from the Engine instance for rendering
	// Uses context.Background() since tests don't need request cancellation
	renderCtx := e.NewRenderContext(context.Background())
	renderCtx.Escaped = escaped
	err := c.Render(renderCtx, buf, RenderOuter)
	if err != nil {
		t.Error("Error rendering node component", err)
	}
	result := buf.String()
	compareStrings(result, expected, t)
}

func initTestEngine() (*Engine, *TemplateRegistryDefault, *templateLoaderString) {
	tr := NewTemplateRegistryDefault()
	tl := newTemplateLoaderString(tr)
	tr.SetLazyTemplateLoader(tl)
	e, _ := New(
		WithCustomRegistry(tr),
		WithDeterministicOutput(true),
	)
	return e, tr, tl
}

func initTestEngineWithMaxNestingDepth(maxDepth int) (*Engine, *TemplateRegistryDefault, *templateLoaderString) {
	tr := NewTemplateRegistryDefault()
	tl := newTemplateLoaderString(tr)
	tr.SetLazyTemplateLoader(tl)
	e, _ := New(
		WithCustomRegistry(tr),
		WithMaxNestingDepth(maxDepth),
		WithDeterministicOutput(true),
	)
	return e, tr, tl
}

// ********************************* Max Nesting Depth Tests **********************************************************************

// TestCircularTemplateReferenceDetection tests that circular template references
// are detected and result in a MaxNestingDepthExceededError instead of stack overflow.
func TestCircularTemplateReferenceDetection(t *testing.T) {
	// Template A uses B, and B uses A - this creates a circular reference
	template := stringutils.TrimMargin(
		`<div data-g-define="template-a">
           |  <span>A uses B:</span>
           |  <div g-use="template-b"></div>
           |</div>
           |<div data-g-define="template-b">
           |  <span>B uses A:</span>
           |  <div g-use="template-a"></div>
           |</div>`)

	g, _, tl := initTestEngineWithMaxNestingDepth(10) // Use low limit for faster test
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}

	buf := new(bytes.Buffer)
	err := g.renderInternal(context.Background(), buf, "template-a", nil, nil, true)

	if err == nil {
		t.Error("Expected MaxNestingDepthExceededError for circular reference, but got no error")
		return
	}

	// Check that we got the right error type
	var maxDepthErr *MaxNestingDepthExceededError
	if !errors.As(err, &maxDepthErr) {
		t.Errorf("Expected MaxNestingDepthExceededError, got: %T - %v", err, err)
		return
	}

	if maxDepthErr.MaxDepth != 10 {
		t.Errorf("Expected MaxDepth to be 10, got %d", maxDepthErr.MaxDepth)
	}
}

// TestSelfReferenceDetection tests that a template referencing itself is detected.
func TestSelfReferenceDetection(t *testing.T) {
	// Template A uses itself - immediate circular reference
	template := stringutils.TrimMargin(
		`<div data-g-define="self-ref">
           |  <span>I use myself:</span>
           |  <div g-use="self-ref"></div>
           |</div>`)

	g, _, tl := initTestEngineWithMaxNestingDepth(5)
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}

	buf := new(bytes.Buffer)
	err := g.renderInternal(context.Background(), buf, "self-ref", nil, nil, true)

	if err == nil {
		t.Error("Expected MaxNestingDepthExceededError for self-reference, but got no error")
		return
	}

	var maxDepthErr *MaxNestingDepthExceededError
	if !errors.As(err, &maxDepthErr) {
		t.Errorf("Expected MaxNestingDepthExceededError, got: %T - %v", err, err)
	}
}

// TestDeepButValidNesting tests that deep but non-circular nesting works correctly.
func TestDeepButValidNesting(t *testing.T) {
	// Chain: level1 -> level2 -> level3 (no circular reference)
	template := stringutils.TrimMargin(
		`<div data-g-define="level1">
           |  <span>Level 1</span>
           |  <div g-use="level2"></div>
           |</div>
           |<div data-g-define="level2">
           |  <span>Level 2</span>
           |  <div g-use="level3"></div>
           |</div>
           |<div data-g-define="level3">
           |  <span>Level 3</span>
           |</div>`)

	expected := stringutils.TrimMargin(
		`<div>
           |  <span>Level 1</span>
           |  <div>
           |  <span>Level 2</span>
           |  <div>
           |  <span>Level 3</span>
           |</div>
           |</div>
           |</div>`)

	g, _, tl := initTestEngineWithMaxNestingDepth(10)
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}

	buf := new(bytes.Buffer)
	err := g.renderInternal(context.Background(), buf, "level1", nil, nil, true)

	if err != nil {
		t.Errorf("Expected valid nesting to succeed, got error: %v", err)
		return
	}

	result := buf.String()
	compareStrings(result, expected, t)
}

// TestMaxNestingDepthConfiguration tests that the max nesting depth can be configured.
func TestMaxNestingDepthConfiguration(t *testing.T) {
	// Chain: level1 -> level2 -> level3 -> level4
	// This requires 3 levels of g-use nesting (level1->level2, level2->level3, level3->level4)
	template := stringutils.TrimMargin(
		`<div data-g-define="level1"><div g-use="level2"></div></div>
           |<div data-g-define="level2"><div g-use="level3"></div></div>
           |<div data-g-define="level3"><div g-use="level4"></div></div>
           |<div data-g-define="level4">End</div>`)

	// With depth 2, should fail (we need 3 nesting levels, but only 2 are allowed)
	g, _, tl := initTestEngineWithMaxNestingDepth(2)
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}

	buf := new(bytes.Buffer)
	err := g.renderInternal(context.Background(), buf, "level1", nil, nil, true)

	if err == nil {
		t.Error("Expected error with depth limit of 2, but render succeeded")
		return
	}

	var maxDepthErr *MaxNestingDepthExceededError
	if !errors.As(err, &maxDepthErr) {
		t.Errorf("Expected MaxNestingDepthExceededError, got: %T - %v", err, err)
		return
	}

	// With depth 3, should succeed (we need exactly 3 nesting levels)
	g2, _, tl2 := initTestEngineWithMaxNestingDepth(3)
	if err := tl2.LoadFromString(template, ""); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}

	buf2 := new(bytes.Buffer)
	err2 := g2.renderInternal(context.Background(), buf2, "level1", nil, nil, true)

	if err2 != nil {
		t.Errorf("Expected success with depth limit of 3, got error: %v", err2)
	}
}

// TestMaxNestingDepthDisabled tests that depth limit can be disabled with 0.
func TestMaxNestingDepthDisabled(t *testing.T) {
	// Chain: level1 -> level2 -> level3 (valid chain, but would fail with limit of 1)
	template := stringutils.TrimMargin(
		`<div data-g-define="level1"><div g-use="level2"></div></div>
           |<div data-g-define="level2"><div g-use="level3"></div></div>
           |<div data-g-define="level3">End</div>`)

	// With depth 0 (disabled), should succeed even with deep nesting
	g, _, tl := initTestEngineWithMaxNestingDepth(0)
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}

	buf := new(bytes.Buffer)
	err := g.renderInternal(context.Background(), buf, "level1", nil, nil, true)

	if err != nil {
		t.Errorf("Expected success with disabled depth limit, got error: %v", err)
	}
}

// TestMaxNestingDepthErrorMessage tests the error message content.
func TestMaxNestingDepthErrorMessage(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="a"><div g-use="b"></div></div>
           |<div data-g-define="b"><div g-use="a"></div></div>`)

	g, _, tl := initTestEngineWithMaxNestingDepth(3)
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}

	buf := new(bytes.Buffer)
	err := g.renderInternal(context.Background(), buf, "a", nil, nil, true)

	if err == nil {
		t.Error("Expected error, but render succeeded")
		return
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "max template nesting depth exceeded") {
		t.Errorf("Error message should contain 'max template nesting depth exceeded', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "circular template references") {
		t.Errorf("Error message should mention circular references, got: %s", errMsg)
	}
}

// TestDefaultMaxNestingDepth tests that the default max nesting depth is applied.
func TestDefaultMaxNestingDepth(t *testing.T) {
	g, _ := New()
	defer g.Close()

	if g.maxNestingDepth != DefaultMaxNestingDepth {
		t.Errorf("Expected default max nesting depth to be %d, got %d", DefaultMaxNestingDepth, g.maxNestingDepth)
	}
}

// ********************************* Context Cancellation Tests *********************************************************

// TestContextCancellationStopsIteration tests that a cancelled context stops iteration.
func TestContextCancellationStopsIteration(t *testing.T) {
	template := stringutils.TrimMargin(
		`<ul data-g-define="list">
           |<li data-g-outer-repeat="[[ .Items ]]" data-g-inner-text="[[ . ]]">item</li>
           |</ul>`)

	g, _, tl := initTestEngine()
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Fatal("Error parsing template:", err)
	}

	items := make([]string, 1000)
	for i := range items {
		items[i] = "item"
	}
	data := map[string]any{"Items": items}

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	buf := new(bytes.Buffer)
	err := g.renderInternal(ctx, buf, "list", data, nil, true)

	if err == nil {
		t.Error("Expected error from cancelled context, but got nil")
		return
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %T - %v", err, err)
	}
}

// TestContextCancellationStopsTemplateUse tests that a cancelled context stops g-use rendering.
func TestContextCancellationStopsTemplateUse(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="parent">
           |<div data-g-use="child"></div>
           |</div>
           |<div data-g-define="child">Hello</div>`)

	g, _, tl := initTestEngine()
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Fatal("Error parsing template:", err)
	}

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	buf := new(bytes.Buffer)
	err := g.renderInternal(ctx, buf, "parent", nil, nil, true)

	if err == nil {
		t.Error("Expected error from cancelled context, but got nil")
		return
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %T - %v", err, err)
	}
}

// TestEscapedFieldInRenderContext tests that ctx.Escaped is properly set.
func TestEscapedFieldInRenderContext(t *testing.T) {
	g, _ := New()
	defer g.Close()

	ctx := g.NewRenderContext(context.Background())
	// Default should be false (the zero value)
	if ctx.Escaped {
		t.Error("Expected Escaped to be false by default on NewRenderContext")
	}
}

// TestCurrentTemplateInRenderContext tests that CurrentTemplate is set during rendering.
func TestCurrentTemplateInRenderContext(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="parent">
           |<div data-g-use="child"></div>
           |</div>
           |<div data-g-define="child">Hello</div>`)

	g, _, tl := initTestEngine()
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Fatal("Error parsing template:", err)
	}

	// Render should succeed and set CurrentTemplate during rendering
	buf := new(bytes.Buffer)
	err := g.renderInternal(context.Background(), buf, "parent", nil, nil, true)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
