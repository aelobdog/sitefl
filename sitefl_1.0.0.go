/*
	MIT License

	Copyright (c) 2020 Ashwin Godbole

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// List of States
const (
	NORMAL int = iota
	ESCAPEFORMATTING
	BOLD
	ITALICS
	UNDERLINE
	LINK
	ALTTEXT
	URL
	IMAGE
	CODE
)

var endings = map[int]byte{
	ESCAPEFORMATTING: '}',
	BOLD:             '*',
	ITALICS:          '/',
	UNDERLINE:        '_',
	ALTTEXT:          ']',
	URL:              ')',
	CODE:             '`',
}

var source string

var current = 0
var peeked = 1

//---------------------------------------------------------
//			LIST OF COMMAND LINE SETTABLE OPTIONS
var preserveNewlines = false

//---------------------------------------------------------

func next() {
	current++
	peeked++
}

func curr() byte {
	return source[current]
}

func peek() byte {
	return source[peeked]
}

func peek2() byte {
	return source[peeked+1]
}

func compile() string {
	var output bytes.Buffer
	ch := curr()
	cState := NORMAL
	for ; current < len(source); next() {
		ch = curr()
		switch cState {
		// ----------------------------------------------------------------------------------------
		case NORMAL:
			switch ch {
			case '{':
				cState = ESCAPEFORMATTING
			case '-':
				if peek() == '-' && peek2() == '-' {
					output.WriteString("\n<hr>\n")
					next()
					next()
				}
			case ';':
				if peek() == ';' {
					output.WriteString("\n<br>\n")
				}
				next()
			case '*':
				if cState == BOLD {
					cState = NORMAL
				} else {
					cState = BOLD
				}
			case '/':
				if cState == ITALICS {
					cState = NORMAL
				} else {
					cState = ITALICS
				}
			case '_':
				if cState == UNDERLINE {
					cState = NORMAL
				} else {
					cState = UNDERLINE
				}
			case '#':
				hNum := 1
				next()
				ch = curr()
				for ch == '#' {
					hNum++
					next()
					ch = curr()
					if hNum == 6 {
						break
					}
				}
				output.WriteString("<h" + strconv.Itoa(hNum) + ">")
				for ch != '\n' && current < len(source) {
					output.WriteByte(ch)
					next()
					if !(current < len(source)) {
						break
					}
					ch = curr()
					if ch == '\n' {
						break
					}
				}
				output.WriteString("</h" + strconv.Itoa(hNum) + ">")
			case '@':
				cState = LINK
			case '!':
				cState = IMAGE
			case '`':
				cState = CODE
			default:
				if ch == '\n' {
					if preserveNewlines {
						output.WriteString("\n<br>\n")
					}
				} else {
					output.WriteByte(ch)
				}
			}
		// ----------------------------------------------------------------------------------------
		case ESCAPEFORMATTING:
			for ch != endings[ESCAPEFORMATTING] && current < len(source) {
				ch = source[current]
				if ch == endings[ESCAPEFORMATTING] {
					break
				}
				output.WriteByte(ch)
				next()
			}
			cState = NORMAL
		// ----------------------------------------------------------------------------------------
		case BOLD:
			output.WriteString("<strong>")
			for ch != endings[BOLD] && current < len(source) {
				ch = source[current]
				if ch == endings[BOLD] {
					break
				}
				output.WriteByte(ch)
				next()
			}
			output.WriteString("</strong>")
			// next()
			cState = NORMAL
		// ----------------------------------------------------------------------------------------
		case ITALICS:
			output.WriteString("<em>")
			for ch != endings[ITALICS] && current < len(source) {
				ch = source[current]
				if ch == endings[ITALICS] {
					break
				}
				output.WriteByte(ch)
				next()
			}
			output.WriteString("</em>")
			// next()
			cState = NORMAL
		// ----------------------------------------------------------------------------------------
		case UNDERLINE:
			output.WriteString("<u>")
			for ch != endings[UNDERLINE] && current < len(source) {
				ch = source[current]
				if ch == endings[UNDERLINE] {
					break
				}
				output.WriteByte(ch)
				next()
			}
			output.WriteString("</u>")
			// next()
			cState = NORMAL
		// ----------------------------------------------------------------------------------------
		case LINK:
			//@[text](link)
			if ch == '[' {
				output.WriteString("\n<a href=\"")
				var url bytes.Buffer
				var alt bytes.Buffer
				next()
				ch = curr()
				for ch != endings[ALTTEXT] && current < len(source) {
					alt.WriteByte(ch)
					next()
					ch = curr()
				}
				next()
				ch = curr()
				if ch != '(' {
					fmt.Println("Improperly formatted link encountered.")
					fmt.Println(string(ch))
					os.Exit(1)
				}
				next()
				ch = curr()
				paropen := 0
				for ch != endings[URL] && current < len(source) {
					if ch == '(' {
						paropen++
					}
					url.WriteByte(ch)
					next()
					ch = curr()
					for ch == ')' && paropen != 0 {
						url.WriteByte(ch)
						next()
						ch = curr()
						paropen--
					}
				}
				next()
				ch = curr()
				cState = NORMAL
				output.WriteString(url.String())
				output.WriteString("\">")
				if alt.String() == "" {
					output.WriteString(url.String())
				} else {
					output.WriteString(alt.String())
				}
				output.WriteString("</a>\n")
			}
		// ----------------------------------------------------------------------------------------
		case IMAGE:
			//![text](link)
			if ch == '[' {
				output.WriteString("\n<img src=\"")
				var url bytes.Buffer
				var alt bytes.Buffer
				var w bytes.Buffer
				var h bytes.Buffer
				next()
				ch = curr()
				for ch != endings[ALTTEXT] && current < len(source) {
					if ch == ':' && peek() == ':' {
						next()
						next()
						ch = curr()
						for ch != ':' || peek() != ':' {
							w.WriteByte(ch)
							next()
							ch = curr()
						}
						next()
						next()
						ch = curr()
						for ch != endings[ALTTEXT] && current < len(source) {
							h.WriteByte(ch)
							next()
							ch = curr()
						}
						break
					}
					alt.WriteByte(ch)
					next()
					ch = curr()
				}
				next()
				ch = curr()
				if ch != '(' {
					fmt.Println("Improperly formatted image-link encountered.")
					fmt.Println(string(ch))
					os.Exit(1)
				}
				next()
				ch = curr()
				paropen := 0
				for ch != endings[URL] && current < len(source) {
					if ch == '(' {
						paropen++
					}
					url.WriteByte(ch)
					next()
					ch = curr()
					for ch == ')' && paropen != 0 {
						url.WriteByte(ch)
						next()
						ch = curr()
						paropen--
					}
				}
				next()
				ch = curr()
				cState = NORMAL
				output.WriteString(url.String())
				// output.WriteString("\">")
				output.WriteString("\" alt=\"")
				if alt.String() == "" {
					output.WriteString(url.String())
				} else {
					output.WriteString(alt.String())
				}
				output.WriteString("\"")
				if w.String() != "" {
					output.WriteString(" width=\"")
					output.WriteString(w.String())
					output.WriteByte('"')
				}
				if h.String() != "" {
					output.WriteString(" height=\"")
					output.WriteString(h.String())
					output.WriteByte('"')
				}
				output.WriteString(">\n")
			}
		// ----------------------------------------------------------------------------------------
		case CODE:
			output.WriteString("\n<pre>\n")
			for ch != endings[CODE] && current < len(source) {
				ch = source[current]
				if ch == endings[CODE] {
					break
				}
				output.WriteByte(ch)
				next()
			}
			output.WriteString("\n</pre>\n")
			// next()
			cState = NORMAL
		}
	}
	return output.String()
}

func writeToFile(content, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Could not open file %q, exiting.", filename)
		os.Exit(1)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		fmt.Printf("Could not write to file %q, exiting.", filename)
	}
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(0)
	}
	optionGiven := false
	if os.Args[1] == "-n" {
		preserveNewlines = true
		optionGiven = true
	} else if os.Args[1] == "-h" {
		usage()
		os.Exit(0)
	} else if os.Args[1][0] == '-' {
		fmt.Printf("Unknown option: %q", os.Args[1])
		os.Exit(0)
	}
	if optionGiven {
		if len(os.Args) != 4 {
			usage()
			os.Exit(0)
		}
	}
	src := 2
	dst := 3
	if !optionGiven {
		if len(os.Args) != 3 {
			usage()
			os.Exit(0)
		}
		src--
		dst--
	}

	if os.Args[src] == "in" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			source += scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	} else {
		scode, err := ioutil.ReadFile(os.Args[src])
		if err != nil {
			fmt.Printf("Could not read file named %q", os.Args[src])
			return
		}
		source = string(scode)
	}
	out := compile()

	if os.Args[dst] == "out" {
		fmt.Println(out)
	} else {
		writeToFile(out, os.Args[dst])
	}
}

func usage() {
	fmt.Println(`
		USAGE:
		------
			sitefl [-n -h] source destination

		OPTIONS:
		--------
			-f : Prints this message
			-n : Preserves new lines

		SOURCE:
		-------
			filename : file to obtain input from
			'in' : get input from stdin
					-> allows user to pipe output from another program into this program
					-> grep -o "something.*something" | sitefl in DESTINATION
					
		DESTINATION:
		------------
			filename : file to send output to
			'out' : send output to stdout
					-> allows the output fo this program to be piped into another program
					-> sitefl SOURCE out | grep "<strong>.*</strong>"
	`)
}
