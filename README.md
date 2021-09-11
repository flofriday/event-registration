# event-registration

![Screenshot](https://user-images.githubusercontent.com/21206831/132953023-4ccf8fb8-0a3a-445c-9802-799b990db510.png)


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

## How it works

After you deployed the app you can [create a QR Code](https://qr-creator.com/url.php)
to the domain where it is running. Attendees of the event scan that code and
register. Once they are registered, a cookie is set, which will only expire
after three days or if the user is deleted from the database. This means all
cookies can be invalidated by deleting the sqlite database file.

For you, the event host, and your colleagues, there is a admin page available under
`/admin`. From there you can export all contacts as JSON or CSV(Excel). The
admin dashboard also shows the last 10 registered persons which means your
colleagues at the entrance don't have to touch attendees phones and can verify
registrations from a save distance on their own devices.

![Screenshot](https://user-images.githubusercontent.com/21206831/132955534-bcb03c1f-82db-4377-9ce5-1386dfa4dd2a.png)

## Licence

Everything in this repository is under the MIT license, except for the
`flyout` template folder. I included those templates to show how this service
was used in production.
