package v4

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/brave/go-update/extension"
	"github.com/stretchr/testify/assert"
)

func TestResponseMarshalJSON(t *testing.T) {
	// Set a constant elapsed days value for consistent test output
	GetElapsedDays = func() int { return 6284 }

	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Empty extension list returns a blank JSON update
	updateResponse := UpdateResponse{}
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)

	// Parse the actual response
	var actual map[string]interface{}
	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify the empty extension case
	resp := actual["response"].(map[string]interface{})
	assert.Equal(t, "4.0", resp["protocol"])
	assert.Equal(t, float64(6284), resp["daystart"].(map[string]interface{})["elapsed_days"])
	assert.Nil(t, resp["apps"])

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
	assert.Equal(t, "4.0", resp["protocol"])
	assert.Equal(t, float64(6284), resp["daystart"].(map[string]interface{})["elapsed_days"])

	apps := resp["apps"].([]interface{})
	assert.Equal(t, 1, len(apps))
	app := apps[0].(map[string]interface{})
	assert.Equal(t, "bfdgpgibhagkpdlnjonhkabjoijopoge", app["appid"])
	assert.Equal(t, "ok", app["status"])

	updateCheck := app["updatecheck"].(map[string]interface{})
	assert.Equal(t, "ok", updateCheck["status"])
	assert.Equal(t, "1.0.0", updateCheck["nextversion"])

	pipelines := updateCheck["pipelines"].([]interface{})
	assert.Equal(t, 1, len(pipelines))
	pipeline := pipelines[0].(map[string]interface{})
	assert.Equal(t, "direct_full", pipeline["pipeline_id"])

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
	assert.Equal(t, "4.0", resp["protocol"])
	assert.Equal(t, float64(6284), resp["daystart"].(map[string]interface{})["elapsed_days"])

	apps = resp["apps"].([]interface{})
	assert.Equal(t, 2, len(apps))

	// First app should be lightThemeExtension
	app = apps[0].(map[string]interface{})
	assert.Equal(t, "ldimlcelhnjgpjjemdjokpgeeikdinbm", app["appid"])

	// Second app should be darkThemeExtension
	app = apps[1].(map[string]interface{})
	assert.Equal(t, "bfdgpgibhagkpdlnjonhkabjoijopoge", app["appid"])
}

func TestWebStoreResponseMarshalJSON(t *testing.T) {
	// No extensions returns blank update response
	updateResponse := WebStoreResponse{}
	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput := `{"gupdate":{"protocol":"4.0","server":"prod","app":null}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single JSON update
	updateResponse = WebStoreResponse{darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput = `{"gupdate":{"protocol":"4.0","server":"prod","app":[{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	// Multiple extensions returns a multiple extension JSON webstore update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = WebStoreResponse{lightThemeExtension, darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput = `{"gupdate":{"protocol":"4.0","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtension.ID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))
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
	expectedOutput := `<gupdate protocol="4.0" server="prod"></gupdate>`
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
	expectedOutput = `<gupdate protocol="4.0" server="prod">
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
	expectedOutput = `<gupdate protocol="4.0" server="prod">
    <app appid="ldimlcelhnjgpjjemdjokpgeeikdinbm" status="ok">
        <updatecheck status="ok" codebase="https://` + extension.GetS3ExtensionBucketHost(lightThemeExtension.ID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx" version="1.0.0" hash_sha256="1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"></updatecheck>
    </app>
    <app appid="bfdgpgibhagkpdlnjonhkabjoijopoge" status="ok">
        <updatecheck status="ok" codebase="https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx" version="1.0.0" hash_sha256="ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"></updatecheck>
    </app>
</gupdate>`
	assert.Equal(t, expectedOutput, xmlData)
}
