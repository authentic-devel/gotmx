package gotmx

import (
	"testing"

	stringutils "github.com/authentic-devel/gotmx/utils"
)

// Tests that after parsing an HTML snippet with a template definition, the template is found
func TestTemplateIsDiscovered(t *testing.T) {
	_, tr, tl := initTestEngine()
	template := stringutils.TrimMargin(
		`<div data-g-define="my-component">
           |  <span>Hello</span>
           |  <span>World</span>
           |</div>`)
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template", err)
	}
	if _, err := tr.GetTemplate("my-component"); err != nil {
		t.Error("Expected template my-component to be found")
	}

	// test again but this time use the short version of the attribute
	template = stringutils.TrimMargin(
		`<div g-define="my-component2">
           |  <span>Hello</span>
           |  <span>World</span>
           |</div>`)
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template", err)
	}
	if _, err := tr.GetTemplate("my-component2"); err != nil {
		t.Error("Expected template my-component2 to be found")
	}
}

// Tests that also templates defined inside templates are correctly discovered
func TestNestedTemplateIsDiscovered(t *testing.T) {
	_, tr, tl := initTestEngine()

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <span>Hello</span>
           |  <div data-g-define="nested-template">
		   |    <span>World</span>		
           |  </div>
		   |  <div g-define="nested-template2">
		   |    <span>World</span>		
           |  </div>
           |</div>`)
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template", err)
	}
	if _, err := tr.GetTemplate("my-template"); err != nil {
		t.Error("Expected template my-template to exist")
	}
	if _, err := tr.GetTemplate("nested-template"); err != nil {
		t.Error("Expected template nested-template to be found")
	}
	if _, err := tr.GetTemplate("nested-template2"); err != nil {
		t.Error("Expected template nested-template2 to be found")
	}
}

// Tests that a template definition is created, even if the g-define attribute is an empty string.
// In that case, the ID of the element is used as the template name
func TestTemplateIsDiscoveredById(t *testing.T) {

	// todo: adapt documentation

	_, tr, tl := initTestEngine()
	template := stringutils.TrimMargin(
		`<div data-g-define id="outer">
           |  <div data-g-define id="inner">
           |    <span>Hello World</span>
           |  </div>
           |</div>`)

	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template", err)
	}
	if _, err := tr.GetTemplate("outer"); err != nil {
		t.Error("Expected template 'outer' to exist")
	}
	if _, err := tr.GetTemplate("inner"); err != nil {
		t.Error("Expected template 'inner' to exist")
	}

}

func TestTemplateWithEmptyNameGivesErrorWhenMissingId(t *testing.T) {

	// todo: adapt documentation

	_, _, tl := initTestEngine()
	template := stringutils.TrimMargin(
		`<div data-g-define>
           |  <span>Hello World</span>
           |</div>`)

	err := tl.LoadFromString(template, "dummy")
	if err == nil {
		t.Error("Expected error when parsing template without id")
	}
	// todo: this test can probably be improved by checking more detailed what error we get

}

// Tests, that a template is still registered (when using g-define), even if the same html node has a g-ignore
// attribute on it. We explicitly define the behavior like that so that users can define nested templates that aren't
// necessarily always rendered in that spot.
func TestTemplateStillDiscoveredEvenIfIgnored(t *testing.T) {
	_, tr, tl := initTestEngine()
	template := stringutils.TrimMargin(
		`<div data-g-define="outer-template" data-g-ignore='outer'>
           |  <div data-g-define="inner-template" data-g-ignore='outer'>
           |    <span>World</span>
           |  </div>
           |</div>`)
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template", err)
	}
	if _, err := tr.GetTemplate("outer-template"); err != nil {
		t.Error("Expected template 'outer-template' to exist")
	}
	if _, err := tr.GetTemplate("inner-template"); err != nil {
		t.Error("Expected template 'inner-template' to exist")
	}
}

// Tests, that if multiple g-define attributes are present (long and short version), it uses the first one
func TestGDefineUsesWhateverGDefineComesFirst(t *testing.T) {
	_, tr, tl := initTestEngine()
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1" g-define="ignore-me">
           |Hello World
           |</div>
           |<div g-define="my-template2" data-g-define="ignore-me" >
           |Hello World
           |</div>
           |`)
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template", err)
	}
	if _, err := tr.GetTemplate("my-template1"); err != nil {
		t.Error("Expected template 'my-template1' to exist")
	}
	if _, err := tr.GetTemplate("my-template2"); err != nil {
		t.Error("Expected template 'my-template2' to exist")
	}

}

// Tests that getting a template that does not exist returns false when calling GetTemplate
func TestTemplateNotFound(t *testing.T) {
	_, tr, tl := initTestEngine()
	template := stringutils.TrimMargin(
		`<div></div>`)
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template", err)
	}
	if _, err := tr.GetTemplate("unknown-template"); err == nil {
		t.Error("Expected template 'unknown-template' to not exist")
	}
}

func TestReturnsErrorIfDuplicateTemplate(t *testing.T) {
	_, _, tl := initTestEngine()
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template"></div>
		   |<div data-g-define="my-template"></div>`)
	if err := tl.LoadFromString(template, "dummy"); err == nil {
		t.Error("Expected error when parsing template with duplicate template name")
	}
	nestedTemplate := stringutils.TrimMargin(
		`<div data-g-define="my-template">
		   |    <div data-g-define="my-template"></div>
           |</div>`)
	if err := tl.LoadFromString(nestedTemplate, "dummy"); err == nil {
		t.Error("Expected error when parsing nested template with duplicate template name")
	}
}

// todo: re-implement support for golang templates
//func TestGoTemplateInAttributesAreDiscovered(t *testing.T) {
//
//	g := NewGoTmx()
//	tr := g.TemplateRegistry.(*TemplateRegistryDefault)
//	template := stringutils.TrimMargin(
//		`<div data-g-define="my-template">
//           |<div id="{{ .Name }}"></div>
//           |<div id="{{ .Status }}"></div>
//           |<div class="Before {{ .MyClass }} After"></div>
//           |</div>`)
//
//	if err := tr.loadFromString(template, "dummy"); err != nil {
//		t.Error("Error parsing template")
//	}
//
//	textTemplates := g.goTextTemplate.Templates()
//	numTextTemplates := len(textTemplates)
//	expected := 3 // calling Templates on a go Text template does not include the root template
//	if numTextTemplates != expected {
//		t.Errorf("Expected %d go text templates to be found. Got %d", expected, numTextTemplates)
//	}
//
//	htmlTemplates := g.goHtmlTemplate.Templates()
//	numHtmlTemplates := len(htmlTemplates)
//	expected = 4 // calling Templates on a go HTML template DOES include the root template, so we have 4 instead of 3
//	if numHtmlTemplates != expected {
//		t.Errorf("Expected %d go HTML templates to be found. Got %d", expected, numHtmlTemplates)
//	}
//}
