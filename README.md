# moonshine

Given four bytes, download a random file from the [UK Web Archive](https://www.webarchive.org.uk/shine), e.g.

```sh
moonshine -ffb d0cf11e0 | xargs wget
```

Full usage:

```text
  -ffb string
        first four bytes of file to find (default "0baddeed")
  -gif
        return a single gif (hex "47494638")
  -list
        list the first five pages from page number
  -page int
        specify a page number to return from, [max: 1000] (default 1)
  -random
        return a random link to a file (default true)
  -sample
        return a sampled list of up to 20 files across the maximum no. results
  -stats
        return statistics for the resource
  -version
        Return version

```

## Sample mode

Sample mode has been added to moonshine to return a more varied selection of the
desired file format example to file format researchers. The results will be
selected across a distribution of all the pages available to the Shine
interface.

To use sample mode you can do the following:

```sh
moonshine --sample --ffb cafebeef | xargs wget
```

## Developing moonshine

### Goreleaser

Testing goreleaser can be done as follows:

* `goreleaser release --skip-publish`

Valid semantic versioning looks as follows:

* `vMM.mm.pp-rc.n`

Where `-rc.n` are optional, e.g. used purely for release candidates.
