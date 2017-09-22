package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

var tokeninfoResp string = `
  {
    "key": "value"
  }
`

func handler(w http.ResponseWriter, r *http.Request) {
	var span opentracing.Span
	ctx, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	// hard code the api name
	url := "/oauth2/tokeninfo"
	if err != nil {
		span = opentracing.GlobalTracer().StartSpan(fmt.Sprintf("HTTP %s %s", r.Method, url))
	} else {
		span = opentracing.GlobalTracer().StartSpan(fmt.Sprintf("HTTP %s %s", r.Method, url), opentracing.ChildOf(ctx))
	}

	// simulate the real time cost job
	time.Sleep(20 * time.Millisecond)
	queryES(span)

	span.SetTag("http.status_code", "200")
	span.SetTag("param.Mozy", "Yes")
	span.LogKV(
		"waited.millis", 1500,
		"note", "something want to say",
	)

	defer span.Finish()
	fmt.Fprint(w, tokeninfoResp)
}

func queryES(parentSpan opentracing.Span) {
	childSpan := opentracing.GlobalTracer().StartSpan(
		"DRIVER:elacticsearch", opentracing.ChildOf(parentSpan.Context()),
	)

	// simulate the real time cost job
	time.Sleep(50 * time.Millisecond)

	childSpan.Finish()
}

func main() {
	// For OpenTracing initializing
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
			// FIXME : hard code the agent host here
			LocalAgentHostPort: "172.17.0.2:5775",
		},
	}
	tracer, closer, _ := cfg.New(
		"Oauth",
		config.Logger(jaeger.StdLogger),
	)
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	http.HandleFunc("/oauth2/tokeninfo", handler)
	http.ListenAndServe(":8080", nil)
}
