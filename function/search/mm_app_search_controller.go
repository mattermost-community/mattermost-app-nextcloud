package search

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/prokhorind/nextcloud/function/oauth"
	"github.com/prokhorind/nextcloud/function/search/models"
)

func FileSearch(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	token := oauth.RefreshToken(creq)
	accessToken := token.AccessToken

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(token)

	fileName := creq.Values["file_name"].(string)

	remoteUrl := creq.Context.OAuth2.OAuth2App.RemoteRootURL
	userId := creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)

	url := fmt.Sprintf("%s%s", remoteUrl, "/remote.php/dav/")

	body := createSearchRequestBody(userId, fileName)

	req, _ := http.NewRequest("SEARCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	xmlResp := models.FileSearchResponseBody{}
	xml.NewDecoder(resp.Body).Decode(&xmlResp)

	asBot := appclient.AsBot(creq.Context)
	//asBot.DM(creq.Context.ActingUser.Id, "Authenticated")

	files := xmlResp.FileResponse

	for _, f := range files {
		ref := f.Href

		hasContetType := false

		for _, p := range f.PropertyStats {
			if len(p.Property.Getcontenttype) != 0 {
				hasContetType = true
				break
			}

		}
		if hasContetType {
			post := model.Post{
				Message:   remoteUrl + ref,
				ChannelId: creq.Context.Channel.Id,
			}
			asBot.CreatePost(&post)
		}
	}

	c.JSON(http.StatusOK, apps.NewDataResponse(nil))
}

func createSearchRequestBody(userName string, fileName string) string {
	body := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
	<d:searchrequest xmlns:d="DAV:" xmlns:oc="http://owncloud.org/ns">
		<d:basicsearch>
			<d:select>
				<d:prop>
					<oc:fileid/>
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
