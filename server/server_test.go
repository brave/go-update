package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brave/go-update/controller"
	"github.com/brave/go-update/extension"
	"github.com/brave/go-update/extension/extensiontest"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

var newExtension1 = extension.Extension{}
var newExtension2 = extension.Extension{}
var handler http.Handler

var contentTypeXML = "application/xml"
var contentTypeJSON = "application/json"

var lightThemeExtensionID = "ldimlcelhnjgpjjemdjokpgeeikdinbm"
var darkThemeExtensionID = "bfdgpgibhagkpdlnjonhkabjoijopoge"

var newExtensionID1 = "newext1eplbcioakkpcpgfkobkghlhen"
var newExtensionID2 = "newext2eplbcioakkpcpgfkobkghlhen"

func init() {
	newExtension1 = extension.Extension{
		ID:          newExtensionID1,
		Blacklisted: false,
		SHA256:      "4c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618",
		Title:       "test",
		Version:     "1.0.0",
	}

	newExtension2 = extension.Extension{
		ID:          newExtensionID2,
		Blacklisted: false,
		SHA256:      "3c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618",
		Title:       "test",
		Version:     "1.0.0",
	}

	// Setup refreshing extensions with a new extension that we'll check for later
	// We maintain a count to make sure the refresh function is called more than just
	// the first time.
	count := 0
	controller.AllExtensionsMap = extension.NewExtensionMap()
	controller.AllExtensionsMap.StoreExtensions(&extension.OfferedExtensions)
	controller.ExtensionUpdaterTimeout = time.Millisecond * 1
	testCtx, logger := setupLogger(context.Background())
	handler = chi.ServerBaseContext(setupRouter(testCtx, logger, true))
	controller.RefreshExtensionsTicker(func() {
		count++
		if count == 1 {
			controller.AllExtensionsMap.Store(newExtensionID1, newExtension1)
		} else if count == 2 {
			controller.AllExtensionsMap.Store(newExtensionID2, newExtension2)
		}
	})
}

func TestPing(t *testing.T) {
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, err := http.Get(server.URL)
	assert.Nil(t, err)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Received non-200 response: %d\n", resp.StatusCode)
	}
	expected := "."
	actual, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	if expected != string(actual) {
		t.Errorf("Expected the message '%s'\n", expected)
	}
}

func testCall(t *testing.T, server *httptest.Server, method string, contentType string, query string,
	requestBody string, expectedResponseCode int, expectedResponse string, redirectLocation string) {
	extensionsURL := fmt.Sprintf("%s/extensions%s", server.URL, query)
	req, err := http.NewRequest(method, extensionsURL, bytes.NewBuffer([]byte(requestBody)))
	assert.Nil(t, err)
	req.Header.Add("Content-Type", contentType)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	assert.Nil(t, err)

	assert.Equal(t, expectedResponseCode, resp.StatusCode)

	// If this is a redirect, ensure the protocol is HTTPS.
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		assert.Equal(t, redirectLocation, location)
	}

	actual, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, expectedResponse, strings.TrimSpace(string(actual)))
}

