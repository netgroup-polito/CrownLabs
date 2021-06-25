package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

const (
	// FormatHeader name of the header used to extract the format.
	FormatHeader = "X-Format"

	// CodeHeader name of the header used as source of the HTTP status code to return.
	CodeHeader = "X-Code"

	// ContentType name of the header that defines the format of the reply.
	ContentType = "Content-Type"

	// OriginalURI name of the header with the original URL from NGINX.
	OriginalURI = "X-Original-URI"

	// Namespace name of the header that contains information about the Ingress namespace.
	Namespace = "X-Namespace"

	// IngressName name of the header that contains the matched Ingress.
	IngressName = "X-Ingress-Name"

	// ServiceName name of the header that contains the matched Service in the Ingress.
	ServiceName = "X-Service-Name"

	// ServicePort name of the header that contains the matched Service port in the Ingress.
	ServicePort = "X-Service-Port"

	// RequestID is a unique ID that identifies the request - same as for backend service.
	RequestID = "X-Request-ID"
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
	418: "I'm a teapot",
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

// DefaultResponseFormat is the default response format used in case no matches are found.
const DefaultResponseFormat = "text/html"

// SupportedResponseFormats is a map associating to each supported format the corresponding file extension.
var SupportedResponseFormats = map[string]string{
	"text/html":        "html",
	"text/plain":       "txt",
	"application/json": "json",
}

// ErrorPageData is the structure containing the values used to fill the response template.
type ErrorPageData struct {
	ErrorCode int
	ErrorMsg  string
}

func main() {
	var (
		httpAddress   string
		templatesPath string
	)

	// Flags initialization
	flag.StringVar(&httpAddress, "http-address", ":8080", "The address the server binds to.")
	flag.StringVar(&templatesPath, "templates-path", "/templates", "The path on disk where the templates are stored")

	klog.InitFlags(nil)
	flag.Parse()

	// Load response templates
	templates, err := loadTemplates(templatesPath)
	if err != nil {
		klog.Fatal("Failed to load response templates: ", err)
	}

	// Configure http handlers
	http.HandleFunc("/", errorHandler(templates))
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	klog.Infof("HTTP server listening at %v", httpAddress)
	if err := http.ListenAndServe(httpAddress, nil); err != nil {
		klog.Fatal("Failed starting the HTTP server: ", err)
	}
}

func errorHandler(templates map[string]*template.Template) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if os.Getenv("DEBUG") != "" {
			w.Header().Set(FormatHeader, r.Header.Get(FormatHeader))
			w.Header().Set(CodeHeader, r.Header.Get(CodeHeader))
			w.Header().Set(ContentType, r.Header.Get(ContentType))
			w.Header().Set(OriginalURI, r.Header.Get(OriginalURI))
			w.Header().Set(Namespace, r.Header.Get(Namespace))
			w.Header().Set(IngressName, r.Header.Get(IngressName))
			w.Header().Set(ServiceName, r.Header.Get(ServiceName))
			w.Header().Set(ServicePort, r.Header.Get(ServicePort))
			w.Header().Set(RequestID, r.Header.Get(RequestID))
		}

		// Get the error code to be displayed
		errCode := r.Header.Get(CodeHeader)
		code, err := strconv.Atoi(errCode)
		if err != nil {
			code = 400
			klog.Warningf("Unexpected error reading return code: %v. Using %v", err, code)
		}

		// Get the error message to be displayed
		message, ok := errorMessages[code]
		if !ok {
			message = errorMessages[0]
			klog.Warningf("Unknown error message for code %v. Using %v", code, message)
		}

		// Get the output format requested by the user
		format := r.Header.Get(FormatHeader)
		if _, ok := SupportedResponseFormats[format]; !ok {
			format = DefaultResponseFormat
		}

		// Configure the response content type
		w.Header().Set(ContentType, format)
		// Set the error code header
		w.WriteHeader(code)

		// Retrieve the appropriate response template
		tmpl := templates[format]

		// Build the ErrorPageData structure
		errorData := ErrorPageData{
			ErrorCode: code,
			ErrorMsg:  message,
		}

		// Serve the response
		klog.Infof("Error Code: %v - URI: '%v' - Ingress: '%v/%v' - Requested format: '%v' - Response format: '%v'",
			code, r.Header.Get(OriginalURI), r.Header.Get(Namespace), r.Header.Get(IngressName), r.Header.Get(FormatHeader), format)
		if err := tmpl.Execute(w, errorData); err != nil {
			klog.Error("Failed to prepare the response from the template: ", err)
			return
		}

		// Update metrics
		duration := time.Since(start).Seconds()
		proto := fmt.Sprintf("%v.%v", r.ProtoMajor, r.ProtoMinor)
		requestCount.WithLabelValues(proto).Inc()
		requestDuration.WithLabelValues(proto).Observe(duration)
	}
}

func loadTemplates(path string) (map[string]*template.Template, error) {
	var err error
	templates := map[string]*template.Template{}

	klog.Infof("Loading templates from %v", path)
	for format, extension := range SupportedResponseFormats {
		templateFile := fmt.Sprintf("%v/error-page.%v.tmpl", path, extension)

		klog.Infof("Loading template %v for format %v", templateFile, format)
		templates[format], err = template.ParseFiles(templateFile)
		if err != nil {
			klog.Errorf("Failed to open template file '%v': %v", templateFile, err)
			return nil, err
		}
		klog.Infof("Template %v correctly loaded", templateFile)
	}

	return templates, nil
}
