package gotmx

import (
	stringutils "github.com/authentic-devel/gotmx/utils"
	"testing"
)

func TestGAttWorksWithLiterals(t *testing.T) {
	data := nodeComponentTestModel{
		StringValue: "primary",
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-att-data-x="my-value">Lorem ipsum</div>
           |  <div data-g-att-data-x="my-value">Lorem ipsum</div>
           |  <div g-att-data-y="[[ .StringValue ]]">Lorem ipsum</div>
           |  <div g-att-data-z="{{ .StringValue }}">Lorem ipsum</div>
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |  <div data-x="my-value">Lorem ipsum</div>
           |  <div data-x="my-value">Lorem ipsum</div>
           |  <div data-y="primary">Lorem ipsum</div>
           |  <div data-z="primary">Lorem ipsum</div>
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}

func TestGAttWorksWithBooleanAttribute(t *testing.T) {
	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div g-att-data-bool="">Lorem ipsum</div>
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |  <div data-bool="">Lorem ipsum</div>
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-template", nil, expected, t)
}

func TestGClass(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "purple",
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <div class="red" g-class="blue">Lorem ipsum</div>
           |  <div g-class="green" class="yellow">Lorem ipsum</div>
           |  <div data-g-class="green" class="yellow">Lorem ipsum</div>
           |  <div g-class="[[ .StringValue ]]" class="yellow">Lorem ipsum</div>
           |  <div g-class="{{ .StringValue }}" class="yellow">Lorem ipsum</div>
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |  <div class="blue">Lorem ipsum</div>
           |  <div class="green">Lorem ipsum</div>
           |  <div class="green">Lorem ipsum</div>
           |  <div class="purple">Lorem ipsum</div>
           |  <div class="purple">Lorem ipsum</div>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}

func TestGHref(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "/production",
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <a href="" g-href="#">Lorem ipsum</a>
           |  <a g-href="/dev" data-g-href="/prod">Lorem ipsum</a>
           |  <a data-g-href="/" href="#">Lorem ipsum</a>
           |  <a g-href="[[ .StringValue ]]" href="/preview">Lorem ipsum</a>
           |  <a g-href="{{ .StringValue }}" href=="/preview">Lorem ipsum</a>
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |  <a href="#">Lorem ipsum</a>
           |  <a href="/prod">Lorem ipsum</a>
           |  <a href="/">Lorem ipsum</a>
           |  <a href="/production">Lorem ipsum</a>
           |  <a href="/production">Lorem ipsum</a>
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}

func TestGSrc(t *testing.T) {

	data := nodeComponentTestModel{
		StringValue: "/production",
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-template">
           |  <img src="" g-src="#" />
           |  <img g-src="/dev" data-g-src="/prod" />
           |  <img data-g-src="/" src="#" />
           |  <img g-src="[[ .StringValue ]]" src="/preview" />
           |  <img g-src="{{ .StringValue }}" src=="/preview" />
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |  <img src="#" />
           |  <img src="/prod" />
           |  <img src="/" />
           |  <img src="/production" />
           |  <img src="/production" />
           |</div>`)
	parseRenderAndCompareTemplate(nil, template, "my-template", data, expected, t)
}
