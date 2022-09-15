package user

type UserSettings struct {
	DisabledCalendars []string `json:"disabled_calendars"`
}

func (u UserSettings) Contains(str string) bool {
	for _, v := range u.DisabledCalendars {
		if v == str {
			return true
		}
	}
	return false
}

func (u UserSettings) RemoveDisabledCalendar(el string) UserSettings {
	temp := u.DisabledCalendars[:0]
	for _, x := range u.DisabledCalendars {
		if !u.Contains(el) {
			temp = append(temp, x)
		}
	}
	return UserSettings{temp}
}

func (u UserSettings) AddDisabledCalendar(el string) UserSettings {
	if u.Contains(el) {
		return u
	}
	return UserSettings{append(u.DisabledCalendars, el)}

}
