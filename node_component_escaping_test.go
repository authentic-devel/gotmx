package gotmx

import (
	stringutils "github.com/authentic-devel/gotmx/utils"
	"testing"
)

// Tests that attributes and text nodes are escaped if requested, in case they are not escaped yet.
// /No matter whether using a literal, a model path or a golang template
func TestTextIsEscapedIfRequested(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: `"&'<>`,
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  These should be escaped by default: ><&'"
           |  Already escaped stuff should remain untouched: .&gt;&lt;&amp;&#39;&#34;
           |  <div attribute="Value is static and does contain special chars like <>&'"></div>
           |  <div attribute='[[ .StringValue ]]'></div>
           |  <div attribute='{{ .StringValue }}'></div>
           |  <!-- Already escapes should be untouched -->
           |  <div attribute="&#34;&amp;&#39;&lt;&gt;"></div>
           |</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |  These should be escaped by default: &gt;&lt;&amp;&#39;&#34;
           |  Already escaped stuff should remain untouched: .&gt;&lt;&amp;&#39;&#34;
           |  <div attribute="Value is static and does contain special chars like &lt;&gt;&amp;&#39;"></div>
           |  <div attribute="&#34;&amp;&#39;&lt;&gt;"></div>
           |  <div attribute="&#34;&amp;&#39;&lt;&gt;"></div>
           |  
           |  <div attribute="&#34;&amp;&#39;&lt;&gt;"></div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

// Tests that text nodes are not escaped if not requested.
func TestTextIsNotEscapedIfNotRequested(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: `"&'<>`,
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  These should not be escaped: ><&'"
           |  This actually will be unescaped because during parsing they will be changed to normal characters: &gt;&lt;&amp;&#39;&#34;
           |  <div attribute="Value is static and does contain special chars like <>&'"></div>
           |  <div attribute='[[ .StringValue ]]'></div>
           |  <div attribute='{{ .StringValue }}'></div>
           |  <!-- Already escapes should be untouched -->
           |  <div attribute="&#34;&amp;&#39;&lt;&gt;"></div>
           |</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |  These should not be escaped: ><&'"
           |  This actually will be unescaped because during parsing they will be changed to normal characters: ><&'"
           |  <div attribute="Value is static and does contain special chars like &lt;&gt;&amp;&#39;"></div>
           |  <div attribute="&#34;&amp;&#39;&lt;&gt;"></div>
           |  <div attribute="&#34;&amp;&#39;&lt;&gt;"></div>
           |  
           |  <div attribute="&#34;&amp;&#39;&lt;&gt;"></div>
           |</div>`)
	parseRenderAndCompareTemplateUnescaped(nil, template, "my-template", data, expectation, t)
}

// Tests, that the content of a g-outer-text attribute is escaped.
func TestGOuterTextIsEscapedIfRequested(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: `"&'<>`,
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-outer-text='{{ .StringValue }}'></div>
           |  <div g-outer-text='[[ .StringValue ]]'></div>
           |  <div g-outer-text="<>&'"></div>
           |</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |  &#34;&amp;&#39;&lt;&gt;
           |  &#34;&amp;&#39;&lt;&gt;
           |  &lt;&gt;&amp;&#39;
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

// Tests that g-outer-text respects the Unescaped() option.
// g-outer-text follows ctx.Escaped since it replaces entire elements and is used
// for raw text extraction (e.g., template references rendered to strings).
func TestGOuterTextIsNotEscapedIfNotRequested(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: `"&'<>`,
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-outer-text='{{ .StringValue }}'></div>
           |  <div g-outer-text='[[ .StringValue ]]'></div>
           |  <div g-outer-text="<>&'"></div>
           |</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |  "&'<>
           |  "&'<>
           |  <>&'
           |</div>`)
	parseRenderAndCompareTemplateUnescaped(nil, template, "my-template", data, expectation, t)
}

// Tests, that the content of a g-inner-text attribute is escaped even when the attribute value does not contain a model
// path or go template expression.
func TestGInnerTextIsEscapedIfRequested(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: `"&'<>`,
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-inner-text='{{ .StringValue }}'></div>
           |  <div g-inner-text='[[ .StringValue ]]'></div>
           |  <div g-inner-text="<>&'"></div>
           |</div>`)

	expectation := stringutils.TrimMargin(
		`<div>
           |  <div>&#34;&amp;&#39;&lt;&gt;</div>
           |  <div>&#34;&amp;&#39;&lt;&gt;</div>
           |  <div>&lt;&gt;&amp;&#39;</div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}

// Tests that g-inner-text always escapes for XSS safety, even with Unescaped().
// This is an intentional security property: g-inner-text is the safe text injection attribute.
// Use g-inner-html for raw HTML injection of trusted content.
func TestGInnerTextAlwaysEscapesEvenWhenUnescaped(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: `"&'<>`,
	}
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-inner-text='{{ .StringValue }}'></div>
           |  <div g-inner-text='[[ .StringValue ]]'></div>
           |  <div g-inner-text="<>&'"></div>
           |</div>`)

	// Even with Unescaped(), g-inner-text still escapes (security)
	expectation := stringutils.TrimMargin(
		`<div>
           |  <div>&#34;&amp;&#39;&lt;&gt;</div>
           |  <div>&#34;&amp;&#39;&lt;&gt;</div>
           |  <div>&lt;&gt;&amp;&#39;</div>
           |</div>`)
	parseRenderAndCompareTemplateUnescaped(nil, template, "my-template", data, expectation, t)
}
