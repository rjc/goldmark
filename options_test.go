package goldmark

import (
	"github.com/yuin/goldmark/parser"
	"testing"
)

func TestAttributeAndAutoHeadingID(t *testing.T) {
	markdown := New(
		WithParserOptions(
			parser.WithAttribute(),
			parser.WithAutoHeadingID(),
		),
	)
	DoTestCaseFile(markdown, "_test/options.txt", t)
}
