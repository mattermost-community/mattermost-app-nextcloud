package calendar

import "encoding/xml"

type UserCalendarsResponse struct {
	XMLName  xml.Name `xml:"multistatus"`
	Text     string   `xml:",chardata"`
	D        string   `xml:"d,attr"`
	S        string   `xml:"s,attr"`
	Cal      string   `xml:"cal,attr"`
	Cs       string   `xml:"cs,attr"`
	Oc       string   `xml:"oc,attr"`
	Nc       string   `xml:"nc,attr"`
	Response []struct {
		Text     string `xml:",chardata"`
		Href     string `xml:"href"`
		Propstat struct {
			Text string `xml:",chardata"`
			Prop struct {
				Text        string `xml:",chardata"`
				Displayname string `xml:"displayname"`
				Getctag     string `xml:"getctag"`
			} `xml:"prop"`
			Status string `xml:"status"`
		} `xml:"propstat"`
	} `xml:"response"`
}

type Value struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
