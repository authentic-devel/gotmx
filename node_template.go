package gotmx

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
)

type AttributeMap map[string]string

type nodeTemplate struct {
	node *html.Node

	name      TemplateName
	namespace Namespace
}

func (n *nodeTemplate) Name() TemplateName {
	return n.name
}

func (n *nodeTemplate) Namespace() Namespace {
	return n.namespace
}

func (n *nodeTemplate) NewRenderable(data any) (Renderable, error) {
	return newNodeComponent(n.node, data), nil
}

func processNode(tr TemplateRegistry, sourceFile string, node *html.Node) error {
	switch node.Type {
	case html.DocumentNode, html.ElementNode:
		if err := processAttributes(tr, sourceFile, node); err != nil {
			return err
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		err := processNode(tr, sourceFile, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func processAttributes(tr TemplateRegistry, sourceFile string, node *html.Node) error {

	var templateAlreadyDefined = false

	for index := range node.Attr {
		// we need to pass the original attribute as a pointer
		// The for loop gives as a copy of the attribute so we need to access by index here.
		attr := &node.Attr[index]

		switch attr.Key {
		case attrDataGDefine, attrGDefine:
			if templateAlreadyDefined {
				continue
			}
			if err := processDefine(tr, sourceFile, node, attr.Val); err != nil {
				return err
			}
			templateAlreadyDefined = true
		case attrDataGAsTemplate, attrGAsTemplate, attrGAsUnsafeTemplate, attrDataGAsUnsafeTemplate:
			if err := processNodeAsTemplate(tr, sourceFile, node, attr); err != nil {
				return err
			}
		default:
			// it is not a g-define attribute but any other regular attribute we check, whether there is a Golang
			// template inside. If there is, we parse it, register it and then change the attribute to a simple
			// template call
			goRegistry, isGoRegistry := tr.(GoTemplateRegistry)

			if isGoRegistry && strings.Contains(attr.Val, "{{") && strings.Contains(attr.Val, "}}") {
				templateName := nextTemplateID()

				if err := goRegistry.RegisterGoTemplate(TemplateName(templateName), attr.Val, sourceFile); err != nil {
					return err
				}
				attr.Val = "[[ :" + templateName + " ]]"
			}
		}
	}
	return nil
}

// Called, when a node has the g-as-template attribute. It takes the innerHTML of the node and registers it as a
// Golang template. if the attribute has a template name, that name is used.
// If it does not have a value, then a random ULID is generated and used as name.
// The actual attribute value is replaced by the ULID of the golang template as a side effect.
func processNodeAsTemplate(tr TemplateRegistry, sourceFile string, node *html.Node, attr *html.Attribute) error {

	golangRegistry, isGolangRegistry := tr.(GoTemplateRegistry)
	if !isGolangRegistry {
		return errors.New("template registry does not support go templates")
	}

	// first we will get the innerHTML of the node, because that will be the template
	innerHTML, err := getInnerHTML(node)
	if err != nil {
		return err
	}
	templateName := attr.Val
	if templateName == "" {
		templateName = nextTemplateID()
		attr.Val = templateName
	}
	if err = golangRegistry.RegisterGoTemplate(TemplateName(templateName), innerHTML, sourceFile); err != nil {
		return err
	}
	return nil
}

func processDefine(tr TemplateRegistry, sourceFile string, node *html.Node, defineAttr string) error {
	if defineAttr == "" {
		attr, exists := findAttribute(node, "id")
		if !exists {
			return errors.New("template with empty g-define must have an id attribute")
		}
		defineAttr = attr

	}
	template := nodeTemplate{
		node:      node,
		name:      TemplateName(defineAttr),
		namespace: Namespace(sourceFile),
	}
	return tr.RegisterTemplate(&template)
}

func findAttribute(node *html.Node, key string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}
