package gotmx

import (
	"bytes"
	"context"
	"testing"

	stringutils "github.com/authentic-devel/gotmx/utils"
)

func TestRenderOuterWorks(t *testing.T) {

	template := `<div id="outer" data-g-define="my-template"><div id="inner">Hello World</div></div>`
	expected := `<div id="outer"><div id="inner">Hello World</div></div>`

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

func TestRenderInnerWorks(t *testing.T) {

	template := `<div data-g-define="my-template"><div>Hello World</div></div>`
	expected := `<div>Hello World</div>`

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

	buf := new(bytes.Buffer)
	// Create a RenderContext from the Engine instance for rendering
	// Uses context.Background() since tests don't need request cancellation
	renderErr := component.Render(g.NewRenderContext(context.Background()), buf, RenderInner)
	if renderErr != nil {
		t.Error("Error rendering component: ", renderErr)
		return
	}
	result := buf.String()
	compareStrings(result, expected, t)
}

func TestRenderInnerWorksWithGInnerText(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
           |<div g-inner-use="my-template2"></div>
           |</div>
           |<div data-g-define id="my-template2" g-inner-text="Hello World from my-template2">Should get replaced</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |Hello World from my-template2
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template1", nil, expectation, t)
}

// Tests that g-inner-use works with g-outer-text defined on the target template.
// What is actually happening is:
// the target component is request to be rendered by using g-inner-use
// When the target component is actually rendered, it normally would skip the surrounding tag because of the g-inner-use.
// But it doesn't even get to that point, because the g-outer-text is evaluated before the element and its children
// are rendered. It has higher priority
func TestGINnerUseWorksWithTargetGOuterText(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
           |<div g-inner-use="my-template2"></div>
           |</div>
           |<div data-g-define id="my-template2" g-outer-text="Hello World from my-template2">Not relevant</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |Hello World from my-template2
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template1", nil, expectation, t)
}

// todo: re-implement support for golang templates
// Tests that a g-outer-repeat works as expected when a g-inner-use is on the same element.
// The expectation is, that the g-inner-use is repeated for each element in the model path, so calling the second
// template multiple times and every time only render its innerHTML.f
//func TestGInnerUseWorksWithGOuterRepeatOnSameElement(t *testing.T) {
//	data := nodeComponentTestModel{
//		StringSlice: []string{"One,", "Two,", "Three,"},
//	}
//	template := stringutils.TrimMargin(
//		`<div data-g-define="my-template1">
//           |<div g-inner-use="my-template2" g-outer-repeat="[[ .StringSlice ]]"></div>
//           |</div>
//           |<div data-g-define="my-template2" data-g-as-template>{{ . }}</div>`)
//
//	expectation := stringutils.TrimMargin(
//		`<div>
//           |One,Two,Three,
//           |</div>`)
//	parseRenderAndCompareTemplate(nil, template, "my-template1", data, expectation, t)
//}

func TestGInnerUseWorksWithSlots(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1" id="root">
           |<div g-inner-use="my-template2">
           |  <article g-use-slot="main">Hello World in the middle</article>
           |</div>
           |</div>
           |<div data-g-define="my-template2">
           |Content before the slot
           |<div data-g-define-slot="main"></div>
           |Content after the slot
           |</div>`)

	expectation := stringutils.TrimMargin(
		`<div id="root">
           |
           |Content before the slot
           |<div><article>Hello World in the middle</article></div>
           |Content after the slot
           |
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template1", nil, expectation, t)
}

// todo: once we have inner-repeat, we also need to test that g-inner-use works with g-inner-repeat on the same element
