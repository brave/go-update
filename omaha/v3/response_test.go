package v3

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/brave/go-update/extension"
	"github.com/stretchr/testify/assert"
)

func TestResponseMarshalJSON(t *testing.T) {
	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Test extensions with different statuses
	updateResponse := UpdateResponse{
		{
			ID:      "test-noupdate-ext",
			Version: "1.0.0",
			Status:  "noupdate",
		},
		{
			ID:      "test-unknown-ext",
			Version: "1.0.0",
			Status:  "error-unknownApplication",
		},
		{
			ID:      "test-restricted-ext",
			Version: "1.0.0",
			Status:  "restricted",
		},
	}
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)

	// Parse the actual response
	var actual map[string]interface{}
	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify the mixed status extensions case
	resp := actual["response"].(map[string]interface{})
	assert.Equal(t, "3.1", resp["protocol"])
	assert.Equal(t, "prod", resp["server"])

	apps := resp["app"].([]interface{})
	assert.Equal(t, 3, len(apps))

	// Check each extension has the correct status and no URLs/manifest for non-"ok" statuses
	expectedStatuses := []string{"noupdate", "error-unknownApplication", "restricted"}
	for i, expectedStatus := range expectedStatuses {
		app := apps[i].(map[string]interface{})
		assert.Equal(t, "ok", app["status"]) // App status is always "ok"

		updateCheck := app["updatecheck"].(map[string]interface{})
		assert.Equal(t, expectedStatus, updateCheck["status"])

		// No URLs or manifest for non-"ok" statuses
		_, hasURLs := updateCheck["urls"]
		_, hasManifest := updateCheck["manifest"]
		assert.False(t, hasURLs, "Non-ok status should not have URLs")
		assert.False(t, hasManifest, "Non-ok status should not have manifest")
	}

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single JSON update
	updateResponse = []extension.Extension{darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)

	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify the single extension case
	resp = actual["response"].(map[string]interface{})
	assert.Equal(t, "3.1", resp["protocol"])
	assert.Equal(t, "prod", resp["server"])

	apps = resp["app"].([]interface{})
	assert.Equal(t, 1, len(apps))
	app := apps[0].(map[string]interface{})
	assert.Equal(t, "bfdgpgibhagkpdlnjonhkabjoijopoge", app["appid"])
	assert.Equal(t, "ok", app["status"])

	updateCheck := app["updatecheck"].(map[string]interface{})
	assert.Equal(t, "ok", updateCheck["status"])

	urls := updateCheck["urls"].(map[string]interface{})
	urlList := urls["url"].([]interface{})
	assert.Equal(t, 1, len(urlList))
	url := urlList[0].(map[string]interface{})
	assert.Contains(t, url["codebase"], "bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx")

	manifest := updateCheck["manifest"].(map[string]interface{})
	assert.Equal(t, "1.0.0", manifest["version"])
	packages := manifest["packages"].(map[string]interface{})
	packageList := packages["package"].([]interface{})
	assert.Equal(t, 1, len(packageList))
	pkg := packageList[0].(map[string]interface{})
	assert.Equal(t, "extension_1_0_0.crx", pkg["name"])
	assert.Equal(t, "ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834", pkg["hash_sha256"])
	assert.Equal(t, true, pkg["required"])

	// Multiple extensions returns a multiple extension JSON update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = []extension.Extension{lightThemeExtension, darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)

	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify the multiple extension case
	resp = actual["response"].(map[string]interface{})
	assert.Equal(t, "3.1", resp["protocol"])
	assert.Equal(t, "prod", resp["server"])

	apps = resp["app"].([]interface{})
	assert.Equal(t, 2, len(apps))

	// First app should be lightThemeExtension
	app = apps[0].(map[string]interface{})
	assert.Equal(t, "ldimlcelhnjgpjjemdjokpgeeikdinbm", app["appid"])

	// Second app should be darkThemeExtension
	app = apps[1].(map[string]interface{})
	assert.Equal(t, "bfdgpgibhagkpdlnjonhkabjoijopoge", app["appid"])
}

