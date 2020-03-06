package golang

import (
	"html"

	"github.com/valyala/bytebufferpool"

	"github.com/qorpress/qorpress/benchmarks/go-template/model"
)

func Navigation(bb *bytebufferpool.ByteBuffer, nav []*model.Navigation) {
	bb.WriteString(`<ul class="navigation">`)
	for _, item := range nav {
		bb.WriteString(`<li><a href="`)
		bb.WriteString(html.EscapeString(item.Link))
		bb.WriteString(`">`)
		bb.WriteString(html.EscapeString(item.Item))
		bb.WriteString(`</a></li>`)

	}
	bb.WriteString(`</ul>`)
}
