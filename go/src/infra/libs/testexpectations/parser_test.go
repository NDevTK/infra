package testexpectations

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
					Bugs:         []string{"crbug.com/12345"},
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Crash"},
				},
				nil,
			},
			{
				"crbug.com/12345 fast/html/keygen.html [ Crash Pass ]",
				&ExpectationStatement{
					Bugs:         []string{"crbug.com/12345"},
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Crash", "Pass"},
				},
				nil,
			},
			{
				"crbug.com/12345 [ Win Debug ] fast/html/keygen.html [ Crash ]",
				&ExpectationStatement{
					Bugs:         []string{"crbug.com/12345"},
					Modifiers:    []string{"Win", "Debug"},
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Crash"},
				},
				nil,
			},
			{
				"crbug.com/12345 [ Win Debug ] fast/html/keygen.html [ Crash Pass ]",
				&ExpectationStatement{
					Bugs:         []string{"crbug.com/12345"},
					Modifiers:    []string{"Win", "Debug"},
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Crash", "Pass"},
				},
				nil,
			},
			{
				"Bug(darin) [ Mac10.9 Debug ] fast/html/keygen.html [ Skip ]",
				&ExpectationStatement{
					Bugs:         []string{"Bug(darin)"},
					Modifiers:    []string{"Mac10.9", "Debug"},
					TestName:     "fast/html/keygen.html",
					Expectations: []string{"Skip"},
				},
				nil,
			},
			{
				"crbug.com/504613 crbug.com/524248 paint/images/image-backgrounds-not-antialiased.html [ Failure ]",
				&ExpectationStatement{
					Bugs:         []string{"crbug.com/504613", "crbug.com/524248"},
					TestName:     "paint/images/image-backgrounds-not-antialiased.html",
					Expectations: []string{"Failure"},
				},
				nil,
			},
			{
				"crbug.com/504613 crbug.com/524248 [ Mac Win ] paint/images/image-backgrounds-not-antialiased.html [ Failure ]",
				&ExpectationStatement{
					Bugs:         []string{"crbug.com/504613", "crbug.com/524248"},
					Modifiers:    []string{"Mac", "Win"},
					TestName:     "paint/images/image-backgrounds-not-antialiased.html",
					Expectations: []string{"Failure"},
				},
				nil,
			},
			{
				"not a valid input line",
				nil,
				fmt.Errorf(`expected LB or IDENT for expectations, but found "valid"`),
			},
		}

		for _, test := range tests {
			p := NewParser(bytes.NewBufferString(test.input))
			stmt, err := p.Parse()
			So(err, ShouldResemble, test.err)
			So(stmt, ShouldResemble, test.expected)

			if test.err != nil {
				continue
			}

			// And test round-trip back into a string.
			So(stmt.String(), ShouldEqual, test.input)
		}
	})
}

func ExampleParser_Parse() {
	URL := "https://chromium.googlesource.com/chromium/src/+/master/third_party/WebKit/LayoutTests/TestExpectations?format=TEXT"

	resp, err := http.Get(URL)
	if err != nil {
		fmt.Printf("Error fetching: %s\n", err)
		return
	}
	defer resp.Body.Close()

	reader := base64.NewDecoder(base64.StdEncoding, resp.Body)
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Printf("Error reading: %s\n", err)
		return
	}

	lines := strings.Split(string(b), "\n")
	stmts := []*ExpectationStatement{}
	for n, line := range lines {
		p := NewParser(bytes.NewBufferString(line))
		stmt, err := p.Parse()
		if err != nil {
			fmt.Printf("Error parsing line %d %q: %s\n", n, line, err)
			return
		}
		stmt.LineNumber = n
		stmt.Original = line
		stmts = append(stmts, stmt)
	}

	fmt.Printf("line count match? %t\n", len(stmts) == len(lines))

	for _, s := range stmts {
		r := s.String()
		if s.Original != r {
			fmt.Printf("%d differs:\n%q\n%q\n", s.LineNumber, s.Original, r)
		}
	}

	// TODO(seanmccullough): Track extra whitespace between test names and
	// expectations in the original lines, or otherwise keep the original text
	// if we haven't edited the semantics of the line.

	// /* Output: */
	// line counts match? true
}
