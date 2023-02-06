package calendar

import (
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
	"github.com/prokhorind/nextcloud/function/oauth"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CalendarPostService interface {
	CreateCalendarPost(option apps.SelectOption) *model.Post
}

type CalendarPostServiceImpl struct {
}

func (c CalendarPostServiceImpl) CreateCalendarPost(option apps.SelectOption) *model.Post {
	post := model.Post{}
	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       "Calendar " + option.Label,
		Description: "Calendar actions",
		Bindings:    []apps.Binding{},
	}

	c.createGetCalendarEventsButton(&commandBinding, option, "Calendar", "Today", "today")
	c.createGetCalendarEventsButton(&commandBinding, option, "Calendar", "Tomorrow", "tomorrow")
	c.createGetCalendarEventsButton(&commandBinding, option, "Calendar", "Select date", "select-date-form")
	c.createCalendarEventsButton(&commandBinding, option, "Calendar", "Create event")

	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)
	return &post

}

func (c CalendarPostServiceImpl) createGetCalendarEventsButton(commandBinding *apps.Binding, option apps.SelectOption, location apps.Location, label string, day string) {
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Submit: apps.NewCall("/get-calendar-events-" + day).WithExpand(apps.Expand{
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			ActingUserAccessToken: apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}).WithState(option),
	})
}

func (c CalendarPostServiceImpl) createCalendarEventsButton(commandBinding *apps.Binding, option apps.SelectOption, location apps.Location, label string) {
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Submit: apps.NewCall("/create-calendar-event-form").WithExpand(apps.Expand{
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			ActingUserAccessToken: apps.ExpandAll,
			ActingUser:            apps.ExpandAll,
		}).WithState(option),
	})
}

func (c CalendarPostServiceImpl) PrepareMeetingDurations() []apps.SelectOption {
	var durations []apps.SelectOption
	durations = append(durations, apps.SelectOption{
		Label: "15 minutes",
		Value: "15 minutes",
	})
	durations = append(durations, apps.SelectOption{
		Label: "30 minutes",
		Value: "30 minutes",
	})
	durations = append(durations, apps.SelectOption{
		Label: "45 minutes",
		Value: "45 minutes",
	})
	durations = append(durations, apps.SelectOption{
		Label: "1 hour",
		Value: "1 hour",
	})
	durations = append(durations, apps.SelectOption{
		Label: "1.5 hours",
		Value: "1.5 hours",
	})
	durations = append(durations, apps.SelectOption{
		Label: "2 hours",
		Value: "2 hours",
	})
	durations = append(durations, apps.SelectOption{
		Label: "All day",
		Value: "All day",
	})
	return durations
}

func (s DetailsViewFormService) createMeetingStartButton(commandBinding *apps.Binding, link string, location apps.Location) {
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    fmt.Sprintf("Join %s Meeting", location),
		Submit:   apps.NewCall("/redirect/meeting").WithState(link),
	})
}

type GetUserByEmailService interface {
	GetUserByEmail(email, etag string) (*model.User, *model.Response, error)
}

type EmailToNicknameCastService struct {
	GetMMUser GetMMUser
}

func (s EmailToNicknameCastService) сastUserEmailsToMMUserNicknames(attendees []*ics.Attendee) string {
	var attendeesNicknames string
	for _, attendee := range attendees {
		attendeesNicknames += s.сastSingleEmailToMMUserNickname(attendee.Email(), attendee.ICalParameters["PARTSTAT"][0])
	}
	if len(attendeesNicknames) != 0 {
		attendeesNicknames = attendeesNicknames[:len(attendeesNicknames)-1]
	}
	return attendeesNicknames
}

type DetailsViewFormService struct {
}

func (s EmailToNicknameCastService) сastSingleEmailToMMUserNickname(email string, status string) string {
	if strings.Contains(email, ":") {
		email = strings.Split(email, ":")[1]
	}
	mmUser, _, err := s.GetMMUser.GetUserByEmail(email, "")
	if err == nil {
		if status == "" {
			return "@" + mmUser.Username + "-" + email + " "
		}
		return "@" + mmUser.Username + "-" + email + "-" + status + " "
	} else {
		return email + "-" + status + " "
	}
}

