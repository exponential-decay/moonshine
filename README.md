# moonshine

Given four bytes, download a random file from the [UK Web Archive](https://www.webarchive.org.uk/shine), e.g.

```go run moonshine.go -ffb d0cf11e0 | xargs wget```

Full usage:
```
Usage of ./moonshine:
  -ffb string
    	first four bytes of file to find (default "0baddeed")
  -list
    	list up to the first five pages results
  -random
    	return a random link to a file (default true)
  -stat
    	stat the resource
  -version
    	Return version.
```