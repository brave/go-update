package v4

import (
	"encoding/json"
	"fmt"

	"github.com/brave/go-update/extension"
	"github.com/go-playground/validator/v10"
)

// UpdateRequest represents an Omaha v4 update request
type UpdateRequest []extension.Extension

// UnmarshalJSON decodes the update server request JSON data
func (r *UpdateRequest) UnmarshalJSON(b []byte) error {
	type CachedItem struct {
		SHA256 string `json:"sha256"`
	}
	type App struct {
		AppID       string       `json:"appid"`
		Version     string       `json:"version"`
		CachedItems []CachedItem `json:"cached_items"`
	}
	type RequestWrapper struct {
		OS           string `json:"@os"`
		Updater      string `json:"@updater"`
		Apps         []App  `json:"apps"`
		Protocol     string `json:"protocol" validate:"required"`
		AcceptFormat string `json:"acceptformat"`
	}
	type JSONRequest struct {
		Request RequestWrapper `json:"request" validate:"required"`
	}

	request := JSONRequest{}
	if err := json.Unmarshal(b, &request); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return fmt.Errorf("request validation failed: %v", validationErrors)
	}

	*r = UpdateRequest{}
	for _, app := range request.Request.Apps {
		fp := ""
		if len(app.CachedItems) > 0 {
			fp = app.CachedItems[0].SHA256
		}
		*r = append(*r, extension.Extension{
			ID:      app.AppID,
			FP:      fp,
			Version: app.Version,
		})
	}

	return nil
}