func (s DetailsViewFormService) createDateForEventInForm(postDTO *CalendarEventPostDTO) string {
	locale := postDTO.creq.Context.ActingUser.Locale
	dateFormatService := DateFormatLocaleService{}
	parsedLocale := dateFormatService.GetLocaleByTag(locale)
	start, _ := postDTO.event.GetStartAt()
	finish, _ := postDTO.event.GetEndAt()

	format := dateFormatService.GetTimeFormatsByLocale(parsedLocale)
	dayFormat := dateFormatService.GetFullFormatsByLocale(parsedLocale)

	return fmt.Sprintf("%s %s-%s", start.In(postDTO.loc).Format(dayFormat), start.In(postDTO.loc).Format(format), finish.In(postDTO.loc).Format(format))
}

func (s CreateCalendarEventPostService) createNameForEvent(name string, postDTO *CalendarEventPostDTO) string {
	locale := postDTO.creq.Context.ActingUser.Locale
	dateFormatService := DateFormatLocaleService{}
	parsedLocale := dateFormatService.GetLocaleByTag(locale)
	start, _ := postDTO.event.GetStartAt()
	finish, _ := postDTO.event.GetEndAt()

	format := dateFormatService.GetTimeFormatsByLocale(parsedLocale)
	dayFormat := dateFormatService.GetFullFormatsByLocale(parsedLocale)
	day := strconv.Itoa(start.Day())
	month := strconv.Itoa(int(start.Month()))
	if len(day) < 2 {
		day = "0" + day
	}
	if len(month) < 2 {
		month = "0" + month
	}
	remoteUrl := postDTO.creq.Context.OAuth2.RemoteRootURL
	calendarUrl := fmt.Sprintf("%s%s%s-%s-%s", remoteUrl, "/apps/calendar/timeGridDay/", strconv.Itoa(start.Year()), month, day)
	return fmt.Sprintf("[%s](%s) %s %s-%s", name, calendarUrl, start.In(postDTO.loc).Format(dayFormat), start.In(postDTO.loc).Format(format), finish.In(postDTO.loc).Format(format))
}

type CalendarTimePostService struct {
}

func (s CalendarTimePostService) PrepareTimeRangeForGetEventsRequest(chosenDate time.Time) (time.Time, time.Time) {
	date := chosenDate.Add(-time.Minute * time.Duration(chosenDate.Minute()))
	date = date.Add(-time.Hour * time.Duration(chosenDate.Hour()))
	date = date.Add(-time.Second * time.Duration(chosenDate.Second()))
	return date.AddDate(0, 0, -1), date.AddDate(0, 0, 1)
}

func (s CalendarTimePostService) GetMMUserLocation(creq apps.CallRequest) *time.Location {
	var timezone string
	var loc *time.Location
	if creq.Context.ActingUser.Timezone["useAutomaticTimezone"] == "false" {
		timezone = creq.Context.ActingUser.Timezone["manualTimezone"]
	} else {
		timezone = creq.Context.ActingUser.Timezone["automaticTimezone"]
	}
	loc, _ = time.LoadLocation(timezone)
	return loc
}

