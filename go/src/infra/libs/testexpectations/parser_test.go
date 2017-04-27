package testexpectations

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Parser", t, func() {
		tests := []struct {
			input    string
			expected *ExpectationStatement
			err      error
		}{
			{
				"",
				&ExpectationStatement{},
				nil,
			},
			{
				"# This is a comment.",
				&ExpectationStatement{
					Comment: "# This is a comment.",
				},
				nil,
			},
			// TODO(seanmccullough): Verify that this is a valid case. The docs say the linter
			// will complain if the line doesn't specify one or more bugs.
			/*
				{
					"fast/html/keygen.html [ Skip ]",
					&ExpectationStatement{
						TestName:     "fast/html/keygen.html",
						Expectations: []string{"Skip"},
					},
					nil,
				},
			*/
			{
				"crbug.com/12345 fast/html/keygen.html [ Crash ]",
				&ExpectationStatement{
					Bugs:         "crbug.com/12345",
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Crash"},
				},
				nil,
			},
			{
				"crbug.com/12345 fast/html/keygen.html [ Crash Pass ]",
				&ExpectationStatement{
					Bugs:         "crbug.com/12345",
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Crash", "Pass"},
				},
				nil,
			},
			{
				"crbug.com/12345 [ Win Debug ] fast/html/keygen.html [ Crash ]",
				&ExpectationStatement{
					Bugs:         "crbug.com/12345",
					Modifiers:    []string{"Win", "Debug"},
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Crash"},
				},
				nil,
			},
			{
				"crbug.com/12345 [ Win Debug ] fast/html/keygen.html [ Crash Pass ]",
				&ExpectationStatement{
					Bugs:         "crbug.com/12345",
					Modifiers:    []string{"Win", "Debug"},
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Crash", "Pass"},
				},
				nil,
			},
			{
				"Bug(darin) [ Mac10.9 Debug ] fast/html/keygen.html [ Skip ]",
				&ExpectationStatement{
					Bugs:         "Bug(darin)",
					Modifiers:    []string{"Mac10.9", "Debug"},
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Skip"},
				},
				nil,
			},
		}

		for _, test := range tests {
			p := NewParser(bytes.NewBufferString(test.input))
			stmt, err := p.Parse()
			So(err, ShouldEqual, test.err)
			So(stmt, ShouldResemble, test.expected)
		}
	})
}
