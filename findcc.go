/*
 * findcc.go
 * Small program to find an n-digit number with a mod-10 checksum in a file
 * by J. Stuart McMurray
 * created 20150115
 * last modified 20150115
 * Copyright (c) 2014 J. Stuart McMurray <kd5pbo@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */
package main

import (
	"flag"
	"fmt"
	"github.com/joeljunstrom/go-luhn"
	"io"
	"os"
	"unicode"
)

/* Usage statement */

func main() { os.Exit(mymain()) }
func mymain() int {
	/* Get the number of digits in the number on the command line */
	numlen := flag.Int("n", 16, "Length of number to find, including "+
		"the check digit.")
	mod10 := flag.Bool("mod10", false, "Use a simple sum modulus 10 "+
		"instead of the Luhn algorithm.")
	quiet := flag.Bool("q", false, "Be quiet; don't print the header.")
	/* Usage statement */
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-q] [-n NN] [filename]",
			os.Args[0])
		fmt.Fprintf(os.Stderr, `

Search for sequences of a set number of ascii digits (controllable by -n) that
either passes validation with the Luhn algorithm or has a final digit that is
equal to the modulus 10 sum of the other digits (with -mod10).  If no filename
is given, the standard input is used.  The offset in the file and line number
where the number was found, as well as the number with its check digit are
printed in a tabular format, separated by whitespace.

Options:
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	/* Work out where to get input */
	input := os.Stdin /* Default to stdin */
	/* Open a file if specified */
	if 1 == flag.NArg() {
		var err error
		if input, err = os.Open(flag.Arg(0)); nil != err {
			fmt.Fprintf(os.Stderr, "Unable to open %v: %v",
				flag.Arg(0), err)
			return -1
		}
	} else if 1 < flag.NArg() {
		fmt.Fprintf(os.Stderr, "Multiple input files are not "+
			"supported.\n")
		return -2
	}

	/* Print the header if we're not quiet */
	if !*quiet {
		fmt.Printf("OFFSET  LINE  NUMBER\n")
	}

	digits := []byte{}  /* Slice to buffer sequential input digits */
	buf := []byte{0x00} /* Read buffer */
	nline := 0          /* Number of newlines read */
	nread := 0          /* Number of bytes read */
	/* Read until EOF */
	for {
		/* Read a byte */
		if n, err := input.Read(buf); nil != err {
			/* Don't whine if we've reached EOF */
			if io.EOF == err {
				return 0
			}
			/* Print any other errors, though */
			fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
			return -3
		} else if 0 == n && nil == err {
			/* Didn't read anything, but no error?  Probably a bug
			somewhere else. */
			fmt.Fprintf(os.Stderr, "Didn't read anything, but "+
				"no error detected.  This shouldn't happen.")
			return -4
		}
		/* Note how many bytes we've read */
		nread++
		/* Note if it's a newline */
		if '\n' == buf[0] {
			nline++
		}
		/* If it's not a digit, clear any waiting digits, try again */
		if !unicode.IsDigit(rune(buf[0])) {
			if 0 < len(digits) {
				digits = []byte{}
			}
			continue
		}
		/* Update the digit buffer with the new digit */
		digits = append(digits, buf...)
		for len(digits) > *numlen { /* Should only loop once */
			digits = digits[1:]
		}
		/* If we have enough, report it if it's a valid checksum */
		if (len(digits) == *numlen) &&
			((*mod10 && mod10Valid(digits)) ||
				(!*mod10 && luhn.Valid(string(digits)))) {
			fmt.Printf("%6v  %4v  %v\n",
				nread-len(digits)-1,
				nline,
				string(digits))

		}
	}
	/* Should never get here */
	fmt.Fprintf(os.Stderr, "Unpossible code execution.  Please debug.\n")
	return -5
}

/* mod10Valid tests whether the input byte array is valid, according to the
help output for -mod10 */
func mod10Valid(digits []byte) bool {
	exp := 0 /* Expected checksum */
	/* Calculate the expected checksum */
	for _, d := range digits[:len(digits)-1] {
		exp = (exp + (int(d) - '0')) % 10
	}
	/* Print the match if we have it */
	return int(digits[len(digits)-1]-'0') == exp
}
