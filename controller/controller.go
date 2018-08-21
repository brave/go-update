package controller

import (
	"encoding/xml"
	"fmt"
	"github.com/brave/go-update/extension"
	"github.com/go-chi/chi"
	"github.com/pressly/lg"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var allExtensionsMap = map[string]extension.Extension{}

// ExtensionsRouter is the router for /extensions endpoints
func ExtensionsRouter(extensions extension.Extensions) chi.Router {
	allExtensionsMap = extension.LoadExtensionsIntoMap(&extensions)
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

		foundExtension, ok := allExtensionsMap[id]
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
		_, ok := allExtensionsMap[updateRequest[0].ID]
		if !ok {
			http.Redirect(w, r, "https://update.googleapis.com/service/update2?braveRedirect=true", http.StatusTemporaryRedirect)
			return
		}
	}
	w.Header().Set("content-type", "application/xml")
	w.WriteHeader(http.StatusOK)
	updateResponse := updateRequest.FilterForUpdates(&allExtensionsMap)
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
