package html

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// A Config struct has configurations for the HTML based renderers.
type Config struct {
	Writer        Writer
	SoftLineBreak bool
	XHTML         bool
}

// NewConfig returns a new Config with defaults.
func NewConfig() Config {
	return Config{
		Writer:        DefaultWriter,
		SoftLineBreak: false,
		XHTML:         false,
	}
}

// SetOption implements renderer.NodeRenderer.SetOption.
func (c *Config) SetOption(name renderer.OptionName, value interface{}) {
	switch name {
	case SoftLineBreak:
		c.SoftLineBreak = value.(bool)
	case XHTML:
		c.XHTML = value.(bool)
	case TextWriter:
		c.Writer = value.(Writer)
	}
}

// An Option interface sets options for HTML based renderers.
type Option interface {
	SetHTMLOption(*Config)
}

// TextWriter is an option name used in WithWriter.
const TextWriter renderer.OptionName = "Writer"

type withWriter struct {
	value Writer
}

func (o *withWriter) SetConfig(c *renderer.Config) {
	c.Options[TextWriter] = o.value
}

func (o *withWriter) SetHTMLOption(c *Config) {
	c.Writer = o.value
}

// WithWriter is a functional option that allow you to set given writer to
// the renderer.
func WithWriter(writer Writer) interface {
	renderer.Option
	Option
} {
	return &withWriter{writer}
}

// SoftLineBreak is an option name used in WithSoftLineBreak.
const SoftLineBreak renderer.OptionName = "SoftLineBreak"

type withSoftLineBreak struct {
}

func (o *withSoftLineBreak) SetConfig(c *renderer.Config) {
	c.Options[SoftLineBreak] = true
}

func (o *withSoftLineBreak) SetHTMLOption(c *Config) {
	c.SoftLineBreak = true
}

// WithSoftLineBreak is a functional option that indicates whether softline breaks
// should be rendered as '<br>'.
func WithSoftLineBreak() interface {
	renderer.Option
	Option
} {
	return &withSoftLineBreak{}
}

// XHTML is an option name used in WithXHTML.
const XHTML renderer.OptionName = "XHTML"

type withXHTML struct {
}

func (o *withXHTML) SetConfig(c *renderer.Config) {
	c.Options[XHTML] = true
}

func (o *withXHTML) SetHTMLOption(c *Config) {
	c.XHTML = true
}

// WithXHTML is a functional option indicates that nodes should be rendered in
// xhtml instead of HTML5.
func WithXHTML() interface {
	Option
	renderer.Option
} {
	return &withXHTML{}
}

// A Renderer struct is an implementation of renderer.NodeRenderer that renders
// nodes as (X)HTML.
type Renderer struct {
	Config
}

// NewRenderer returns a new Renderer with given options.
func NewRenderer(opts ...Option) renderer.NodeRenderer {
	r := &Renderer{
		Config: NewConfig(),
	}

	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// Render implements renderer.NodeRenderer.Render.
func (r *Renderer) Render(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	switch node := n.(type) {

	// blocks

	case *ast.Document:
		return r.renderDocument(writer, source, node, entering), nil
	case *ast.Heading:
		return r.renderHeading(writer, source, node, entering), nil
	case *ast.Blockquote:
		return r.renderBlockquote(writer, source, node, entering), nil
	case *ast.CodeBlock:
		return r.renderCodeBlock(writer, source, node, entering), nil
	case *ast.FencedCodeBlock:
		return r.renderFencedCodeBlock(writer, source, node, entering), nil
	case *ast.HTMLBlock:
		return r.renderHTMLBlock(writer, source, node, entering), nil
	case *ast.List:
		return r.renderList(writer, source, node, entering), nil
	case *ast.ListItem:
		return r.renderListItem(writer, source, node, entering), nil
	case *ast.Paragraph:
		return r.renderParagraph(writer, source, node, entering), nil
	case *ast.TextBlock:
		return r.renderTextBlock(writer, source, node, entering), nil
	case *ast.ThemanticBreak:
		return r.renderThemanticBreak(writer, source, node, entering), nil
	// inlines

	case *ast.AutoLink:
		return r.renderAutoLink(writer, source, node, entering), nil
	case *ast.CodeSpan:
		return r.renderCodeSpan(writer, source, node, entering), nil
	case *ast.Emphasis:
		return r.renderEmphasis(writer, source, node, entering), nil
	case *ast.Image:
		return r.renderImage(writer, source, node, entering), nil
	case *ast.Link:
		return r.renderLink(writer, source, node, entering), nil
	case *ast.RawHTML:
		return r.renderRawHTML(writer, source, node, entering), nil
	case *ast.Text:
		return r.renderText(writer, source, node, entering), nil
	}
	return ast.WalkContinue, renderer.NotSupported
}
func (r *Renderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		r.Writer.RawWrite(w, line.Value(source))
	}
}

