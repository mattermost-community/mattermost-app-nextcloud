package file

import (
	"encoding/xml"
	"fmt"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
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

func sendFiles(f FileResponse, creq *apps.CallRequest) {
	ref := f.Href
	refs := strings.Split(ref, "/")
	fileName := refs[len(refs)-1]
	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	asBot := appclient.AsBot(creq.Context)

	hasContentType := false

	for _, p := range f.PropertyStats {
		if len(p.Property.Getcontenttype) != 0 {
			hasContentType = true
			break
		}
	}
	if hasContentType {
		post := model.Post{
			Message:   fmt.Sprintf("[Download](%s) %s", remoteUrl+ref, fileName),
			ChannelId: creq.Context.Channel.Id,
		}
		asBot.CreatePost(&post)
	}
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
