package fourasdemain

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"appengine"
	"appengine/datastore"
)

func init() {
	http.Handle("/api/subscribe", appHandler(subscribe))
}

const ourEmail = `"Nicolas Grilly pour Fouras Demain" <contact@fourasdemain.fr>`

var assetDir = PackageDir()

type Subscriber struct {
	Email     string
	Referrer  string
	FormURI   string
	UserAgent string
	IP        string
	Date      time.Time
}

func subscribe(ctx appengine.Context, r *http.Request) interface{} {
	s := Subscriber{
		r.FormValue("email"),
		r.FormValue("referrer"),
		r.Header.Get("Referer"),
		r.Header.Get("User-Agent"),
		remoteHost(r),
		time.Now(),
	}

	if _, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "subscriber", nil), &s); err != nil {
		return err
	}

	// TODO: remove the logo
	// TODO: set Reply-To to s.Email
	err := SendEmail(
		ctx,
		ourEmail,
		[]string{ourEmail},
		"Fouras Demain - Nouvelle inscription",
		MustRenderTemplate(assetDir+"/notify_form_submit.html", structToMap(s)),
		"notification",
	)
	if err != nil {
		return err
	}

	err = SendEmail(
		ctx,
		ourEmail,
		[]string{s.Email},
		"Merci pour votre inscription sur www.fourasdemain.fr",
		MustRenderTemplate(assetDir+"/welcome.html", nil),
		"confirmation",
	)
	if err != nil {
		return err
	}

	return struct{}{}
}

// structToMap convert a struct to a map of string to interface{}.
func structToMap(s interface{}) map[string]interface{} {
	data := map[string]interface{}{}
	sv := reflect.ValueOf(s)
	st := sv.Type()
	for i := 0; i < st.NumField(); i++ {
		data[st.Field(i).Name] = sv.Field(i).Interface()
	}
	return data
}

type appHandler func(appengine.Context, *http.Request) interface{}

var internalServerErrorResponse = []byte(`{"error": "Internal Server Error"}`)

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	response := fn(ctx, r)

	w.Header().Set("Content-Type", "application/json")

	err, _ := response.(error)
	if err != nil {
		ctx.Errorf("handler error: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(internalServerErrorResponse)
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		ctx.Errorf("cannot encode response as JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(internalServerErrorResponse)
		return
	}

	w.Write(b)
}

func remoteHost(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Return the full remote address if we are unable to strip the port
		return r.RemoteAddr
	}
	return host
}

// PackageDir returns the path to the directory containing the source code of its caller.
// It is useful to locate assets stored alongside the code using them.
//
// Implementation notes:
// Within the Google App Engine toolchain, the filename returned by runtime.Caller is relative to the project root.
// But within the standard Go toolchain, the filename returned by runtime.Caller is an absolute path.
// As a consequence, this works only if the source code is available at runtime in the same location as at build time.
// We could fix this by making all paths relative with -gcflags '-trimpath=...' (Google App Engine does that).
// We could also fix this by letting PackageDir compute the project base dir using runtime.Caller and then compute a relative path using filepath.Rel().
//
func PackageDir() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("cannot recover caller file name")
	}
	return filepath.ToSlash(filepath.Dir(filename))
}

func MustRenderTemplate(filename string, data interface{}) []byte {
	// TODO: Scan and parse all templates at program initialization
	t := template.Must(template.ParseFiles(filename))
	var b bytes.Buffer
	err := t.Execute(&b, data)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}
