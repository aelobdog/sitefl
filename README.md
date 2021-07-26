# sitefl
SiTeFL (Simple Text Format Language) is a markdown like text formatting language / text-to-html tool written from scratch in Go

[ SiTeFL now supports nested formatting ]

## Syntax:
```
- Headings    :   # ... ######

- Bold        :   *bold*

- Italics     :   /italics/

- Underline   :   _underline_

- Code        :   `code`

- Images      :   ![Alt-text](link to image)

- Link        :   @[Display-text::width::height](link url)

- NewLine     :   ;;

- Horizontal  :   ---

- Escape		  :		{ *bold* will appear as is (including the '*'s) }
  Formatting
```

USAGE:
------
	sitefl [-OPTIONS] [stylesheet, template] source destination

OPTIONS:
--------
	f : Prints this message
	n : Preserves new lines
	s : attach Stylesheet
	t : use template
	w : wrap the output within a div (id = 'unit')

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

EXAMPLE:
--------
( assuming you are in this directory )

### if you want to use the template,

	sitefl -ts defaults/templateHTML.html defaults/templateCSS.css testFile index.html
	

