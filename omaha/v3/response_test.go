package v3

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/brave/go-update/extension"
	"github.com/stretchr/testify/assert"
)

func TestResponseMarshalJSON(t *testing.T) {
	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Empty extension list returns a blank JSON update
	updateResponse := Response{}
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput := `{"response":{"protocol":"3.1","server":"prod","app":null}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single JSON update
	updateResponse = []extension.Extension{darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput = `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","required":true}]}}}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	// Multiple extensions returns a multiple extension JSON update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = []extension.Extension{lightThemeExtension, darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput = `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(lightThemeExtension.ID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + extension.GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","required":true}]}}}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))
}

func TestWebStoreResponseMarshalJSON(t *testing.T) {
	// No extensions returns blank update response
	updateResponse := WebStoreResponse{}
	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)
	expectedOutput := `{"gupdate":{"protocol":"3.1","server":"prod","app":null}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
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
	updateResponse := Response{}
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
