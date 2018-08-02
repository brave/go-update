package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/brave/go-update/extension/extensiontest"
	"github.com/go-chi/chi"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var handler http.Handler

func init() {
	handler = chi.ServerBaseContext(setupRouter(setupLogger(context.Background())))
}

func TestPing(t *testing.T) {
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Received non-200 response: %d\n", resp.StatusCode)
	}
	expected := "."
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if expected != string(actual) {
		t.Errorf("Expected the message '%s'\n", expected)
	}
}

func testCall(t *testing.T, server *httptest.Server, requestBody string, expectedResponseCode int, expectedResponse string) {
	extensionsURL := fmt.Sprintf("%s/extensions", server.URL)
	req, err := http.NewRequest("POST", extensionsURL, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatalf("Error with POST: %v\n", err)
	}
	req.Header.Add("Content-Type", "application/xml")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error with running request: %v\n", err)
	}

	extensiontest.AssertEqual(t, resp.StatusCode, expectedResponseCode)

	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response: %v\n", err)
	}

	extensiontest.AssertEqual(t, strings.TrimSpace(string(actual)), expectedResponse)
}

func TestUpdateExtensions(t *testing.T) {
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
	testCall(t, server, requestBody, http.StatusOK, expectedResponse)

	// Unsupported protocol version
	requestBody =
		`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="2.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
			<app appid="aomjjhallfgjeglblehebfpbcfeobpgk">
				<updatecheck codebase="https://s3.amazonaws.com/brave-extensions/release/aomjjhallfgjeglblehebfpbcfeobpgk/extension_4_5_9_90.crx" version="4.5.9.90"/>
			</app>
		</request>`
	expectedResponse = "Error reading body request version: 2.0 not supported"
	testCall(t, server, requestBody, http.StatusBadRequest, expectedResponse)

	// Not XML
	requestBody = "For the king!"
	expectedResponse = "Error reading body EOF"
	testCall(t, server, requestBody, http.StatusBadRequest, expectedResponse)

	// Malformed XML
	requestBody = "<This way! No, that way!"
	expectedResponse = "Error reading body XML syntax error on line 1: attribute name without = in element"
	testCall(t, server, requestBody, http.StatusBadRequest, expectedResponse)

	// Different XML schema
	requestBody = "<text>For the alliance!</text>"
	expectedResponse = "Error reading body expected element type <request> but have <text>"
	testCall(t, server, requestBody, http.StatusBadRequest, expectedResponse)

	// Empty body request
	requestBody = ""
	expectedResponse = "Error reading body EOF"
	testCall(t, server, requestBody, http.StatusBadRequest, expectedResponse)

	lightThemeExtension := extensiontest.ExtensionRequestFnFor("ldimlcelhnjgpjjemdjokpgeeikdinbm")

	// Single extension out of date
	requestBody = lightThemeExtension("0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://s3.amazonaws.com/brave-extensions/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/Brave Light Theme"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, requestBody, http.StatusOK, expectedResponse)

	// Single extension same version
	requestBody = lightThemeExtension("1.0.0")
	expectedResponse = "<response protocol=\"3.1\" server=\"prod\"></response>"
	testCall(t, server, requestBody, http.StatusOK, expectedResponse)

	// Single extension greater version
	requestBody = lightThemeExtension("2.0.0")
	expectedResponse = "<response protocol=\"3.1\" server=\"prod\"></response>"
	testCall(t, server, requestBody, http.StatusOK, expectedResponse)

	lightAndDarkThemeRequest := extensiontest.ExtensionRequestFnForTwo("ldimlcelhnjgpjjemdjokpgeeikdinbm", "bfdgpgibhagkpdlnjonhkabjoijopoge")

	// Multiple components with none out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "70.0.0")
	expectedResponse = "<response protocol=\"3.1\" server=\"prod\"></response>"
	testCall(t, server, requestBody, http.StatusOK, expectedResponse)

	// Only one components out of date
	requestBody = lightAndDarkThemeRequest("0.0.0", "70.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://s3.amazonaws.com/brave-extensions/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/Brave Light Theme"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, requestBody, http.StatusOK, expectedResponse)

	// Other component of 2 out of date
	requestBody = lightAndDarkThemeRequest("70.0.0", "0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://s3.amazonaws.com/brave-extensions/release/bfdgpgibhagkpdlnjonhkabjoijopoge/Brave Dark Theme"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, requestBody, http.StatusOK, expectedResponse)

	// Both components need updates
	requestBody = lightAndDarkThemeRequest("0.0.0", "0.0.0")
	expectedResponse = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://s3.amazonaws.com/brave-extensions/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/Brave Light Theme"></url>
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
                <url codebase="https://s3.amazonaws.com/brave-extensions/release/bfdgpgibhagkpdlnjonhkabjoijopoge/Brave Dark Theme"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	testCall(t, server, requestBody, http.StatusOK, expectedResponse)

	// Unkonwn extension ID goes to Google server
	requestBody = extensiontest.ExtensionRequestFnFor("aaaaaaaaaaaaaaaaaaaa")("0.0.0")
	expectedResponse = ""
	testCall(t, server, requestBody, http.StatusFound, expectedResponse)
}
