package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// FormatHeader name of the header used to extract the format
	FormatHeader = "X-Format"

	// CodeHeader name of the header used as source of the HTTP status code to return
	CodeHeader = "X-Code"

	// ContentType name of the header that defines the format of the reply
	ContentType = "Content-Type"

	// OriginalURI name of the header with the original URL from NGINX
	OriginalURI = "X-Original-URI"

	// Namespace name of the header that contains information about the Ingress namespace
	Namespace = "X-Namespace"

	// IngressName name of the header that contains the matched Ingress
	IngressName = "X-Ingress-Name"

	// ServiceName name of the header that contains the matched Service in the Ingress
	ServiceName = "X-Service-Name"

	// ServicePort name of the header that contains the matched Service port in the Ingress
	ServicePort = "X-Service-Port"

	// RequestId is a unique ID that identifies the request - same as for backend service
	RequestId = "X-Request-ID"

	// ServierPortVar is the name of the environment variable indicating
	// the port the http server is listening to.
	ServerPortVar = "ERROR_SERVER_PORT"

	// TemplatesPathVar is the name of the environment variable indicating
	// the location on disk of templates served by the handler.
	TemplatesPathVar = "ERROR_TEMPLATES_PATH"

	// StaticPathVar is the name of the environment variable indicating
	// the location on disk of static files served by the handler.
	StaticPathVar = "ERROR_STATIC_PATH"

	// StaticServePathVar is the name of the environment variable indicating
	// the path where the static resources are served by the handler.
	StaticServePathVar = "ERROR_STATIC_SERVE_PATH"

	// StaticUriVar is the name of the environment variable indicating URI where the static
	// associated with the error pages are hosted.
	StaticUriVar = "ERROR_STATIC_URI"
)

var errorMessages = map[int]string{
	0:   "Unknown Error",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Payload Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	416: "Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I\"m a teapot",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	510: "Not Extended",
	511: "Network Authentication Required",
}

func main() {

	// Variable initializations
	serverPort := "8080"
	if os.Getenv(ServerPortVar) != "" {
		serverPort = os.Getenv(ServerPortVar)
	}

	templatesPath := "/www/templates"
	if os.Getenv(TemplatesPathVar) != "" {
		templatesPath = os.Getenv(TemplatesPathVar)
	}
	log.Printf("[I] Local templates path: %v", templatesPath)

	staticPath := "/www/static"
	if os.Getenv(StaticPathVar) != "" && os.Getenv(StaticPathVar) != "/" {
		staticPath = os.Getenv(StaticPathVar)
	}
	log.Printf("[I] Local static resources path: %v", staticPath)

	StaticServePath := "/error-page"
	if os.Getenv(StaticServePathVar) != "" {
		StaticServePath = os.Getenv(StaticServePathVar)
	}
	log.Printf("{I] Static resources path: %v", staticPath)

	staticUri := StaticServePath
	if os.Getenv(StaticUriVar) != "" {
		staticUri = os.Getenv(StaticUriVar)
	}
	log.Printf("[I] Static resources URI: %v", staticUri)

	// Configure http handlers
	http.HandleFunc("/", errorHandler(templatesPath, staticUri))
	http.Handle(StaticServePath+"/", http.StripPrefix(StaticServePath+"/", http.FileServer(FileSystem{http.Dir(staticPath)})))
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("[I] Starting serving resources at :%v", serverPort)
	http.ListenAndServe(fmt.Sprintf(":"+serverPort), nil)
}

func errorHandler(templatesPath string, staticUri string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defaultFormat := "text/html"
		defaultExt := "html"

		if os.Getenv("DEBUG") != "" {
			w.Header().Set(FormatHeader, r.Header.Get(FormatHeader))
			w.Header().Set(CodeHeader, r.Header.Get(CodeHeader))
			w.Header().Set(ContentType, r.Header.Get(ContentType))
			w.Header().Set(OriginalURI, r.Header.Get(OriginalURI))
			w.Header().Set(Namespace, r.Header.Get(Namespace))
			w.Header().Set(IngressName, r.Header.Get(IngressName))
			w.Header().Set(ServiceName, r.Header.Get(ServiceName))
			w.Header().Set(ServicePort, r.Header.Get(ServicePort))
			w.Header().Set(RequestId, r.Header.Get(RequestId))
		}

		// Get the output format requested by the user
		format := r.Header.Get(FormatHeader)
		ext := defaultExt
		if format == "text/html" {
			ext = "html"
		} else if format == "application/json" {
			ext = "json"
		} else {
			format = defaultFormat
		}

		w.Header().Set(ContentType, format)

		// Open the error template
		templateFile := fmt.Sprintf("%v/error-page.%v.tmpl", templatesPath, ext)
		template, err := template.ParseFiles(templateFile)
		if err != nil {
			log.Printf("[E] Failed to open template file '%v': %v", templateFile, err)
			http.NotFound(w, r)
			return
		}

		// Get the error code to be displayed
		errCode := r.Header.Get(CodeHeader)
		code, err := strconv.Atoi(errCode)
		if err != nil {
			code = 400
			log.Printf("[E] Unexpected error reading return code: %v. Using %v", err, code)
		}

		// Get the error message to be displayed
		message, ok := errorMessages[code]
		if !ok {
			message = errorMessages[0]
			log.Printf("[E] Unknown error message for code %v. Using %v", code, message)
		}

		// Build the ErrorPageData structure
		errorData := ErrorPageData{
			ErrorCode: code,
			ErrorMsg:  message,
			StaticUri: staticUri,
		}

		// Set the error code header
		w.WriteHeader(code)

		// Serve the response
		log.Printf("[I] Error Code: %v - URI: %v - Ingress: %v/%v - Format: %v",
			code, r.Header.Get(OriginalURI), r.Header.Get(Namespace), r.Header.Get(IngressName), format)
		template.Execute(w, errorData)

		// Update metrics
		duration := time.Now().Sub(start).Seconds()

		proto := strconv.Itoa(r.ProtoMajor)
		proto = fmt.Sprintf("%s.%s", proto, strconv.Itoa(r.ProtoMinor))

		requestCount.WithLabelValues(proto).Inc()
		requestDuration.WithLabelValues(proto).Observe(duration)
	}
}

type ErrorPageData struct {
	ErrorCode int
	ErrorMsg  string
	StaticUri string
}

// Custom FileSystem wrapper to prevent showing directory information when serving static files
// https://gist.github.com/hauxe/f88a87f4037bca23f04f6d100f6e08d4#file-http_static_custom_http_server-go
type FileSystem struct {
	fs http.FileSystem
}

func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}
