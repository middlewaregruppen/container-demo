package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"

	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

var S *http.Server

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ata_increase_me_clicks",
		Help: "Times increase me link has been pressed",
	})
	gauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ata_request_load",
		Help: "Request Load",
	})
)

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

	pflag.Parse()

	// Recommended configuration for dev.
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		}}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(
		"awesome",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()

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
	r.HandleFunc("/authentication", AuthHandler)
	r.Handle("/metrics", promhttp.Handler())

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

func AuthHandler(rw http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	p := r.Form.Get("password")

	secret, err := ioutil.ReadFile("/secret-password/password")

	if err != nil {

		rw.Write([]byte("Can not find secret in /secret-password. Make sure it has been mounted"))
	}

	log.Printf("u: %s, p %s ", secret, p)

	ss := fmt.Sprintf("%s", secret)

	ss = strings.TrimSuffix(ss, "\n")

	if ss == p {

		rw.Write([]byte("<h1>You are authenticated</h1> Have fun!"))

	} else {
		rw.Write([]byte("<h1>You are not allowed to access this. Wrong Password.</h1> Are you trying to hxx0er me?"))
	}

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
		for i := 0; i < lines; i++ {
			log.Printf("Logging a lot: %d ", i)

		}
		d := time.Since(start)
		res := fmt.Sprintf("Logged %d lines in %.2f seconds", lines, d.Seconds())

		rw.Write([]byte(res))

	case "log1000":
		lines := 1000
		start := time.Now()
		for i := 0; i < lines; i++ {
			log.Printf("Logging a lot: %d ", i)

		}
		d := time.Since(start)
		res := fmt.Sprintf("Logged %d lines in %.2f seconds", lines, d.Seconds())

		rw.Write([]byte(res))

	case "log10000":
		lines := 10000
		start := time.Now()
		for i := 0; i < lines; i++ {
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

	case "metrics-increase":
		opsProcessed.Inc()

		rw.Write([]byte("clicks has been increased"))

	case "metrics-gauge-10":
		gauge.Set(10)
		rw.Write([]byte("ata_request_load set to 10"))

	case "metrics-gauge-50":
		gauge.Set(50)
		rw.Write([]byte("ata_request_load set to 50"))

	case "metrics-gauge-90":
		gauge.Set(90)
		rw.Write([]byte("ata_request_load set to 90"))

	case "tracing-flow1":
		span, ctx := opentracing.StartSpanFromContext(r.Context(), "awesome_business_function")
		defer span.Finish()

		time.Sleep(200 * time.Millisecond)

		if !BusinessFunction(ctx) {

			rw.Write([]byte("Request failed!"))

		} else {
			rw.Write([]byte("Request successful!"))
		}

	}

}

func BusinessFunction(ctx context.Context) bool {

	span, ctx := opentracing.StartSpanFromContext(ctx, "fetching_data")
	defer span.Finish()
	time.Sleep(100 * time.Millisecond)

	rand.Int()
	if rand.Intn(3) < 2 {
		span.SetTag("db.host", "dbserver1.middleware.se")
		re := rand.Intn(3)
		a := time.Duration(re)
		time.Sleep(a * time.Second)

		span.SetTag("db.records.retrieved", re*100)

		span2, _ := opentracing.StartSpanFromContext(ctx, "api_call")
		span2.SetTag("api.endpoint", "api.middleware.se")

		rea := rand.Intn(3)
		aa := time.Duration(rea)

		time.Sleep(aa * time.Millisecond * 100)
		span2.SetTag("api.response", 200)

		return true
	}

	span.SetTag("db.host", "dbserver2.middleware.se")
	span.SetTag("error", true)
	a := time.Duration(rand.Intn(4))
	time.Sleep(a * time.Second)

	return false

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