func (s DetailsViewFormService) CreateViewButton(commandBinding *apps.Binding, location apps.Location, organizer string, label string, postDTO *CalendarEventPostDTO, formTitle string, icsLink string) {
	event := postDTO.event
	property := event.GetProperty(ics.ComponentPropertyDescription)
	var description string
	if property == nil {
		description = ""
	} else {
		description = property.Value
	}
	zoomLinks, googleMeetLinks := s.getZoomAndGoogleMeetLinksFromDescription(description)
	service := EmailToNicknameCastService{GetMMUser: postDTO.bot}

	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Form: &apps.Form{
			Title: formTitle,
			Fields: []apps.Field{
				{
					Type:       apps.FieldTypeText,
					Name:       "Date",
					Label:      "Date",
					ReadOnly:   true,
					Value:      s.createDateForEventInForm(postDTO),
					IsRequired: true,
				},
				{
					Type:        apps.FieldTypeText,
					Name:        "Description",
					Label:       "Description",
					ReadOnly:    true,
					Value:       description,
					TextSubtype: apps.TextFieldSubtypeTextarea,
				},
				{
					Type:                apps.FieldTypeStaticSelect,
					Name:                "Attendees",
					Label:               "Attendees",
					SelectIsMulti:       true,
					Value:               s.prepareAttendeeStaticSelect(service.сastUserEmailsToMMUserNicknames(event.Attendees())),
					SelectStaticOptions: s.prepareAttendeeStaticSelect(service.сastUserEmailsToMMUserNicknames(event.Attendees())),
					ReadOnly:            true,
				},
				{
					Type:       apps.FieldTypeText,
					Name:       "Organizer",
					Label:      "Organizer",
					ReadOnly:   true,
					IsRequired: true,
					Value:      service.сastSingleEmailToMMUserNickname(organizer, ""),
				},
			},
			Submit: apps.NewCall("/do-nothing"),
		},
	})
	i := len(commandBinding.Bindings) - 1
	if len(zoomLinks) != 0 {
		commandBinding.Bindings[i].Form.Fields = append(commandBinding.Bindings[i].Form.Fields, apps.Field{
			Type:        apps.FieldTypeText,
			Name:        "ZoomUrl",
			Label:       "ZoomLink",
			Value:       zoomLinks,
			ReadOnly:    true,
			IsRequired:  true,
			TextSubtype: apps.TextFieldSubtypeURL,
		})
		s.createMeetingStartButton(commandBinding, strings.Split(zoomLinks, " ")[0], "Zoom")
	}
	if len(googleMeetLinks) != 0 {
		commandBinding.Bindings[i].Form.Fields = append(commandBinding.Bindings[i].Form.Fields, apps.Field{
			Type:        apps.FieldTypeText,
			Name:        "GoogleMeetUrl",
			Label:       "Google-Meet-Link",
			Value:       googleMeetLinks,
			ReadOnly:    true,
			IsRequired:  true,
			TextSubtype: apps.TextFieldSubtypeURL,
		})
		s.createMeetingStartButton(commandBinding, strings.Split(googleMeetLinks, " ")[0], "Google Meet")
	}
	commandBinding.Bindings[i].Form.Fields = append(commandBinding.Bindings[i].Form.Fields, apps.Field{
		Type:        apps.FieldTypeText,
		Name:        "Event-Import",
		Label:       "Event-Import",
		Description: "Use this link to import event in your calendars",
		Value:       icsLink,
		ReadOnly:    true,
		IsRequired:  true,
		TextSubtype: apps.TextFieldSubtypeURL,
	})
}

func (s DetailsViewFormService) getZoomAndGoogleMeetLinksFromDescription(description string) (string, string) {
	zoomPattern := regexp.MustCompile(`https:\/\/[\w-]*\.?zoom.us\/(j|my)\/[\d\w?=-]+`)
	googleMeetPattern := regexp.MustCompile(`https?:\/\/(.+?\.)?meet\.google\.com(\/[A-Za-z0-9\-]*)?`)
	zoomLinks := zoomPattern.FindAllString(description, -1)
	googleMeetLinks := googleMeetPattern.FindAllString(description, -1)
	return strings.Join(zoomLinks, " "), strings.Join(googleMeetLinks, " ")
}

func (s DetailsViewFormService) prepareAttendeeStaticSelect(attendees string) []apps.SelectOption {
	options := make([]apps.SelectOption, 0)
	for _, a := range strings.Split(attendees, " ") {
		options = append(options, apps.SelectOption{Label: a, Value: a})
	}
	return options
}

type GetMMUser interface {
	GetUser(userId, etag string) (*model.User, *model.Response, error)
	GetUserByEmail(email, etag string) (*model.User, *model.Response, error)
	DMPost(userID string, post *model.Post) (*model.Post, error)
	GetUsersByIds(userIds []string) ([]*model.User, *model.Response, error)
}

type CreateCalendarEventPostService struct {
	GetMMUser GetMMUser
}

