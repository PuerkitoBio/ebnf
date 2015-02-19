// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scanner

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode/utf8"
)

// A StringReader delivers its data one string segment at a time via Read.
type StringReader struct {
	data []string
	step int
}

func (r *StringReader) Read(p []byte) (n int, err error) {
	if r.step < len(r.data) {
		s := r.data[r.step]
		n = copy(p, s)
		r.step++
	} else {
		err = io.EOF
	}
	return
}

func readRuneSegments(t *testing.T, segments []string) {
	got := ""
	want := strings.Join(segments, "")
	s := new(Scanner).Init(&StringReader{data: segments})
	for {
		ch := s.Next()
		if ch == EOF {
			break
		}
		got += string(ch)
	}
	if got != want {
		t.Errorf("segments=%v got=%s want=%s", segments, got, want)
	}
}

var segmentList = [][]string{
	{},
	{""},
	{"日", "本語"},
	{"\u65e5", "\u672c", "\u8a9e"},
	{"\U000065e5", " ", "\U0000672c", "\U00008a9e"},
	{"\xe6", "\x97\xa5\xe6", "\x9c\xac\xe8\xaa\x9e"},
	{"Hello", ", ", "World", "!"},
	{"Hello", ", ", "", "World", "!"},
}

func TestRegexpErrors(t *testing.T) {
	data := `/non-terminated
	// comment
	/eof`

	s := new(Scanner).Init(strings.NewReader(data))
	want := "regexp literal not terminated"
	cnt := 0
	s.Error = func(s *Scanner, msg string) {
		if msg != want {
			t.Errorf("want %q, got %q", want, msg)
		}
		cnt++
	}
	for tok := s.Scan(); tok != EOF; tok = s.Scan() {
	}
	if cnt != 2 {
		t.Errorf("want %d errors, got %d", 2, cnt)
	}
}

func TestRegexp(t *testing.T) {
	cases := []string{
		`/a/`,
		`/ab/`,
		`/abc/`,
		`/\//`,
		`/[a-zA-Z]+[0-9]*/`,
		`/[a-zA-Z]+\n?/`,
		`/[a-zA-Z]+\n?\/\\/`,
	}
	for i, c := range cases {
		if i != 3 {
			continue
		}
		s := new(Scanner).Init(strings.NewReader(c))
		if tok := s.Scan(); tok != Regexp {
			t.Errorf("%q: want token %q, got %q", c, TokenString(Regexp), TokenString(tok))
		}
		if text := s.TokenText(); text != c {
			t.Errorf("want %q, got %q", c, text)
		}
		if n := s.ErrorCount; n > 0 {
			t.Errorf("%d errors", n)
		}
	}
}

func TestNext(t *testing.T) {
	for _, s := range segmentList {
		readRuneSegments(t, s)
	}
}

type token struct {
	tok  rune
	text string
}

var f100 = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

