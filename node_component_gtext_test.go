package gotmx

import (
	stringutils "github.com/authentic-devel/gotmx/utils"
	"testing"
)

//********************************* g-inner-text ***********************************************************************
//**********************************************************************************************************************

func TestGInnerTextWorks(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: `Hello World with special chars > < & " '`,
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           | <!-- long version -->
           | <div data-g-inner-text="Hello World">This should get replaced by the text</div>
           | <!-- short version -->
           | <div g-inner-text="Hello World">This should get replaced by the text</div>
           | <!-- With model path -->
           | <div g-inner-text="[[ .StringValue ]]">This should get replaced by the text</div>
           | <!-- With Golang template -->
           | <div g-inner-text="{{ .StringValue }}">This should get replaced by the text</div>
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           | 
           | <div>Hello World</div>
           | 
           | <div>Hello World</div>
           | 
           | <div>Hello World with special chars &gt; &lt; &amp; &#34; &#39;</div>
           | 
           | <div>Hello World with special chars &gt; &lt; &amp; &#34; &#39;</div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}

func TestGInnerTextLongVersionHasHigherPriority(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template" g-inner-text="ignore-me" data-g-inner-text="Hello World">
           |This should get replaced by the text
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>Hello World</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}

// Tests that the g-inner-text attribute is evaluated after the g-if attribute
func TestGInnerTextIsEvaluatedAfterIf(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-if="[[ .BooleanValue ]]" g-inner-text="Hello World">This should get replaced by the text</div>
           |</div>`)
	expectationForTrue := stringutils.TrimMargin(
		`<div>
           |  <div>Hello World</div>
           |</div>`)
	expectationForFalse := stringutils.TrimMargin(
		`<div>
           |  
           |</div>`)
	data := nodeComponentTestModel{
		BooleanValue: true,
	}

	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectationForTrue, t)
	data.BooleanValue = false
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectationForFalse, t)
}

func TestGInnerTextWorksWithIgnoreOuter(t *testing.T) {

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-ignore="outer" g-inner-text="Hello World">Lorem ipsum</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |  
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}
func TestGInnerTextWorksWithIgnoreInner(t *testing.T) {

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-ignore="inner" g-inner-text="Hello World">Lorem ipsum</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |  <div></div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}
func TestGInnerTextWorksWithIgnoreOuterOnly(t *testing.T) {

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-ignore="outer-only" g-inner-text="Hello World">Lorem ipsum</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |  Hello World
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}

//********************************* g-outer-text ***********************************************************************
//**********************************************************************************************************************

func TestGOuterTextWorks(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: `Hello World with special chars > < & " '`,
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           | <!-- long version -->
           | <div data-g-outer-text="Hello World">This should get replaced by the text</div>
           | <!-- short version -->
           | <div g-outer-text="Hello World">This should get replaced by the text</div>
           | <!-- With model path -->
           | <div g-outer-text="[[ .StringValue ]]">This should get replaced by the text</div>
           | <!-- With Golang template -->
           | <div g-outer-text="{{ .StringValue }}">This should get replaced by the text</div>
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           | 
           | Hello World
           | 
           | Hello World
           | 
           | Hello World with special chars &gt; &lt; &amp; &#34; &#39;
           | 
           | Hello World with special chars &gt; &lt; &amp; &#34; &#39;
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}

func TestGOuterTextIsIgnoredForAnyIgnore(t *testing.T) {

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div g-ignore="outer" g-outer-text="Hello World">Lorem ipsum</div>
           |<div g-ignore="inner" g-outer-text="Hello World">Lorem ipsum</div>
           |<div g-ignore="outer-only" g-outer-text="Hello World">Lorem ipsum</div>
           |<div g-ignore="invalid" g-outer-text="Hello World">Lorem ipsum</div>
           |<div g-ignore g-outer-text="Hello World">Lorem ipsum</div>
           |<div data-g-ignore g-outer-text="Hello World">Lorem ipsum</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |
           |
           |
           |
           |
           |
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}

func TestGOuterTextWorksWithGIf(t *testing.T) {

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div data-g-if="false" g-outer-text="Hello World" >Lorem ipsum</div>
           |---
           |<div data-g-if="true" g-outer-text="Hello World" >Lorem ipsum</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |
           |---
           |Hello World
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expectation, t)
}

func TestGOuterTextWorksWithGRepeat(t *testing.T) {

	data := nodeComponentTestModel{
		StringSlice: []string{"String 1\n", "Spacial chars > < & '\n", "String 3"},
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |<div data-g-outer-repeat="[[ .StringSlice ]]"  g-outer-text="[[ . ]]" >Lorem ipsum</div>
           |</div>`)
	expectation := stringutils.TrimMargin(
		`<div>
           |String 1
           |Spacial chars &gt; &lt; &amp; &#39;
           |String 3
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expectation, t)
}
