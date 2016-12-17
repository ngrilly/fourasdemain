Install the Google App Engine SDK for Go:

https://cloud.google.com/appengine/docs/go/download

Install Hugo:

	$ brew install hugo

Install CSS inliner:

	$ go get github.com/aymerick/douceur

Create and complete a file named `env` in the project root directory:

    MAILGUN_API_KEY=

To start the local server:

	$ make serve

To deploy to Google App Engine:

	$ make deploy
