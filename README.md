# sitefl
SiTeFL (Simple Text Format Language) is a markdown like text formatting language / text-to-html tool written from scratch in Go

Current issues :
  nested formatting does not work.
  it will require quite a bit of rewriting, so it will get fixed when I'm feeling like ;P

## Syntax:
```
- Headings  :   # ... ######
- Bold      :   *bold*
- Italics   :   /italics/
- Underline :   _underline_
- Code      :   `code`       --> works when split across multiple lines too
- Images    :   ![Alt-text](link to image)
- Link      :   @[Display-text::width::height](link url)
- NewLine   :   ;;
- Horizontal:   ---
```

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
