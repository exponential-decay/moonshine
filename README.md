# moonshine

Given four bytes, download a random file from the [UK Web Archive](https://www.webarchive.org.uk/shine), or the *Archives Unleashed* [Warclight](http://warclight.archivesunleashed.org) project e.g.

```./moonshine -ffb d0cf11e0 | xargs wget```

or Warclight:

```./moonshine -warclight -gif | xargs wget```

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
      specify a page number to return from (default 1)
  -random
      return a random link to a file (default true)
  -stat
      stat the resource
  -version
      Return version
  -warclight
      Use Warclight instead of Shine
```

## Developing moonshine

### Goreleaser

Testing goreleaser can be done as follows:

* `goreleaser release --skip-publish`

Valid semantic versioning looks as follows:

* `vMM.mm.pp-rc.n`

Where `-rc.n` are optional, e.g. used purely for release candidates.
