package dynamics

import (
	"fmt"
)

// UploadFile uploads a file to a specified entity in Dynamics 365
func (d *D365) UploadFile(entityID, field, filename string, fileData []byte) error {
    endpoint := fmt.Sprintf("new_fakturas(%s)/%s", entityID, field)
    resp, err := d.Resty.R().
        SetHeader("Authorization", fmt.Sprintf("Bearer %v", d.AccessToken)).
        SetHeader("Content-Type", "application/octet-stream").
        SetHeader("x-ms-file-name", filename).
        SetBody(fileData).
        Put(d.URL + "/api/data/v9.2/" + endpoint)

    if err != nil {
        return fmt.Errorf("error uploading file: %v", err)
    }

    if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
        return fmt.Errorf("error uploading file: %v", resp.String())
    }

    return nil
}
