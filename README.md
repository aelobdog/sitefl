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