func (r *Renderer) renderDocument(w util.BufWriter, source []byte, n *ast.Document, entering bool) ast.WalkStatus {
	// nothing to do
	return ast.WalkContinue
}

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, n *ast.Heading, entering bool) ast.WalkStatus {
	if entering {
		w.WriteString("<h")
		w.WriteByte("0123456"[n.Level])
		if n.ID != nil {
			w.WriteString(` id="`)
			w.Write(n.ID)
			w.WriteByte('"')
		}
		w.WriteByte('>')
	} else {
		w.WriteString("</h")
		w.WriteByte("0123456"[n.Level])
		w.WriteString(">\n")
	}
	return ast.WalkContinue
}

func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, n *ast.Blockquote, entering bool) ast.WalkStatus {
	if entering {
		w.WriteString("<blockquote>\n")
	} else {
		w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, n *ast.CodeBlock, entering bool) ast.WalkStatus {
	if entering {
		w.WriteString("<pre><code>")
		r.writeLines(w, source, n)
	} else {
		w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, n *ast.FencedCodeBlock, entering bool) ast.WalkStatus {
	if entering {
		w.WriteString("<pre><code")
		if n.Info != nil {
			segment := n.Info.Segment
			info := segment.Value(source)
			i := 0
			for ; i < len(info); i++ {
				if info[i] == ' ' {
					break
				}
			}
			language := info[:i]
			w.WriteString(" class=\"language-")
			r.Writer.Write(w, language)
			w.WriteString("\"")
		}
		w.WriteByte('>')
		r.writeLines(w, source, n)
	} else {
		w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue
}

func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, n *ast.HTMLBlock, entering bool) ast.WalkStatus {
	if entering {
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			w.Write(line.Value(source))
		}
	} else {
		if n.HasClosure() {
			closure := n.ClosureLine
			w.Write(closure.Value(source))
		}
	}
	return ast.WalkContinue
}

func (r *Renderer) renderList(w util.BufWriter, source []byte, n *ast.List, entering bool) ast.WalkStatus {
	tag := "ul"
	if n.IsOrdered() {
		tag = "ol"
	}
	if entering {
		w.WriteByte('<')
		w.WriteString(tag)
		if n.IsOrdered() && n.Start != 1 {
			fmt.Fprintf(w, " start=\"%d\">\n", n.Start)
		} else {
			w.WriteString(">\n")
		}
	} else {
		w.WriteString("</")
		w.WriteString(tag)
		w.WriteString(">\n")
	}
	return ast.WalkContinue
}

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, n *ast.ListItem, entering bool) ast.WalkStatus {
	if entering {
		w.WriteString("<li>")
		fc := n.FirstChild()
		if fc != nil {
			if _, ok := fc.(*ast.TextBlock); !ok {
				w.WriteByte('\n')
			}
		}
	} else {
		w.WriteString("</li>\n")
	}
	return ast.WalkContinue
}

