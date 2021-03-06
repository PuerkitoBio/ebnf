// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ebnf

import (
	"bytes"
	"strings"
	"testing"
)

var goodGrammars = []string{
	`Program = .`,

	`Program = foo .
	 foo = "foo" .`,

	`Program = Foo .
	 Foo = /foo/ | Bar .
	 Bar = "z" .`,

	`Program = "a" | "b" "c" .`,

	`Program = "a" … "z" .`,
	`Program = "(" … ")" .`,

	`Program = /[a-x]/ .`,

	`Program = Song .
	 Song = { Note } .
	 Note = Do | (Re | Mi | Fa | So | La) | Ti .
	 Do = "c" .
	 Re = "d" .
	 Mi = "e" .
	 Fa = "f" .
	 So = /g|G/ .
	 La = "a" .
	 Ti = ti .
	 ti = "b" .`,

	`Program = Start | Empty .
	  Start = ( Token [ Token "," Token ] ) .
	  Token = empty .
	  Empty = . 
	  empty = . // with a trailing comment`,
}

var badGrammars = []string{
	`Program = | .`,
	`Program = | b .`,
	`Program = a … b .`,
	`Program = "a" … .`,
	`Program = "a" … /\n/ .`,
	`Program = … "b" .`,
	`Program = () .`,
	`Program = [] .`,
	`Program = {} .`,
}

var invalidGrammars = []string{
	`Program = /\pzz/ .`,
}

func checkGood(t *testing.T, src string) {
	grammar, err := Parse("", bytes.NewBuffer([]byte(src)))
	if err != nil {
		t.Errorf("Parse(%s) failed: %v", src, err)
		return
	}
	if err = Verify(grammar, "Program"); err != nil {
		t.Errorf("Verify(%s) failed: %v", src, err)
	}
	for k, v := range grammar {
		if strings.ToLower(k) == "empty" {
			if v.Expr != nil {
				t.Errorf("Non-empty production named `empty`: %v", v)
			}
		}
	}
}

func checkBad(t *testing.T, src string) {
	_, err := Parse("", bytes.NewBuffer([]byte(src)))
	if err == nil {
		t.Errorf("Parse(%s) should have failed", src)
	}
	t.Log(err)
}

func checkInvalid(t *testing.T, src string) {
	grammar, err := Parse("", bytes.NewBuffer([]byte(src)))
	if err != nil {
		t.Errorf("Parse(%s) failed: %v", src, err)
		return
	}
	if err = Verify(grammar, "Program"); err == nil {
		t.Errorf("Verify(%s) succeeded, wanted error", src)
	}
}

func TestGrammars(t *testing.T) {
	for _, src := range goodGrammars {
		checkGood(t, src)
	}
	for _, src := range badGrammars {
		checkBad(t, src)
	}
	for _, src := range invalidGrammars {
		checkInvalid(t, src)
	}
}
