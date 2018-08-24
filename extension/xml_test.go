package extension

import (
	"encoding/xml"
	"github.com/brave/go-update/extension/extensiontest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateResponseMarshalXML(t *testing.T) {
	allExtensionsMap := LoadExtensionsIntoMap(&OfferedExtensions)
	// Empty extension list returns a blank XML update
	updateResponse := UpdateResponse{}
	xmlData, err := xml.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput := `<response protocol="3.1" server="prod"></response>`
	assert.Equal(t, expectedOutput, string(xmlData))

	darkThemeExtension, ok := allExtensionsMap["bfdgpgibhagkpdlnjonhkabjoijopoge"]
	assert.True(t, ok)

	// Single extension list returns a single XML update
	updateResponse = []Extension{darkThemeExtension}
	xmlData, err = xml.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput = `<response protocol="3.1" server="prod">
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://s3.amazonaws.com/brave-extensions/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"></url>
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
	lightThemeExtension, ok := allExtensionsMap["ldimlcelhnjgpjjemdjokpgeeikdinbm"]
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap["bfdgpgibhagkpdlnjonhkabjoijopoge"]
	assert.True(t, ok)
	updateResponse = []Extension{lightThemeExtension, darkThemeExtension}
	xmlData, err = xml.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://s3.amazonaws.com/brave-extensions/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"></url>
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
                <url codebase="https://s3.amazonaws.com/brave-extensions/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"></url>
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

func TestUpdateRequestUnmarshalXML(t *testing.T) {
	// Empty data returns an error
	updateRequest := UpdateRequest{}
	err := xml.Unmarshal([]byte(""), &updateRequest)
	assert.NotNil(t, err, "UnmarshalXML should return an error for empty content")

	// Malformed XML returns an error
	err = xml.Unmarshal([]byte("<"), &updateRequest)
	assert.NotNil(t, err, "UnmarshalXML should return an error for malformed XML")

	// Wrong schema returns an error
	err = xml.Unmarshal([]byte("<text>For the king!</text>"), &updateRequest)
	assert.NotNil(t, err, "UnmarshalXML should return an error for wrong XML Schema")

	// No extensions XML with proper schema, no error with 0 extensions returned
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
		  <hw physmemory="16"/>
		  <os platform="Mac OS X" version="10.11.6" arch="x86_64"/>
		</request>`)
	err = xml.Unmarshal(data, &updateRequest)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(updateRequest))

	onePasswordID := "aomjjhallfgjeglblehebfpbcfeobpgk" // #nosec
	onePasswordVersion := "4.7.0.90"
	onePasswordRequest := extensiontest.ExtensionRequestFnFor(onePasswordID)
	data = []byte(onePasswordRequest(onePasswordVersion))
	err = xml.Unmarshal(data, &updateRequest)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(updateRequest))
	assert.Equal(t, onePasswordID, updateRequest[0].ID)
	assert.Equal(t, onePasswordVersion, updateRequest[0].Version)

	pdfJSID := "jdbefljfgobbmcidnmpjamcbhnbphjnb"
	pdfJSVersion := "1.0.0"
	twoExtensionRequest := extensiontest.ExtensionRequestFnForTwo(onePasswordID, pdfJSID)
	data = []byte(twoExtensionRequest(onePasswordVersion, pdfJSVersion))
	err = xml.Unmarshal(data, &updateRequest)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(updateRequest))
	assert.Equal(t, onePasswordID, updateRequest[0].ID)
	assert.Equal(t, onePasswordVersion, updateRequest[0].Version)
	assert.Equal(t, pdfJSID, updateRequest[1].ID)
	assert.Equal(t, pdfJSVersion, updateRequest[1].Version)

	// Check for unsupported protocol version
	data = []byte(`<request protocol="2.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64"/>`)
	err = xml.Unmarshal(data, &updateRequest)
	assert.NotNil(t, err, "Unrecognized protocol should have an error")
}

func TestWebStoreUpdateResponseMarshalXML(t *testing.T) {
	// No extensions returns blank update response
	updateResponse := WebStoreUpdateResponse{}
	allExtensionsMap := LoadExtensionsIntoMap(&OfferedExtensions)
	xmlData, err := xml.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput := `<gupdate protocol="3.1" server="prod"></gupdate>`
	assert.Equal(t, expectedOutput, string(xmlData))

	darkThemeExtension, ok := allExtensionsMap["bfdgpgibhagkpdlnjonhkabjoijopoge"]
	assert.True(t, ok)

	// Single extension list returns a single XML update
	updateResponse = WebStoreUpdateResponse{darkThemeExtension}
	xmlData, err = xml.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput = `<gupdate protocol="3.1" server="prod">
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge" status="ok">
        <updatecheck status="ok" codebase="https://s3.amazonaws.com/brave-extensions/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx" version="1.0.0" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"></updatecheck>
    </app>
</gupdate>`
	assert.Equal(t, expectedOutput, string(xmlData))

	// Multiple extensions returns a multiple extension XML webstore update
	lightThemeExtension, ok := allExtensionsMap["ldimlcelhnjgpjjemdjokpgeeikdinbm"]
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap["bfdgpgibhagkpdlnjonhkabjoijopoge"]
	assert.True(t, ok)
	updateResponse = WebStoreUpdateResponse{lightThemeExtension, darkThemeExtension}
	xmlData, err = xml.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput = `<gupdate protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm" status="ok">
        <updatecheck status="ok" codebase="https://s3.amazonaws.com/brave-extensions/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx" version="1.0.0" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"></updatecheck>
    </app>
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge" status="ok">
        <updatecheck status="ok" codebase="https://s3.amazonaws.com/brave-extensions/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx" version="1.0.0" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"></updatecheck>
    </app>
</gupdate>`
	assert.Equal(t, expectedOutput, string(xmlData))
}
