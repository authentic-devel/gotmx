package gotmx

import (
	"testing"

	stringutils "github.com/authentic-devel/gotmx/utils"
)

func TestGAttIfWorksWithLiterals(t *testing.T) {

	data := nodeComponentTestModel{
		BooleanValue:  true,
		BooleanValue2: false,
	}

	template := stringutils.TrimMargin(
		`<div data-g-define="my-templateWithTrue">
           |  <div g-att-class="primary" g-attif-class="true">Lorem ipsum</div>
           |  <div g-att-class="primary" g-attif-class="false">Lorem ipsum</div>
           |  <div class="secondary" g-attif-class="[[ .BooleanValue ]]">Lorem ipsum</div>
           |  <div class="secondary" g-attif-class="[[ .BooleanValue2 ]]">Lorem ipsum</div>
           |  <button g-attif-enabled="[[ .BooleanValue ]]">Lorem ipsum</button>
           |  <button g-attif-enabled="[[ .BooleanValue2 ]]">Lorem ipsum</button>
           |</div>`)
	expected := stringutils.TrimMargin(
		`<div>
           |  <div class="primary">Lorem ipsum</div>
           |  <div>Lorem ipsum</div>
           |  <div class="secondary">Lorem ipsum</div>
           |  <div>Lorem ipsum</div>
           |  <button enabled="">Lorem ipsum</button>
           |  <button>Lorem ipsum</button>
           |</div>`)

	parseRenderAndCompareTemplate(nil, template, "my-templateWithTrue", data, expected, t)
}
