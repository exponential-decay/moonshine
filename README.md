# moonshine

Given four bytes, download a random file from the [UK Web Archive](https://www.webarchive.org.uk/shine), e.g.

```./moonshine -ffb d0cf11e0 | xargs wget```

Full usage:

```text
Usage of ./moonshine:
  -ffb string
    	first four bytes of file to find (default "0baddeed")
  -gif
    	return a single gif
  -list
    	list the first five pages from page number
  -page int
    	specify a page number to return from, [max: 9000] (default 1)
  -random
    	return a random link to a file (default true)
  -stats
    	return statistics for the resource
  -version
    	Return version

```

## Developing moonshine

### Goreleaser

Testing goreleaser can be done as follows:

* `goreleaser release --skip-publish`

Valid semantic versioning looks as follows:

* `vMM.mm.pp-rc.n`

Where `-rc.n` are optional, e.g. used purely for release candidates.
