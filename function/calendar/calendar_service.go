package calendar

import (
	"encoding/xml"
	"errors"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	icalTimestampFormatUtcLocal = "20060102T150405"
	icalTimestampFormatUtc      = "20060102T150405Z"
)

type CalendarService interface {
	CreateEvent(body string) (*http.Response, error)
	GetUrl() string
	GetCalendarEvent() (string, error)
	DeleteUserEvent() (*http.Response, error)
	GetUserCalendars() []apps.SelectOption
	GetCalendarEvents(event CalendarEventRequestRange) []CalendarEventData
	UpdateAttendeeStatus(cal *ics.Calendar, user *model.User, status string) (string, error)
	AddButtonsToEvents(commandBinding apps.Binding, status string, path string) apps.Binding
}

type CalendarServiceImpl struct {
	calendarRequestService CalendarRequestService
}

func (c CalendarServiceImpl) GetUrl() string {
	return c.calendarRequestService.getUrl()

}

func (c CalendarServiceImpl) CreateEvent(body string) (*http.Response, error) {
	return c.calendarRequestService.createEvent(body)
}

func (c CalendarServiceImpl) GetUserCalendars() []apps.SelectOption {

	selectOptions := make([]apps.SelectOption, 0)
	calendarsResponse, err := c.calendarRequestService.getUserCalendars()

	if err != nil {
		return selectOptions
	}

	for _, r := range calendarsResponse.Response {

		calendarName := r.Propstat.Prop.Displayname

		if len(calendarName) > 0 {
			splitUrl := strings.Split(r.Href, "/")
			val := splitUrl[len(splitUrl)-2]
			selectOption := apps.SelectOption{
				Label: calendarName,
				Value: val,
			}
			selectOptions = append(selectOptions, selectOption)
		}
	}
	return selectOptions
}

func (c CalendarServiceImpl) GetCalendarEvents(event CalendarEventRequestRange) []CalendarEventData {
	log.Infof("Sending get calendar events request with a date range from %s to %s", event.From.String(), event.To.String())
	calendarEventData := make([]CalendarEventData, 0)

	resp, err := c.calendarRequestService.getCalendarEvents(event)
	if err != nil {
		return calendarEventData
	}

	for _, r := range resp.Response {
		eventData := CalendarEventData{}
		eventData.CalendarStr = r.Propstat.Prop.CalendarData
		eventData.CalendarId = getEventUrlByResponse(r.Href)
		calendarEventData = append(calendarEventData, eventData)
	}
	return calendarEventData
}

func getEventUrlByResponse(href string) string {
	return strings.Split(href, "/")[6]
}

func (c CalendarServiceImpl) UpdateAttendeeStatus(cal *ics.Calendar, user *model.User, status string) (string, error) {
	if cal == nil {
		return "", errors.New("this event is no longer valid")
	}
	for _, e := range cal.Events() {
		for _, a := range e.Attendees() {
			if user.Email == a.Email() {
				log.Infof("Updating event status in the event with id: %s for the user with an email %s", e.Id(), a.Email())
				a.ICalParameters["PARTSTAT"] = []string{status}
				break
			}
		}
	}
	return cal.Serialize(), nil
}

func (c CalendarServiceImpl) AddButtonsToEvents(commandBinding apps.Binding, status string, path string) apps.Binding {
	log.Info("Adding buttons to the event post")
	var label string
	if len(status) != 0 && status != "NEEDS-ACTION" {
		label = status
	} else {
		label = "Going?"
	}
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: "Going",
		Label:    label,
		Bindings: make([]apps.Binding, 0),
	})
	i := len(commandBinding.Bindings) - 1
	if status != "ACCEPTED" {
		commandBinding.Bindings[i].Bindings = append(commandBinding.Bindings[i].Bindings, apps.Binding{
			Location: "Accept",
			Label:    "Accept",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "accepted")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
				ActingUser:            apps.ExpandAll,
			}),
		})
		log.Info("Accept button added")
	}
	if status != "DECLINED" {
		commandBinding.Bindings[i].Bindings = append(commandBinding.Bindings[i].Bindings, apps.Binding{
			Location: "Decline",
			Label:    "Decline",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "declined")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
				ActingUser:            apps.ExpandAll,
			}),
		})
		log.Info("Declined button added")
	}

	if status != "TENTATIVE" {
		commandBinding.Bindings[i].Bindings = append(commandBinding.Bindings[i].Bindings, apps.Binding{
			Location: "Tentative",
			Label:    "Tentative",
			Submit: apps.NewCall(fmt.Sprintf("%s/%s", path, "tentative")).WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
				ActingUser:            apps.ExpandAll,
			}),
		})
		log.Info("Tentative button added")
	}
	return commandBinding
}

func (c CalendarServiceImpl) DeleteUserEvent() (*http.Response, error) {
	return c.calendarRequestService.deleteUserEvent()
}

func (c CalendarServiceImpl) GetCalendarEvent() (string, error) {
	return c.calendarRequestService.getCalendarEvent()
}

