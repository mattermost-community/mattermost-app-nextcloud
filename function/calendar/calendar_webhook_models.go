package calendar

import (
	ics "github.com/arran4/golang-ical"
	"time"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type WebhookCalendarRequest struct {
	Path   string `json:"path"`
	Values struct {
		Data struct {
			CalendarData struct {
				ID                                             string      `json:"id"`
				Principaluri                                   string      `json:"principaluri"`
				URI                                            string      `json:"uri"`
				DAVDisplayname                                 string      `json:"{DAV:}displayname"`
				HTTPAppleComNsIcalCalendarColor                string      `json:"{http://apple.com/ns/ical/}calendar-color"`
				HTTPAppleComNsIcalCalendarOrder                int         `json:"{http://apple.com/ns/ical/}calendar-order"`
				HTTPCalendarserverOrgNsGetctag                 string      `json:"{http://calendarserver.org/ns/}getctag"`
				HTTPNextcloudComNsDeletedAt                    interface{} `json:"{http://nextcloud.com/ns}deleted-at"`
				HTTPNextcloudComNsOwnerDisplayname             string      `json:"{http://nextcloud.com/ns}owner-displayname"`
				HTTPSabredavOrgNsSyncToken                     string      `json:"{http://sabredav.org/ns}sync-token"`
				UrnIetfParamsXMLNsCaldavCalendarDescription    interface{} `json:"{urn:ietf:params:xml:ns:caldav}calendar-description"`
				UrnIetfParamsXMLNsCaldavCalendarTimezone       string      `json:"{urn:ietf:params:xml:ns:caldav}calendar-timezone"`
				UrnIetfParamsXMLNsCaldavScheduleCalendarTransp struct {
				} `json:"{urn:ietf:params:xml:ns:caldav}schedule-calendar-transp"`
				UrnIetfParamsXMLNsCaldavSupportedCalendarComponentSet struct {
				} `json:"{urn:ietf:params:xml:ns:caldav}supported-calendar-component-set"`
			} `json:"calendarData"`
			CalendarID int    `json:"calendarId"`
			EventType  string `json:"eventType"`
			ObjectData struct {
				Calendardata   string `json:"calendardata"`
				Calendarid     string `json:"calendarid"`
				Classification int    `json:"classification"`
				Component      string `json:"component"`
				Etag           string `json:"etag"`
				ID             string `json:"id"`
				Lastmodified   string `json:"lastmodified"`
				Size           int    `json:"size"`
				URI            string `json:"uri"`
			} `json:"objectData"`
			Shares []interface{} `json:"shares"`
		} `json:"data"`
		Headers struct {
			AcceptEncoding      string `json:"Accept-Encoding"`
			ContentLength       string `json:"Content-Length"`
			ContentType         string `json:"Content-Type"`
			MattermostSessionID string `json:"Mattermost-Session-Id"`
			UserAgent           string `json:"User-Agent"`
			XForwardedFor       string `json:"X-Forwarded-For"`
			XForwardedProto     string `json:"X-Forwarded-Proto"`
			XWebhookSignature   string `json:"X-Webhook-Signature"`
		} `json:"headers"`
		HTTPMethod string `json:"httpMethod"`
		RawQuery   string `json:"rawQuery"`
	} `json:"values"`
	apps.Context `json:"context"`
}

type AppBindingsEmbeddedMessage struct {
	Location    string `json:"location"`
	Label       string `json:"label"`
	AppID       string `json:"app_id"`
	Description string `json:"description"`
	Bindings    []struct {
		Location string `json:"location"`
		Label    string `json:"label"`
		Submit   string `json:"submit,omitempty"`
		Bindings []struct {
			Location string `json:"location"`
			Label    string `json:"label"`
			Submit   string `json:"submit"`
		} `json:"bindings,omitempty"`
	} `json:"bindings"`
}

type CalendarEventDto struct {
	CalendarId     string          `json:"calendar_id"`
	ID             string          `json:"id"`
	Summary        string          `json:"summary"`
	Description    string          `json:"description"`
	Start          string          `json:"start_date"`
	End            string          `json:"end_date"`
	OrganizerEmail string          `json:"organizer"`
	Attendees      []*ics.Attendee `json:"attendees"`
}

func (e CalendarEventDto) GetFormattedStartDate(outputFormat string) string {
	date, error := time.Parse(icalTimestampFormatUtc, e.Start)

	if error != nil {
		date, _ := time.Parse(icalTimestampFormatUtcLocal, e.Start)
		return date.Format(outputFormat)
	}

	return date.Format(outputFormat)
}

func (e CalendarEventDto) GetFormattedEndDate(outputFormat string) string {
	date, error := time.Parse(icalTimestampFormatUtc, e.End)

	if error != nil {
		date, _ := time.Parse(icalTimestampFormatUtcLocal, e.End)
		return date.Format(outputFormat)
	}

	return date.Format(outputFormat)
}
