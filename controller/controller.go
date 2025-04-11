package controller

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brave/go-update/extension"
	"github.com/brave/go-update/omaha"
	"github.com/brave/go-update/omaha/protocol"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"github.com/pressly/lg"
)

// PDFJSExtensionID will be used to add an exception to pass the request for
// PDF viewer extension install from chrome web store to the extension updater proxy
var PDFJSExtensionID = "oemmndcbldboiebfnladdacbdfmadadm"

// WidivineExtensionID is used to add an exception to pass the request for widivine
// directly to google servers
var WidivineExtensionID = "oimompecagnajdejgnnjijobebaeigek"

// AllExtensionsMap holds a mapping of extension ID to extension object.
// This list for tests is populated by extensions.OfferedExtensions.
// For normal operations of this server it is obtained from the AWS config
// of the host machine for DynamoDB.
var AllExtensionsMap = extension.NewExtensionMap()

// ExtensionUpdaterTimeout is the amount of time to wait between getting new updates from DynamoDB for the list of extensions
var ExtensionUpdaterTimeout = time.Minute * 10

// ProtocolFactory is the factory used to create protocol handlers
var ProtocolFactory = &omaha.DefaultFactory{}

func initExtensionUpdatesFromDynamoDB() {
	awsConfig := &aws.Config{}
	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		log.Printf("failed to connect to new session %v\n", err)
		sentry.CaptureException(err)
		return
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	params := &dynamodb.ScanInput{
		TableName: aws.String("Extensions"),
	}

	// For most use cases, you probably wouldn't want to scan all entries; however,
	// for our use case we have a read only small number of items, that are infrequently
	// updated, usually less than daily by an external tool, and very often queried.
	result, err := svc.Scan(params)
	if err != nil {
		log.Printf("failed to make Scan API call %v\n", err)
		sentry.CaptureException(err)
		return
	}

	// Update the extensions map
	for _, item := range result.Items {
		id := *item["ID"].S

		ext := extension.Extension{
			ID:          id,
			Blacklisted: *item["Disabled"].BOOL,
			SHA256:      *item["SHA256"].S,
			Title:       *item["Title"].S,
			Version:     *item["Version"].S,
		}

		if plist := item["PatchList"]; plist != nil {
			var pinfo map[string]*extension.PatchInfo
			if err := dynamodbattribute.UnmarshalMap(plist.M, &pinfo); err != nil {
				log.Printf("failed to parse PatchList %v\n", err)
				sentry.CaptureException(err)
			} else {
				ext.PatchList = pinfo
			}
		}

		AllExtensionsMap.Store(id, ext)

	}
}

// RefreshExtensionsTicker updates the list of extensions by
// calling the specified extensionMapUpdater function
func RefreshExtensionsTicker(extensionMapUpdater func()) {
	extensionMapUpdater()
	ticker := time.NewTicker(ExtensionUpdaterTimeout)
	go func() {
		for range ticker.C {
			extensionMapUpdater()
		}
	}()
}

// ExtensionsRouter is the router for /extensions endpoints
func ExtensionsRouter(_ extension.Extensions, testRouter bool) chi.Router {
	if !testRouter {
		RefreshExtensionsTicker(initExtensionUpdatesFromDynamoDB)
	}

	r := chi.NewRouter()
	r.Post("/", UpdateExtensions)
	r.Get("/", WebStoreUpdateExtension)
	r.Get("/test", PrintExtensions)
	return r
}

// PrintExtensions is just used for troubleshooting to see what the internal list of extensions DB holds
// It simply prints out text for all extensions when visiting /extensions/test.
// Since our internally maintained list is always small by design, this is not a big deal for performance.
func PrintExtensions(w http.ResponseWriter, r *http.Request) {
	log := lg.Log(r.Context())
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	data, err := AllExtensionsMap.MarshalJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error in marshal %v", err), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		log.Errorf("Error writing response for printing extensions: %v", err)
	}
}

