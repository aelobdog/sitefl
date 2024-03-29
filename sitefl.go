/*
	MIT License

	Copyright (c) 2020-2022 Ashwin Godbole

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
	"strings"
)

// States
const (
	normal int = iota
	escape
	bold
	italics
	underline
	link
	alttext
	url
	image
	code
	heading
)

// map of names to characters, for flexibility
var blockChars = map[string]byte{
	"esc":         '\\',
	"escapeBegin": '{',
	"escapeEnd":   '}',
	"bold":        '*',
	"italics":     '/',
	"underline":   '_',
	"line":        '-',
	"newline":     ';',
	"link":        '@',
	"image":       '!',
	"code":        '`',
	"heading":     '#',
    "listElement": '+',
}

// Stack to keep track of the open states
var openStates []int

func lastState() int {
	return openStates[len(openStates)-1]
}

func pushState(state int) {
	openStates = append(openStates, state)
}

func remLastState() {
	openStates = openStates[0 : len(openStates)-1]
}

var source string
var output string
var wrapBeg string
var wrapEnd string
var ch byte

var current = 0
var peeked = 1
var hNum int

var preserveNewLines = false

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
	var compiled bytes.Buffer
	for ; current < len(source); next() {
		ch = curr()
		switch ch {
		case blockChars["esc"]:
			next()
			ch = curr()
			compiled.WriteByte(ch)

		case blockChars["escapeBegin"]:
			next()
			ch = curr()
			for ch != blockChars["escapeEnd"] && current < len(source) {
				if ch == blockChars["esc"] {
					next()
					ch = curr()
				}
				compiled.WriteByte(ch)
				next()
				ch = curr()
				if ch == blockChars["escapeEnd"] {
					break
				}
			}

		case blockChars["bold"]:
			if len(openStates) > 0 && lastState() == bold {
				compiled.WriteString("</strong>")
				remLastState()
			} else {
				compiled.WriteString("<strong>")
				pushState(bold)
			}

		case blockChars["italics"]:
			if len(openStates) > 0 && lastState() == italics {
				compiled.WriteString("</em>")
				remLastState()
			} else {
				compiled.WriteString("<em>")
				pushState(italics)
			}

		case blockChars["underline"]:
			if len(openStates) > 0 && lastState() == underline {
				compiled.WriteString("</u>")
				remLastState()
			} else {
				compiled.WriteString("<u>")
				pushState(underline)
			}

		case blockChars["line"]:
			if peek() == blockChars["line"] && peek2() == blockChars["line"] {
				if len(openStates) == 0 {
					next()
					next()
					ch = curr()
					compiled.WriteString("<hr>")
				} else {
					compiled.WriteByte(ch)
				}
			} else {
				compiled.WriteByte(ch)
			}

        case blockChars["listElement"]:
            println("here")
            compiled.WriteString("<li>")
            next()
            ch = curr()
            start := current
            end := current
            for ch != '\n' {
                end++
                next()
                ch = curr()
            }
            compiled.WriteString(source[start:end])
            compiled.WriteString("</li>")

		case blockChars["newline"]:
			if peek() == blockChars["newline"] {
				next()
				compiled.WriteString("<br>")
			} else {
				compiled.WriteByte(ch)
			}

		case blockChars["heading"]:
			hNum = 1
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
			current--
			peeked--
			compiled.WriteString("<h" + strconv.Itoa(hNum) + ">")
			pushState(heading)

		case '\n':
			if len(openStates) > 0 && lastState() == heading {
				remLastState()
				compiled.WriteString("</h" + strconv.Itoa(hNum) + ">")
				hNum = 0
			} else if preserveNewLines {
				compiled.WriteString("<br>")
			} else {
				compiled.WriteByte(ch)
			}

		case blockChars["link"]:
			next()
			ch = curr()
			if ch == '[' {
				compiled.WriteString("<a href=\"")
				var url bytes.Buffer
				var alt bytes.Buffer
				next()
				ch = curr()
				for ch != ']' && current < len(source) {
					alt.WriteByte(ch)
					next()
					ch = curr()
				}
				next()
				ch = curr()
				if ch != '(' {
					println("Improperly formatted link found.")
					println(string(ch))
					os.Exit(1)
				}
				next()
				ch = curr()
				paropen := 0
				for ch != ')' && current < len(source) {
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
				// next()
				// ch = curr()
				compiled.WriteString(url.String())
				compiled.WriteString("\">")
				if alt.String() == "" {
					compiled.WriteString(url.String())
				} else {
					compiled.WriteString(alt.String())
				}
				compiled.WriteString("</a>")
			}

		case blockChars["image"]:
			next()
			ch = curr()
			if ch == '[' {
				compiled.WriteString("\n<img src=\"")
				var url bytes.Buffer
				var alt bytes.Buffer
				var w bytes.Buffer
				var h bytes.Buffer
				next()
				ch = curr()
				for ch != ']' && current < len(source) {
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
						for ch != ']' && current < len(source) {
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
					println("Improperly formatted image-link encountered.")
					println(string(ch))
					os.Exit(1)
				}
				next()
				ch = curr()
				paropen := 0
				for ch != ')' && current < len(source) {
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
				compiled.WriteString(url.String())
				compiled.WriteString("\" alt=\"")
				if alt.String() == "" {
					compiled.WriteString(url.String())
				} else {
					compiled.WriteString(alt.String())
				}
				compiled.WriteString("\"")
				if w.String() != "" {
					compiled.WriteString(" width=\"")
					compiled.WriteString(w.String())
					compiled.WriteByte('"')
				}
				if h.String() != "" {
					compiled.WriteString(" height=\"")
					compiled.WriteString(h.String())
					compiled.WriteByte('"')
				}
				compiled.WriteString(">\n")
			}

		case blockChars["code"]:
			next()
			ch = curr()
			compiled.WriteString("<pre>")

            // if the code block looks like `::filename`, then read the contents of the file
            // and use that as the source
            if ch == ':' && peek() == ':' {
                next()
                next()
                start := current
                end := current
                for ch != blockChars["code"] && current < len(source) {
                    end ++
					next()
					ch = curr()
                }
                filename := source[start:end]
                file, err := os.Open(filename)
                if err != nil {
                    fmt.Println("error: unable to open file '", filename, "'")
                }
                defer file.Close()
                scanner := bufio.NewScanner(file)
                lineno := 1
                for scanner.Scan() {
                    display := fmt.Sprintf("%3d| %s\n", lineno, scanner.Text())
                    compiled.WriteString(display)
                    lineno ++
                }
            } else {
                ch = curr()
                lineno := 1
                number := fmt.Sprintf("%3d| ", lineno)
                if ch == '\n' {
                    lineno --
                } else {
                    compiled.WriteString(number)
                }
                for ch != blockChars["code"] && current < len(source) {
                    if ch == '\n' {
                        if peek() == blockChars["code"] {
                            next()
                            ch = curr()
                            break
                        }

                        lineno++
                        number = fmt.Sprintf("\n%3d| ", lineno)
                        compiled.WriteString(number)
                        next()
                        ch = curr()

                    } else {
                        if ch == '\\' && peek() == blockChars["code"] {
                            next()
                            ch = curr()
                        }
                        compiled.WriteByte(ch)
                        next()
                        ch = curr()
                    }

                    if ch == blockChars["code"] || current >= len(source) {
                        break
                    }
                }
            }
			compiled.WriteString("</pre>")

		default:
			compiled.WriteByte(ch)

		}
	}
	return compiled.String()
}

func writeToFile(content, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		println("Could not open file, exiting.")
		os.Exit(1)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		println("Could not write to file exiting.")
	}
}

func usage() {
	println(`
USAGE:
------
	sitefl [-OPTIONS] [stylesheet, template] source destination

OPTIONS:
--------
	f : Prints this message
	n : Preserves new lines
	s : attach Stylesheet
	t : use template
	w : wrap the output in a div (id = 'unit')

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

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(0)
	}

	optionsGiven := false
	html := 1
	css := 1
	src := 2
	dst := 3
	var htmlBeg string
	var htmlend string

	if os.Args[1][:1] == "-" {
		optionsGiven = true
		options := os.Args[1][1:]
		for _, v := range options {
			switch string(v) {
			case "n":
				preserveNewLines = true
				if len(options) == 1 {
					if len(os.Args) != 4 {
						usage()
						os.Exit(0)
					}
				}
			case "h":
				usage()
				os.Exit(0)
			case "t":
				if len(os.Args) < 5 {
					usage()
					os.Exit(0)
				}
				if css == 2 {
					html = 3
				} else {
					html = 2
				}
				src++
				dst++
			case "s":
				if len(os.Args) < 5 {
					usage()
					os.Exit(0)
				}
				if html == 2 {
					css = 3
				} else {
					css = 2
				}
				src++
				dst++
			case "w":
				wrapBeg = "<div id='unit'>\n"
				wrapEnd = "\n</div>"
			default:
				fmt.Printf("Unknown option : %q", v)
			}
		}
	}

	if !optionsGiven {
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
			println("Could not read file named", os.Args[src])
			return
		}
		source = string(scode)
	}

	output := wrapBeg + compile() + wrapEnd

	if html != 1 {
		templateHTML, err := ioutil.ReadFile(os.Args[html])
		if err != nil {
			println("Could not read file named", os.Args[html])
			return
		}
		htmlData := string(templateHTML)
		htmlBeg = htmlData[:strings.Index(htmlData, "ent\">")+5]
		htmlend = htmlData[strings.Index(htmlData, "ent\">")+5:]
	}

	if css != 1 {
		htmlBeg = htmlBeg[:strings.Index(htmlBeg, "href=\"")+6] + os.Args[css] + htmlBeg[strings.Index(htmlBeg, "href=\"")+6:]
	}

	if html != 1 {
		output = htmlBeg + output + htmlend
	}

	if os.Args[dst] == "out" {
		println(output)
	} else {
		writeToFile(output, os.Args[dst])
	}
}
