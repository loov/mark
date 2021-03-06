package html

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"strconv"
	"strings"

	"github.com/loov/mark"
)

var (
	linkTemplate  = template.Must(template.New("").Parse(`<a href="{{.Href}}">{{.Title}}</a>`))
	imageTemplate = template.Must(template.New("").Parse(`<figure><img src="{{.Href}}" alt="{{.Title}}" title="{{.Title}}">{{if .Title}}<figcaption>{{.Title}}</figcaption>{{end}}</figure>`))
)

func exec(t *template.Template, data interface{}) string {
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

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
	case mark.SoftBreak:
		return "\n"
	case mark.HardBreak:
		return "<br>"
	case mark.Link:
		return exec(linkTemplate, map[string]interface{}{
			"Href":  el.Href,
			"Title": template.HTML(ConvertParagraph(&el.Title)),
		})
	case mark.Image:
		return exec(imageTemplate, map[string]interface{}{
			"Href":  el.Href,
			"Title": template.HTML(ConvertParagraph(&el.Alt)),
		})
	default:
		panic(fmt.Errorf("unimplemented: %#+v", inline))
	}
}

func ConvertParagraph(el *mark.Paragraph) (r string) {
	for _, item := range el.Items {
		r += ConvertInline(item)
	}
	return r
}

func ConvertBlock(block mark.Block) (r string) {
	switch el := block.(type) {
	case *mark.Sequence:
		for _, item := range *el {
			r += ConvertBlock(item)
		}
		return r
	case *mark.Modifier:
		starttag := "<div class=\"" + template.JSEscapeString(el.Class) + "\">"
		for _, item := range el.Content {
			r += ConvertBlock(item)
		}
		return starttag + r + "</div>"
	case *mark.Code:
		starttag := "<pre><code>"
		if el.Language != "" {
			starttag = "<pre><code class=\"language-" +
				template.JSEscapeString(el.Language) +
				"\">"
		}

		return starttag +
			html.EscapeString(strings.Join(el.Lines, "\n")) +
			"</code></pre>"
	case *mark.Paragraph:
		return "<p>" + ConvertParagraph(el) + "</p>"
	case *mark.Separator:
		if el.Title.IsEmpty() {
			return "<hr>"
		}
		return "<div class=\"separator\">" +
			ConvertParagraph(&el.Title) +
			"</div>"

	case *mark.List:
		for _, seq := range el.Content {
			if len(seq) == 0 {
				r += "<li></li>"
				continue
			}

			if p, ok := seq[0].(*mark.Paragraph); ok {
				r += "<li>" + ConvertParagraph(p) + "</li>"
			} else {
				r += "<li>" + ConvertBlock(&seq) + "</li>"
			}
		}

		if el.Ordered {
			return "<ol>" + r + "</ol>"
		}
		return "<ul>" + r + "</ul>"

	case *mark.Section:
		ht := "h" + strconv.Itoa(el.Level)
		return "<section>" +
			"<" + ht + ">" + ConvertParagraph(&el.Title) + "</" + ht + ">" +
			ConvertBlock(&el.Content) +
			"</section>"
	case *mark.Quote:
		return "<blockquote>" + ConvertBlock(&el.Content) + "</blockquote>"
	default:
		panic(fmt.Errorf("unimplemented: %#+v", block))
	}
}

func Convert(seq mark.Sequence) string {
	return ConvertBlock(&seq)
}
