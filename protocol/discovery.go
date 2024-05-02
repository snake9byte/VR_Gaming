package protocol

import (
	"bytes"
	"html/template"
	"strings"
	"time"
)

type DiscoveryRequest struct {
	RemoteHost string
	Action     string
	Host       string
	Man        string
	Mx         string
	St         string
}

func trimQuotes(input string) string {
	return strings.TrimLeft(strings.TrimRight(input, "\""), "\"")
}

// IsDeviceRequest returns true if the ST urn maps to a Belkin device
func (dp *DiscoveryRequest) IsDeviceRequest() bool {
	return "urn:Belkin:device:**" == dp.St
}

func (dp *DiscoveryRequest) parseLine(line string) {
	if strings.Contains(line, "M-SEARCH *") {
		dp.Action = line
	} else if strings.Contains(line, ":") {

		parts := strings.SplitN(line, ":", 2)
		value := strings.TrimSpace(parts[1])

		switch parts[0] {
		case "HOST":
			dp.Host = value
		case "MAN":
			dp.Man = trimQuotes(value)
		case "MX":
			dp.Mx = value
		case "ST":
			dp.St = value
		default:
			// unknown line or empty line
		}
	}
}

// ParseDiscoveryRequest parses a uPnP discovery packet into a DiscoveryRequest
func ParseDiscoveryRequest(packet []byte) (req *DiscoveryRequest) {
	req = &DiscoveryRequest{}

	buf := &bytes.Buffer{}
	for _, b := range packet {
		switch b {
		case 13:
		// does nothing
		case 10:
			req.parseLine(string(buf.Bytes()))
			buf.Reset()
			continue
		default:
			buf.WriteByte(b)
		}
	}
	return req
}

// DiscoveryResponseParams is the parameters for the discovery response template
type DiscoveryResponseParams struct {
	DeviceName string
	ServerIP   string
	ServerPort int
	UUID       string
	UU         string
	Date       string
}

var drTemplateText = `HTTP/1.1 200 OK
CACHE-CONTROL: max-age=86400
DATE: {{.Date}}
EXT:
LOCATION: http://{{.ServerIP}}:{{.ServerPort}}/settings.xml
OPT: "http://schemas.upnp.org/upnp/1/0/"; ns=01
01-NLS: {{.UUID}}
SERVER: Unspecified, UPnP/1.0, Unspecified
X-User-Agent: redsonic
ST: urn:Belkin:device:**
USN: uuid:Socket-1_0-{{.UU}}::urn:Belkin:device:**

`
var drTemplate = template.Must(template.New("d").Parse(drTemplateText))

// BuildDiscoveryResponse creates the response packet to a discovery request
func BuildDiscoveryResponse(parms *DiscoveryResponseParams) []byte {

	parms.Date = time.Now().UTC().Format(time.RFC1123)

	b := &bytes.Buffer{}

	drTemplate.Execute(b, parms)

	return b.Bytes()
}
