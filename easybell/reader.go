package easybell

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"time"
)

// ErrBadFilter indicates that the filter values for a CallLogReader are invalid.
// If you receive this error you can correct the filter values and call [CallLogReader.Read] again.
var ErrBadFilter = errors.New("bad filter")

// NewCallLogReader creates a new reader that reads the call log entries in the specified time frame.
func NewCallLogReader(c *Client, start, end time.Time) *CallLogReader {
	return &CallLogReader{
		Client: c,
		Start:  start,
		End:    end,
	}
}

// CallLogReader implements in interface to read the easyBell call log.
type CallLogReader struct {
	// The Client provides the HTTP credentials for the reader.
	Client *Client

	// Filter options. Must not be modified after first Read call
	Start         time.Time
	End           time.Time
	NumberFilter  string
	PartnerFilter string
	Direction     string
	Type          string
	Kind          string

	// PageSize determines the number of call log entries that are fetched in a single go.
	// A maximum value is not documented.
	// The page size must not be changed after the first Read call.
	PageSize int

	buf     []*CallLogEntry
	i       int
	curPage int
}

// Reset resets r to the specified time frame.
// Calling Reset discards any data in the read buffer. The next call to r.Read will fetch the first page of the new time frame.
func (r *CallLogReader) Reset(start, end time.Time) {
	r.buf = r.buf[:0]
	r.i = 0
	r.curPage = 0
	r.Start = start
	r.End = end
}

// Read reads the next call log entry and returns it.
// If the read buffer is empty this method fetches the next page of calls from the easyBell API.
// If all calls have been read, the error will be io.EOF.
func (r *CallLogReader) Read() (*CallLogEntry, error) {
	if r.i >= len(r.buf) {
		if err := r.nextPage(); err != nil {
			return nil, err
		}
		if len(r.buf) == 0 {
			return nil, io.EOF
		}
	}
	entry := r.buf[r.i]
	r.i++
	return entry, nil
}

// nextPage fetches the next page of calls from the easyBell API.
func (r *CallLogReader) nextPage() error {
	r.curPage++
	r.i = 0

	query := url.Values{}
	query.Set("s", strconv.FormatInt(r.Start.Unix(), 10))
	query.Set("e", strconv.FormatInt(r.End.Unix(), 10))
	query.Set("page", strconv.Itoa(r.curPage))

	if r.PageSize > 0 {
		query.Set("size", strconv.Itoa(r.PageSize))
	}
	if r.NumberFilter != "" {
		query.Set("filter_RUFNUMMER", r.NumberFilter)
	}
	if r.PartnerFilter != "" {
		query.Set("filter_PARTNER", r.PartnerFilter)
	}
	if r.Direction != "" {
		query.Set("filter_RICHTUNG", string(r.Direction))
	}
	if r.Type != "" {
		query.Set("filter_TYPE", string(r.Type))
	}
	if r.Kind != "" {
		query.Set("filter_ART", string(r.Kind))
	}
	resp, err := r.Client.httpClient.Get("https://login.easybell.de/call-log/ajax?" + query.Encode())
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	decoder := json.NewDecoder(resp.Body)
	// There is a "last_page" field in the response.
	// However, it does not seem to be trustworthy.
	// It may always refer to the number of pages for a page size of 10
	// but in order to be safe we ignore it completely.
	page := struct {
		LastPage int             `json:"last_page"`
		Data     []*CallLogEntry `json:"data"`
	}{Data: r.buf}
	if err = decoder.Decode(&page); err != nil {
		return err
	}
	r.buf = page.Data
	if page.LastPage == 0 {
		return ErrBadFilter
	}
	return nil
}

// ReadUsage reads all calls from r and aggregates the used call minutes into a Usage value.
// When all calls have been read, the error will be nil.
// In particular io.EOF is not considered an error for this function.
func (r *CallLogReader) ReadUsage() (u Usage, err error) {
	var entry *CallLogEntry
	for {
		if entry, err = r.Read(); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return
		}
		switch entry.Kind {
		case CallKindNational:
			u.National += entry.Duration
		case CallKindMobile:
			u.Mobile += entry.Duration
		case CallKindInternational:
			u.Other += entry.Duration
		default:
			u.Other += entry.Duration
		}
	}
}

// Usage is a simple struct that holds information about used phone minutes.
type Usage struct {
	National time.Duration
	Mobile   time.Duration
	Other    time.Duration
}

// Total calculates the total phone time of u.
func (u Usage) Total() time.Duration {
	return u.National + u.Mobile + u.Other
}
