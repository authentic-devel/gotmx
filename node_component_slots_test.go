package gotmx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	stringutils "github.com/authentic-devel/gotmx/utils"
)

//********************************* Slotting children into Components **************************************************
//**********************************************************************************************************************

func TestBasicSlotsWork(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-define-slot="slot1">This should get replaced</div>
           |  <div g-define-slot="">This is the default slot and the content should also get replaced</div>
           |  <div g-define-slot="slotIgnored" data-g-define-slot="slot2"></div>
           |</div>`)

	expected := stringutils.TrimMargin(
		`<div>
           |  <div>Hello World</div>
           |  <div>Hello World2</div>
           |  <div>Hello WorldHello World2</div>
           |</div>`)

	g, _, tl := initTestEngine()
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template: ", err)
		return
	}

	component, err := g.createComponent("my-template", nil)
	if err != nil {
		t.Error("Expected component to exist, got error:", err)
		return
	}
	asHasChildren, ok := component.(HasChildren)
	if !ok {
		t.Error("Expected component to be HasChildren")
		return
	}
	child1, _ := NewStringLiteralTemplate("my-template1", "Hello World", "dummy").NewRenderable(nil)
	child2, _ := NewStringLiteralTemplate("my-template2", "Hello World2", "dummy").NewRenderable(nil)

	asHasChildren.AddChild("slot1", child1)
	asHasChildren.AddChild("", child2)
	asHasChildren.AddChild("slot2", child1)
	asHasChildren.AddChild("slot2", child2)

	buf := new(bytes.Buffer)
	// Create a RenderContext from the Engine instance for rendering
	// Uses context.Background() since tests don't need request cancellation
	renderErr := component.Render(g.NewRenderContext(context.Background()), buf, RenderOuter)
	if renderErr != nil {
		t.Error("Error rendering component: ", renderErr)
		return
	}
	result := buf.String()
	compareStrings(result, expected, t)

}

//********************************* g-use ******************************************************************************
//**********************************************************************************************************************

func TestGUseWorks(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "my-template2",
	}

	// This contains various version of defining g-use.
	// - The long attribute
	// - the short attribute
	// - testing with literal
	// - testing with model path
	// - testing with Golang  template
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
           |<div g-use="[[ .StringValue ]]">Will be replaced</div>
           |<div g-use="{{ .StringValue }}">Will be replaced</div>
           |<div g-use="my-template2">Will be replaced</div>
           |<div data-g-use="[[ .StringValue ]]">Will be replaced</div>
           |<div data-g-use="{{ .StringValue }}">Will be replaced</div>
           |<div data-g-use="my-template2">Will be replaced</div>
           |</div>
           |<div data-g-define id="my-template2">Hello World</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |<div id="my-template2">Hello World</div>
		   |<div id="my-template2">Hello World</div>
		   |<div id="my-template2">Hello World</div>
		   |<div id="my-template2">Hello World</div>
		   |<div id="my-template2">Hello World</div>
		   |<div id="my-template2">Hello World</div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template1", data, expectation, t)

}

// Tests that g-use returns a ComponentNotFoundError if the value references an invalid / unknown template.
func TestGUseReturnsErrorWithInvalidModelPath(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "not-existing",
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
           |<div g-use="[[ .StringValue ]]">Will be replaced</div>
           |</div>`)

	g, _, tl := initTestEngine()
	if err := tl.LoadFromString(template, ""); err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	tpl, err := g.registry.GetTemplate("my-template1")
	if err != nil {
		t.Fatalf("Error getting template: %v", err)
	}

	c, err := tpl.NewRenderable(data)
	if err != nil {
		t.Fatalf("Error creating component: %v", err)
	}

	var buf bytes.Buffer
	renderCtx := g.NewRenderContext(context.Background())
	renderCtx.Escaped = true
	err = c.Render(renderCtx, &buf, RenderOuter)
	if err == nil {
		t.Fatal("Expected error when referencing non-existent component, but got none")
	}

	var compErr *ComponentNotFoundError
	if !errors.As(err, &compErr) {
		t.Errorf("Expected ComponentNotFoundError, got: %T (%v)", err, err)
	}
	if compErr.ComponentRef != "not-existing" {
		t.Errorf("Expected ComponentRef to be 'not-existing', got: %q", compErr.ComponentRef)
	}
}

