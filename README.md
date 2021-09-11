# event-registration

![out](https://user-images.githubusercontent.com/21206831/132853558-d948faa4-dd40-4730-809a-5267b1890de5.png)

This service is developed for the
[Flyout Event](https://spaceteam.at/flyout/?lang=en) organized by the
[TU Wien SpaceTeam](https://spaceteam.at/?lang=en).
However, it should also be easily adaptable to other events.

## Features

- Storage of contact data according to the current situation in Austria:
  - first and last name
  - phone number and if available email (this app makes it a requirement)
  - Date and time of registration
- Easy setup, configuration and deployment
- Simple storage (sqlite) and export(csv) via the Webinterface.
- Great documentation

## Setup native

1. First you need to install Go 1.16 or later, a C compiler, and sqlite.
2. Edit config.toml (the file has comments to guide you)
3. Compile with `go build` and run with `./event-registration`

The app is now available at http://localhost:3000/

## Setup with docker

1. First you need to install Docker
2. Edit config.toml (the file has comments to guide you)
3. Build the docker image with `docker build -t events .`
4. Run the container with: `docker run -it -p 3000:3000 --rm --name events-container events`

The app is now available at http://localhost:3000/

## Licence

Everything in this repository is under the MIT license, except for the
`flyout` template folder. I included those templates to show how this service
was used in production.
