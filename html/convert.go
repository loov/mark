package html

import (
	"fmt"
	"html"
	"strconv"

	"github.com/loov/mark"
)

func ConvertInline(inline mark.Inline) (r string) {
	switch el := inline.(type) {
	case mark.Text:
		return html.EscapeString(string(el))
	case mark.Emphasis:
		for _, x := range el {
			r += ConvertInline(x)
		}
		return "<em>" + r + "</em>"
	case mark.Bold:
		for _, x := range el {
			r += ConvertInline(x)
		}
		return "<b>" + r + "</b>"
	case mark.CodeSpan:
		x := html.EscapeString(string(el))
		return "<code>" + x + "</code>"
	case mark.LineBreak:
		return "<br>"
	default:
		panic(fmt.Errorf("unimplemented: %#+v", inline))
	}
}

func ConvertBlock(block mark.Block) (r string) {
	switch el := block.(type) {
	case *mark.Sequence:
		for _, item := range *el {
			r += ConvertBlock(item)
		}
		return r
	case *mark.Paragraph:
		for _, item := range el.Items {
			r += ConvertInline(item)
		}
		return "<p>" + r + "</p>"
	case *mark.Section:
		ht := "h" + strconv.Itoa(el.Level)
		return "<section>" +
			"<" + ht + ">" + ConvertBlock(&el.Title) + "</" + ht + ">" +
			ConvertBlock(&el.Content) +
			"</section>"
	default:
		panic(fmt.Errorf("unimplemented: %#+v", block))
	}
}

func Convert(seq mark.Sequence) string {
	return ConvertBlock(&seq)
}
