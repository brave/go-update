package extension

import (
	"github.com/brave/go-update/extension/extensiontest"
	"testing"
)

func TestMarshalXML(t *testing.T) {
	// Empty extension list returns a blank XML update
	extensions := []Extension{}
	xml, err := MarshalXML(extensions)
	extensiontest.AssertEqual(t, err, nil)
	expectedOutput := `<response protocol="3.1" server="prod"></response>`
	extensiontest.AssertEqual(t, string(xml), expectedOutput)

	darkThemeExtension, err := Contains(OfferedExtensions, "bfdgpgibhagkpdlnjonhkabjoijopoge")
	extensiontest.AssertEqual(t, err, nil)

	// Single extension list returns a single XML update
	extensions = []Extension{darkThemeExtension}
	xml, err = MarshalXML(extensions)
	extensiontest.AssertEqual(t, err, nil)
	expectedOutput = `<response protocol="3.1" server="prod">
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
	extensiontest.AssertEqual(t, string(xml), expectedOutput)

	// Multiple extensions returns a multiple extension XML update
	lightThemeExtension, err := Contains(OfferedExtensions, "ldimlcelhnjgpjjemdjokpgeeikdinbm")
	extensiontest.AssertEqual(t, err, nil)
	darkThemeExtension, err = Contains(OfferedExtensions, "bfdgpgibhagkpdlnjonhkabjoijopoge")
	extensiontest.AssertEqual(t, err, nil)
	extensions = []Extension{lightThemeExtension, darkThemeExtension}
	xml, err = MarshalXML(extensions)
	extensiontest.AssertEqual(t, err, nil)
	expectedOutput = `<response protocol="3.1" server="prod">
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
	extensiontest.AssertEqual(t, string(xml), expectedOutput)
}

func TestUnmarshalXML(t *testing.T) {
	// Empty data returns an error
	_, err := UnmarshalXML([]byte(""))
	if err == nil {
		t.Fatalf("UnmarshalXML should return an error for empty content")
	}
	// Malformed XML returns an error
	_, err = UnmarshalXML([]byte("<"))
	if err == nil {
		t.Fatalf("UnmarshalXML should return an error for malformed XML")
	}
	// Wrong schema returns an error
	_, err = UnmarshalXML([]byte("<text>For the king!</text>"))
	if err == nil {
		t.Fatalf("UnmarshalXML should return an error for wrong XML Schema")
	}
	// No extensions XML with proper schema, no error with 0 extensions returned
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
		  <hw physmemory="16"/>
		  <os platform="Mac OS X" version="10.11.6" arch="x86_64"/>
		</request>`)
	extensions, err := UnmarshalXML(data)
	extensiontest.AssertEqual(t, err, nil)
	extensiontest.AssertEqual(t, len(extensions), 0)

	onePasswordID := "aomjjhallfgjeglblehebfpbcfeobpgk" // #nosec
	onePasswordVersion := "4.7.0.90"
	onePasswordRequest := extensiontest.ExtensionRequestFnFor(onePasswordID)
	data = []byte(onePasswordRequest(onePasswordVersion))
	extensions, err = UnmarshalXML(data)
	extensiontest.AssertEqual(t, err, nil)
	extensiontest.AssertEqual(t, len(extensions), 1)
	extensiontest.AssertEqual(t, extensions[0].ID, onePasswordID)
	extensiontest.AssertEqual(t, extensions[0].Version, onePasswordVersion)

	pdfJSID := "jdbefljfgobbmcidnmpjamcbhnbphjnb"
	pdfJSVersion := "1.0.0"
	twoExtnesionRequest := extensiontest.ExtensionRequestFnForTwo(onePasswordID, pdfJSID)
	data = []byte(twoExtnesionRequest(onePasswordVersion, pdfJSVersion))
	extensions, err = UnmarshalXML(data)
	extensiontest.AssertEqual(t, err, nil)
	extensiontest.AssertEqual(t, len(extensions), 2)
	extensiontest.AssertEqual(t, extensions[0].ID, onePasswordID)
	extensiontest.AssertEqual(t, extensions[0].Version, onePasswordVersion)
	extensiontest.AssertEqual(t, extensions[1].ID, pdfJSID)
	extensiontest.AssertEqual(t, extensions[1].Version, pdfJSVersion)

	// Check for unsupported protocol version
	data = []byte(`<request protocol="2.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64"/>`)
	_, err = UnmarshalXML(data)
	if err == nil {
		t.Fatalf("Unrecognized protocol should have an error")
	}
}
