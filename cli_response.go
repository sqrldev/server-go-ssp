package ssp

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
)

//

// TIF bitflags
const (
	// the identity has been seen before
	TIFIDMatch = 0x1
	// the previous identity is a known identity
	TIFPreviousIDMatch = 0x2
	// the IP address of the current request and the original Nut request match
	TIFIPMatched = 0x4
	// the SQRL account is disabled
	TIFSQRLDisabled = 0x8
	// the ClientBody.Cmd is not recognized
	TIFFunctionNotSupported = 0x10
	// used for all the random server errors like failures to connect to datastores
	TIFTransientError = 0x20
	// the specific ClientBody.Cmd could not be completed for any reason
	TIFCommandFailed = 0x40
	// the client sent bad or unrecognized data or signature validation failed
	TIFClientFailure = 0x80
	// The owner of the Nut doesn't match this request
	TIFBadIDAssociation = 0x100
	// The IDK has been rekeyed to a newer one
	TIFIdentitySuperseded = 0x200
)

// TIFDesc description of the TIF bits
var TIFDesc = map[uint32]string{
	TIFIDMatch:              "ID Matched",
	TIFPreviousIDMatch:      "Previous ID Matched",
	TIFIPMatched:            "IP Matched",
	TIFSQRLDisabled:         "Identity disabled",
	TIFFunctionNotSupported: "Command not recognized",
	TIFTransientError:       "Server Error",
	TIFCommandFailed:        "Command failed",
	TIFClientFailure:        "Bad client request",
	TIFBadIDAssociation:     "Mismatch of nut to idk",
	TIFIdentitySuperseded:   "Identity superseded by newer one",
}

// CliResponse encodes a response to the SQRL client
// As specified https://www.grc.com/sqrl/semantics.htm
type CliResponse struct {
	Version []int
	Nut     Nut
	TIF     uint32
	Qry     string
	URL     string
	Sin     string
	Suk     string
	Ask     *Ask
	Can     string

	// HoardCache is not serialized but the encoded response is saved here
	// so we can check it in the next request
	HoardCache *HoardCache
}

// Ask holds optional response to queries that
// will be shown to the user from the SQRL client
type Ask struct {
	Message string `json:"message"`
	Button1 string `json:"button1,omitempty"`
	URL1    string `json:"url1,omitempty"`
	Button2 string `json:"button2,omitempty"`
	URL2    string `json:"url2,omitempty"`
}

// ParseAsk parses the special Ask format
func ParseAsk(askString string) *Ask {
	encparts := strings.Split(askString, "~")
	parts := make([]string, len(encparts))
	for i, e := range encparts {
		b, _ := Sqrl64.DecodeString(e)
		parts[i] = string(b)
	}
	ask := &Ask{
		Message: parts[0],
	}
	if len(parts) > 1 {
		ask.Button1, ask.URL1 = splitButton(parts[1])
	}
	if len(parts) > 2 {
		ask.Button2, ask.URL2 = splitButton(parts[2])
	}
	return ask
}

func splitButton(buttonString string) (string, string) {
	semi := strings.Index(buttonString, ";")
	if semi == -1 {
		return buttonString, ""
	}
	button := buttonString[0:semi]
	url := buttonString[semi+1:]
	return button, url
}

// Encode creates the tilde and semicolon separated ask format
func (a *Ask) Encode() string {
	delimited := make([]string, 1)
	delimited[0] = Sqrl64.EncodeToString([]byte(a.Message))
	button := encodeButton(a.Button1, a.URL1)
	if button != "" {
		delimited = append(delimited, button)
	}
	button = encodeButton(a.Button2, a.URL2)
	if button != "" {
		delimited = append(delimited, button)
	}
	return strings.Join(delimited, "~")
}

func encodeButton(button, url string) string {
	if button != "" {
		urlappend := ""
		if url != "" {
			urlappend = fmt.Sprintf(";%v", url)
		}
		return Sqrl64.EncodeToString([]byte(fmt.Sprintf("%s%s", removeSemi(button), urlappend)))
	}
	return ""
}

func removeTilde(v string) string {
	return strings.Replace(v, "~", "", -1)
}

func removeSemi(v string) string {
	return strings.Replace(v, ";", "", -1)
}

// NewCliResponse creates a minimal valid CliResponse object
func NewCliResponse(nut Nut, qry string) *CliResponse {
	return &CliResponse{
		Version: []int{1},
		Nut:     nut,
		Qry:     qry,
	}
}

