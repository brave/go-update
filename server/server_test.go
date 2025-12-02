package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brave/go-update/controller"
	"github.com/brave/go-update/extension"
	"github.com/brave/go-update/extension/extensiontest"
	"github.com/brave/go-update/logger"
	"github.com/stretchr/testify/assert"
)

var (
	newExtension1 = extension.Extension{}
	newExtension2 = extension.Extension{}
	handler       http.Handler
)

var (
	contentTypeXML  = "application/xml"
	contentTypeJSON = "application/json"
)

var (
	lightThemeExtensionID = "ldimlcelhnjgpjjemdjokpgeeikdinbm"
	darkThemeExtensionID  = "bfdgpgibhagkpdlnjonhkabjoijopoge"
)

var (
	newExtensionID1 = "newext1eplbcioakkpcpgfkobkghlhen"
	newExtensionID2 = "newext2eplbcioakkpcpgfkobkghlhen"
)

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
	serverCtx, log := logger.Setup(context.Background())
	_, router := setupRouter(serverCtx, true)

	// Create a middleware that adds the context with logger to each request
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the logger to the context
		ctx := logger.WithContext(r.Context(), log)
		router.ServeHTTP(w, r.WithContext(ctx))
	})

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
	actual, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	if expected != string(actual) {
		t.Errorf("Expected the message '%s'\n", expected)
	}
}

