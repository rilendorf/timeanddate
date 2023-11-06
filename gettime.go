package timeanddate

import (
	// "gopkg.in/resty.v1"
	"github.com/PuerkitoBio/goquery"

	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"log"
)

var (
	ErrDeserialize = errors.New("Failed to deserialize")
)

type deserializeErr struct {
	Err  error
	Part string

	Data string
}

func (e *deserializeErr) Error() string {
	if e.Data != "" {
		return "deserialize " + e.Part + ": " + e.Err.Error() + " (" + e.Data + ")"
	}
	return "deserialize " + e.Part + ": " + e.Err.Error()
}

func wrap(err error, part string, data string) error {
	return &deserializeErr{
		Err:  err,
		Part: part,
		Data: data,
	}
}

type Position struct {
	Latitude  float32
	Longitude float32
}

func (c *Position) String() string {
	if c == nil {
		return "<nil>"
	}

	return fmt.Sprintf("Lat: %.2f; Lon: %.2f", c.Latitude, c.Longitude)
}

// e.g. 49°58'N / 9°09'E
func (c *Position) UnmarshalText(d []byte) error {
	var LatDeg, LatMin, LonDeg, LonMin float32
	var LatDir, LonDir string

	str := strings.ReplaceAll(string(d), "°", " ")
	str = strings.ReplaceAll(str, "'", " ")

	_, err := fmt.Sscanf(str, "%f %f %s / %f %f %s",
		&LatDeg, &LatMin, &LatDir,
		&LonDeg, &LonMin, &LonDir,
	)
	if err != nil {
		return err
	}

	c.Latitude = LatDeg + LatMin/60
	c.Longitude = LonDeg + LonMin/60

	if LatDir == "S" {
		c.Latitude *= -1
	}

	if LonDir == "W" {
		c.Latitude *= -1
	}

	return nil
}

type Currency struct {
	Name string
	Code string
}

// Euro (EUR)
func (c *Currency) UnmarshalText(d []byte) error {
	spl := strings.SplitN(string(d), " (", 2)

	if len(spl) != 2 {
		return ErrDeserialize
	}

	c.Name = spl[0]
	if len(spl[1]) > 1 {
		c.Code = spl[1][:len(spl[1])-1]
	}

	return nil
}

func (s *Currency) String() string {
	if s == nil {
		return "<nil>"
	}

	return s.Name + " (" + s.Code + ")"
}

type State struct {
	Name string
	Code string
}

// Bavaria (BY)
func (c *State) UnmarshalText(d []byte) error {
	spl := strings.SplitN(string(d), " (", 2)

	c.Name = spl[0]
	if len(spl) != 2 {
		return nil
	}

	if len(spl[1]) > 1 {
		c.Code = spl[1][:len(spl[1])-1]
	}

	return nil
}

func (s *State) String() string {
	if s == nil {
		return "<nil>"
	}

	return s.Name + " (" + s.Code + ")"
}

type Time struct {
	Hours, Minutes, Seconds int
}

func (t *Time) String() string {
	return fmt.Sprintf("%d:%d:%d", t.Hours, t.Minutes, t.Seconds)
}

func (t *Time) UnmarshalText(d []byte) error {
	_, err := fmt.Sscanf(string(d), "%d:%d:%d",
		&t.Hours, &t.Minutes, &t.Seconds,
	)

	return err
}

type Date struct {
	Weekday   time.Weekday
	Month     time.Month
	Day, Year int
}

func (t *Date) String() string {
	return fmt.Sprintf("%s, %d, %s %d", t.Weekday, t.Day, t.Month, t.Year)
}

func (t *Date) UnmarshalText(d []byte) error {
	// storing successive space-separated values into successive arguments
	str := strings.ReplaceAll(string(d), ",", " ")

	var month string
	var weekday string

	_, err := fmt.Sscanf(str, "%s  %d %s %d",
		&weekday, &t.Day, &month, &t.Year)

	t.Month = UnmarshalMonth(month)
	t.Weekday = UnmarshalWeekday(weekday)

	return err
}

func UnmarshalWeekday(s string) time.Weekday {
	switch s {
	case "Sunday":
		return time.Sunday
	case "Monday":
		return time.Monday
	case "Tuesday":
		return time.Tuesday
	case "Wednesday":
		return time.Wednesday
	case "Thursday":
		return time.Thursday
	case "Friday":
		return time.Friday
	case "Saturday":
		return time.Saturday
	}

	return 0
}

