# pindish

Configuration
-------------

Configure your .env file with:
HOST=0.0.0.0
PORT=8080
DATABASE_URL=
PINDISH_APP_ID=
PINDISH_APP_SECRET=

Building
--------
go get
go build
go install
source .env
$GOPATH/bin/pindish

Set your APP_ID and APP_SECRET according to your pinterest developer account

DATABASE_URL is a postgres database URL, heroku sets this if you add their database

Deploy to heroku or wherever and the API will be alive

If deploying to heroku, do a push to your heroku master branch to deploy the app
Be sure to push your configuration!

Pindish uses 12 factor, so all configuration comes from environment

License
-------

MIT license, 2019