func (s CreateCalendarEventPostService) CreateCalendarEventPost(postDTO *CalendarEventPostDTO) *model.Post {
	var name, organizer, eventStatus string
	for _, e := range postDTO.event.Properties {
		if e.BaseProperty.IANAToken == "ORGANIZER" {
			organizer = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "SUMMARY" {
			name = e.BaseProperty.Value
		}
		if e.BaseProperty.IANAToken == "STATUS" {
			eventStatus = e.BaseProperty.Value
		}
	}

	userId := postDTO.creq.Context.OAuth2.User.(map[string]interface{})["user_id"].(string)
	remoteUrl := postDTO.creq.Context.OAuth2.OAuth2App.RemoteRootURL
	reqUrl := fmt.Sprintf("%s/remote.php/dav/calendars/%s/%s/%s", remoteUrl, userId, postDTO.calendarId, postDTO.eventId)

	post := model.Post{}
	commandBinding := apps.Binding{
		Location:    "embedded",
		AppID:       "nextcloud",
		Label:       s.createNameForEvent(name, postDTO),
		Description: "Going?",
		Bindings:    []apps.Binding{},
	}
	calendarService := CalendarServiceImpl{}

	if eventStatus == "CANCELLED" {
		commandBinding.Label = fmt.Sprintf("Cancelled ~~%s~~", commandBinding.Label)
		commandBinding.Description = fmt.Sprintf("~~%s~~", commandBinding.Description)
		m1 := make(map[string]interface{})
		m1["app_bindings"] = []apps.Binding{commandBinding}

		post.SetProps(m1)

		return &post
	}

	if strings.Contains(organizer, ":") {
		organizer = strings.Split(organizer, ":")[1]
	}
	organizerEmail := postDTO.creq.Context.ActingUser.Email
	status := s.FindAttendeeStatus(*postDTO.event, postDTO.creq.Context.ActingUser.Id)

	if organizerEmail != organizer {
		path := fmt.Sprintf("/users/%s/calendars/%s/events/%s/status", userId, postDTO.calendarId, postDTO.eventId)
		commandBinding = calendarService.AddButtonsToEvents(commandBinding, string(status), path)
	}

	deletePath := fmt.Sprintf("/delete-event/%s/events/%s", postDTO.calendarId, postDTO.eventId)
	detailButtonService := DetailsViewFormService{}
	detailButtonService.CreateViewButton(&commandBinding, "view-details", organizer, "View Details", postDTO, name, reqUrl)
	s.сreateDeleteButton(&commandBinding, "Delete", "Delete", deletePath)
	m1 := make(map[string]interface{})
	m1["app_bindings"] = []apps.Binding{commandBinding}

	post.SetProps(m1)

	return &post
}

func (s CreateCalendarEventPostService) FindAttendeeStatus(event ics.VEvent, userId string) ics.ParticipationStatus {
	user, _, _ := s.GetMMUser.GetUser(userId, "")
	for _, a := range event.Attendees() {
		if user.Email == a.Email() {
			return a.ParticipationStatus()
		}
	}
	return ""
}

func (s CreateCalendarEventPostService) сreateDeleteButton(commandBinding *apps.Binding, location apps.Location, label string, deletePath string) {
	expand := apps.Expand{
		OAuth2App:             apps.ExpandAll,
		OAuth2User:            apps.ExpandAll,
		ActingUserAccessToken: apps.ExpandAll,
		ActingUser:            apps.ExpandAll,
	}
	commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
		Location: location,
		Label:    label,
		Submit:   apps.NewCall(deletePath).WithExpand(expand),
	})
}

type OauthService interface {
	RefreshToken() oauth.Token
}

type GetEventsService struct {
	CalendarService                CalendarService
	CalendarTimePostService        CalendarTimePostService
	CreateCalendarEventPostService CreateCalendarEventPostService
	GetMMUser                      GetMMUser
}

func (s GetEventsService) GetUserEvents(creq apps.CallRequest, date time.Time, calendar string) error {
	loc := s.CalendarTimePostService.GetMMUserLocation(creq)

	mmUserId := creq.Context.ActingUser.Id

	from, to := s.CalendarTimePostService.PrepareTimeRangeForGetEventsRequest(date)
	eventRange := CalendarEventRequestRange{
		From: from,
		To:   to,
	}

	events, eventIds := s.CalendarService.GetCalendarEvents(eventRange)
	calendarEvents := make([]ics.VEvent, 0)

	for i := 0; i < len(events); i++ {
		cal, _ := ics.ParseCalendar(strings.NewReader(events[i]))
		event := *cal.Events()[0]
		if len(event.Properties) != 0 {
			calendarEvents = append(calendarEvents, event)
		}
	}

	dailyCalendarEvents := make([]ics.VEvent, 0)

	for _, e := range calendarEvents {
		at, _ := e.GetStartAt()
		endAt, _ := e.GetEndAt()
		localStartTime := at.In(loc)
		localEndTime := endAt.In(loc)
		if localStartTime.Day() == date.Day() || localEndTime.Day() == date.Day() {
			dailyCalendarEvents = append(dailyCalendarEvents, e)
		}
	}

	if len(dailyCalendarEvents) == 0 {
		return errors.New("You don`t have events at this day")
	}

	for i, e := range dailyCalendarEvents {
		postDto := CalendarEventPostDTO{&e, s.GetMMUser, calendar, eventIds[i], loc, creq}
		post := s.CreateCalendarEventPostService.CreateCalendarEventPost(&postDto)
		s.GetMMUser.DMPost(mmUserId, post)
	}

	return nil
}