func TestUpdateExtensionsXML(t *testing.T) {
	server := httptest.NewServer(handler)
	defer server.Close()

	// No extensions
	requestBody := `
		<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
		  <hw physmemory="16"/>
		  <os platform="Mac OS X" version="10.11.6" arch="x86_64"/>
		</request>`
	expectedResponse := "<response protocol=\"3.1\" server=\"prod\"></response>"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Unsupported protocol version
	requestBody =
		`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="2.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
			<app appid="aomjjhallfgjeglblehebfpbcfeobpgk">
				<updatecheck codebase="https://` + extension.GetS3ExtensionBucketHost("aomjjhallfgjeglblehebfpbcfeobpgk") + `/release/aomjjhallfgjeglblehebfpbcfeobpgk/extension_4_5_9_90.crx" version="4.5.9.90"/>
			</app>
		</request>`
	expectedResponse = "Error reading body request version: 2.0 not supported"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Not XML
	requestBody = "For the king!"
	expectedResponse = "Error reading body EOF"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Malformed XML
	requestBody = "<This way! No, that way!"
	expectedResponse = "Error reading body XML syntax error on line 1: attribute name without = in element"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Different XML schema
	requestBody = "<text>For the alliance!</text>"
	expectedResponse = "Error reading body expected element type <request> but have <text>"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Empty body request
	requestBody = ""
	expectedResponse = "Error reading body EOF"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	lightThemeExtension := extensiontest.ExtensionRequestFnForXML("ldimlcelhnjgpjjemdjokpgeeikdinbm")

	// Single extension out of date
	requestBody = lightThemeExtension("0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single extension same version
	requestBody = lightThemeExtension("1.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="noupdate"></updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single extension greater version
	requestBody = lightThemeExtension("2.0.0")
	expectedResponse = "<response protocol=\"3.1\" server=\"prod\"></response>"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	lightAndDarkThemeRequest := extensiontest.ExtensionRequestFnForTwoXML("ldimlcelhnjgpjjemdjokpgeeikdinbm", "bfdgpgibhagkpdlnjonhkabjoijopoge")

	// Multiple components with none out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "70.0.0")
	expectedResponse = "<response protocol=\"3.1\" server=\"prod\"></response>"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Only one components out of date
	requestBody = lightAndDarkThemeRequest("0.0.0", "70.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Other component of 2 out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(darkThemeExtensionID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Both components need updates
	requestBody = lightAndDarkThemeRequest("0.0.0", "0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(darkThemeExtensionID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Unknown extension ID goes to Google server via componentupdater proxy
	requestBody = extensiontest.ExtensionRequestFnForXML("aaaaaaaaaaaaaaaaaaaa")("0.0.0")
	expectedResponse = ""
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusTemporaryRedirect, expectedResponse, "https://componentupdater.brave.com/service/update2")

	// Unknown extension ID goes to Google server via componentupdater proxy
	// and preserves query params
	requestBody = extensiontest.ExtensionRequestFnForXML("aaaaaaaaaaaaaaaaaaaa")("0.0.0")
	expectedResponse = ""
	testCall(t, server, http.MethodPost, contentTypeXML, "?test=hi", requestBody, http.StatusTemporaryRedirect, expectedResponse, "https://componentupdater.brave.com/service/update2?test=hi")

	// Requests for widevine should use update.googleapis.com directly without any proxy
	requestBody = extensiontest.ExtensionRequestFnForXML(controller.WidivineExtensionID)("0.0.0")
	expectedResponse = ""
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusTemporaryRedirect, expectedResponse, "https://update.googleapis.com/service/update2")

	// Make sure a huge request body does not crash the server
	data := make([]byte, 1024*1024*11) // 11 MiB
	_, err := rand.Read(data)
	assert.Nil(t, err)
	requestBody = string(data)
	expectedResponse = "Request too large"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Single new extension out of date that was added in by the refresh timer
	requestBody = extensiontest.ExtensionRequestFnForXML("newext1eplbcioakkpcpgfkobkghlhen")("0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="newext1eplbcioakkpcpgfkobkghlhen">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(newExtensionID1) + `/release/newext1eplbcioakkpcpgfkobkghlhen/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="4c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single second new extension out of date that was added in by the refresh timer
	requestBody = extensiontest.ExtensionRequestFnForXML("newext2eplbcioakkpcpgfkobkghlhen")("0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="newext2eplbcioakkpcpgfkobkghlhen">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(newExtensionID2) + `/release/newext2eplbcioakkpcpgfkobkghlhen/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="3c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")
}

func getQueryParams(extension *extension.Extension) string {
	return `x=id%3D` + extension.ID + `%26v%3D` + extension.Version
}

func TestWebStoreUpdateExtensionXML(t *testing.T) {
	server := httptest.NewServer(handler)
	defer server.Close()

	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Empty query param request, no extensions.
	requestBody := ""
	query := ""
	expectedResponse := `<gupdate protocol="3.1" server="prod"></gupdate>`
	testCall(t, server, http.MethodGet, contentTypeXML, query, requestBody, http.StatusOK, expectedResponse, "")

	// Extension that we handle which is outdated should produce a response
	outdatedLightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	outdatedLightThemeExtension.Version = "0.0.0"
	assert.True(t, ok)
	query = "?" + getQueryParams(&outdatedLightThemeExtension)
	expectedResponse = `<gupdate protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm" status="ok">
        <updatecheck status="ok" codebase="https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx" version="1.0.0" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"></updatecheck>
    </app>
</gupdate>`
	testCall(t, server, http.MethodGet, contentTypeXML, query, requestBody, http.StatusOK, expectedResponse, "")

	// Multiple extensions that we handle which are outdated should produce a response
	outdatedDarkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	outdatedDarkThemeExtension.Version = "0.0.0"
	query = "?" + getQueryParams(&outdatedLightThemeExtension) + "&" + getQueryParams(&outdatedDarkThemeExtension)
	expectedResponse = `<gupdate protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm" status="ok">
        <updatecheck status="ok" codebase="https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx" version="1.0.0" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"></updatecheck>
    </app>
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge" status="ok">
        <updatecheck status="ok" codebase="https://` + extension.GetS3ExtensionBucketHost(darkThemeExtensionID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx" version="1.0.0" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"></updatecheck>
    </app>
</gupdate>`
	testCall(t, server, http.MethodGet, contentTypeXML, query, requestBody, http.StatusOK, expectedResponse, "")

	// Extension that we handle which is up to date should NOT produce an update but still be successful
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	query = "?" + getQueryParams(&lightThemeExtension)
	expectedResponse = `<gupdate protocol="3.1" server="prod"></gupdate>`
	testCall(t, server, http.MethodGet, contentTypeXML, query, requestBody, http.StatusOK, expectedResponse, "")

	// Unknown extension ID goes to Google server
	unknownExtension := extension.Extension{
		ID:      "aaaaaaaaaaaaaaaaaaaa",
		Version: "0.0.0",
	}
	query = "?" + getQueryParams(&unknownExtension)
	expectedResponse = `<a href="https://extensionupdater.brave.com/service/update2/crx?x=id%3Daaaaaaaaaaaaaaaaaaaa%26v%3D0.0.0">Temporary Redirect</a>.`
	testCall(t, server, http.MethodGet, contentTypeXML, query, requestBody, http.StatusTemporaryRedirect, expectedResponse, "https://extensionupdater.brave.com/service/update2/crx?x=id%3Daaaaaaaaaaaaaaaaaaaa%26v%3D0.0.0")

	// Unknown extension ID with multiple extensions, we try to handle ourselves.
	unknownExtension = extension.Extension{
		ID:      "aaaaaaaaaaaaaaaaaaaa",
		Version: "0.0.0",
	}
	unknownExtension2 := extension.Extension{
		ID:      "bbaaaaaaaaaaaaaaaaaa",
		Version: "0.0.0",
	}
	query = "?" + getQueryParams(&unknownExtension) + "&" + getQueryParams(&unknownExtension2)
	expectedResponse = `<gupdate protocol="3.1" server="prod"></gupdate>`
	testCall(t, server, http.MethodGet, contentTypeXML, query, requestBody, http.StatusOK, expectedResponse, "")
}

func TestUpdateExtensionsJSON(t *testing.T) {
	jsonPrefix := ")]}'\n"
	server := httptest.NewServer(handler)
	defer server.Close()

	// No extensions
	requestBody := `{"request":{"protocol":"3.1","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`
	expectedResponse := jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":null}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Unsupported protocol version
	requestBody = `{"request":{"protocol":"2.0","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`
	expectedResponse = "Error reading body request version: 2.0 not supported"
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Not JSON
	requestBody = "For the king!"
	expectedResponse = "Error reading body invalid character 'F' looking for beginning of value"
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Malformed JSON
	requestBody = "{request"
	expectedResponse = "Error reading body invalid character 'r' looking for beginning of object key string"
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	lightThemeExtension := extensiontest.ExtensionRequestFnForJSON("ldimlcelhnjgpjjemdjokpgeeikdinbm")

	// Single extension out of date
	requestBody = lightThemeExtension("0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single extension same version
	requestBody = lightThemeExtension("1.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"noupdate"}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single extension greater version
	requestBody = lightThemeExtension("2.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":null}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	lightAndDarkThemeRequest := extensiontest.ExtensionRequestFnForTwoJSON("ldimlcelhnjgpjjemdjokpgeeikdinbm", "bfdgpgibhagkpdlnjonhkabjoijopoge")

	// Multiple components with none out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "70.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":null}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Only one components out of date
	requestBody = lightAndDarkThemeRequest("0.0.0", "70.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Other component of 2 out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtensionID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Both components need updates
	requestBody = lightAndDarkThemeRequest("0.0.0", "0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtensionID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Unknown extension ID goes to Google server via componentupdater proxy
	requestBody = extensiontest.ExtensionRequestFnForJSON("aaaaaaaaaaaaaaaaaaaa")("0.0.0")
	expectedResponse = ""
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusTemporaryRedirect, expectedResponse, "https://componentupdater.brave.com/service/update2/json")

	// Unknown extension ID goes to Google server via componentupdater proxy
	// and preserves query params
	requestBody = extensiontest.ExtensionRequestFnForJSON("aaaaaaaaaaaaaaaaaaaa")("0.0.0")
	expectedResponse = ""
	testCall(t, server, http.MethodPost, contentTypeJSON, "?test=hi", requestBody, http.StatusTemporaryRedirect, expectedResponse, "https://componentupdater.brave.com/service/update2/json?test=hi")

	// Make sure a huge request body does not crash the server
	data := make([]byte, 1024*1024*11) // 11 MiB
	_, err := rand.Read(data)
	assert.Nil(t, err)
	requestBody = string(data)
	expectedResponse = "Request too large"
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Single new extension out of date that was added in by the refresh timer
	requestBody = extensiontest.ExtensionRequestFnForJSON("newext1eplbcioakkpcpgfkobkghlhen")("0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"newext1eplbcioakkpcpgfkobkghlhen","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(newExtensionID1) + `/release/newext1eplbcioakkpcpgfkobkghlhen/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","hash_sha256":"4c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single second new extension out of date that was added in by the refresh timer
	requestBody = extensiontest.ExtensionRequestFnForJSON("newext2eplbcioakkpcpgfkobkghlhen")("0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"newext2eplbcioakkpcpgfkobkghlhen","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(newExtensionID2) + `/release/newext2eplbcioakkpcpgfkobkghlhen/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","hash_sha256":"3c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")
}

func TestWebStoreUpdateExtensionJSON(t *testing.T) {
	server := httptest.NewServer(handler)
	defer server.Close()

	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Empty query param request, no extensions.
	requestBody := ""
	query := ""
	expectedResponse := `{"gupdate":{"protocol":"3.1","server":"prod","app":null}}`
	testCall(t, server, http.MethodGet, contentTypeJSON, query, requestBody, http.StatusOK, expectedResponse, "")

	// Extension that we handle which is outdated should produce a response
	outdatedLightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	outdatedLightThemeExtension.Version = "0.0.0"
	assert.True(t, ok)
	query = "?" + getQueryParams(&outdatedLightThemeExtension)
	expectedResponse = `{"gupdate":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"}}]}}`
	testCall(t, server, http.MethodGet, contentTypeJSON, query, requestBody, http.StatusOK, expectedResponse, "")

	// Multiple extensions that we handle which are outdated should produce a response
	outdatedDarkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	outdatedDarkThemeExtension.Version = "0.0.0"
	query = "?" + getQueryParams(&outdatedLightThemeExtension) + "&" + getQueryParams(&outdatedDarkThemeExtension)
	expectedResponse = `{"gupdate":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtensionID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"}}]}}`
	testCall(t, server, http.MethodGet, contentTypeJSON, query, requestBody, http.StatusOK, expectedResponse, "")

	// Extension that we handle which is up to date should NOT produce an update but still be successful
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	query = "?" + getQueryParams(&lightThemeExtension)
	expectedResponse = `{"gupdate":{"protocol":"3.1","server":"prod","app":null}}`
	testCall(t, server, http.MethodGet, contentTypeJSON, query, requestBody, http.StatusOK, expectedResponse, "")

	// Unknown extension ID goes to Google server
	unknownExtension := extension.Extension{
		ID:      "aaaaaaaaaaaaaaaaaaaa",
		Version: "0.0.0",
	}
	query = "?" + getQueryParams(&unknownExtension)
	expectedResponse = "<a href=\"https://extensionupdater.brave.com/service/update2/crx?x=id%3Daaaaaaaaaaaaaaaaaaaa%26v%3D0.0.0\">Temporary Redirect</a>."
	testCall(t, server, http.MethodGet, contentTypeJSON, query, requestBody, http.StatusTemporaryRedirect, expectedResponse, "https://extensionupdater.brave.com/service/update2/crx?x=id%3Daaaaaaaaaaaaaaaaaaaa%26v%3D0.0.0")

	// Unknown extension ID with multiple extensions, we try to handle ourselves.
	unknownExtension = extension.Extension{
		ID:      "aaaaaaaaaaaaaaaaaaaa",
		Version: "0.0.0",
	}
	unknownExtension2 := extension.Extension{
		ID:      "bbaaaaaaaaaaaaaaaaaa",
		Version: "0.0.0",
	}
	query = "?" + getQueryParams(&unknownExtension) + "&" + getQueryParams(&unknownExtension2)
	expectedResponse = `{"gupdate":{"protocol":"3.1","server":"prod","app":null}}`
	testCall(t, server, http.MethodGet, contentTypeJSON, query, requestBody, http.StatusOK, expectedResponse, "")
}

func TestPrintExtensions(t *testing.T) {
	server := httptest.NewServer(handler)
	defer server.Close()

	testURL := fmt.Sprintf("%s/extensions/test", server.URL)
	req, err := http.NewRequest(http.MethodGet, testURL, bytes.NewBuffer([]byte("")))
	assert.Nil(t, err)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	actual, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(actual), "ldimlcelhnjgpjjemdjokpgeeikdinbm"))

	// Clear out the extensions map.
	controller.AllExtensionsMap = extension.NewExtensionMap()
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	actual, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, string(actual), "{}")
}
