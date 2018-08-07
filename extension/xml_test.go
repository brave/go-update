package extension

import (
	"encoding/xml"
	"github.com/brave/go-update/extension/extensiontest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshalXML(t *testing.T) {
	// Empty extension list returns a blank XML update
	extensions := Extensions{}
	xmlData, err := xml.Marshal(&extensions)
	assert.Nil(t, err)
	expectedOutput := `<response protocol="3.1" server="prod"></response>`
	assert.Equal(t, expectedOutput, string(xmlData))

	darkThemeExtension, err := OfferedExtensions.Contains("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.Nil(t, err)

	// Single extension list returns a single XML update
	extensions = []Extension{darkThemeExtension}
	xmlData, err = xml.Marshal(&extensions)
	assert.Nil(t, err)
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
	assert.Equal(t, expectedOutput, string(xmlData))

	// Multiple extensions returns a multiple extension XML update
	lightThemeExtension, err := OfferedExtensions.Contains("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.Nil(t, err)
	darkThemeExtension, err = OfferedExtensions.Contains("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.Nil(t, err)
	extensions = []Extension{lightThemeExtension, darkThemeExtension}
	xmlData, err = xml.Marshal(&extensions)
	assert.Nil(t, err)
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
	assert.Equal(t, expectedOutput, string(xmlData))
}

func TestUnmarshalXML(t *testing.T) {
	// Empty data returns an error
	extensions := Extensions{}
	err := xml.Unmarshal([]byte(""), &extensions)
	assert.NotNil(t, err, "UnmarshalXML should return an error for empty content")

	// Malformed XML returns an error
	err = xml.Unmarshal([]byte("<"), &extensions)
	assert.NotNil(t, err, "UnmarshalXML should return an error for malformed XML")

	// Wrong schema returns an error
	err = xml.Unmarshal([]byte("<text>For the king!</text>"), &extensions)
	assert.NotNil(t, err, "UnmarshalXML should return an error for wrong XML Schema")

	// No extensions XML with proper schema, no error with 0 extensions returned
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
		  <hw physmemory="16"/>
		  <os platform="Mac OS X" version="10.11.6" arch="x86_64"/>
		</request>`)
	err = xml.Unmarshal(data, &extensions)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(extensions))

	onePasswordID := "aomjjhallfgjeglblehebfpbcfeobpgk" // #nosec
	onePasswordVersion := "4.7.0.90"
	onePasswordRequest := extensiontest.ExtensionRequestFnFor(onePasswordID)
	data = []byte(onePasswordRequest(onePasswordVersion))
	err = xml.Unmarshal(data, &extensions)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(extensions))
	assert.Equal(t, onePasswordID, extensions[0].ID)
	assert.Equal(t, onePasswordVersion, extensions[0].Version)

	pdfJSID := "jdbefljfgobbmcidnmpjamcbhnbphjnb"
	pdfJSVersion := "1.0.0"
	twoExtnesionRequest := extensiontest.ExtensionRequestFnForTwo(onePasswordID, pdfJSID)
	data = []byte(twoExtnesionRequest(onePasswordVersion, pdfJSVersion))
	err = xml.Unmarshal(data, &extensions)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(extensions))
	assert.Equal(t, onePasswordID, extensions[0].ID)
	assert.Equal(t, onePasswordVersion, extensions[0].Version)
	assert.Equal(t, pdfJSID, extensions[1].ID)
	assert.Equal(t, pdfJSVersion, extensions[1].Version)

	// Check for unsupported protocol version
	data = []byte(`<request protocol="2.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64"/>`)
	err = xml.Unmarshal(data, &extensions)
	assert.NotNil(t, err, "Unrecognized protocol should have an error")
}
