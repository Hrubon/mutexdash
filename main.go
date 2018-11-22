package main

import (
	"fmt"
	"github.com/Hrubon/mutexdash/model"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"html/template"
	"log"
	"net/http"
	"os"
)

/* Command-line options definition */

var opts struct {
	EtcdEndpoints []string `short:"e" long:"etcd-eps" required:"true" description:"URL describing etcd REST endpoint. This option can appear multiple times."`
	EtcdTimeout int `short:"t" long:"etcd-to" default:"5" description:"etcd receive timeout - how many seconds to wait for a response."`
	EtcdRootNs string `short:"n" long:"etcd-root-ns" default:"/mutexes" description:"etcd root namespace - where to look for mutexes."`
	SkipCheck bool `short:"s" long:"skip-test" description:"Skip testing the connection to the etcd upstream on startup."`
	ListenOn string `short:"l" long:"listen-on" required:"true" description:"Socket on which to listen for incomming HTTP connections - e.g. 127.0.0.1:8080."`
}

func usage() {
	fmt.Println("usage: ./mutexdash")
	fmt.Println("	-e <etcd upstream endpoint list>")
	fmt.Println("	-t <etcd receive timeout in seconds>")
	fmt.Println("	-n <etcd mutex root namespace>")
	fmt.Println("	-s whether to skip initial etcd upstream check")
	fmt.Println("	-l <http listening socket - in form address:port>")

}

/* HTTP error-handling */

func httpError(w http.ResponseWriter, httpCode int, err error) {
	err = errors.Wrapf(err, "HTTP status %d", httpCode)
	http.Error(w, err.Error(), httpCode)
}

func assertHttpMethod(w http.ResponseWriter, r *http.Request, method string) bool {
if r.Method != method {
		msg := fmt.Sprintf("HTTP method '%s' is not allowed, use '%s' method instead", r.Method, method)
		httpError(w, http.StatusMethodNotAllowed, errors.New(msg))
		return false
	}
	return true
}

/* Main entry point - web server controller */

func main() {
	// Parse command-line options
    _, err := flags.Parse(&opts)
	if err != nil {
		usage()
		os.Exit(1)
	}

	// Init model and perform upstream check
	model := model.NewModel(opts.EtcdEndpoints, opts.EtcdTimeout, opts.EtcdRootNs)
	if !opts.SkipCheck {
		err = model.TestConnection()
		if err != nil {
			err = errors.Wrap(err, "Exiting due to a problem with the etcd upstream")
			fmt.Println(err)
			os.Exit(2)
		}
	}

	// Setup web server controller
	http.HandleFunc("/",
		func (w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/mutex/list", 301)
		})
	http.HandleFunc("/mutex/list",
		func (w http.ResponseWriter, r *http.Request) {
			if !assertHttpMethod(w, r, "GET") {
				return
			}
			mList, err := model.ListMutexes()
			if err != nil {
				httpError(w, http.StatusInternalServerError, err)
				return
			}
			t, err := template.ParseFiles("mutexlist.html")
			if err != nil {
				httpError(w, http.StatusInternalServerError, err)
				return
			}
			err = t.Execute(w, mList)
			if err != nil {
				httpError(w, http.StatusInternalServerError, err)
				return
			}
		})
	http.HandleFunc("/mutex/unlock",
		func (w http.ResponseWriter, r *http.Request) {
			if !assertHttpMethod(w, r, "POST") {
				return
			}
			etcdPath := r.PostFormValue("etcdPath")
			if etcdPath == "" {
				err := errors.New("'etcdPath' field is missing in the POST request body")
				httpError(w, http.StatusInternalServerError, err)
				return
			}
			err := model.UnlockMutex(etcdPath)
			if err != nil {
				httpError(w, http.StatusInternalServerError, err)
				return
			}
			prevService := r.PostFormValue("prevService")
			if prevService != "" {
				http.Redirect(w, r, "/mutex/list#" + prevService, 303)
			} else {
				http.Redirect(w, r, "/mutex/list", 303)
			}
		})

	// Start listening for HTTP requests
	fmt.Fprintln(os.Stderr, "MutexDash v1.0")
	log.Fatal(http.ListenAndServe(opts.ListenOn, nil))
}