var tokenList = []token{
	{Comment, "// line comments"},
	{Comment, "//"},
	{Comment, "////"},
	{Comment, "// comment"},
	{Comment, "// /* comment */"},
	{Comment, "// // comment //"},
	{Comment, "//" + f100},

	{Comment, "// general comments"},
	{Comment, "/**/"},
	{Comment, "/***/"},
	{Comment, "/* comment */"},
	{Comment, "/* // comment */"},
	{Comment, "/* /* comment */"},
	{Comment, "/*\n comment\n*/"},
	{Comment, "/*" + f100 + "*/"},

	{Comment, "// identifiers"},
	{Ident, "a"},
	{Ident, "a0"},
	{Ident, "foobar"},
	{Ident, "abc123"},
	{Ident, "LGTM"},
	{Ident, "_"},
	{Ident, "_abc123"},
	{Ident, "abc123_"},
	{Ident, "_abc_123_"},
	{Ident, "_äöü"},
	{Ident, "_本"},
	{Ident, "äöü"},
	{Ident, "本"},
	{Ident, "a۰۱۸"},
	{Ident, "foo६४"},
	{Ident, "bar９８７６"},
	{Ident, f100},

	{Comment, "// decimal ints"},
	{Int, "0"},
	{Int, "1"},
	{Int, "9"},
	{Int, "42"},
	{Int, "1234567890"},

	{Comment, "// octal ints"},
	{Int, "00"},
	{Int, "01"},
	{Int, "07"},
	{Int, "042"},
	{Int, "01234567"},

	{Comment, "// hexadecimal ints"},
	{Int, "0x0"},
	{Int, "0x1"},
	{Int, "0xf"},
	{Int, "0x42"},
	{Int, "0x123456789abcDEF"},
	{Int, "0x" + f100},
	{Int, "0X0"},
	{Int, "0X1"},
	{Int, "0XF"},
	{Int, "0X42"},
	{Int, "0X123456789abcDEF"},
	{Int, "0X" + f100},

	{Comment, "// floats"},
	{Float, "0."},
	{Float, "1."},
	{Float, "42."},
	{Float, "01234567890."},
	{Float, ".0"},
	{Float, ".1"},
	{Float, ".42"},
	{Float, ".0123456789"},
	{Float, "0.0"},
	{Float, "1.0"},
	{Float, "42.0"},
	{Float, "01234567890.0"},
	{Float, "0e0"},
	{Float, "1e0"},
	{Float, "42e0"},
	{Float, "01234567890e0"},
	{Float, "0E0"},
	{Float, "1E0"},
	{Float, "42E0"},
	{Float, "01234567890E0"},
	{Float, "0e+10"},
	{Float, "1e-10"},
	{Float, "42e+10"},
	{Float, "01234567890e-10"},
	{Float, "0E+10"},
	{Float, "1E-10"},
	{Float, "42E+10"},
	{Float, "01234567890E-10"},

	{Comment, "// chars"},
	{Char, `' '`},
	{Char, `'a'`},
	{Char, `'本'`},
	{Char, `'\a'`},
	{Char, `'\b'`},
	{Char, `'\f'`},
	{Char, `'\n'`},
	{Char, `'\r'`},
	{Char, `'\t'`},
	{Char, `'\v'`},
	{Char, `'\''`},
	{Char, `'\000'`},
	{Char, `'\777'`},
	{Char, `'\x00'`},
	{Char, `'\xff'`},
	{Char, `'\u0000'`},
	{Char, `'\ufA16'`},
	{Char, `'\U00000000'`},
	{Char, `'\U0000ffAB'`},

	{Comment, "// strings"},
	{String, `" "`},
	{String, `"a"`},
	{String, `"本"`},
	{String, `"\a"`},
	{String, `"\b"`},
	{String, `"\f"`},
	{String, `"\n"`},
	{String, `"\r"`},
	{String, `"\t"`},
	{String, `"\v"`},
	{String, `"\""`},
	{String, `"\000"`},
	{String, `"\777"`},
	{String, `"\x00"`},
	{String, `"\xff"`},
	{String, `"\u0000"`},
	{String, `"\ufA16"`},
	{String, `"\U00000000"`},
	{String, `"\U0000ffAB"`},
	{String, `"` + f100 + `"`},

	{Comment, "// raw strings"},
	{String, "``"},
	{String, "`\\`"},
	{String, "`" + "\n\n/* foobar */\n\n" + "`"},
	{String, "`" + f100 + "`"},

	{Comment, "// individual characters"},
	// NUL character is not allowed
	{'\x01', "\x01"},
	{' ' - 1, string(' ' - 1)},
	{'+', "+"},
	{'\\', "\\"},
	{'.', "."},
	{'~', "~"},
	{'(', "("},

	{Comment, "// regular expression literals"},
	{Regexp, `/a/`},
}

