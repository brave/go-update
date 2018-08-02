package extension

import (
	"encoding/xml"
	"strings"
)

// MarshalXML marshals the passed Extensions into XML output
// which is compatible with Google's component update server.
func MarshalXML(extensions Extensions) ([]byte, error) {
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
	for _, extension := range extensions {
		app := App{AppID: extension.ID}
		app.UpdateCheck = UpdateCheck{Status: "ok"}
		url := "https://s3.amazonaws.com/brave-extensions/release/" + extension.ID + "/" + extension.Title
		extensionName := "extension_" + strings.Replace(extension.Version, ".", "_", -1) + ".crx"
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
	output, err := xml.MarshalIndent(response, "", "    ")
	return output, err
}

// UnmarshalXML unmarshals the passed component update server XML data and
// unmarshals it into an array of Extension.
func UnmarshalXML(data []byte) (Extensions, error) {
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
	extensions := Extensions{}

	request := Request{}
	err := xml.Unmarshal(data, &request)
	if err != nil {
		return extensions, err
	}

	for _, app := range request.App {
		extensions = append(extensions, Extension{
			ID:      app.AppID,
			Version: app.Version,
		})
	}

	return extensions, err
}
