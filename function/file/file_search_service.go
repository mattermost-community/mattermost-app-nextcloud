package file

import (
	"encoding/xml"
	"fmt"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"net/http"
	"strings"
)

func sendFileSearchRequest(url string, body string, accessToken string) FileSearchResponseBody {
	req, _ := http.NewRequest("SEARCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	xmlResp := FileSearchResponseBody{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)
	return xmlResp
}

func createSearchRequestBody(userName string, fileName string) string {
	body := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
	<d:searchrequest xmlns:d="DAV:" xmlns:oc="http://owncloud.org/ns">
		<d:basicsearch>
			<d:select>
				<d:prop>
					<oc:fileid/>
					<d:resourcetype/>
					<d:displayname/>
					<d:getcontenttype/>
					<d:getetag/>
					<oc:size/>
				</d:prop>
			</d:select>
			<d:from>
				<d:scope>
					<d:href>/files/%s</d:href>
					<d:depth>infinity</d:depth>
				</d:scope>
			</d:from>
			<d:where>

			<d:like>
                <d:prop>
					<d:displayname/>
                </d:prop>
                <d:literal>%s</d:literal>
            </d:like>
			</d:where>
			<d:orderby/>
	   </d:basicsearch>
   </d:searchrequest>`, userName, "%"+fileName+"%")

	return body
}

type SearchSelectOptions interface {
	CreateFileSelectOptions(files []FileResponse) []apps.SelectOption
	CreateFolderSelectOptions(resp FileSearchResponseBody, userId string, rootLabel string, rootValue string) ([]apps.SelectOption, apps.SelectOption)
}

type SearchSelectOptionsImpl struct {
}

func (SearchSelectOptionsImpl) CreateFileSelectOptions(files []FileResponse) []apps.SelectOption {
	fileSelectOptions := make([]apps.SelectOption, 0)

	for _, f := range files {

		hasContentType := false
		for _, p := range f.PropertyStats {
			if len(p.Property.Getcontenttype) != 0 {
				hasContentType = true
				break
			}
		}

		if !hasContentType {
			continue
		}
		ref := f.Href
		displayname := f.PropertyStats[0].Property.Displayname
		refs := strings.Split(ref, "/")
		r := strings.NewReplacer("%20", " ")

		var sharingPath string
		if len(refs) > 6 {
			sharingPath = r.Replace("/" + strings.Join(refs[5:len(refs)-1], "/") + "/" + displayname)
		} else {
			sharingPath = "/" + displayname
		}

		option := apps.SelectOption{Label: sharingPath[1:], Value: sharingPath}
		fileSelectOptions = append(fileSelectOptions, option)
	}
	return fileSelectOptions
}

func (SearchSelectOptionsImpl) CreateFolderSelectOptions(resp FileSearchResponseBody, userId string, rootLabel string, rootValue string) ([]apps.SelectOption, apps.SelectOption) {
	folderSelectOptions := make([]apps.SelectOption, 0)
	for _, f := range resp.FileResponse {
		hasContentType := false

		for _, p := range f.PropertyStats {
			if len(p.Property.Getcontenttype) != 0 {
				hasContentType = true
				break
			}
		}
		if !hasContentType {
			split := strings.Split(f.Href, "/remote.php/dav/files/"+userId)[1]
			option := apps.SelectOption{Label: split[1 : len(split)-1], Value: split}
			folderSelectOptions = append(folderSelectOptions, option)
		}
	}
	defaultSelectOption := apps.SelectOption{Label: rootLabel, Value: rootValue}
	folderSelectOptions = append(folderSelectOptions, defaultSelectOption)
	return folderSelectOptions, defaultSelectOption
}
