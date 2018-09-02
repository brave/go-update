package controller

import (
	"encoding/xml"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/brave/go-update/extension"
	"github.com/getsentry/raven-go"
	"github.com/go-chi/chi"
	"github.com/pressly/lg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AllExtensionsMap holds a mapping of extension ID to extension object.
// This list for tests is populated by extensions.OfferedExtensions.
// For normal operaitons of this server it is obtained from the AWS config
// of the host machine for DynamoDB.
var AllExtensionsMap = map[string]extension.Extension{}

// ExtensionUpdaterTimeout is the amount of time to wait between getting new updates from DynamoDB for the list of extensions
var ExtensionUpdaterTimeout = time.Minute * 10

func initExtensionUpdatesFromDynamoDB() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		log.Printf("failed to connect to new session %v\n", err)
		raven.CaptureError(err, nil)
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
		raven.CaptureError(err, nil)
		return
	}

	// Update the extensions map
	for _, item := range result.Items {
		id := *item["ID"].S
		AllExtensionsMap[id] = extension.Extension{
			ID:          id,
			Blacklisted: *item["Disabled"].BOOL,
			SHA256:      *item["SHA256"].S,
			Title:       *item["Title"].S,
			Version:     *item["Version"].S,
		}
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
func ExtensionsRouter(extensions extension.Extensions) chi.Router {
	RefreshExtensionsTicker(initExtensionUpdatesFromDynamoDB)
	r := chi.NewRouter()
	r.Post("/", UpdateExtensions)
	r.Get("/", WebStoreUpdateExtension)
	return r
}

// WebStoreUpdateExtension is the handler for updating extensions made via the GET HTTP methhod.
// Get requests look like this:
// /extensions?os=mac&arch=x64&os_arch=x86_64&nacl_arch=x86-64&prod=chromiumcrx&prodchannel=&prodversion=69.0.54.0&lang=en-US&acceptformat=crx2,crx3&x=id%3Doemmndcbldboiebfnladdacbdfmadadm%26v%3D0.0.0.0%26installedby%3Dpolicy%26uc%26ping%3Dr%253D-1%2526e%253D1"
// The query parameter x contains the encoded extension information, there can be more than one x parameter.
func WebStoreUpdateExtension(w http.ResponseWriter, r *http.Request) {
	log := lg.Log(r.Context())
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Errorf("Error closing body stream: %v", err)
		}
	}()

	xValues := r.URL.Query()["x"]
	webStoreResponse := extension.WebStoreUpdateResponse{}
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
			http.Error(w, fmt.Sprintf("No extension ID specified."), http.StatusBadRequest)
			return
		}

		foundExtension, ok := AllExtensionsMap[id]
		if !ok && len(xValues) == 1 {
			http.Redirect(w, r, "https://clients2.google.com/service/update2/crx?"+r.URL.RawQuery+"&braveRedirect=true", http.StatusTemporaryRedirect)
			return
		}
		if extension.CompareVersions(v, foundExtension.Version) < 0 {
			webStoreResponse = append(webStoreResponse, extension.Extension{
				ID:      foundExtension.ID,
				Version: foundExtension.Version,
				SHA256:  foundExtension.SHA256,
			})
		}
	}

	w.Header().Set("content-type", "application/xml")
	w.WriteHeader(http.StatusOK)
	data, err := xml.Marshal(&webStoreResponse)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error in marshal XML %v", err), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}

// UpdateExtensions is the handler for updating extensions
func UpdateExtensions(w http.ResponseWriter, r *http.Request) {
	log := lg.Log(r.Context())
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Errorf("Error closing body stream: %v", err)
		}
	}()

	limit := int64(1024 * 1024 * 10) // 10MiB
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, limit))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
		return
	}
	if len(body) == int(limit) {
		http.Error(w, "Request too large", http.StatusBadRequest)
		return
	}

	updateRequest := extension.UpdateRequest{}
	err = xml.Unmarshal(body, &updateRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading body %v", err), http.StatusBadRequest)
		return
	}
	// Special case, if there's only 1 extension in the request and it is not something
	// we know about, redirect the client to google component update server.
	if len(updateRequest) == 1 {
		_, ok := AllExtensionsMap[updateRequest[0].ID]
		if !ok {
			http.Redirect(w, r, "https://update.googleapis.com/service/update2?braveRedirect=true", http.StatusTemporaryRedirect)
			return
		}
	}
	w.Header().Set("content-type", "application/xml")
	w.WriteHeader(http.StatusOK)
	updateResponse := updateRequest.FilterForUpdates(&AllExtensionsMap)
	data, err := xml.Marshal(&updateResponse)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error in marshal XML %v", err), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}
