package v3

import (
	"encoding/json"
	"encoding/xml"
	"strings"

	"github.com/brave/go-update/extension"
)

// Response represents an Omaha v3 update response
type Response extension.UpdateResponse

// GetUpdateStatus determines the update status based on extension data
func GetUpdateStatus(extension extension.Extension) string {
	if extension.Status == "noupdate" {
		return "noupdate"
	}
	return "ok"
}

// MarshalJSON encodes the extension list into response JSON
func (r *Response) MarshalJSON() ([]byte, error) {
	type URL struct {
		Codebase     string `json:"codebase,omitempty"`
		CodebaseDiff string `json:"codebasediff,omitempty"`
	}
	type URLs struct {
		URLs []URL `json:"url"`
	}
	type Package struct {
		Name       string `json:"name"`
		NameDiff   string `json:"namediff,omitempty"`
		SizeDiff   int    `json:"sizediff,omitempty"`
		FP         string `json:"fp"`
		SHA256     string `json:"hash_sha256"`
		DiffSHA256 string `json:"hashdiff_sha256,omitempty"`
		Required   bool   `json:"required"`
	}
	type Packages struct {
		Package []Package `json:"package"`
	}
	type Manifest struct {
		Version  string   `json:"version"`
		Packages Packages `json:"packages"`
	}
	type UpdateCheck struct {
		Status   string    `json:"status"`
		URLs     *URLs     `json:"urls,omitempty"`
		Manifest *Manifest `json:"manifest,omitempty"`
	}
	type App struct {
		AppID       string      `json:"appid"`
		Status      string      `json:"status"`
		UpdateCheck UpdateCheck `json:"updatecheck"`
	}
	type ResponseWrapper struct {
		Protocol string `json:"protocol"`
		Server   string `json:"server"`
		Apps     []App  `json:"app"`
	}
	type JSONResponse struct {
		Response ResponseWrapper `json:"response"`
	}

	response := ResponseWrapper{}
	response.Protocol = "3.1"
	response.Server = "prod"
	for _, ext := range *r {
		app := App{AppID: ext.ID, Status: "ok"}
		patchInfo, pInfoFound := ext.PatchList[ext.FP]
		app.UpdateCheck = UpdateCheck{Status: GetUpdateStatus(ext)}
		extensionName := "extension_" + strings.Replace(ext.Version, ".", "_", -1) + ".crx"
		url := "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/" + extensionName
		diffURL := "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/patches/" + ext.SHA256 + "/"
		if app.UpdateCheck.Status == "ok" {
			if app.UpdateCheck.URLs == nil {
				app.UpdateCheck.URLs = &URLs{
					URLs: []URL{},
				}
			}
			app.UpdateCheck.URLs.URLs = append(app.UpdateCheck.URLs.URLs, URL{
				Codebase: url,
			})

			app.UpdateCheck.Manifest = &Manifest{
				Version: ext.Version,
			}

			pkg := Package{
				Name:     extensionName,
				SHA256:   ext.SHA256,
				FP:       ext.SHA256,
				Required: true,
			}

			// Only v3.1 supports diffs
			if pInfoFound {
				app.UpdateCheck.URLs.URLs = append(app.UpdateCheck.URLs.URLs, URL{
					CodebaseDiff: diffURL,
				})
				pkg.NameDiff = patchInfo.Namediff
				pkg.DiffSHA256 = patchInfo.Hashdiff
				pkg.SizeDiff = patchInfo.Sizediff
			}

			app.UpdateCheck.Manifest.Packages.Package = append(app.UpdateCheck.Manifest.Packages.Package, pkg)
		}

		response.Apps = append(response.Apps, app)
	}

	jsonResponse := JSONResponse{}
	jsonResponse.Response = response
	return json.Marshal(jsonResponse)
}

