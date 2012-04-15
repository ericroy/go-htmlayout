### What is it?
Go-htmlayout (gohl) is a Go wrapper of the [HTMLayout](http://www.terrainformatica.com/htmlayout/) library by TerraInformatica.  HTMLayout is a fast, lightweight, and embeddable HTML/CSS rendering component for Windows.

### Project status
Go-htmlayout is written for Go1.  Most of the HTMLayout API has been wrapped (especially dom element related stuff), but there is still more of the HTMLayout API that I haven't tackled yet.  Tests are also incomplete at this point.  In short, this wrapper is probably not ready for serious production use.

### Getting started
Clone the repo.  In the main directory of the repository, run the following to fetch the latest version of HTMLayout:

```bash
python get_htmlayout.py
```

Though the tests provide an example of how to get started using gohl, there are probably better examples of window creation/management elsewhere on the internet.