// Tests, that the children oof an element with a g-use attribute are ignored, if the referenced template has a no default
// slot
func TestGUseChildrenIgnoredIfTargetHasNoDefaultSlot(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
		   |<div g-use="my-template2">
           |  <p>This paragraph should be ignored, because the target component does not have a default slot</p>
		   |</div>
		   |</div>
		   |<div data-g-define id="my-template2">Hello World</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
		   |<div id="my-template2">Hello World</div>
		   |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template1", nil, expectation, t)
}

// Tests, that
// - the children oof an element with a g-use attribute are placed in the correct slot of the target component
// - the long version and the short version of g-use-slot works
// - g-use-slot can use literals, model paths and Golang templates
// - only whitespace text nodes are not added to the default slot
func TestGUseChildrenAreSlottedCorrectly(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "slot1",
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
		   |<div g-use="my-template2">
           |  <p>This paragraph should go into the default slot</p>
           |  <p g-use-slot="slot1">in slot1</p>
           |  <p g-use-slot="{{ .StringValue }}">also in slot1</p>
           |  <p g-use-slot="[[ .StringValue ]]">last in slot1</p>
           |  <!-- The whitespace after this should not be added to the default slot -->
           |
           |  <p data-g-use-slot="slot2">This paragraph should go into slot2</p>
		   |</div>
		   |</div>
		   |<div data-g-define id="my-template2">
           | <div g-define-slot="slot2" id="slot2">
           | </div>
           | <div data-g-define-slot="slot1" id="slot1">
           | </div>
           | <div data-g-define-slot id="defaultSlot">
           | This content should get replaced by the slot content 
           | </div>
           | <p>This paragraph should also be present</p>
           |</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
		   |<div id="my-template2">
           | <div id="slot2"><p>This paragraph should go into slot2</p></div>
           | <div id="slot1"><p>in slot1</p><p>also in slot1</p><p>last in slot1</p></div>
           | <div id="defaultSlot"><p>This paragraph should go into the default slot</p></div>
           | <p>This paragraph should also be present</p>
           |</div>
		   |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template1", data, expectation, t)
}

// Tests, that when a template has multiple slots with the same name, that the slotted children will be rendered in
// all of them.
func TestMultipleSlotsRendersMultipleTimes(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
		   |<div g-use="my-template2">
           |<p g-use-slot="slot1">in slot1</p>
           |</div>
           |</div>
           |
           |<div data-g-define id="my-template2">
           |<div g-define-slot="slot1"></div>
           |<hr>
           |<div g-define-slot="slot1"></div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
		   |<div id="my-template2">
           |<div><p>in slot1</p></div>
           |<hr />
           |<div><p>in slot1</p></div>
           |</div>
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template1", nil, expectation, t)
}

func TestGUseWorksWithCustomRenderables(t *testing.T) {

	g, tr, _ := initTestEngine()
	customTemplate := myCustomTemplate{
		name: "custom",
		text: "Hello World from custom Renderable",
	}
	err := tr.RegisterTemplate(customTemplate)
	if err != nil {
		t.Error("Error registering custom template", err)
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
           |<div g-use="custom">Will be replaced</div>
           |</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |Hello World from custom Renderable
           |</div>`)
	parseRenderAndCompareTemplate(g, template, "my-template1", nil, expectation, t)
}

type myCustomTemplate struct {
	name TemplateName
	text string
}

func (c myCustomTemplate) Name() TemplateName { return c.name }
func (c myCustomTemplate) Namespace() Namespace {
	return ""
}
func (c myCustomTemplate) NewRenderable(data any) (Renderable, error) {
	return c, nil
}

func (c myCustomTemplate) Render(_ *RenderContext, writer io.Writer, _ RenderType) error {
	_, err := writer.Write([]byte(c.text))
	return err
}