type CalendarRequestService interface {
	getUrl() string
	getCalendarEvent() (string, error)
	getUserCalendars() (UserCalendarsResponse, error)
	deleteUserEvent() (*http.Response, error)
	getCalendarEvents(event CalendarEventRequestRange) (UserCalendarEventsResponse, error)
	createEvent(body string) (*http.Response, error)
}

type CalendarRequestServiceImpl struct {
	Url   string
	Token string
}

func (c CalendarRequestServiceImpl) getUrl() string {
	return c.Url
}

func (c CalendarRequestServiceImpl) getUserCalendars() (UserCalendarsResponse, error) {

	body :=
		`<d:propfind xmlns:d="DAV:" xmlns:cs="http://calendarserver.org/ns/">
	<d:prop>
	   <d:displayname />
	   <cs:getctag />
	</d:prop>
  </d:propfind>`

	req, _ := http.NewRequest("PROPFIND", c.Url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	log.Info("Sending user get user calendar request")
	client := retryClient.StandardClient()
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during getting of the user calendars. Error: %s", err)
		return UserCalendarsResponse{}, err
	}

	if resp.StatusCode != http.StatusMultiStatus {
		log.Errorf("getUserCalendars request failed with status %s", resp.Status)
		respErr := fmt.Errorf("getUserCalendars request failed with code %d", resp.StatusCode)
		return UserCalendarsResponse{}, respErr
	}

	xmlResp := UserCalendarsResponse{}
	xmlError := xml.NewDecoder(resp.Body).Decode(&xmlResp)
	if xmlError != nil {
		log.Errorf("Error during xml decoding %s", xmlError.Error())
		return UserCalendarsResponse{}, xmlError
	}

	return xmlResp, nil
}

func (c CalendarRequestServiceImpl) deleteUserEvent() (*http.Response, error) {
	req, _ := http.NewRequest("DELETE", c.Url, nil)
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	log.Info("Sending a delete request")
	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during deleting of the event. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusNoContent {
		log.Errorf("getCalendarEvents request failed with status %s", resp.Status)
		respErr := fmt.Errorf("getCalendarEvents request failed with code %d", resp.StatusCode)
		return nil, respErr
	}

	return resp, nil
}

func (c CalendarRequestServiceImpl) getCalendarEvents(event CalendarEventRequestRange) (UserCalendarEventsResponse, error) {

	from := event.From.UTC().Format(icalTimestampFormatUtc)
	to := event.To.UTC().Format(icalTimestampFormatUtc)

	body := fmt.Sprintf(`<c:calendar-query xmlns:c="urn:ietf:params:xml:ns:caldav"
    xmlns:cs="http://calendarserver.org/ns/"
    xmlns:ca="http://apple.com/ns/ical/" 
    xmlns:d="DAV:">                                                            
    <d:prop>                
        <c:calendar-data />
    </d:prop>  
        <c:filter>
        <c:comp-filter name="VCALENDAR">
            <c:comp-filter name="VEVENT">
                <c:time-range start="%s" end="%s"/>
            </c:comp-filter>
        </c:comp-filter>
    </c:filter>
</c:calendar-query> `, from, to)

	req, _ := http.NewRequest("REPORT", c.Url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("Depth", "1")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during getting of the calendar events. Error: %s", err)
		return UserCalendarEventsResponse{}, err
	}

	if resp.StatusCode != http.StatusMultiStatus {
		log.Errorf("getCalendarEvents request failed with status %s", resp.Status)
		respErr := fmt.Errorf("getCalendarEvents request failed with code %d", resp.StatusCode)
		return UserCalendarEventsResponse{}, respErr
	}

	xmlResp := UserCalendarEventsResponse{}
	xmlError := xml.NewDecoder(resp.Body).Decode(&xmlResp)
	if xmlError != nil {
		log.Errorf("Error during xml decoding %s", xmlError.Error())
		return UserCalendarEventsResponse{}, xmlError
	}

	return xmlResp, err

}

func (c CalendarRequestServiceImpl) createEvent(body string) (*http.Response, error) {

	req, _ := http.NewRequest("PUT", c.Url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/calendar; charset=UTF-8")
	req.Header.Set("Depth", "0")
	req.Header.Set("X-NC-CalDAV-Webcal-Caching", "On")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	log.Info("Sending create event request to Nextcloud")
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during creating of the event. Error: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		log.Errorf("createEvent request failed with status %s", resp.Status)
		respErr := fmt.Errorf("createEvent request failed with code %d", resp.StatusCode)
		return nil, respErr
	}

	return resp, nil
}
func (c CalendarRequestServiceImpl) getCalendarEvent() (string, error) {
	req, _ := http.NewRequest("GET", c.Url, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	log.Info("Sending get calendar event request")
	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_RETRIES"))
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries

	client := retryClient.StandardClient()
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Errorf("Error during refreshing of the token. Error: %s", err)
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("getCalendarEvent request failed with status %s", resp.Status)
		respErr := fmt.Errorf("getCalendarEvent request failed with code %d", resp.StatusCode)
		return "", respErr
	}

	event, parsingErr := io.ReadAll(resp.Body)
	if parsingErr != nil {
		log.Errorf("Error during parsing of the event. Error: %s", err)
		return "", err
	}

	return string(event), nil
}
