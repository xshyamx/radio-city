# RadioCity Podcast Feed #

Generate podcast rss feeds by scraping RadioCity website

## Configurations ##

The number of worker goroutines is configured by setting the environment variable `WC_COUNT` the default value is 10

## Build & Run ##

Install/update godep 

```sh
go get -u github.com/tools/godep
```

Build the executable & run

``` sh
go build && ./radio-city
```

or run for development using

``` sh
go run main.go scrape.go feed.go
```

### Generate master feed json ###

Run the `master.go` in the master folder

``` sh
cd master && go run master.go
```