// WebStoreUpdateExtension is the handler for updating extensions made via the GET HTTP method.
// Get requests look like this:
// /extensions?os=mac&arch=x64&os_arch=x86_64&nacl_arch=x86-64&prod=chromiumcrx&prodchannel=&prodversion=69.0.54.0&lang=en-US&acceptformat=crx2,crx3&x=id%3Doemmndcbldboiebfnladdacbdfmadadm%26v%3D0.0.0.0%26installedby%3Dpolicy%26uc%26ping%3Dr%253D-1%2526e%253D1"
// The query parameter x contains the encoded extension information, there can be more than one x parameter.
func WebStoreUpdateExtension(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("content-type")

	log := lg.Log(r.Context())
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Errorf("Error closing body stream: %v", err)
		}
	}()

	xValues := r.URL.Query()["x"]
	webStoreResponse := extension.Extensions{}

	AllExtensionsMap.RLock()
	defer AllExtensionsMap.RUnlock()

	for _, x := range xValues {
		unescaped, err := url.QueryUnescape(x)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error unescaping query parameters: %v", err), http.StatusBadRequest)
			return
		}
		values, err := url.ParseQuery(unescaped)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing query parameters: %v", err), http.StatusBadRequest)
			return
		}

		id := strings.Trim(values.Get("id"), "[]")
		v := values.Get("v")
		if len(id) == 0 {
			http.Error(w, "No extension ID specified.", http.StatusBadRequest)
			return
		}

		foundExtension, ok := AllExtensionsMap.Load(id)
		if (!ok || id == PDFJSExtensionID) && len(xValues) == 1 {
			http.Redirect(w, r, "https://extensionupdater.brave.com/service/update2/crx?"+r.URL.RawQuery, http.StatusTemporaryRedirect)
			return
		}

		// We dont have any Brave Extensions yet, so this part of the code is not tested
		if extension.CompareVersions(v, foundExtension.Version) < 0 {
			webStoreResponse = append(webStoreResponse, extension.Extension{
				ID:      foundExtension.ID,
				Version: foundExtension.Version,
				SHA256:  foundExtension.SHA256,
			})
		}
	}

	w.WriteHeader(http.StatusOK)

	// It is impossible to determine the response protocol version for WebStoreUpdateExtension calls.
	// The incoming request is GET and does not include any information about the protocol version,
	// therefore 3.1 is always used for WebStoreUpdateExtension responses.
	protocolHandler, err := ProtocolFactory.CreateProtocol("3.1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating protocol handler: %v", err), http.StatusInternalServerError)
		return
	}

	data, err := protocolHandler.FormatWebStoreResponse(webStoreResponse, contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error formatting response: %v", err), http.StatusInternalServerError)
		return
	}

	// Set content type
	if protocol.IsJSONRequest(contentType) {
		w.Header().Set("content-type", "application/json")
	} else {
		w.Header().Set("content-type", "application/xml")
	}

	_, err = w.Write(data)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}

// UpdateExtensions is the handler for updating extensions
func UpdateExtensions(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("content-type")
	jsonPrefix := []byte(")]}'\n")

	log := lg.Log(r.Context())
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Errorf("Error closing body stream: %v", err)
		}
	}()

	limit := int64(1024 * 1024 * 10) // 10MiB
	body, err := io.ReadAll(io.LimitReader(r.Body, limit))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
		return
	}
	if len(body) == int(limit) {
		http.Error(w, "Request too large", http.StatusBadRequest)
		return
	}

	// Special case for empty body
	if len(body) == 0 {
		http.Error(w, "Error parsing request: EOF", http.StatusBadRequest)
		return
	}

	// Use protocol handler to format response
	protocolVersion, err := protocol.DetectProtocolVersion(body, contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
		return
	}

	// The validation now happens inside CreateProtocol, so we don't need a separate check here
	protocolHandler, err := ProtocolFactory.CreateProtocol(protocolVersion)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
		return
	}

	// Parse the request
	updateRequest, err := protocolHandler.ParseRequest(body, contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
		return
	}

	// Special case, if there's only 1 extension in the request and it is not something
	// we know about, redirect the client to google component update server.
	if len(updateRequest) == 1 {
		_, ok := AllExtensionsMap.Load(updateRequest[0].ID)
		if !ok {
			queryString := ""
			if len(r.URL.RawQuery) != 0 {
				queryString = "?" + r.URL.RawQuery
			}
			host := extension.GetComponentUpdaterHost()
			if updateRequest[0].ID == WidivineExtensionID {
				host = "update.googleapis.com"
			}
			if protocol.IsJSONRequest(contentType) {
				http.Redirect(w, r, "https://"+host+"/service/update2/json"+queryString, http.StatusTemporaryRedirect)
			} else {
				http.Redirect(w, r, "https://"+host+"/service/update2"+queryString, http.StatusTemporaryRedirect)
			}
			return
		}
	}

	// Set content type header
	if protocol.IsJSONRequest(contentType) {
		w.Header().Set("content-type", "application/json")
	} else {
		w.Header().Set("content-type", "application/xml")
	}

	w.WriteHeader(http.StatusOK)

	// Use the generic FilterForUpdates function
	updateResponse := extension.FilterForUpdates(updateRequest, AllExtensionsMap)

	// Use the same protocol version for response as the request for v4
	// Otherwise default to 3.1 for backward compatibility
	responseProtocolVersion := "3.1"
	if protocolVersion == "4.0" {
		responseProtocolVersion = protocolVersion
	}

	responseProtocolHandler, err := ProtocolFactory.CreateProtocol(responseProtocolVersion)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating response protocol handler: %v", err), http.StatusInternalServerError)
		return
	}

	data, err := responseProtocolHandler.FormatUpdateResponse(updateResponse, contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error formatting response: %v", err), http.StatusInternalServerError)
		return
	}

	if protocol.IsJSONRequest(contentType) {
		data = append(jsonPrefix, data...)
	}

	_, err = w.Write(data)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}
