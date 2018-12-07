package main

import (
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"log"
	"net/http"
	"time"

	"crypto/rand"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"encoding/json"
)

var S *http.Server

var RespondToHealth bool

var (
	startupTime time.Duration
	listen      string
)

var Data [][]byte

func init() {

	pflag.DurationVar(&startupTime, "boot-time", 1*time.Second, "time it takes to start up the service")
	pflag.StringVar(&listen, "listen", ":8080", "http listen")

}

type Message struct {
	Test string
}

func main() {

	log.Printf(`
   ___ _      ________  ___________  __  _______
  / _ | | /| / / __/  |/  / __/ __ \/  |/  / __/
 / __ | |/ |/ / _// /|_/ /\ \/ /_/ / /|_/ / _/
/_/ |_|__/|__/___/_/  /_/___/\____/_/  /_/___/  `)

	log.Printf("Loading Resources ... ")

	time.Sleep(startupTime)

	RespondToHealth = true

	r := mux.NewRouter()

	r.PathPrefix("/ui").Handler(http.StripPrefix("/ui", http.FileServer(http.Dir("./ui"))))
	r.HandleFunc("/action/{action}", ActionHandler)
	r.HandleFunc("/", InfoHandler)
	r.HandleFunc("/health", HealthHandler)

	S = &http.Server{
		Addr:           listen,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Started server")

	log.Fatal(S.ListenAndServe())

}

func ActionHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch vars["action"] {

	case "kill":

		go func() {
			time.Sleep(time.Second * 5)
			S.Close()
			panic("I just died in your arms tonight ")

		}()

		rw.Write([]byte("Application will be killed in 5 seconds"))

	case "malloc20mb":
		log.Printf("Allocating 20mb to existing %d Mb", len(Data)/2048*2)

		for i := 0; i < 1024*20; i++ {
			kb := make([]byte, 1024)
			rand.Read(kb)
			Data = append(Data, kb)
		}

		res := fmt.Sprintf("Allocated 20mb. Size now: %d Mb", len(Data)/2048*2)

		rw.Write([]byte(res))

	case "livenessoff":
		RespondToHealth = false

		rw.Write([]byte("Letting /health time out from now on"))

	case "fileinfo":
		nofiles := 0
		var size int64
		var files []string
		filepath.Walk("/", func(path string, info os.FileInfo, err error) error {

			if strings.HasPrefix("/dev", path) {
				return nil
			}
			if strings.HasPrefix("/proc", path) {
				return nil
			}

			if err != nil {
				return nil
			}
			files = append(files, info.Name())
			nofiles++
			size = size + info.Size()
			return nil
		})

		res := fmt.Sprintf("Found %d files. Size: %d Mb", nofiles, size/1024/1024)

		rw.Write([]byte(res))

	case "log100":
		lines := 100
		start := time.Now()
		for i :=0; i <lines ; i++ {
			log.Printf("Logging a lot: %d ", i)

		}
		d := time.Since(start)
		res := fmt.Sprintf("Logged %d lines in %.2f seconds", lines, d.Seconds())

		rw.Write([]byte(res))


	case "log1000":
		lines := 1000
		start := time.Now()
		for i :=0; i <lines ; i++ {
			log.Printf("Logging a lot: %d ", i)

		}
		d := time.Since(start)
		res := fmt.Sprintf("Logged %d lines in %.2f seconds", lines, d.Seconds())

		rw.Write([]byte(res))

	case "log10000":
		lines := 10000
		start := time.Now()
		for i :=0; i <lines ; i++ {
			log.Printf("Logging a lot: %d ", i)

		}
		d := time.Since(start)
		res := fmt.Sprintf("Logged %d lines in %.2f seconds", lines, d.Seconds())

		rw.Write([]byte(res))

	case "cpusmall":
		const testBytes = `{ "Test": "value" }`
		iter := int64(700000)
		start := time.Now()
		p := &Message{}
		for i := int64(1); i < iter; i++ {
			json.NewDecoder(strings.NewReader(testBytes)).Decode(p)
		}
		d := time.Since(start)
		res := fmt.Sprintf("[small]. Took %.2f seconds", d.Seconds())
		rw.Write([]byte(res))



	case "cpumedium":
		const testBytes = `{ "Test": "value" }`
		iter := int64(2000000)
		start := time.Now()
		p := &Message{}
		for i := int64(1); i < iter; i++ {
			json.NewDecoder(strings.NewReader(testBytes)).Decode(p)
		}
		d := time.Since(start)
		res := fmt.Sprintf("[medium]. Took %.2f seconds", d.Seconds())
		rw.Write([]byte(res))


	case "cpularge":
		const testBytes = `{ "Test": "value" }`
		iter := int64(8000000)
		start := time.Now()
		p := &Message{}
		for i := int64(1); i < iter; i++ {
			json.NewDecoder(strings.NewReader(testBytes)).Decode(p)
		}
		d := time.Since(start)
		res := fmt.Sprintf("[large]. Took %.2f seconds", d.Seconds())
		rw.Write([]byte(res))






	}



}

type Info struct {
	Hostname   string `json:"hostname"`
	ClientAddr string `json:"client"`
}

func InfoHandler(rw http.ResponseWriter, r *http.Request) {
	i := Info{}
	i.Hostname, _ = os.Hostname()

	i.ClientAddr = r.RemoteAddr

	t, err := ioutil.ReadFile("ui/index.html")
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("test").Parse(string(t))

	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(rw, i)
	if err != nil {
		panic(err)
	}

}

func HealthHandler(rw http.ResponseWriter, r *http.Request) {

	if !RespondToHealth {
		time.Sleep(30 * time.Minute)

	}

	rw.Write([]byte("All good!"))

}