func testCall(t *testing.T, server *httptest.Server, method string, contentType string, query string,
	requestBody string, expectedResponseCode int, expectedResponse string, redirectLocation string,
) {
	extensionsURL := fmt.Sprintf("%s/extensions%s", server.URL, query)
	req, err := http.NewRequest(method, extensionsURL, bytes.NewBuffer([]byte(requestBody)))
	assert.Nil(t, err)
	req.Header.Add("Content-Type", contentType)

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
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

func TestUpdateExtensionsXMLV3(t *testing.T) {
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
	requestBody = `
		<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="2.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
			<app appid="aomjjhallfgjeglblehebfpbcfeobpgk">
				<updatecheck codebase="https://` + extension.GetS3ExtensionBucketHost("aomjjhallfgjeglblehebfpbcfeobpgk") + `/release/aomjjhallfgjeglblehebfpbcfeobpgk/extension_4_5_9_90.crx" version="4.5.9.90"/>
			</app>
		</request>`
	expectedResponse = "Error parsing request: unsupported protocol version: 2.0"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Not XML
	requestBody = "For the king!"
	expectedResponse = "Error parsing request: error parsing XML: EOF"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Malformed XML
	requestBody = "<This way! No, that way!"
	expectedResponse = "Error parsing request: error parsing XML: XML syntax error on line 1: attribute name without = in element"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Different XML schema
	requestBody = "<text>For the alliance!</text>"
	expectedResponse = "Error parsing request: error parsing XML: expected element type <request> but have <text>"
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Empty body request
	requestBody = ""
	expectedResponse = "Error parsing request: EOF"
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
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="noupdate"></updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	lightAndDarkThemeRequest := extensiontest.ExtensionRequestFnForTwoXML("ldimlcelhnjgpjjemdjokpgeeikdinbm", "bfdgpgibhagkpdlnjonhkabjoijopoge")

	// Multiple components with none out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "70.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="noupdate"></updatecheck>
    </app>
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge">
        <updatecheck status="noupdate"></updatecheck>
    </app>
</response>`
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
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge">
        <updatecheck status="noupdate"></updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Other component of 2 out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="noupdate"></updatecheck>
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
	requestBody = extensiontest.ExtensionRequestFnForXML(controller.WidevineExtensionID)("0.0.0")
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

	// Test mixed extension statuses in XML - one outdated, one current, one unknown
	threeExtensionXMLRequest := func(lightVersion, darkVersion, unknownVersion string) string {
		return `
		<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
		  <hw physmemory="16"/>
		  <os platform="Mac OS X" version="10.11.6" arch="x86_64"/>
		  <app appid="` + lightThemeExtensionID + `" version="` + lightVersion + `" installsource="ondemand">
		    <updatecheck />
		    <ping rd="-2" ping_freshness="" />
		  </app>
		  <app appid="` + darkThemeExtensionID + `" version="` + darkVersion + `" installsource="ondemand">
		    <updatecheck />
		    <ping rd="-2" ping_freshness="" />
		  </app>
		  <app appid="unknown-test-extension" version="` + unknownVersion + `" installsource="ondemand">
		    <updatecheck />
		    <ping rd="-2" ping_freshness="" />
		  </app>
		</request>`
	}

	// Mixed statuses XML: outdated (ok), current (noupdate), unknown (error-unknownApplication)
	requestBody = threeExtensionXMLRequest("0.0.0", "70.0.0", "1.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="` + lightThemeExtensionID + `">
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
    <app appid="` + darkThemeExtensionID + `">
        <updatecheck status="noupdate"></updatecheck>
    </app>
    <app appid="unknown-test-extension">
        <updatecheck status="error-unknownApplication"></updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Test with blacklisted extension XML
	originalAllExtensionsMapXML := controller.AllExtensionsMap
	controller.AllExtensionsMap = extension.NewExtensionMap()
	controller.AllExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Get and blacklist the light theme extension
	lightExtXML, ok := controller.AllExtensionsMap.Load(lightThemeExtensionID)
	assert.True(t, ok)
	lightExtXML.Blacklisted = true
	controller.AllExtensionsMap.Store(lightThemeExtensionID, lightExtXML)

	// Test blacklisted extension returns restricted status in XML
	requestBody = extensiontest.ExtensionRequestFnForXML(lightThemeExtensionID)("0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="` + lightThemeExtensionID + `">
        <updatecheck status="restricted"></updatecheck>
    </app>
</response>`
	testCall(t, server, http.MethodPost, contentTypeXML, "", requestBody, http.StatusOK, expectedResponse, "")

	// Restore original extensions map
	controller.AllExtensionsMap = originalAllExtensionsMapXML
}

func getQueryParams(extension *extension.Extension) string {
	return `x=id%3D` + extension.ID + `%26v%3D` + extension.Version
}

func TestWebStoreUpdateExtensionV3XML(t *testing.T) {
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

func TestUpdateExtensionsV3JSON(t *testing.T) {
	jsonPrefix := ")]}'\n"
	server := httptest.NewServer(handler)
	defer server.Close()

	// No extensions
	requestBody := `{"request":{"protocol":"3.1","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`
	// The server responds with a different response than what's documented externally
	expectedResponse := jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":null}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Unsupported protocol version
	requestBody = `{"request":{"protocol":"2.0","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`
	expectedResponse = "Error parsing request: unsupported protocol version: 2.0"
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Not JSON
	requestBody = "For the king!"
	expectedResponse = "Error parsing request: error parsing JSON request: invalid character 'F' looking for beginning of value"
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Malformed JSON
	requestBody = "{request"
	expectedResponse = "Error parsing request: error parsing JSON request: invalid character 'r' looking for beginning of object key string"
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	lightThemeExtension := extensiontest.ExtensionRequestFnForJSON("ldimlcelhnjgpjjemdjokpgeeikdinbm")

	// Single extension out of date
	requestBody = lightThemeExtension("0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single extension same version
	requestBody = lightThemeExtension("1.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"noupdate"}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single extension greater version
	requestBody = lightThemeExtension("2.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"noupdate"}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	lightAndDarkThemeRequest := extensiontest.ExtensionRequestFnForTwoJSON("ldimlcelhnjgpjjemdjokpgeeikdinbm", "bfdgpgibhagkpdlnjonhkabjoijopoge")

	// Multiple components with none out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "70.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"noupdate"}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"noupdate"}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Only one components out of date
	requestBody = lightAndDarkThemeRequest("0.0.0", "70.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"noupdate"}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Other component of 2 out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"noupdate"}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtensionID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Both components need updates
	requestBody = lightAndDarkThemeRequest("0.0.0", "0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtensionID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","required":true}]}}}}]}}`
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
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"newext1eplbcioakkpcpgfkobkghlhen","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(newExtensionID1) + `/release/newext1eplbcioakkpcpgfkobkghlhen/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"4c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","hash_sha256":"4c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Single second new extension out of date that was added in by the refresh timer
	requestBody = extensiontest.ExtensionRequestFnForJSON("newext2eplbcioakkpcpgfkobkghlhen")("0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"newext2eplbcioakkpcpgfkobkghlhen","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(newExtensionID2) + `/release/newext2eplbcioakkpcpgfkobkghlhen/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"3c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","hash_sha256":"3c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Test mixed extension statuses in a single request - one outdated, one current, one unknown
	threeExtensionRequest := func(lightVersion, darkVersion, unknownVersion string) string {
		return `{"request":{"protocol":"3.1","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"},"app":[{"appid":"` + lightThemeExtensionID + `","installsource":"ondemand","ping":{"r":-2},"updatecheck":{},"version":"` + lightVersion + `"},{"appid":"` + darkThemeExtensionID + `","installsource":"ondemand","ping":{"r":-2},"updatecheck":{},"version":"` + darkVersion + `"},{"appid":"unknown-test-extension","installsource":"ondemand","ping":{"r":-2},"updatecheck":{},"version":"` + unknownVersion + `"}]}}`
	}

	// Mixed statuses: outdated (ok), current (noupdate), unknown (error-unknownApplication)
	requestBody = threeExtensionRequest("0.0.0", "70.0.0", "1.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"` + lightThemeExtensionID + `","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtensionID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}},{"appid":"` + darkThemeExtensionID + `","status":"ok","updatecheck":{"status":"noupdate"}},{"appid":"unknown-test-extension","status":"ok","updatecheck":{"status":"error-unknownApplication"}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Test with blacklisted extension
	originalAllExtensionsMap := controller.AllExtensionsMap
	controller.AllExtensionsMap = extension.NewExtensionMap()
	controller.AllExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Get and blacklist the light theme extension
	lightExt, ok := controller.AllExtensionsMap.Load(lightThemeExtensionID)
	assert.True(t, ok)
	lightExt.Blacklisted = true
	controller.AllExtensionsMap.Store(lightThemeExtensionID, lightExt)

	// Test blacklisted extension returns restricted status
	requestBody = lightThemeExtension("0.0.0")
	expectedResponse = jsonPrefix + `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"` + lightThemeExtensionID + `","status":"ok","updatecheck":{"status":"restricted"}}]}}`
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, expectedResponse, "")

	// Restore original extensions map
	controller.AllExtensionsMap = originalAllExtensionsMap
}

func TestWebStoreUpdateExtensionV3JSON(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/extensions/all", server.URL)
	req, err := http.NewRequest(http.MethodGet, testURL, bytes.NewBuffer([]byte("")))
	assert.Nil(t, err)
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
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
	controller.AllExtensionsCache.Invalidate() // Invalidate cache after clearing extensions
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	actual, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, string(actual), "{}")
}

func testCallAndParseJSON(t *testing.T, server *httptest.Server, method string, contentType string, query string,
	requestBody string, expectedResponseCode int, redirectLocation string,
) map[string]interface{} {
	extensionsURL := fmt.Sprintf("%s/extensions%s", server.URL, query)
	req, err := http.NewRequest(method, extensionsURL, bytes.NewBuffer([]byte(requestBody)))
	assert.Nil(t, err)
	req.Header.Add("Content-Type", contentType)

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	assert.Nil(t, err)

	assert.Equal(t, expectedResponseCode, resp.StatusCode)

	// If this is a redirect, ensure the location is as expected
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		assert.Equal(t, redirectLocation, location)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	// For JSON responses with our prefix, parse and return the object
	if contentType == contentTypeJSON && len(body) > 6 && string(body[:5]) == ")]}'\n" {
		var result map[string]interface{}
		err = json.Unmarshal(body[5:], &result)
		assert.Nil(t, err)

		return result
	}

	// For non-JSON responses, return nil
	return nil
}

// Extension represents an extension with ID and version for testing
type AppVersionPair struct {
	ID      string
	Version string
}

// buildUpdateV4JSON creates a JSON request body for extension update protocol
func buildUpdateV4JSON(protocol string, apps []AppVersionPair) string {
	baseRequest := map[string]interface{}{
		"request": map[string]interface{}{
			"@os":            "mac",
			"@updater":       "chromecrx",
			"acceptformat":   "crx3,download,puff,run,xz,zucc",
			"protocol":       protocol,
			"version":        "chrome-53.0.2785.116",
			"prodversion":    "137.0.7115.0",
			"requestid":      "{2c047e22-fe09-44d0-883e-28c1d8db4762}",
			"sessionid":      "{b3296be1-ffae-4833-bcf0-31a6c4603ec6}",
			"lang":           "en-GB",
			"updaterchannel": "canary",
			"prodchannel":    "canary",
			"updaterversion": "137.0.7115.0",
			"arch":           "arm64",
			"nacl_arch":      "arm",
			"dedup":          "cr",
			"domainjoined":   false,
			"ismachine":      false,
			"hw": map[string]interface{}{
				"physmemory": 64,
				"avx":        false,
				"sse":        false,
				"sse2":       false,
				"sse3":       false,
				"sse41":      false,
				"sse42":      false,
				"ssse3":      false,
			},
			"os": map[string]interface{}{
				"arch":     "arm64",
				"platform": "Mac OS X",
				"version":  "15.4.0",
			},
		},
	}

	// Add apps if provided
	if len(apps) > 0 {
		appsList := make([]map[string]interface{}, len(apps))
		for i, app := range apps {
			appsList[i] = map[string]interface{}{
				"appid":       app.ID,
				"version":     app.Version,
				"enabled":     true,
				"installedby": "internal",
				"lang":        "en-GB",
				"ping": map[string]interface{}{
					"r": -1,
				},
				"updatecheck": map[string]interface{}{},
			}
		}
		baseRequest["request"].(map[string]interface{})["apps"] = appsList
	}

	jsonBytes, err := json.Marshal(baseRequest)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}
	return string(jsonBytes)
}

func TestUpdateExtensionsV4JSON(t *testing.T) {
	server := httptest.NewServer(handler)
	defer server.Close()

	// Reset the extensions map to start from a clean state
	controller.AllExtensionsMap = extension.NewExtensionMap()
	controller.AllExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// No extensions
	requestBody := buildUpdateV4JSON("4.0", []AppVersionPair{})
	result := testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj := result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])
	daystart, ok := respObj["daystart"].(map[string]interface{})
	assert.True(t, ok, "daystart should be present")
	assert.NotNil(t, daystart["elapsed_days"], "elapsed_days should be present")
	apps, ok := respObj["apps"]
	assert.True(t, ok, "apps should be present")
	assert.Nil(t, apps, "apps should be null for no extensions")

	// Unsupported protocol version
	requestBody = buildUpdateV4JSON("4.77", []AppVersionPair{})
	expectedResponse := "Error parsing request: unsupported protocol version: 4.77"
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusBadRequest, expectedResponse, "")

	// Single extension out of date
	lightThemeExtensionID := "ldimlcelhnjgpjjemdjokpgeeikdinbm"
	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: lightThemeExtensionID, Version: "0.0.0"},
	})
	result = testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj = result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])

	appsInterface, ok := respObj["apps"]
	assert.True(t, ok, "apps should be present")
	appsArray, ok := appsInterface.([]interface{})
	assert.True(t, ok, "apps should be an array")
	assert.Equal(t, 1, len(appsArray), "apps should contain 1 item")

	appInterface := appsArray[0]
	app, ok := appInterface.(map[string]interface{})
	assert.True(t, ok, "app should be a map")
	assert.Equal(t, lightThemeExtensionID, app["appid"])
	assert.Equal(t, "ok", app["status"])

	updatecheckInterface, ok := app["updatecheck"]
	assert.True(t, ok, "updatecheck should be present")
	updatecheck, ok := updatecheckInterface.(map[string]interface{})
	assert.True(t, ok, "updatecheck should be a map")
	assert.Equal(t, "ok", updatecheck["status"])

	// Single extension same version
	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: lightThemeExtensionID, Version: "1.0.0"},
	})
	result = testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj = result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])

	appsInterface, ok = respObj["apps"]
	assert.True(t, ok, "apps should be present")
	appsArray, ok = appsInterface.([]interface{})
	assert.True(t, ok, "apps should be an array")
	assert.Equal(t, 1, len(appsArray), "apps should contain 1 item")

	appInterface = appsArray[0]
	app, ok = appInterface.(map[string]interface{})
	assert.True(t, ok, "app should be a map")
	assert.Equal(t, lightThemeExtensionID, app["appid"])
	assert.Equal(t, "ok", app["status"])

	updatecheckInterface, ok = app["updatecheck"]
	assert.True(t, ok, "updatecheck should be present")
	updatecheck, ok = updatecheckInterface.(map[string]interface{})
	assert.True(t, ok, "updatecheck should be a map")
	assert.Equal(t, "noupdate", updatecheck["status"])

	// Single extension greater version
	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: lightThemeExtensionID, Version: "2.0.0"},
	})
	result = testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj = result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])

	appsInterface, ok = respObj["apps"]
	assert.True(t, ok, "apps should be present")
	appsArray, ok = appsInterface.([]interface{})
	assert.True(t, ok, "apps should be an array")
	assert.Equal(t, 1, len(appsArray), "apps should contain 1 item")

	appInterface = appsArray[0]
	app, ok = appInterface.(map[string]interface{})
	assert.True(t, ok, "app should be a map")
	assert.Equal(t, lightThemeExtensionID, app["appid"])
	assert.Equal(t, "ok", app["status"])

	updatecheckInterface, ok = app["updatecheck"]
	assert.True(t, ok, "updatecheck should be present")
	updatecheck, ok = updatecheckInterface.(map[string]interface{})
	assert.True(t, ok, "updatecheck should be a map")
	assert.Equal(t, "noupdate", updatecheck["status"])

	// Multiple extensions test - create a request with light and dark theme extensions
	darkThemeExtensionID := "bfdgpgibhagkpdlnjonhkabjoijopoge"
	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: lightThemeExtensionID, Version: "0.0.0"},
		{ID: darkThemeExtensionID, Version: "0.0.0"},
	})
	result = testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj = result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])

	appsInterface, ok = respObj["apps"]
	assert.True(t, ok, "apps should be present")
	appsArray, ok = appsInterface.([]interface{})
	assert.True(t, ok, "apps should be an array")
	assert.Equal(t, 2, len(appsArray), "apps should contain 2 items")

	extensionIDs := make(map[string]bool)
	for _, appItem := range appsArray {
		appMap, isMap := appItem.(map[string]interface{})
		assert.True(t, isMap, "app item should be a map")
		extensionIDs[appMap["appid"].(string)] = true
	}
	assert.True(t, extensionIDs[lightThemeExtensionID], "response should include light theme extension")
	assert.True(t, extensionIDs[darkThemeExtensionID], "response should include dark theme extension")

	// Only one extension out of date
	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: lightThemeExtensionID, Version: "0.0.0"},
		{ID: darkThemeExtensionID, Version: "70.0.0"},
	})
	result = testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj = result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])

	appsInterface, ok = respObj["apps"]
	assert.True(t, ok, "apps should be present")
	appsArray, ok = appsInterface.([]interface{})
	assert.True(t, ok, "apps should be an array")
	assert.Equal(t, 2, len(appsArray), "apps should contain 2 items")

	app1Object := appsArray[0]
	app1, ok := app1Object.(map[string]interface{})
	assert.True(t, ok, "app should be a map")
	assert.Equal(t, lightThemeExtensionID, app1["appid"])

	app2Object := appsArray[1]
	app2, ok := app2Object.(map[string]interface{})
	assert.True(t, ok, "app should be a map")
	assert.Equal(t, darkThemeExtensionID, app2["appid"])

	extensionIDs = make(map[string]bool)
	for _, appItem := range appsArray {
		appMap, isMap := appItem.(map[string]interface{})
		assert.True(t, isMap, "app item should be a map")
		extensionIDs[appMap["appid"].(string)] = true
	}
	assert.True(t, extensionIDs[lightThemeExtensionID], "response should include light theme extension")
	assert.True(t, extensionIDs[darkThemeExtensionID], "response should include dark theme extension")

	// Unknown extension ID goes to Google server via componentupdater proxy
	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: "aaaaaaaaaaaaaaaaaaaa", Version: "0.0.0"},
	})
	expectedResponse = ""
	testCall(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusTemporaryRedirect, expectedResponse, "https://componentupdater.brave.com/service/update2/json")

	// Multiple extensions with unknown extension should return error status (not redirect)
	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: lightThemeExtensionID, Version: "0.0.0"},
		{ID: "unknownextensionid123", Version: "1.0.0"},
	})
	result = testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj = result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])

	appsInterface, ok = respObj["apps"]
	assert.True(t, ok, "apps should be present")
	appsArray, ok = appsInterface.([]interface{})
	assert.True(t, ok, "apps should be an array")
	assert.Equal(t, 2, len(appsArray), "apps should contain 2 items")

	// Check both extensions are returned with appropriate statuses
	extensionStatuses := make(map[string]string)
	for _, appItem := range appsArray {
		appMap, isMap := appItem.(map[string]interface{})
		assert.True(t, isMap, "app item should be a map")
		appID := appMap["appid"].(string)
		updateCheck := appMap["updatecheck"].(map[string]interface{})
		status := updateCheck["status"].(string)
		extensionStatuses[appID] = status
	}

	assert.Equal(t, "ok", extensionStatuses[lightThemeExtensionID], "Known extension should have ok status")
	assert.Equal(t, "error-unknownApplication", extensionStatuses["unknownextensionid123"], "Unknown extension should have error status")

	// Test mixed extension statuses - one needing update, one up-to-date, one unknown
	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: lightThemeExtensionID, Version: "0.0.0"},  // Needs update -> ok
		{ID: darkThemeExtensionID, Version: "1.0.0"},   // Up-to-date -> noupdate
		{ID: "anothernewunknownext", Version: "1.0.0"}, // Unknown -> error-unknownApplication
	})
	result = testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj = result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])

	appsInterface, ok = respObj["apps"]
	assert.True(t, ok, "apps should be present")
	appsArray, ok = appsInterface.([]interface{})
	assert.True(t, ok, "apps should be an array")
	assert.Equal(t, 3, len(appsArray), "apps should contain 3 items")

	// Verify all three extensions with different statuses
	extensionStatuses = make(map[string]string)
	for _, appItem := range appsArray {
		appMap, isMap := appItem.(map[string]interface{})
		assert.True(t, isMap, "app item should be a map")
		appID := appMap["appid"].(string)
		updateCheck := appMap["updatecheck"].(map[string]interface{})
		status := updateCheck["status"].(string)
		extensionStatuses[appID] = status
	}

	assert.Equal(t, "ok", extensionStatuses[lightThemeExtensionID], "Outdated extension should have ok status")
	assert.Equal(t, "noupdate", extensionStatuses[darkThemeExtensionID], "Up-to-date extension should have noupdate status")
	assert.Equal(t, "error-unknownApplication", extensionStatuses["anothernewunknownext"], "Unknown extension should have error status")

	// Test restricted/blacklisted extension scenario
	// First, set up a blacklisted extension in the extensions map
	controller.AllExtensionsMap = extension.NewExtensionMap()
	controller.AllExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Get an extension and mark it as blacklisted
	restrictedExt, ok := controller.AllExtensionsMap.Load(lightThemeExtensionID)
	assert.True(t, ok, "Should find light theme extension")
	restrictedExt.Blacklisted = true
	controller.AllExtensionsMap.Store(lightThemeExtensionID, restrictedExt)

	requestBody = buildUpdateV4JSON("4.0", []AppVersionPair{
		{ID: lightThemeExtensionID, Version: "0.0.0"}, // Blacklisted extension
	})
	result = testCallAndParseJSON(t, server, http.MethodPost, contentTypeJSON, "", requestBody, http.StatusOK, "")

	respObj = result["response"].(map[string]interface{})
	assert.Equal(t, "4.0", respObj["protocol"])

	appsInterface, ok = respObj["apps"]
	assert.True(t, ok, "apps should be present")
	appsArray, ok = appsInterface.([]interface{})
	assert.True(t, ok, "apps should be an array")
	assert.Equal(t, 1, len(appsArray), "apps should contain 1 item")

	appInterface = appsArray[0]
	app, ok = appInterface.(map[string]interface{})
	assert.True(t, ok, "app should be a map")
	assert.Equal(t, lightThemeExtensionID, app["appid"])
	assert.Equal(t, "ok", app["status"])

	updatecheckInterface, ok = app["updatecheck"]
	assert.True(t, ok, "updatecheck should be present")
	updatecheck, ok = updatecheckInterface.(map[string]interface{})
	assert.True(t, ok, "updatecheck should be a map")
	assert.Equal(t, "restricted", updatecheck["status"], "Blacklisted extension should have restricted status")

	// Reset extensions map back to clean state for any subsequent tests
	controller.AllExtensionsMap = extension.NewExtensionMap()
	controller.AllExtensionsMap.StoreExtensions(&extension.OfferedExtensions)
}
