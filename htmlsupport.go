package gotmx

import (
	"bufio"
	"bytes"
	"io"

	"golang.org/x/net/html"
)

func getInnerHTML(node *html.Node) (string, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	w := bufio.NewWriter(buf)
	if err := renderInnerHTML(w, node); err != nil {
		return "", err
	}
	if err := w.Flush(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderInnerHTML(writer io.Writer, node *html.Node) error {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			if _, err := io.WriteString(writer, child.Data); err != nil {
				return err
			}
		case html.ElementNode:
			if err := renderElementNode(writer, child); err != nil {
				return err
			}
		default:
			// everything else we ignore for now
		}
	}
	return nil
}

func renderElementNode(writer io.Writer, n *html.Node) error {
	isVoid, err := renderOpeningTag(writer, n)
	if err != nil {
		return err
	}
	if isVoid {
		return nil
	}
	if err = renderInnerHTML(writer, n); err != nil {
		return err
	}
	return renderClosingTag(writer, n)
}

func renderOpeningTag(writer io.Writer, n *html.Node) (bool, error) {
	if _, err := io.WriteString(writer, "<"); err != nil {
		return false, err
	}
	if _, err := io.WriteString(writer, n.Data); err != nil {
		return false, err
	}
	if err := renderAttributes(writer, n); err != nil {
		return false, err
	}

	// if the element is a void element (an element that cannot have children), then we can close it now and return
	// because there are no children
	isVoidElement := voidElements[n.Data]
	if isVoidElement {
		if n.FirstChild != nil {
			return true, &VoidElementChildError{Element: n.Data}
		}
		_, err := io.WriteString(writer, " />")
		return true, err
	}
	_, err := io.WriteString(writer, ">")
	if err != nil {
		return false, err
	}
	return false, nil
}

// ----------------------------------------------------------------------------------------------------------------------
func renderClosingTag(writer io.Writer, n *html.Node) error {
	if _, err := io.WriteString(writer, "</"); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, n.Data); err != nil {
		return err
	}
	_, err := io.WriteString(writer, ">")
	return err
}

func renderAttributes(writer io.Writer, node *html.Node) error {
	for _, a := range node.Attr {
		if _, err := io.WriteString(writer, " "); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, a.Key); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, `="`); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, a.Val); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, `"`); err != nil {
			return err
		}
	}
	return nil
}

// Section 12.1.2, "Elements", gives this list of void elements. Void elements
// are those that can't have any contents.
// Note: copied from golang html package
var voidElements = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"keygen": true, // "keygen" has been removed from the spec, but are kept here for backwards compatibility.
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

// Note: copied from golang html package
func childTextNodesAreLiteral(n *html.Node) bool {
	// Per WHATWG HTML 13.3, if the parent of the current node is a style,
	// script, xmp, iframe, noembed, noframes, or plaintext element, and the
	// current node is a text node, append the value of the node's data
	// literally. The specification is not explicit about it, but we only
	// enforce this if we are in the HTML namespace (i.e. when the namespace is
	// "").
	// NOTE: we also always include noscript elements, although the
	// specification states that they should only be rendered as such if
	// scripting is enabled for the node (which is not something we track).
	if n.Namespace != "" {
		return false
	}
	switch n.Data {
	case "iframe", "noembed", "noframes", "noscript", "plaintext", "script", "style", "xmp":
		return true
	default:
		return false
	}

}