func (r *Renderer) renderParagraph(w util.BufWriter, source []byte, n *ast.Paragraph, entering bool) ast.WalkStatus {
	if entering {
		w.WriteString("<p>")
	} else {
		w.WriteString("</p>\n")
	}
	return ast.WalkContinue
}

func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, n *ast.TextBlock, entering bool) ast.WalkStatus {
	if !entering {
		if _, ok := n.NextSibling().(ast.Node); ok && n.FirstChild() != nil {
			w.WriteByte('\n')
		}
		return ast.WalkContinue
	}
	return ast.WalkContinue
}

func (r *Renderer) renderThemanticBreak(w util.BufWriter, source []byte, n *ast.ThemanticBreak, entering bool) ast.WalkStatus {
	if !entering {
		return ast.WalkContinue
	}
	if r.XHTML {
		w.WriteString("<hr />\n")
	} else {
		w.WriteString("<hr>\n")
	}
	return ast.WalkContinue
}

func (r *Renderer) renderAutoLink(w util.BufWriter, source []byte, n *ast.AutoLink, entering bool) ast.WalkStatus {
	if !entering {
		return ast.WalkContinue
	}
	w.WriteString(`<a href="`)
	segment := n.Value.Segment
	value := segment.Value(source)
	if n.AutoLinkType == ast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(value), []byte("mailto:")) {
		w.WriteString("mailto:")
	}
	w.Write(util.EscapeHTML(util.URLEscape(value, false)))
	w.WriteString(`">`)
	w.Write(util.EscapeHTML(value))
	w.WriteString(`</a>`)
	return ast.WalkContinue
}

func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, n *ast.CodeSpan, entering bool) ast.WalkStatus {
	if entering {
		w.WriteString("<code>")
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				r.Writer.RawWrite(w, value[:len(value)-1])
				if c != n.LastChild() {
					r.Writer.RawWrite(w, []byte(" "))
				}
			} else {
				r.Writer.RawWrite(w, value)
			}
		}
		return ast.WalkSkipChildren
	}
	w.WriteString("</code>")
	return ast.WalkContinue
}

func (r *Renderer) renderEmphasis(w util.BufWriter, source []byte, n *ast.Emphasis, entering bool) ast.WalkStatus {
	tag := "em"
	if n.Level == 2 {
		tag = "strong"
	}
	if entering {
		w.WriteByte('<')
		w.WriteString(tag)
		w.WriteByte('>')
	} else {
		w.WriteString("</")
		w.WriteString(tag)
		w.WriteByte('>')
	}
	return ast.WalkContinue
}

func (r *Renderer) renderLink(w util.BufWriter, source []byte, n *ast.Link, entering bool) ast.WalkStatus {
	if entering {
		w.WriteString("<a href=\"")
		w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		w.WriteByte('"')
		if n.Title != nil {
			w.WriteString(` title="`)
			r.Writer.Write(w, n.Title)
			w.WriteByte('"')
		}
		w.WriteByte('>')
	} else {
		w.WriteString("</a>")
	}
	return ast.WalkContinue
}
func (r *Renderer) renderImage(w util.BufWriter, source []byte, n *ast.Image, entering bool) ast.WalkStatus {
	if !entering {
		return ast.WalkContinue
	}
	w.WriteString("<img src=\"")
	w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
	w.WriteString(`" alt="`)
	w.Write(n.Text(source))
	w.WriteByte('"')
	if n.Title != nil {
		w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		w.WriteByte('"')
	}
	if r.XHTML {
		w.WriteString(" />")
	} else {
		w.WriteString(">")
	}
	return ast.WalkSkipChildren
}

func (r *Renderer) renderRawHTML(w util.BufWriter, source []byte, n *ast.RawHTML, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, n *ast.Text, entering bool) ast.WalkStatus {
	if !entering {
		return ast.WalkContinue
	}
	segment := n.Segment
	if n.IsRaw() {
		w.Write(segment.Value(source))
	} else {
		r.Writer.Write(w, segment.Value(source))
		if n.HardLineBreak() || (n.SoftLineBreak() && r.SoftLineBreak) {
			if r.XHTML {
				w.WriteString("<br />\n")
			} else {
				w.WriteString("<br>\n")
			}
		} else if n.SoftLineBreak() {
			w.WriteByte('\n')
		}
	}
	return ast.WalkContinue
}

