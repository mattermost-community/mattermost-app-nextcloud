package file

import (
	"testing"
)

func TestCreateFolderSelectOptions(t *testing.T) {
	pr := property{Getcontenttype: ""}
	propertySt := propertyStat{Property: pr}
	fileResponse := FileResponse{}
	fileResponse.Href = "/remote.php/dav/files/user/Folder/"
	fileResponse.PropertyStats = []propertyStat{
		propertySt,
	}
	searchRespBody := FileSearchResponseBody{FileResponse: []FileResponse{fileResponse}}

	testedInstance := SearchSelectOptionsImpl{}

	expectedDefaultLabel := "root_label"
	expectedDefaultValue := "root_value"
	options, defaultOption := testedInstance.CreateFolderSelectOptions(searchRespBody, "user", expectedDefaultLabel, expectedDefaultValue, false)

	if expectedDefaultLabel != defaultOption.Label || expectedDefaultValue != defaultOption.Value {
		t.Error("Wrong default option created")
	}

	if len(options) < 2 {
		t.Error("Wrong number of folder options created")
	}
}

func TestCreateFileSelectOptions(t *testing.T) {
	pr := property{Getcontenttype: "img/png", Displayname: "img.png"}
	propertySt := propertyStat{Property: pr}
	fileResponse := FileResponse{}
	fileResponse.Href = "/remote.php/dav/files/user/Folder/"
	fileResponse.PropertyStats = []propertyStat{
		propertySt,
	}

	testedInstance := SearchSelectOptionsImpl{}

	options := testedInstance.CreateFileSelectOptions([]FileResponse{fileResponse})

	if len(options) != 1 {
		t.Error("Wrong number of file options created")
	}
	expectedLabel := "Folder/img.png"
	expectedValue := "/Folder/img.png"

	if expectedLabel != options[0].Label || expectedValue != options[0].Value {
		t.Error("Wrong file select option created")

	}
}
