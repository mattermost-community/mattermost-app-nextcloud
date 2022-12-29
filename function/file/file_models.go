package file

import (
	"encoding/xml"
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type FileSearchResponseBody struct {
	XMLName      xml.Name       `xml:"multistatus"`
	Text         string         `xml:",chardata"`
	D            string         `xml:"d,attr"`
	S            string         `xml:"s,attr"`
	Oc           string         `xml:"oc,attr"`
	Nc           string         `xml:"nc,attr"`
	FileResponse []FileResponse `xml:"response"`
}

type FileResponse struct {
	Text          string         `xml:",chardata"`
	Href          string         `xml:"href"`
	PropertyStats []propertyStat `xml:"propstat"`
}

type propertyStat struct {
	Text     string   `xml:",chardata"`
	Status   string   `xml:"status"`
	Property property `xml:"prop"`
}

type property struct {
	Text           string `xml:",chardata"`
	Fileid         string `xml:"fileid"`
	Getcontenttype string `xml:"getcontenttype"`
	Getetag        string `xml:"getetag"`
	Size           string `xml:"size"`
	Displayname    string `xml:"displayname"`
}

type DynamicSelectResponse struct {
	Items []apps.SelectOption `json:"items"`
}

type SharedFilesResponseBody struct {
	XMLName xml.Name `xml:"ocs"`
	Text    string   `xml:",chardata"`
	Meta    OcsMeta  `xml:"meta"`
	Data    struct {
		Text    string           `xml:",chardata"`
		Element []FileShareModel `xml:"element"`
	} `xml:"data"`
}

type SharedFileResponseBody struct {
	XMLName xml.Name       `xml:"ocs"`
	Text    string         `xml:",chardata"`
	Meta    OcsMeta        `xml:"meta"`
	Data    FileShareModel `xml:"data"`
}

type OcsMeta struct {
	Text       string `xml:",chardata"`
	Status     string `xml:"status"`
	Statuscode string `xml:"statuscode"`
	Message    string `xml:"message"`
}

type FileShareModel struct {
	Text                 string `xml:",chardata"`
	ID                   string `xml:"id"`
	ShareType            string `xml:"share_type"`
	UidOwner             string `xml:"uid_owner"`
	DisplaynameOwner     string `xml:"displayname_owner"`
	Permissions          string `xml:"permissions"`
	CanEdit              string `xml:"can_edit"`
	CanDelete            string `xml:"can_delete"`
	Stime                string `xml:"stime"`
	Parent               string `xml:"parent"`
	Expiration           string `xml:"expiration"`
	Token                string `xml:"token"`
	UidFileOwner         string `xml:"uid_file_owner"`
	Note                 string `xml:"note"`
	Label                string `xml:"label"`
	DisplaynameFileOwner string `xml:"displayname_file_owner"`
	Path                 string `xml:"path"`
	ItemType             string `xml:"item_type"`
	Mimetype             string `xml:"mimetype"`
	HasPreview           string `xml:"has_preview"`
	StorageID            string `xml:"storage_id"`
	Storage              string `xml:"storage"`
	ItemSource           string `xml:"item_source"`
	FileSource           string `xml:"file_source"`
	FileParent           string `xml:"file_parent"`
	FileTarget           string `xml:"file_target"`
	ShareWith            string `xml:"share_with"`
	ShareWithDisplayname string `xml:"share_with_displayname"`
	Password             string `xml:"password"`
	SendPasswordByTalk   string `xml:"send_password_by_talk"`
	URL                  string `xml:"url"`
	MailSend             string `xml:"mail_send"`
	HideDownload         string `xml:"hide_download"`
	Attributes           string `xml:"attributes"`
}

type FileShareRequestBody struct {
	Path      string `json:"path"`
	ShareType int32  `json:"shareType"`
}
