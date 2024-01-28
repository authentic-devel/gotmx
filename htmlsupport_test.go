package gotmx

import (
	stringutils "github.com/authentic-devel/gotmx/utils"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

func TestGetInnerHtml(t *testing.T) {
	expected := stringutils.TrimMargin(
		`{{ .SomeValue }}
	    | <span>Hello</span> <span>World {{ .SomeOtherValue}}</span> 
        | and some specialChars like ><"'&.
        | In addition here are some void elements:
        |  <area />
        |  <base />
        |  <br />
        |  <table><colgroup><col /></colgroup></table>
        |  <embed />
        |  <hr />
        |  <img src="dummyImg" />
		|  <input type="text" />
        |  <link />
        |  <meta />
        |  <param />
        |  <source />
        |  <track />
        |  <wbr />`)
	template := "<body>" + expected + "</body>"

	reader := strings.NewReader(template)
	node, err := html.Parse(reader)
	if err != nil {
		t.Error("Error parsing template", err)
	}

	// the root node will be the document node. To get the actual node we want, we need  to travel
	// down the tree firstChild (Html), lastChild (body)
	bodyNode := node.FirstChild.LastChild

	innerHtml, err := getInnerHTML(bodyNode)
	if err != nil {
		t.Error("Error getting inner html", err)
	}
	compareStrings(innerHtml, expected, t)

}
