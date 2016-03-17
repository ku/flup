package main

import (
	"flag"
	"fmt"
	"github.com/ku/flickgo"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func startDaemon(port int, ch chan string, authFileName string) {
	var authToken string

	uploadQueue := make(chan string, 64*1024)

	flag.Parse()

	apikey := flag.Arg(0)
	secret := flag.Arg(1)

	httpClient := &http.Client{}
	fl := flickgo.New(apikey, secret, httpClient)

	var frob string

	http.HandleFunc("/auth/start", func(w http.ResponseWriter, r *http.Request) {
		frob = fl.GetFrob()
		res := flickgo.SignedURL(secret, apikey, "auth", map[string]string{
			"frob":  frob,
			"perms": "delete",
		})

		t, _ := template.New("foo").Parse(`
			<h1>Needs your authorization</h1>
			<ol>
				<li><a href="{{.}}" target="_blank">Open flickr to authorize the app</a>
				<li><a href="./done">Continue to save the token after authorizing the app</a>
		`)
		t.Execute(w, template.HTML(res))

	})
	http.HandleFunc("/auth/done", func(w http.ResponseWriter, r *http.Request) {
		token, user, err := fl.GetToken(frob)
		fmt.Println(token, user, err)
		ioutil.WriteFile(authFileName, []byte(token), 0600)
		ch <- token

		t, _ := template.New("foo").Parse(`
			<h1>Token saved successfully</h1>
			<div>
			Add files to upload with 
			<pre><code>
				fullpathname=` + "`echo $file | sed 's/%/%25/g' | sed 's/ /%20/g'`" + `
				curl "http://localhost:58080/queue/add?file="$fullpathname
			</code></pre>
			or something like
			<pre><code>
			 find ` + "`pwd`" + ` -type f -print0  | xargs -0 flup/add2queue.sh     
			</code></pre>
			</div>
		`)
		t.Execute(w, nil)

	})

	http.HandleFunc("/queue/add", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		files := r.Form["file"]
		for _, file := range files {
			uploadQueue <- file
			fmt.Println("added", file)
		}
	})

	go func() {
		fmt.Println("waiting ch")
		authToken = <-ch
		fmt.Println("authToken", authToken)
		fl.AuthToken = authToken
	}()

	// uploaders

	go func() {
		mutex := make(chan int, 4)

		for {
			mutex <- 1
			go func() {
				fullpathname := <-uploadQueue

				components := strings.Split(fullpathname, "/")
				filename := components[len(components)-1]

				name := filename
				args := map[string]string{
				//"tags": "test tag",
				}

				bytes, _ := ioutil.ReadFile(fullpathname)

				//fmt.Println("start uploading", name)
				ticket, err := fl.Upload(name, bytes, args)
				if err == nil {
					cmd := exec.Command("metatag", "-a", "Blue", fullpathname)
					err := cmd.Run()
					if err != nil {
						fmt.Println(err)
					}
				} else {
				}
				fmt.Println(name, ticket, err)
				<-mutex
			}()
		}
	}()

	fmt.Printf("starting at port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}

func authTokenChannel() chan string {
	authFileName := ".key"

	ch := make(chan string, 1)

	_, err := os.Stat(authFileName)
	if err == nil {
		buf, _ := ioutil.ReadFile(authFileName)
		token := string(buf)
		fmt.Println(token)
		ch <- token
		fmt.Println(token)
	} else {
		go func() {
			time.Sleep(time.Millisecond * 100)
			cmd := exec.Command("/usr/bin/open", "http://localhost:58080/auth/start")
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				close(ch)
			}
		}()
	}
	return ch
}

func main() {
	port := 58080
	ch := authTokenChannel()

	startDaemon(port, ch, ".key")
}
