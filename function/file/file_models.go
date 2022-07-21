package file

import "encoding/xml"

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