func readWhile(source []byte, index [2]int, pred func(byte) bool) (int, bool) {
	j := index[0]
	ok := false
	for ; j < index[1]; j++ {
		c1 := source[j]
		if pred(c1) {
			ok = true
			continue
		}
		break
	}
	return j, ok
}

// A Writer interface wirtes textual contents to a writer.
type Writer interface {
	// Write writes given source to writer with resolving references and unescaping
	// backslash escaped characters.
	Write(writer util.BufWriter, source []byte)

	// RawWrite wirtes given source to writer without resolving references and
	// unescaping backslash escaped characters.
	RawWrite(writer util.BufWriter, source []byte)
}

type defaultWriter struct {
}

var htmlEscaleTable = [256][]byte{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, []byte("&quot;"), nil, nil, nil, []byte("&amp;"), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, []byte("&lt;"), nil, []byte("&gt;"), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}

func escapeRune(writer util.BufWriter, r rune) {
	if r < 256 {
		v := htmlEscaleTable[byte(r)]
		if v != nil {
			writer.Write(v)
			return
		}
	}
	writer.WriteRune(util.ToValidRune(r))
}

func (d *defaultWriter) RawWrite(writer util.BufWriter, source []byte) {
	n := 0
	l := len(source)
	for i := 0; i < l; i++ {
		v := htmlEscaleTable[source[i]]
		if v != nil {
			writer.Write(source[i-n : i])
			n = 0
			writer.Write(v)
			continue
		}
		n++
	}
	if n != 0 {
		writer.Write(source[l-n:])
	}
}

func (d *defaultWriter) Write(writer util.BufWriter, source []byte) {
	escaped := false
	ok := false
	limit := len(source)
	n := 0
	for i := 0; i < limit; i++ {
		c := source[i]
		if escaped {
			if util.IsPunct(c) {
				d.RawWrite(writer, source[n:i-1])
				n = i
				escaped = false
				continue
			}
		}
		if c == '&' {
			pos := i
			next := i + 1
			if next < limit && source[next] == '#' {
				nnext := next + 1
				nc := source[nnext]
				// code point like #x22;
				if nnext < limit && nc == 'x' || nc == 'X' {
					start := nnext + 1
					i, ok = readWhile(source, [2]int{start, limit}, util.IsHexDecimal)
					if ok && i < limit && source[i] == ';' {
						v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 16, 32)
						d.RawWrite(writer, source[n:pos])
						n = i + 1
						escapeRune(writer, rune(v))
						continue
					}
					// code point like #1234;
				} else if nc >= '0' && nc <= '9' {
					start := nnext
					i, ok = readWhile(source, [2]int{start, limit}, util.IsNumeric)
					if ok && i < limit && i-start < 8 && source[i] == ';' {
						v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 0, 32)
						d.RawWrite(writer, source[n:pos])
						n = i + 1
						escapeRune(writer, rune(v))
						continue
					}
				}
			} else {
				start := next
				i, ok = readWhile(source, [2]int{start, limit}, util.IsAlphaNumeric)
				// entity reference
				if ok && i < limit && source[i] == ';' {
					name := util.BytesToReadOnlyString(source[start:i])
					entity, ok := util.LookUpHTML5EntityByName(name)
					if ok {
						d.RawWrite(writer, source[n:pos])
						n = i + 1
						d.RawWrite(writer, entity.Characters)
						continue
					}
				}
			}
			i = next - 1
		}
		if c == '\\' {
			escaped = true
			continue
		}
		escaped = false
	}
	d.RawWrite(writer, source[n:len(source)])
}

// DefaultWriter is a default implementation of the Writer.
var DefaultWriter = &defaultWriter{}