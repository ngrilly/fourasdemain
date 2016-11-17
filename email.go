package fourasdemain

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"appengine"
	"appengine/urlfetch"
	"github.com/pkg/errors"
)

const mailgunEndpoint = "https://api.mailgun.net/v3/fourasdemain.fr/messages"

var mailgunAPIKey = os.Getenv("MAILGUN_API_KEY")

// SendEmail sends an email.
func SendEmail(ctx appengine.Context, from string, to []string, subject string, msg []byte, inline string, tag string) error {
	reqBody := &bytes.Buffer{}

	w := multipart.NewWriter(reqBody)
	addString(w, "from", from)
	addString(w, "to", to...)
	addString(w, "subject", subject)
	addString(w, "o:tracking", "yes")
	addString(w, "o:tag", tag)
	addBytes(w, "html", msg)
	if inline != "" {
		addFile(w, "inline", inline)
	}
	w.Close()

	req, err := http.NewRequest("POST", mailgunEndpoint, reqBody)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetBasicAuth("api", mailgunAPIKey)

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not reach mailgun")
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read body")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("mailgun error: %s\n%s", resp.Status, respBody)
	}
	return nil
}

func addString(w *multipart.Writer, fieldname string, values ...string) {
	for _, v := range values {
		err := w.WriteField(fieldname, v)
		if err != nil {
			panic(err)
		}
	}
}

func addBytes(w *multipart.Writer, fieldname string, values ...[]byte) {
	for _, v := range values {
		part, err := w.CreateFormField(fieldname)
		if err != nil {
			panic(err)
		}
		_, err = part.Write(v)
		if err != nil {
			panic(err)
		}
	}
}

func addFile(w *multipart.Writer, fieldname, filename string) {
	part, err := w.CreateFormFile(fieldname, filename)
	if err != nil {
		panic(err)
	}

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(part, f)
	if err != nil {
		panic(err)
	}
}