func UnmarshalMonth(s string) time.Month {
	switch s {
	case "January":
		return time.January
	case "February":
		return time.February
	case "March":
		return time.March
	case "April":
		return time.April
	case "May":
		return time.May
	case "June":
		return time.June
	case "July":
		return time.July
	case "August":
		return time.August
	case "September":
		return time.September
	case "October":
		return time.October
	case "November":
		return time.November
	case "December":
		return time.December
	}

	return 0
}

type TimeAndDate struct {
	Country    string    // e.g. Germany
	State      *State    // e.g. Bavaria (BY)
	Province   *State    // e.g. Bavaria (BY)
	Position   *Position // e.g. 49°58'N / 9°09'E
	Elevation  int       // in meters
	Currency   *Currency // e.g. Euro (EUR)
	Languages  []string  // e.g. German
	AccessCode string    //  e.g. +49

	Time *Time // e.g. 12:54:53
	Date *Date // e.g. Monday, 23 October 2023
}

// returns time in UTC because easier // TODO: get data from time/zone/
func (t *TimeAndDate) TimeTime() time.Time {
	// Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location
	return time.Date(t.Date.Year, t.Date.Month, t.Date.Day,
		t.Time.Hours, t.Time.Minutes, t.Time.Seconds, 0, time.UTC)
}

func Get(path string) (r *TimeAndDate, err error) {
	return DefaultClient.Get(path)
}

func (c *Client) Get(path string) (r *TimeAndDate, err error) {
	url := "https://www.timeanddate.com/" + path
	res, err := c.resty.R().Get(url)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.Body()))
	if err != nil {
		return
	}

	// 0: key, 1: value
	var data [][2]string

	// content
	doc.Find("table.table.table--left.table--inner-borders-rows").Each(func(i int, s *goquery.Selection) {
		// keys
		keys := s.Find("th")
		data = make([][2]string, len(keys.Nodes))
		keys.Each(func(i int, s *goquery.Selection) {
			if len(data) > i { // boundscheck
				data[i][0] = s.Text()
			}
		})

		// values
		values := s.Find("td")

		values.Each(func(i int, s *goquery.Selection) {
			if len(data) > i { // boundscheck
				data[i][1] = s.Text()
			}
		})
	})

	tad := new(TimeAndDate)
	var value string

	for i := 0; i < len(data); i++ {
		value = data[i][1]

		switch data[i][0] { // keys

		case "Country: ":
			tad.Country = value
			break

		case "State: ":
			tad.State = new(State)
			err = tad.State.UnmarshalText([]byte(value))
			if err != nil {
				return r, wrap(err, "state", value)
			}
			break

		case "Lat/Long: ":
			tad.Position = new(Position)
			err = tad.Position.UnmarshalText([]byte(value))
			if err != nil {
				return r, wrap(err, "position", value)
			}
			break

		case "Elevation: ":
			_, err = fmt.Sscanf(value, "%d m", &tad.Elevation)
			if err != nil {
				return r, wrap(err, "elevation", value)
			}
			break

		case "Currency: ":
			tad.Currency = new(Currency)
			err = tad.Currency.UnmarshalText([]byte(value))
			if err != nil {
				return r, wrap(err, "currency", value)
			}
			break

		case "Languages: ":
			// can have multible (e.g. Vancouver)
			tad.Languages = strings.Split(value, ", ")
			break

		case "Country Code: ":
			tad.AccessCode = value
			break

		case "Province: ":
			tad.Province = new(State)
			err = tad.Province.UnmarshalText([]byte(value))
			if err != nil {
				return r, wrap(err, "state", value)
			}
			break

		default:
			log.Printf("unhandled key '%s'", data[i][0])
		}
	}

	// get time & date:
	timeSel := doc.Find("span#ct.h1")
	if len(timeSel.Nodes) == 0 {
		return r, wrap(ErrDeserialize, "time", "len0")
	}

	tad.Time = new(Time)
	err = tad.Time.UnmarshalText([]byte(timeSel.Text()))
	if err != nil {
		return r, wrap(err, "time", timeSel.Text())
	}

	dateSel := doc.Find("span#ctdat")
	if len(dateSel.Nodes) == 0 {
		return r, wrap(ErrDeserialize, "date", "len0")
	}

	tad.Date = new(Date)
	err = tad.Date.UnmarshalText([]byte(dateSel.Text()))
	if err != nil {
		return r, wrap(err, "date", dateSel.Text())
	}

	return tad, err
}
