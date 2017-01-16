# OpenGroupware iCalendar Service

[![GitHub release](https://img.shields.io/github/release/geo-data/ogo-ical.svg)](https://github.com/geo-data/ogo-ical/releases/latest)
[![Travis CI](https://img.shields.io/travis/geo-data/ogo-ical.svg)](https://travis-ci.org/geo-data/ogo-ical)
[![Go Report Card](https://goreportcard.com/badge/github.com/geo-data/ogo-ical)](https://goreportcard.com/report/github.com/geo-data/ogo-ical)
[![GoDoc](https://img.shields.io/badge/documentation-godoc-blue.svg)](https://godoc.org/github.com/geo-data/ogo-ical)

This repository provides a server which exposes an existing OpenGroupware
database as an iCalendar service.  This makes it suitable for integrating
OpenGroupware installations with external calendar services which require
resources in iCalendar format.

## Usage

Let's assume the services is installed and configured to run on <http://localhost:8080>  To use it you need to point your calendar software to the iCalendar feed at <http://localhost:8080> and add one or more filters (you can't download the whole calendar at once for security reasons).  You can filter by username and/or case insensitive matching of event titles.  Examples are:

* <http://localhost:8080?user=me> find all events the user **me** is involved in.

* <http://localhost:8080?match=my%20project> find all events with the string **my project** in the title.

* <http://localhost:8080?user=me&match=my%20project> find all **my project** events that the user **me** is involved in.

* <http://localhost:8080?user=me&user=you> find all events that either users **me** or **you** are involved in.

Feeds show future events and past events up to a month ago. Note that feeds are
read only - you can't update OpenGroupware from your external calendar.

## Configuration

```
$ ogo-ical -h
Usage of ./ogo-ical:
  -address string
        server address (default ":8080")
  -dsn string
        postgresql Data Source Name
  -version
        display version information
```

The server can also be configured by setting the `$OGO_ICAL_DSN` and
`$OGO_ICAL_ADDRESS` environment variables which set the `-dsn` and `-address`
flags respectively.

## Installation

### Binary download

You can download a self contained `ogo-ical` binary compiled for Linux x86_64
from the [latest release](https://github.com/geo-data/ogo-ical/releases/latest).

### Via Docker

The latest version is available from the Docker Registry at
[geodata/ogo-ical:latest](https://hub.docker.com/r/geodata/ogo-ical).  It can be
run as follows:

```
docker run -d -e OGO_ICAL_ADDRESS=my.server:80 -e "OGO_ICAL_DSN=host=postgres dbname=ogo" geodata/ogo-ical:latest
```

### From source

Install [Go](https://golang.org/) and simply:

```
go get github.com/geo-data/ogo-ical
```

This should install the `ogo-ical` binary under `$GOPATH/bin`.

## License

[![license](https://img.shields.io/github/license/geo-data/ogo-ical.svg)](https://github.com/geo-data/ogo-ical/blob/master/LICENSE)

MIT - See the file `LICENSE` for details.
