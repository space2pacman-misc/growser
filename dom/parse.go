package dom

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// Node — упрощённый узел DOM-дерева.
type Node struct {
	Tag      string            // "div", "p", "#text"
	Attrs    map[string]string // атрибуты
	Text     string            // для текстовых узлов
	Children []*Node
	Parent   *Node
}

// Parse превращает HTML-байты в дерево Node.
func Parse(data []byte) (*Node, error) {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	return convert(doc, nil), nil
}

func convert(n *html.Node, parent *Node) *Node {
	node := &Node{Parent: parent}

	switch n.Type {
	case html.ElementNode:
		node.Tag = n.Data
		node.Attrs = make(map[string]string)
		for _, a := range n.Attr {
			node.Attrs[a.Key] = a.Val
		}
	case html.TextNode:
		node.Tag = "#text"
		node.Text = strings.TrimSpace(n.Data)
		// Пропускаем пустые текстовые узлы
		if node.Text == "" {
			return nil
		}
	case html.DocumentNode:
		node.Tag = "#document"
	default:
		return nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if child := convert(c, node); child != nil {
			node.Children = append(node.Children, child)
		}
	}
	return node
}

// Print печатает дерево с отступами — для отладки.
func (n *Node) Print(depth int) {
	indent := strings.Repeat("  ", depth)

	if n.Tag == "#text" {
		fmt.Printf("%s\"%s\"\n", indent, n.Text)
		return
	}

	fmt.Printf("%s<%s", indent, n.Tag)
	for k, v := range n.Attrs {
		fmt.Printf(" %s=\"%s\"", k, v)
	}
	fmt.Println(">")

	for _, c := range n.Children {
		c.Print(depth + 1)
	}
}

// Find ищет первый узел по тегу (обход в глубину).
func (n *Node) Find(tag string) *Node {
	if n.Tag == tag {
		return n
	}
	for _, c := range n.Children {
		if found := c.Find(tag); found != nil {
			return found
		}
	}
	return nil
}

// TextContent собирает весь текст в поддереве.
func (n *Node) TextContent() string {
	if n.Tag == "#text" {
		return n.Text + " "
	}
	var sb strings.Builder
	for _, c := range n.Children {
		sb.WriteString(c.TextContent())
	}
	return sb.String()
}