func makeSource(pattern string) *bytes.Buffer {
	var buf bytes.Buffer
	for _, k := range tokenList {
		fmt.Fprintf(&buf, pattern, k.text)
	}
	return &buf
}

func checkTok(t *testing.T, s *Scanner, line int, got, want rune, text string) {
	if got != want {
		t.Fatalf("tok = %s, want %s for %q", TokenString(got), TokenString(want), text)
	}
	if s.Line != line {
		t.Errorf("line = %d, want %d for %q", s.Line, line, text)
	}
	stext := s.TokenText()
	if stext != text {
		t.Errorf("text = %q, want %q", stext, text)
	} else {
		// check idempotency of TokenText() call
		stext = s.TokenText()
		if stext != text {
			t.Errorf("text = %q, want %q (idempotency check)", stext, text)
		}
	}
}

func countNewlines(s string) int {
	n := 0
	for _, ch := range s {
		if ch == '\n' {
			n++
		}
	}
	return n
}

func testScan(t *testing.T) {
	s := new(Scanner).Init(makeSource(" \t%s\n"))
	tok := s.Scan()
	line := 1
	for _, k := range tokenList {
		if k.tok != Comment {
			checkTok(t, s, line, tok, k.tok, k.text)
			tok = s.Scan()
		}
		line += countNewlines(k.text) + 1 // each token is on a new line
	}
	checkTok(t, s, line, tok, EOF, "")
}

func TestScan(t *testing.T) {
	testScan(t)
}

func TestPosition(t *testing.T) {
	src := makeSource("\t\t\t\t%s\n")
	s := new(Scanner).Init(src)
	s.Scan()
	pos := Position{"", 4, 1, 5}
	for _, k := range tokenList {
		if k.tok != Comment {
			if s.Offset != pos.Offset {
				t.Errorf("offset = %d, want %d for %q", s.Offset, pos.Offset, k.text)
			}
			if s.Line != pos.Line {
				t.Errorf("line = %d, want %d for %q", s.Line, pos.Line, k.text)
			}
			if s.Column != pos.Column {
				t.Errorf("column = %d, want %d for %q", s.Column, pos.Column, k.text)
			}
		}
		pos.Offset += 4 + len(k.text) + 1     // 4 tabs + token bytes + newline
		pos.Line += countNewlines(k.text) + 1 // each token is on a new line
		if k.tok != Comment {
			s.Scan()
		}
	}
	// make sure there were no token-internal errors reported by scanner
	if s.ErrorCount != 0 {
		t.Errorf("%d errors", s.ErrorCount)
	}
}

func TestScanNext(t *testing.T) {
	const BOM = '\uFEFF'
	BOMs := string(BOM)
	s := new(Scanner).Init(strings.NewReader(BOMs + "if a == bcd /* com" + BOMs + "ment */ {\n\ta += c\n}" + BOMs + "// line comment ending in eof"))
	checkTok(t, s, 1, s.Scan(), Ident, "if") // the first BOM is ignored
	checkTok(t, s, 1, s.Scan(), Ident, "a")
	checkTok(t, s, 1, s.Scan(), '=', "=")
	checkTok(t, s, 0, s.Next(), '=', "")
	checkTok(t, s, 0, s.Next(), ' ', "")
	checkTok(t, s, 0, s.Next(), 'b', "")
	checkTok(t, s, 1, s.Scan(), Ident, "cd")
	checkTok(t, s, 1, s.Scan(), '{', "{")
	checkTok(t, s, 2, s.Scan(), Ident, "a")
	checkTok(t, s, 2, s.Scan(), '+', "+")
	checkTok(t, s, 0, s.Next(), '=', "")
	checkTok(t, s, 2, s.Scan(), Ident, "c")
	checkTok(t, s, 3, s.Scan(), '}', "}")
	checkTok(t, s, 3, s.Scan(), BOM, BOMs)
	checkTok(t, s, 3, s.Scan(), -1, "")
	if s.ErrorCount != 0 {
		t.Errorf("%d errors", s.ErrorCount)
	}
}