// MarshalXML encodes the extension list into response XML
func (r *Response) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
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
		XMLName  xml.Name  `xml:"updatecheck"`
		URLs     *URLs     `xml:"urls,omitempty"`
		Status   string    `xml:"status,attr"`
		Manifest *Manifest `xml:"manifest,omitempty"`
	}
	type App struct {
		XMLName     xml.Name `xml:"app"`
		AppID       string   `xml:"appid,attr"`
		UpdateCheck UpdateCheck
	}
	type ResponseWrapper struct {
		XMLName  xml.Name `xml:"response"`
		Protocol string   `xml:"protocol,attr"`
		Server   string   `xml:"server,attr"`
		Apps     []App
	}
	response := ResponseWrapper{}
	response.Protocol = "3.1"
	response.Server = "prod"
	for _, ext := range *r {
		app := App{AppID: ext.ID}
		app.UpdateCheck = UpdateCheck{Status: GetUpdateStatus(ext)}
		extensionName := "extension_" + strings.Replace(ext.Version, ".", "_", -1) + ".crx"
		url := "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/" + extensionName
		if app.UpdateCheck.Status == "ok" {
			if app.UpdateCheck.URLs == nil {
				app.UpdateCheck.URLs = &URLs{
					URLs: []URL{},
				}
			}
			app.UpdateCheck.URLs.URLs = append(app.UpdateCheck.URLs.URLs, URL{
				Codebase: url,
			})
			app.UpdateCheck.Manifest = &Manifest{
				Version: ext.Version,
			}
			pkg := Package{
				Name:     extensionName,
				SHA256:   ext.SHA256,
				Required: true,
			}
			app.UpdateCheck.Manifest.Packages.Package = append(app.UpdateCheck.Manifest.Packages.Package, pkg)
		}
		response.Apps = append(response.Apps, app)
	}
	e.Indent("", "    ")
	err := e.EncodeElement(response, xml.StartElement{Name: xml.Name{Local: "response"}})
	return err
}

// AsExtensionResponse converts v3.Response to extension.UpdateResponse
func (r *Response) AsExtensionResponse() extension.UpdateResponse {
	return extension.UpdateResponse(*r)
}

// WebStoreResponse represents a web store update response
type WebStoreResponse extension.WebStoreUpdateResponse

// MarshalJSON encodes the extension list into response JSON
func (r *WebStoreResponse) MarshalJSON() ([]byte, error) {
	type UpdateCheck struct {
		Status   string `json:"status"`
		Codebase string `json:"codebase"`
		Version  string `json:"version"`
		SHA256   string `json:"hash_sha256"`
	}
	type App struct {
		AppID       string      `json:"appid"`
		Status      string      `json:"status"`
		UpdateCheck UpdateCheck `json:"updatecheck"`
	}
	type GUpdate struct {
		Protocol string `json:"protocol"`
		Server   string `json:"server"`
		Apps     []App  `json:"app"`
	}
	type JSONGUpdate struct {
		GUpdate GUpdate `json:"gupdate"`
	}
	response := GUpdate{}
	response.Protocol = "3.1"
	response.Server = "prod"

	for _, ext := range *r {
		extensionName := "extension_" + strings.Replace(ext.Version, ".", "_", -1) + ".crx"
		app := App{
			AppID:  ext.ID,
			Status: "ok",
			UpdateCheck: UpdateCheck{
				Status:   "ok",
				SHA256:   ext.SHA256,
				Version:  ext.Version,
				Codebase: "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/" + extensionName,
			},
		}
		response.Apps = append(response.Apps, app)
	}
	jsonGupdate := JSONGUpdate{}
	jsonGupdate.GUpdate = response

	return json.Marshal(jsonGupdate)
}

// MarshalXML encodes the extension list into response XML
func (r *WebStoreResponse) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
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

	for _, ext := range *r {
		extensionName := "extension_" + strings.Replace(ext.Version, ".", "_", -1) + ".crx"
		app := App{
			AppID:  ext.ID,
			Status: "ok",
			UpdateCheck: UpdateCheck{
				Status:   "ok",
				SHA256:   ext.SHA256,
				Version:  ext.Version,
				Codebase: "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/" + extensionName,
			},
		}
		response.Apps = append(response.Apps, app)
	}
	e.Indent("", "    ")
	err := e.EncodeElement(response, xml.StartElement{Name: xml.Name{Local: "gupdate"}})
	return err
}

// AsExtensionWebStoreResponse converts v3.WebStoreResponse to extension.WebStoreUpdateResponse
func (r *WebStoreResponse) AsExtensionWebStoreResponse() extension.WebStoreUpdateResponse {
	return extension.WebStoreUpdateResponse(*r)
}
