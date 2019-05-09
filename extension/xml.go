package extension

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// MarshalXML encodes the extension list into response XML
func (updateResponse *UpdateResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type URL struct {
		XMLName  xml.Name `xml:"url"`
		Codebase string   `xml:"codebase,attr"`
	}
	type URLs struct {
		XMLName xml.Name `xml:"urls"`
		URLs    []URL
	}
	type Package struct {
		XMLName  xml.Name `xml:"package"`
		Name     string   `xml:"name,attr"`
		SHA256   string   `xml:"hash_sha256,attr"`
		Required bool     `xml:"required,attr"`
	}
	type Packages struct {
		XMLName xml.Name `xml:"packages"`
		Package []Package
	}
	type Manifest struct {
		XMLName  xml.Name `xml:"manifest"`
		Version  string   `xml:"version,attr"`
		Packages Packages
	}
	type UpdateCheck struct {
		XMLName  xml.Name `xml:"updatecheck"`
		URLs     URLs
		Status   string `xml:"status,attr"`
		Manifest Manifest
	}
	type App struct {
		XMLName     xml.Name `xml:"app"`
		AppID       string   `xml:"appid,attr"`
		UpdateCheck UpdateCheck
	}
	type Response struct {
		XMLName  xml.Name `xml:"response"`
		Protocol string   `xml:"protocol,attr"`
		Server   string   `xml:"server,attr"`
		Apps     []App
	}
	response := Response{}
	response.Protocol = "3.1"
	response.Server = "prod"
	for _, extension := range *updateResponse {
		app := App{AppID: extension.ID}
		app.UpdateCheck = UpdateCheck{Status: GetUpdateStatus(extension)}
		extensionName := "extension_" + strings.Replace(extension.Version, ".", "_", -1) + ".crx"
		url := "https://" + GetS3ExtensionBucketHost() + "/release/" + extension.ID + "/" + extensionName
		app.UpdateCheck.URLs.URLs = append(app.UpdateCheck.URLs.URLs, URL{
			Codebase: url,
		})
		app.UpdateCheck.Manifest = Manifest{
			Version: extension.Version,
		}
		pkg := Package{
			Name:     extensionName,
			SHA256:   extension.SHA256,
			Required: true,
		}
		app.UpdateCheck.Manifest.Packages.Package = append(app.UpdateCheck.Manifest.Packages.Package, pkg)
		response.Apps = append(response.Apps, app)
	}
	e.Indent("", "    ")
	err := e.EncodeElement(response, xml.StartElement{Name: xml.Name{Local: "response"}})
	return err
}

// MarshalXML encodes the extension list into response XML
func (updateResponse *WebStoreUpdateResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type UpdateCheck struct {
		XMLName  xml.Name `xml:"updatecheck"`
		Status   string   `xml:"status,attr"`
		Codebase string   `xml:"codebase,attr"`
		Version  string   `xml:"version,attr"`
		SHA256   string   `xml:"hash_sha256,attr"`
	}
	type App struct {
		XMLName     xml.Name `xml:"app"`
		AppID       string   `xml:"appid,attr"`
		Status      string   `xml:"status,attr"`
		UpdateCheck UpdateCheck
	}
	type GUpdate struct {
		XMLName  xml.Name `xml:"gupdate"`
		Protocol string   `xml:"protocol,attr"`
		Server   string   `xml:"server,attr"`
		Apps     []App
	}
	response := GUpdate{}
	response.Protocol = "3.1"
	response.Server = "prod"

	for _, extension := range *updateResponse {
		extensionName := "extension_" + strings.Replace(extension.Version, ".", "_", -1) + ".crx"
		app := App{
			AppID:  extension.ID,
			Status: "ok",
			UpdateCheck: UpdateCheck{
				Status:   "ok",
				SHA256:   extension.SHA256,
				Version:  extension.Version,
				Codebase: "https://" + GetS3ExtensionBucketHost() + "/release/" + extension.ID + "/" + extensionName,
			},
		}
		response.Apps = append(response.Apps, app)
	}
	e.Indent("", "    ")
	err := e.EncodeElement(response, xml.StartElement{Name: xml.Name{Local: "gupdate"}})
	return err
}

// UnmarshalXML decodes the update server request XML data for a list of extensions
func (updateRequest *UpdateRequest) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type UpdateCheck struct {
		XMLName xml.Name `xml:"updatecheck"`
	}
	type App struct {
		XMLName     xml.Name `xml:"app"`
		AppID       string   `xml:"appid,attr"`
		UpdateCheck UpdateCheck
		Version     string `xml:"version,attr"`
	}
	type Request struct {
		XMLName  xml.Name `xml:"request"`
		App      []App    `xml:"app"`
		Protocol string   `xml:"protocol,attr"`
	}

	request := Request{}
	err := d.DecodeElement(&request, &start)
	if err != nil {
		return err
	}

	*updateRequest = UpdateRequest{}
	for _, app := range request.App {
		*updateRequest = append(*updateRequest, Extension{
			ID:      app.AppID,
			Version: app.Version,
		})
	}

	if request.Protocol != "3.0" && request.Protocol != "3.1" {
		err = fmt.Errorf("request version: %v not supported", request.Protocol)
	}

	return err
}