func testError(t *testing.T, src, pos, msg string, tok rune) {
	s := new(Scanner).Init(strings.NewReader(src))
	errorCalled := false
	s.Error = func(s *Scanner, m string) {
		if !errorCalled {
			// only look at first error
			if p := s.Pos().String(); p != pos {
				t.Errorf("pos = %q, want %q for %q", p, pos, src)
			}
			if m != msg {
				t.Errorf("msg = %q, want %q for %q", m, msg, src)
			}
			errorCalled = true
		}
	}
	tk := s.Scan()
	if tk != tok {
		t.Errorf("tok = %s, want %s for %q", TokenString(tk), TokenString(tok), src)
	}
	if !errorCalled {
		t.Errorf("error handler not called for %q", src)
	}
	if s.ErrorCount == 0 {
		t.Errorf("count = %d, want > 0 for %q", s.ErrorCount, src)
	}
}

func TestError(t *testing.T) {
	testError(t, "\x00", "1:1", "illegal character NUL", 0)
	testError(t, "\x80", "1:1", "illegal UTF-8 encoding", utf8.RuneError)
	testError(t, "\xff", "1:1", "illegal UTF-8 encoding", utf8.RuneError)

	testError(t, "a\x00", "1:2", "illegal character NUL", Ident)
	testError(t, "ab\x80", "1:3", "illegal UTF-8 encoding", Ident)
	testError(t, "abc\xff", "1:4", "illegal UTF-8 encoding", Ident)

	testError(t, `"a`+"\x00", "1:3", "illegal character NUL", String)
	testError(t, `"ab`+"\x80", "1:4", "illegal UTF-8 encoding", String)
	testError(t, `"abc`+"\xff", "1:5", "illegal UTF-8 encoding", String)

	testError(t, "`a"+"\x00", "1:3", "illegal character NUL", String)
	testError(t, "`ab"+"\x80", "1:4", "illegal UTF-8 encoding", String)
	testError(t, "`abc"+"\xff", "1:5", "illegal UTF-8 encoding", String)

	testError(t, `'\"'`, "1:3", "illegal char escape", Char)
	testError(t, `"\'"`, "1:3", "illegal char escape", String)

	testError(t, `01238`, "1:6", "illegal octal number", Int)
	testError(t, `01238123`, "1:9", "illegal octal number", Int)
	testError(t, `0x`, "1:3", "illegal hexadecimal number", Int)
	testError(t, `0xg`, "1:3", "illegal hexadecimal number", Int)
	testError(t, `'aa'`, "1:4", "illegal char literal", Char)

	testError(t, `'`, "1:2", "literal not terminated", Char)
	testError(t, `'`+"\n", "1:2", "literal not terminated", Char)
	testError(t, `"abc`, "1:5", "literal not terminated", String)
	testError(t, `"abc`+"\n", "1:5", "literal not terminated", String)
	testError(t, "`abc\n", "2:1", "literal not terminated", String)
	testError(t, `/*/`, "1:4", "comment not terminated", EOF)
}

// An errReader returns (0, err) where err is not io.EOF.
type errReader struct{}

func (errReader) Read(b []byte) (int, error) {
	return 0, io.ErrNoProgress // some error that is not io.EOF
}

func TestIOError(t *testing.T) {
	s := new(Scanner).Init(errReader{})
	errorCalled := false
	s.Error = func(s *Scanner, msg string) {
		if !errorCalled {
			if want := io.ErrNoProgress.Error(); msg != want {
				t.Errorf("msg = %q, want %q", msg, want)
			}
			errorCalled = true
		}
	}
	tok := s.Scan()
	if tok != EOF {
		t.Errorf("tok = %s, want EOF", TokenString(tok))
	}
	if !errorCalled {
		t.Errorf("error handler not called")
	}
}