func TestWebStoreResponseMarshalJSON(t *testing.T) {
	// Test WebStore response with realistic extensions
	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// WebStore response with actual extension (WebStore format always shows updates)
	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	updateResponse := WebStoreResponse{darkThemeExtension}
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput := `{"gupdate":{"protocol":"3.1","server":"prod","app":[{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single JSON update
	updateResponse = WebStoreResponse{darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput = `{"gupdate":{"protocol":"3.1","server":"prod","app":[{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	// Multiple extensions returns a multiple extension JSON webstore update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = WebStoreResponse{lightThemeExtension, darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput = `{"gupdate":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtension.ID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))
}

func TestResponseMarshalXML(t *testing.T) {
	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Empty extension list returns a blank XML update
	updateResponse := UpdateResponse{}
	var buf strings.Builder
	encoder := xml.NewEncoder(&buf)
	err := updateResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "response"}})
	assert.Nil(t, err)
	encoder.Flush()
	xmlData := buf.String()
	expectedOutput := `<response protocol="3.1" server="prod"></response>`
	assert.Equal(t, expectedOutput, xmlData)

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single XML update
	updateResponse = []extension.Extension{darkThemeExtension}
	buf.Reset()
	encoder = xml.NewEncoder(&buf)
	err = updateResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "response"}})
	assert.Nil(t, err)
	encoder.Flush()
	xmlData = buf.String()
	expectedOutput = `<response protocol="3.1" server="prod">
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	assert.Equal(t, expectedOutput, xmlData)

	// Multiple extensions returns a multiple extension XML update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = []extension.Extension{lightThemeExtension, darkThemeExtension}
	buf.Reset()
	encoder = xml.NewEncoder(&buf)
	err = updateResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "response"}})
	assert.Nil(t, err)
	encoder.Flush()
	xmlData = buf.String()
	expectedOutput = `<response protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm">
        <updatecheck status="ok">
            <urls>
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(lightThemeExtension.ID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"></url>
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
                <url codebase="https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"></url>
            </urls>
            <manifest version="1.0.0">
                <packages>
                    <package name="extension_1_0_0.crx" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834" required="true"></package>
                </packages>
            </manifest>
        </updatecheck>
    </app>
</response>`
	assert.Equal(t, expectedOutput, xmlData)
}

func TestWebStoreResponseMarshalXML(t *testing.T) {
	// No extensions returns blank update response
	updateResponse := WebStoreResponse{}
	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	var buf strings.Builder
	encoder := xml.NewEncoder(&buf)
	err := updateResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "gupdate"}})
	assert.Nil(t, err)
	encoder.Flush()
	xmlData := buf.String()
	expectedOutput := `<gupdate protocol="3.1" server="prod"></gupdate>`
	assert.Equal(t, expectedOutput, xmlData)

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single XML update
	updateResponse = WebStoreResponse{darkThemeExtension}
	buf.Reset()
	encoder = xml.NewEncoder(&buf)
	err = updateResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "gupdate"}})
	assert.Nil(t, err)
	encoder.Flush()
	xmlData = buf.String()
	expectedOutput = `<gupdate protocol="3.1" server="prod">
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge" status="ok">
        <updatecheck status="ok" codebase="https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx" version="1.0.0" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"></updatecheck>
    </app>
</gupdate>`
	assert.Equal(t, expectedOutput, xmlData)

	// Multiple extensions returns a multiple extension XML webstore update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = WebStoreResponse{lightThemeExtension, darkThemeExtension}
	buf.Reset()
	encoder = xml.NewEncoder(&buf)
	err = updateResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "gupdate"}})
	assert.Nil(t, err)
	encoder.Flush()
	xmlData = buf.String()
	expectedOutput = `<gupdate protocol="3.1" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm" status="ok">
        <updatecheck status="ok" codebase="https://` + extension.GetS3ExtensionBucketHost(lightThemeExtension.ID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx" version="1.0.0" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"></updatecheck>
    </app>
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge" status="ok">
        <updatecheck status="ok" codebase="https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx" version="1.0.0" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"></updatecheck>
    </app>
</gupdate>`
	assert.Equal(t, expectedOutput, xmlData)
}
