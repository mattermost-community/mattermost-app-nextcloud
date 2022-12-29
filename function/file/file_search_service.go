package file

import (
	"encoding/xml"
	"fmt"
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
