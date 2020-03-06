// This file is generated by gorazor 1.2.2
// DON'T modified manually
// Should edit source file and re-generate: gorazor/tpl/helper/navigation.gohtml

package helper

import (
	"github.com/SlinSo/goTemplateBenchmark/model"
	"github.com/sipin/gorazor/gorazor"
	"io"
	"strings"
)

// Navigation generates gorazor/tpl/helper/navigation.gohtml
func Navigation(nav []*model.Navigation) string {
	var _b strings.Builder
	RenderNavigation(&_b, nav)
	return _b.String()
}

// RenderNavigation render gorazor/tpl/helper/navigation.gohtml
func RenderNavigation(_buffer io.StringWriter, nav []*model.Navigation) {
	// Line: 6
	_buffer.WriteString("\n<ul class=\"navigation\">")

	for _, item := range nav {

		// Line: 10
		_buffer.WriteString("<li><a href=\"")
		// Line: 10
		_buffer.WriteString(gorazor.HTMLEscStr(item.Link))
		// Line: 10
		_buffer.WriteString("\">")
		// Line: 10
		_buffer.WriteString(gorazor.HTMLEscStr(item.Item))
		// Line: 10
		_buffer.WriteString("</a></li>")

	}

	// Line: 12
	_buffer.WriteString("\n</ul>")

}
