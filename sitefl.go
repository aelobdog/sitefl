/*
	MIT License

	Copyright (c) 2022 Ashwin Godbole

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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
    "bufio"
    "strings"
)

var blockChars = map[string]byte{
	"escape":     '\\', // impl
	"bold":        '*', // impl
	"italics":     '/', // impl
	"underline":   '_', // impl
    "strike":      '~', // impl
	"line":        '-', // impl
	"newline":     ';', // impl
	"link":        '@',
	"image":       '!',
	"code":        '`',
	"heading":     '#', // impl
    "listElement": '+', // impl
}

type SiteflObject struct {
    input []byte
    output bytes.Buffer
    currentPos int
    renderNL bool
    lineNo int
}

func NewSO() *SiteflObject {
    return &SiteflObject{currentPos: 0, lineNo: 0}
}

func (so *SiteflObject) advanceChar() {
    if so.currentPos < len(so.input) {
        if so.currentChar() == '\n' {
            so.lineNo++
        }
        so.currentPos++
    }
}

func (so *SiteflObject) currentChar() byte {
    if so.currentPos < len(so.input) {
        return so.input[so.currentPos]
    }
    return 0
}

func (so *SiteflObject) peekChar() byte {
    if so.currentPos < len(so.input) - 1 {
        return so.input[so.currentPos + 1]
    }
    return 0
}

func (so *SiteflObject) peekpeekChar() byte {
    if so.currentPos < len(so.input) - 2 {
        return so.input[so.currentPos + 2]
    }
    return 0
}

// process the input byte sequence till 'end' is encountered.
// this function is used for processing text decoration options
// like bold, italics, underline, strikethrough
func (so *SiteflObject) processTill(end byte, normal bool) {
    lineno := 1
    number := fmt.Sprintf("%3d| ", lineno)
    if !normal {
        if so.currentChar() == '\n' {
            lineno --
        } else {
            so.output.WriteString(number)
        }
    }
    for ; so.currentPos < len(so.input) && so.currentChar() != end; so.advanceChar() {
        if !normal {
            if so.currentChar() == '\n' {
                if so.peekChar() == blockChars["code"] {
                    so.advanceChar()
                    break
                }

                lineno++
                number = fmt.Sprintf("\n%3d| ", lineno)
                so.output.WriteString(number)
                // so.advanceChar()

            } else {
                if so.currentChar() == '\\' && so.peekChar() == blockChars["code"] {
                    so.advanceChar()
                    so.output.WriteByte(so.currentChar())
                    // very dirty, think of something to replace this
                    so.input[so.currentPos] = '.'
                } else {
                    so.output.WriteByte(so.currentChar())
                }
                // so.advanceChar()
            }

            if  so.currentChar() == blockChars["code"] || so.currentPos >= len(so.input) {
                break
            }
        } else {
            switch so.currentChar() {

            case blockChars["escape"]:
                so.output.WriteByte(so.peekChar())
                so.advanceChar()

                // NOTE: move to its own function ?
            case blockChars["bold"]:
                so.advanceChar()
                so.output.WriteString("<b>")
                so.processTill(blockChars["bold"], true)
                so.output.WriteString("</b>")

                // NOTE: move to its own function ?
            case blockChars["italics"]:
                so.advanceChar()
                so.output.WriteString("<i>")
                so.processTill(blockChars["italics"], true)
                so.output.WriteString("</i>")

                // NOTE: move to its own function ?
            case blockChars["underline"]:
                so.advanceChar()
                so.output.WriteString("<u>")
                so.processTill(blockChars["underline"], true)
                so.output.WriteString("</u>")

                // NOTE: move to its own function ?
            case blockChars["strike"]:
                so.advanceChar()
                so.output.WriteString("<s>")
                so.processTill(blockChars["strike"], true)
                so.output.WriteString("</s>")

            default:
                so.output.WriteByte(so.currentChar())
            }
        }
    }
}

// proess the input from start to finish, deals with top level
// formatting and delegates specific formatting to processTill
func (so *SiteflObject) processInput() {
    for so.currentPos < len(so.input) {
        switch so.currentChar() {

        case blockChars["escape"]:
            so.output.WriteByte(so.peekChar())
            so.advanceChar()

        case '\n':
            if so.renderNL {
                so.output.WriteString("<br>\n")
            }
            so.output.WriteByte('\n')

        case blockChars["newline"]:
            so.output.WriteString("<br>\n")

        case blockChars["heading"]:
            level := so.peekChar() - '0'
            if level < 1 || level > 6 {
                fmt.Println("Error (line", so.lineNo,"): found heading with level not in [1, 6]")
            }
            so.advanceChar()
            char := string(so.currentChar())
            so.advanceChar()

            so.output.WriteString("<h" + char + ">")
            so.processTill('\n', true)
            so.output.WriteString("</h" + char + ">\n")

        case blockChars["listElement"]:
            so.advanceChar()
            so.output.WriteString("<li>")
            so.processTill('\n', true)
            so.output.WriteString("</li>\n")

        case blockChars["code"]:
            so.advanceChar()
            so.output.WriteString("<pre>")

            if so.currentChar() == ':' && so.peekChar() == ':' {
                so.advanceChar()
                so.advanceChar()
                start := so.currentPos
                end := start
                for so.currentChar() != blockChars["code"] && 
                    so.currentPos < len(so.input) {
                    end ++
                    so.advanceChar()
                }
                filename := so.input[start:end]
                file, err := os.Open(string(filename))
                if err != nil {
                    fmt.Println("Error (line",
                                so.lineNo, 
                                ") unable to open file '",
                                filename, "'")
                }
                defer file.Close()
                scanner := bufio.NewScanner(file)
                lineno := 1
                for scanner.Scan() {
                    display := fmt.Sprintf("%3d| %s\n", lineno, scanner.Text())
                    so.output.WriteString(display)
                    lineno ++
                }
            } else {
                so.processTill(blockChars["code"], false)
            }
            so.output.WriteString("</pre>\n")

        case blockChars["line"]:
            if  so.peekChar() == blockChars["line"] &&
                so.peekpeekChar() == blockChars["line"] {
                so.advanceChar()
                so.advanceChar()
                so.output.WriteString("<hr>\n")
            } else {
                so.output.WriteByte(so.currentChar())
            }

        // NOTE: move to its own function ?
        case blockChars["bold"]:
            so.advanceChar()
            so.output.WriteString("<b>")
            so.processTill(blockChars["bold"], true)
            so.output.WriteString("</b>")

        // NOTE: move to its own function ?
        case blockChars["italics"]:
            so.advanceChar()
            so.output.WriteString("<i>")
            so.processTill(blockChars["italics"], true)
            so.output.WriteString("</i>")

        // NOTE: move to its own function ?
        case blockChars["underline"]:
            so.advanceChar()
            so.output.WriteString("<u>")
            so.processTill(blockChars["underline"], true)
            so.output.WriteString("</u>")

        // NOTE: move to its own function ?
        case blockChars["strike"]:
            so.advanceChar()
            so.output.WriteString("<s>")
            so.processTill(blockChars["strike"], true)
            so.output.WriteString("</s>")

        default:
            so.output.WriteByte(so.currentChar())
        }

        if so.currentPos >= len(so.input) - 1 {
            break;
        }
        so.advanceChar()
    }
}

func writeToFile(text string, f *os.File) {
    _, err := f.WriteString(text)
	if err != nil {
        println("Error: could not write to file", os.Args[2])
	}
}

func main() {
	if len(os.Args) < 3 {
		os.Exit(0)
	}
    scode, err := ioutil.ReadFile(os.Args[1])

    so := NewSO()
    so.input = scode

    if err != nil {
        fmt.Println("Error: could not read file `", os.Args[1], "`")
    }

    var htmlBeg string
    var htmlEnd string

    if len(os.Args) >= 4 {
		templateHTML, err := ioutil.ReadFile(os.Args[3])
		if err != nil {
			println("Could not read file named", os.Args[3])
			return
		}
		htmlData := string(templateHTML)
		htmlBeg = htmlData[:strings.Index(htmlData, "ent\">")+5]
		htmlEnd = htmlData[strings.Index(htmlData, "ent\">")+5:]
    }

    if len(os.Args) >= 5 {
        htmlBeg = htmlBeg[:strings.Index(htmlBeg, "href=\"") + 6] + os.Args[4] + htmlBeg[strings.Index(htmlBeg, "href=\"") + 6:]
    }

	f, err := os.Create(os.Args[2])
	if err != nil {
        println("Error: could not open file", os.Args[2])
		os.Exit(1)
	}
	defer f.Close()

    if len(os.Args) == 6 {
        so.renderNL = os.Args[5] == "true"
    }

    so.processInput()
    writeToFile(htmlBeg, f)
    writeToFile(so.output.String(), f)
    writeToFile(htmlEnd, f)
}
