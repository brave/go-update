package extension

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarshalJSON encodes the extension list into response JSON
func (updateResponse *UpdateResponse) MarshalJSON() ([]byte, error) {
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
	type Response struct {
		Protocol string `json:"protocol"`
		Server   string `json:"server"`
		Apps     []App  `json:"app"`
	}
	type JSONResponse struct {
		Response Response `json:"response"`
	}

	response := Response{}
	response.Protocol = "3.1"
	response.Server = "prod"
	for _, extension := range *updateResponse {
		app := App{AppID: extension.ID, Status: "ok"}
		patchInfo, pInfoFound := extension.PatchList[extension.FP]
		app.UpdateCheck = UpdateCheck{Status: GetUpdateStatus(extension)}
		extensionName := "extension_" + strings.Replace(extension.Version, ".", "_", -1) + ".crx"
		url := "https://" + GetS3ExtensionBucketHost(extension.ID) + "/release/" + extension.ID + "/" + extensionName
		diffURL := "https://" + GetS3ExtensionBucketHost(extension.ID) + "/release/" + extension.ID + "/patches/" + extension.SHA256 + "/"
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
				Version: extension.Version,
			}

			pkg := Package{
				Name:     extensionName,
				SHA256:   extension.SHA256,
				FP:       extension.SHA256,
				Required: true,
			}

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

// MarshalJSON encodes the extension list into response JSON
func (updateResponse *WebStoreUpdateResponse) MarshalJSON() ([]byte, error) {
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

	for _, extension := range *updateResponse {
		extensionName := "extension_" + strings.Replace(extension.Version, ".", "_", -1) + ".crx"
		app := App{
			AppID:  extension.ID,
			Status: "ok",
			UpdateCheck: UpdateCheck{
				Status:   "ok",
				SHA256:   extension.SHA256,
				Version:  extension.Version,
				Codebase: "https://" + GetS3ExtensionBucketHost(extension.ID) + "/release/" + extension.ID + "/" + extensionName,
			},
		}
		response.Apps = append(response.Apps, app)
	}
	jsonGupdate := JSONGUpdate{}
	jsonGupdate.GUpdate = response

	return json.Marshal(jsonGupdate)
}

// UnmarshalJSON decodes the update server request JSON data for a list of extensions
func (updateRequest *UpdateRequest) UnmarshalJSON(b []byte) error {
	type Package struct {
		FP string `json:"fp"`
	}
	type Packages struct {
		Package []Package `json:"package"`
	}
	type App struct {
		AppID       string       `json:"appid"`
		FP          string       `json:"fp,omitempty"`           // For 3.x
		CachedItems []CachedItem `json:"cached_items,omitempty"` // For 4.0
		Version     string       `json:"version"`
		Packages    Packages     `json:"packages"`
	}
	type Request struct {
		OS       string `json:"@os"`
		Updater  string `json:"@updater,omitempty"`  // For 3.x
		Updaters string `json:"@updaters,omitempty"` // For 4.0
		App      []App  `json:"app,omitempty"`       // For 3.x
		Apps     []App  `json:"apps,omitempty"`      // For 4.0
		Protocol string `json:"protocol"`
	}
	type JSONRequest struct {
		Request Request `json:"request"`
	}

	request := JSONRequest{}
	if err := json.Unmarshal(b, &request); err != nil {
		return err
	}

	// Validate protocol
	protocol := request.Request.Protocol
	if protocol != "3.0" && protocol != "3.1" && protocol != "4.0" {
		return fmt.Errorf("request version: %v not supported", protocol)
	}

	*updateRequest = UpdateRequest{}

	var appsToProcess []App
	if protocol == "4.0" { // Use 'apps' for protocol 4.0
		appsToProcess = request.Request.Apps
	} else { // Fallback to 'app' for 3.x
		appsToProcess = request.Request.App
	}

	for _, app := range appsToProcess {
		fp := app.FP
		// spec discrepancy: FP might be set within a "package" object (v3) instead of the "app" object (v3.1)
		// https://github.com/google/omaha/blob/main/doc/ServerProtocolV3.md#package-request
		// https://chromium.googlesource.com/chromium/src.git/+/master/docs/updater/protocol_3_1.md#update-checks-body-update-check-request-objects-update-check-request-3
		if fp == "" && len(app.Packages.Package) > 0 {
			fp = app.Packages.Package[0].FP
		}
		*updateRequest = append(*updateRequest, Extension{
			ID:      app.AppID,
			FP:      fp,
			Version: app.Version,
			// Protocol is not stored on the extension itself anymore
		})
	}

	return nil
}
