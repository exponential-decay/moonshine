# moonshine

Given four bytes, download a random file from the [UK Web Archive](https://www.webarchive.org.uk/shine), e.g.

```go run moonshine.go -ffb d0cf11e0 | xargs wget```

Full usage:
```
Usage of ./moonshine:
  -ffb string
      first four bytes of file to find (default "0baddeed")
  -gif
      return a single gif from the UKWA
  -list
      list the first five pages from page number
  -page int
      specify a page number to return from (default 1)
  -random
      return a random link to a file (default true)
  -stat
      stat the resource
  -version
      Return version
```