func checkPos(t *testing.T, got, want Position) {
	if got.Offset != want.Offset || got.Line != want.Line || got.Column != want.Column {
		t.Errorf("got offset, line, column = %d, %d, %d; want %d, %d, %d",
			got.Offset, got.Line, got.Column, want.Offset, want.Line, want.Column)
	}
}

func checkNextPos(t *testing.T, s *Scanner, offset, line, column int, char rune) {
	if ch := s.Next(); ch != char {
		t.Errorf("ch = %s, want %s", TokenString(ch), TokenString(char))
	}
	want := Position{Offset: offset, Line: line, Column: column}
	checkPos(t, s.Pos(), want)
}

func checkScanPos(t *testing.T, s *Scanner, offset, line, column int, char rune) {
	want := Position{Offset: offset, Line: line, Column: column}
	checkPos(t, s.Pos(), want)
	if ch := s.Scan(); ch != char {
		t.Errorf("ch = %s, want %s", TokenString(ch), TokenString(char))
		if string(ch) != s.TokenText() {
			t.Errorf("tok = %q, want %q", s.TokenText(), string(ch))
		}
	}
	checkPos(t, s.Position, want)
}

func TestPos(t *testing.T) {
	// corner case: empty source
	s := new(Scanner).Init(strings.NewReader(""))
	checkPos(t, s.Pos(), Position{Offset: 0, Line: 1, Column: 1})
	s.Peek() // peek doesn't affect the position
	checkPos(t, s.Pos(), Position{Offset: 0, Line: 1, Column: 1})

	// corner case: source with only a newline
	s = new(Scanner).Init(strings.NewReader("\n"))
	checkPos(t, s.Pos(), Position{Offset: 0, Line: 1, Column: 1})
	checkNextPos(t, s, 1, 2, 1, '\n')
	// after EOF position doesn't change
	for i := 10; i > 0; i-- {
		checkScanPos(t, s, 1, 2, 1, EOF)
	}
	if s.ErrorCount != 0 {
		t.Errorf("%d errors", s.ErrorCount)
	}

	// corner case: source with only a single character
	s = new(Scanner).Init(strings.NewReader("本"))
	checkPos(t, s.Pos(), Position{Offset: 0, Line: 1, Column: 1})
	checkNextPos(t, s, 3, 1, 2, '本')
	// after EOF position doesn't change
	for i := 10; i > 0; i-- {
		checkScanPos(t, s, 3, 1, 2, EOF)
	}
	if s.ErrorCount != 0 {
		t.Errorf("%d errors", s.ErrorCount)
	}

	// positions after calling Next
	s = new(Scanner).Init(strings.NewReader("  foo६४  \n\n本語\n"))
	checkNextPos(t, s, 1, 1, 2, ' ')
	s.Peek() // peek doesn't affect the position
	checkNextPos(t, s, 2, 1, 3, ' ')
	checkNextPos(t, s, 3, 1, 4, 'f')
	checkNextPos(t, s, 4, 1, 5, 'o')
	checkNextPos(t, s, 5, 1, 6, 'o')
	checkNextPos(t, s, 8, 1, 7, '६')
	checkNextPos(t, s, 11, 1, 8, '४')
	checkNextPos(t, s, 12, 1, 9, ' ')
	checkNextPos(t, s, 13, 1, 10, ' ')
	checkNextPos(t, s, 14, 2, 1, '\n')
	checkNextPos(t, s, 15, 3, 1, '\n')
	checkNextPos(t, s, 18, 3, 2, '本')
	checkNextPos(t, s, 21, 3, 3, '語')
	checkNextPos(t, s, 22, 4, 1, '\n')
	// after EOF position doesn't change
	for i := 10; i > 0; i-- {
		checkScanPos(t, s, 22, 4, 1, EOF)
	}
	if s.ErrorCount != 0 {
		t.Errorf("%d errors", s.ErrorCount)
	}
}
