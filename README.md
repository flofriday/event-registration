# event-registration

**WIP**: do not use!

This service is developed for the
[Flyout Event](https://spaceteam.at/flyout/?lang=en) organized by the
[SpaceTeam](https://spaceteam.at/?lang=en).
However, it should also be easily adaptable to other events.

## Goals

**Note:** At the moment none of those are implemented

- Storage of contact data according to the current situation in Austria:
  - first and last name
  - phone number and if available email (this app makes it a requirement)
  - Date and time of registration
- Easy setup, configuration and deployment
- Simple datastorage (sqlite) and export(csv, pdf) via the Webinterface.
- Great documentation

## Setup

1. First you need to install Go, a C compiler, and sqlite.
2. Edit config.toml (the file has comments to guide you)
3. Run `go run main.go`
