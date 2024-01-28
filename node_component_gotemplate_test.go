package gotmx

import (
	"testing"

	stringutils "github.com/authentic-devel/gotmx/utils"
)

// TODO: need more tests for g-as-template
//   How in Combination with g-inner-text?
//   How in combination with g-use?
//   How in combination with g-use-slot?
//   Does it work on the same node as g-define?
//   How does it behave with g-ignore on the same node?
//   What if short and long version are on the same node? (long version has prio)

func TestBasicGolangTemplates(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue:   "The String Value",
		StringValue2:  `> < & " '`,
		BooleanValue:  true,
		BooleanValue2: false,
		StringSlice:   []string{"String 1", "String 2", `Special characters > < & " '`},
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template" g-as-template="my-golang-html-template">
          |{{ .StringValue -}}
          |{{- if .BooleanValue }} should be shown{{end -}}
          |{{- if .BooleanValue2 -}} should not be shown{{- end}}
          |Special characters from the model should be escaped: {{ .StringValue2 }}
          |Unless they are literals in the template > < & " '
          |<ul>
          |{{range .StringSlice -}}
          |<li>{{.}}</li>
          |{{end -}}
          |</ul>
          |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
          |The String Value should be shown
          |Special characters from the model should be escaped: &gt; &lt; &amp; &#34; &#39;
          |Unless they are literals in the template > &lt; & " '
          |<ul>
          |<li>String 1</li>
          |<li>String 2</li>
          |<li>Special characters &gt; &lt; &amp; &#34; &#39;</li>
          |</ul>
          |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}

func TestGolangTemplateIsDiscoveredInsideGAsTemplate(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template" g-as-template>
          |{{define "golang-sub-template" }}
          |Content of the sub template
          |{{end}}
          |{{ .StringValue }}
          |<ul>
          |{{range .StringSlice}}
          |<li>{{.}}</li>
          |{{end}}
          |</ul>
          |</div>`)
	_, tr, tl := initTestEngine()
	if err := tl.LoadFromString(template, "dummy"); err != nil {
		t.Error("Error parsing template", err)
	}
	templates := tr.goTextTemplate.Templates()
	numTemplates := len(templates)
	if numTemplates != 2 {
		t.Error("Expected 2 templates to be found, but found", numTemplates)
	}

	if tr.goHtmlTemplate.Lookup("golang-sub-template") == nil {
		t.Error("Expected template 'golang-sub-template' to be found")
	}

	// the same templates must also be registered in the text version of the templates
	templates = tr.goTextTemplate.Templates()
	numTemplates = len(templates)
	if numTemplates != 2 {
		t.Error("Expected 2 templates to be found, but found", numTemplates)
	}

	if tr.goTextTemplate.Lookup("golang-sub-template") == nil {
		t.Error("Expected template 'golang-sub-template' to be found")
	}

	// imp: also need to test whether the whole template was registered correctly which is harder since it has
	// a random name
}

func TestBasicGolangTextTemplates(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue:   "The String Value",
		StringValue2:  `> < & " '`,
		BooleanValue:  true,
		BooleanValue2: false,
		StringSlice:   []string{"String 1", "String 2", `Special characters > < & " '`},
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template" g-as-unsafe-template="my-golang-inner-text-template">
         |{{ .StringValue -}}
         |{{- if .BooleanValue }} should be shown{{end -}}
         |{{- if .BooleanValue2 -}} should not be shown{{- end}}
         |Special characters from the model should NOT be escaped: {{ .StringValue2 }}
         |As well as literals in the template > < & " '
         |<ul>
         |{{range .StringSlice -}}
         |<li>{{.}}</li>
         |{{end -}}
         |</ul>
         |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
         |The String Value should be shown
         |Special characters from the model should NOT be escaped: > < & " '
         |As well as literals in the template > < & " '
         |<ul>
         |<li>String 1</li>
         |<li>String 2</li>
         |<li>Special characters > < & " '</li>
         |</ul>
         |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}

// Tests that a Golang template also works, even it has no explicit name
func TestGolangTemplatesWithoutExplicitName(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: "The String Value",
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template" g-as-template>
         |{{ .StringValue }}
         |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
         |The String Value
         |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}

func TestGolangTemplatesCanReferenceEachOtherInAsTemplate(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: `Hello World with special characters > < & " '`,
	}
	template1 := stringutils.TrimMargin(
		`<div data-g-define="my-template1" g-as-template="my-golang-html-template1">
        |{{- template "my-golang-html-template2" . -}}
        |</div>`)
	template2 := stringutils.TrimMargin(
		`<div data-g-define="my-template2" g-as-template="my-golang-html-template2">
        |{{ .StringValue }}
        |</div>`)

	expected := stringutils.TrimMargin(
		`<div>
        |Hello World with special characters &gt; &lt; &amp; &#34; &#39;
        |</div>`)

	g, _, tl := initTestEngine()

	if err := tl.LoadFromString(template1, "dummy"); err != nil {
		t.Error("Error parsing template1", err)
	}
	if err := tl.LoadFromString(template2, "dummy"); err != nil {
		t.Error("Error parsing template2", err)
	}
	renderAndCompareTemplate(g, "my-template1", data, true, expected, t)
}

// Tests, that a golang template within an attribute can reference any other Golang template
func TestGolangTemplatesCanReferenceEachOtherInAttribute(t *testing.T) {
	template1 := stringutils.TrimMargin(
		`<div data-g-define="my-template1">
         |<div g-as-template="other_golang_template" g-ignore>Hello World</div>
         |<div g-inner-text='{{- template "other_golang_template" . -}}'></div>
         |</div>

         |`)

	expected := stringutils.TrimMargin(
		`<div>
         |
         |<div>Hello World</div>
         |</div>`)

	g, _, tl := initTestEngine()
	if err := tl.LoadFromString(template1, "dummy"); err != nil {
		t.Error("Error parsing template1", err)
	}
	renderAndCompareTemplate(g, "my-template1", nil, true, expected, t)
}

func TestGolangTemplatesWithCustomFunction(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: "World",
	}
	g, tr, _ := initTestEngine()
	tr.RegisterFunc("MyFunc", func(value string) (string, error) {
		return "Hello " + value, nil
	})

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1" g-as-template="my-golang-html-template">
         |{{-  (MyFunc .StringValue ) -}}
         |</div>`)

	expected := `<div>Hello World</div>`

	parseRenderAndCompareTemplate(g, template, "my-template1", data, expected, t)

}

func TestGolangTemplatesCanCallGotmxTemplate(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: `World with special chars > < & \" '`,
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template1" g-as-unsafe-template="my-golang-html-template">
         |{{-  (GTemplate "my-template2" .) -}}
         |</div>
         |<div data-g-define="my-template2">
		   |Hello <span g-inner-text="[[ .StringValue ]]"></span>
         |</div>`)

	expected := stringutils.TrimMargin(
		`<div><div>
         |Hello <span>World with special chars &gt; &lt; &amp; \&#34; &#39;</span>
         |</div></div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template1", data, expected, t)
}

func TestGolangTemplatesCanCallGotmxTextTemplate(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: `special chars > < & \" '`,
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template" g-as-template="my-golang-html-template">
         |{{-  (GTextTemplate "my-template2" .) -}}
         |</div>
         |<div data-g-define="my-template2" g-inner-text="All should be escaped since it is called as a text template"></div>`)

	expected := stringutils.TrimMargin(
		`<div>&lt;div&gt;All should be escaped since it is called as a text template&lt;/div&gt;</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}
