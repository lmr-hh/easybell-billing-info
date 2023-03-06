package easybell

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"
)

const (
	NationalMinutePrice = 0.0083
	MobileMinutePrice   = 0.0824
)

type CallDirection string

const (
	AnyCall            CallDirection = "*"
	AnyOutbound        CallDirection = "1"
	AnyInbound         CallDirection = "2"
	SuccessfulOutbound CallDirection = "11"
	SuccessfulInbound  CallDirection = "21"
	FailedOutbound     CallDirection = "12"
	FailedInbound      CallDirection = "22"
)

type CallType string

const (
	AnyCallType        CallType = "*"
	ForwardCallType    CallType = "forward"
	RegularCallType    CallType = "call"
	VoiceboxCallType   CallType = "voicebox"
	Fax2MailCallType   CallType = "fax2mail"
	SMSCallType        CallType = "sms"
	ConferenceCallType CallType = "conference"
)

type CallKind string

const (
	AnyCallKind           CallKind = "*"
	MobileCallKind        CallKind = "mobile"
	NationalCallKind      CallKind = "national"
	InternationalCallKind CallKind = "international"
)

type CallLogFilter struct {
	NumberFilter    string
	PartnerFilter   string
	DirectionFilter CallDirection
	CallTypeFilter  CallType
	CallKindFilter  CallKind
}

type Client struct {
	httpClient *http.Client
}

func NewClient() Client {
	return Client{
		httpClient: &http.Client{
			Jar: &cookiejar.Jar{},
		},
	}
}

func (e *Client) ensureHttpClient() {
	if e.httpClient == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic(err)
		}
		e.httpClient = &http.Client{
			Jar: jar,
		}
	}
}

func (e *Client) Login(username string, password string) error {
	e.ensureHttpClient()
	resp, err := e.httpClient.PostForm("https://login.easybell.de/login", url.Values{
		"id":       []string{username},
		"password": []string{password},
	})
	if err == nil {
		return resp.Body.Close()
	}
	return err
}

func (e *Client) Logout() error {
	e.ensureHttpClient()
	resp, err := e.httpClient.Get("https://login.easybell.de/logout")
	if err == nil {
		return resp.Body.Close()
	}
	return err
}

func (e *Client) GetCallLogPage(start time.Time, end time.Time, size int, page int, filter *CallLogFilter) (*CallLogPage, error) {
	e.ensureHttpClient()

	query := url.Values{}

	query.Set("s", strconv.FormatInt(start.Unix(), 10))
	query.Set("e", strconv.FormatInt(end.Unix(), 10))
	query.Set("size", strconv.Itoa(size))
	query.Set("page", strconv.Itoa(page))

	if filter.NumberFilter != "" {
		query.Set("filter_RUFNUMMER", filter.NumberFilter)
	}
	if filter.PartnerFilter != "" {
		query.Set("filter_PARTNER", filter.PartnerFilter)
	}
	if filter.DirectionFilter != "" {
		query.Set("filter_RICHTUNG", string(filter.DirectionFilter))
	}
	if filter.CallTypeFilter != "" {
		query.Set("filter_TYPE", string(filter.CallTypeFilter))
	}
	if filter.CallKindFilter != "" {
		query.Set("filter_ART", string(filter.CallKindFilter))
	}
	resp, err := e.httpClient.Get("https://login.easybell.de/call-log/ajax?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var callLog CallLogPage
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&callLog)
	if err != nil {
		return nil, err
	}
	return &callLog, nil
}

func (e *Client) IterateCallLog(start time.Time, end time.Time, filter *CallLogFilter, handler func(entry *CallLogEntry) bool) error {
	lastPage := 1
	for page := 1; page <= lastPage; page++ {
		callLogPage, err := e.GetCallLogPage(start, end, 1000, page, filter)
		if err != nil {
			return err
		}
		lastPage = callLogPage.LastPage
		for _, entry := range callLogPage.Calls {
			if !handler(entry) {
				return nil
			}
		}
	}
	return nil
}

func (e *Client) CollectCalls(start time.Time, end time.Time, filter *CallLogFilter) (national int, mobile int, other int, err error) {
	err = e.IterateCallLog(start, end, filter, func(entry *CallLogEntry) bool {
		if entry.Kind == NationalCallKind {
			national += entry.Minutes
		} else if entry.Kind == MobileCallKind {
			mobile += entry.Minutes
		} else if entry.Kind == InternationalCallKind {
			other += entry.Minutes
		}
		return true
	})
	return
}