// WithIDMatch set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithIDMatch() *CliResponse {
	cr.TIF |= TIFIDMatch
	return cr
}

// ClearIDMatch set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) ClearIDMatch() *CliResponse {
	cr.TIF = cr.TIF &^ TIFIDMatch
	return cr
}

// WithPreviousIDMatch set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithPreviousIDMatch() *CliResponse {
	cr.TIF |= TIFPreviousIDMatch
	return cr
}

// ClearPreviousIDMatch clears the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) ClearPreviousIDMatch() *CliResponse {
	cr.TIF = cr.TIF &^ TIFPreviousIDMatch
	return cr
}

// WithIPMatch set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithIPMatch() *CliResponse {
	cr.TIF |= TIFIPMatched
	return cr
}

// WithSQRLDisabled set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithSQRLDisabled() *CliResponse {
	cr.TIF |= TIFSQRLDisabled
	return cr
}

// WithFunctionNotSupported set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithFunctionNotSupported() *CliResponse {
	cr.TIF |= TIFFunctionNotSupported
	return cr
}

// WithTransientError set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithTransientError() *CliResponse {
	cr.TIF |= TIFTransientError
	return cr
}

// WithClientFailure set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithClientFailure() *CliResponse {
	cr.TIF |= TIFClientFailure
	return cr
}

// WithCommandFailed set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithCommandFailed() *CliResponse {
	cr.TIF |= TIFCommandFailed
	return cr
}

// WithBadIDAssociation set the appropriate TIF bits on this response.
// Returns the object for easier chaining (not immutability).
func (cr *CliResponse) WithBadIDAssociation() *CliResponse {
	cr.TIF |= TIFBadIDAssociation
	return cr
}

func (cr *CliResponse) WithIdentitySuperseded() *CliResponse {
	cr.TIF |= TIFIdentitySuperseded
	return cr
}

// Encode writes the response as the CRNL format and
// encodes it using Sqrl64 encoding.
func (cr *CliResponse) Encode() []byte {
	var b bytes.Buffer

	// TODO be less lazy and support ranges
	b.WriteString("ver=")
	for i, v := range cr.Version {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(strconv.Itoa(v))
	}
	b.WriteString("\r\n")

	b.WriteString(fmt.Sprintf("nut=%v\r\n", cr.Nut))

	b.WriteString(fmt.Sprintf("tif=%x\r\n", cr.TIF))

	b.WriteString(fmt.Sprintf("qry=%v\r\n", cr.Qry))

	if cr.URL != "" {
		b.WriteString(fmt.Sprintf("url=%v\r\n", cr.URL))
	}

	if cr.Sin != "" {
		b.WriteString(fmt.Sprintf("sin=%v\r\n", cr.Sin))
	}

	if cr.Suk != "" {
		b.WriteString(fmt.Sprintf("suk=%v\r\n", cr.Suk))
	}

	if cr.Ask != nil {
		b.WriteString(fmt.Sprintf("ask=%v\r\n", cr.Ask.Encode()))
	}

	if cr.Can != "" {
		b.WriteString(fmt.Sprintf("can=%v\r\n", cr.Can))
	}

	encoded := Sqrl64.EncodeToString(b.Bytes())
	log.Printf("Encoded response: <%v>", encoded)
	return []byte(encoded)
}

// ParseCliResponse parses a server response
func ParseCliResponse(body []byte) (*CliResponse, error) {
	decoded := make([]byte, Sqrl64.DecodedLen(len(body)))
	_, err := Sqrl64.Decode(decoded, body)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	params, err := ParseSqrlQuery(string(decoded))
	if err != nil {
		return nil, fmt.Errorf("couldn't parse response: %v", err)
	}

	tifbig, err := strconv.ParseUint(params["tif"], 16, 32)
	if err != nil {
		return nil, fmt.Errorf("can't parse tif: %v", err)
	}

	return &CliResponse{
		Version: []int{1},
		Nut:     Nut(params["nut"]),
		TIF:     uint32(tifbig),
		Qry:     params["qry"],
		URL:     params["url"],
		Sin:     params["sin"],
		Suk:     params["suk"],
		Ask:     ParseAsk(params["ask"]),
		Can:     params["can"],
	}, nil
}
