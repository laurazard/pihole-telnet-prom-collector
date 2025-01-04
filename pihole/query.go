package pihole

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// example: 1734952435 AAAA cc-api-data.adobe.io 192.168.1.14 1 0 4 0 N/A -1 N/A#0 ""
// from https://github.com/pi-hole/FTL/blob/61a211f1c187206f5ff901afae657968114fde15/src/api/api.c#L1055-L1070
type Query struct {
	Timestamp    time.Time  // 1734952435
	QueryType    string     // AAAA
	Domain       string     // cc-api-data.adobe.io
	ClientIP     string     // 192.168.1.14
	Status       StatusType // 1
	DNSSEC       int        // 0
	ReplyType    ReplyType  // 4
	DelayMs      float64    // 0
	CNAMEDomain  string     // N/A *only set if status == 2
	RegexMatchID int        // -1
	Upstream     string     // N/A#0 / 192.168.1.8#5335
	QueryEDE     string     // ???
}

type StatusType uint

const (
	Unknown StatusType = iota
	BlockedGravity
	AllowedForwarded
	AllowedCache
	BlockedRegex
	BlockedBlacklist
	BlockedUpstream
	BlockedUpstreamZeroReply
	BlockedUpstreamNXDomain
	BlockedGravityCNAMEInspection
	BlockedRegexCNAMEInspection
	BlockedBlacklistCNAMEInspection
	Retried
	RetriedIgnored // this may happen during ongoing DNSSEC validation
	AlreadyForwarded
	BlockedDBBusy
	BlockedSpecialDomain // such as mozilla canary/apple private relay
	RepliedStaleCache
)

func (s StatusType) String() string {
	switch s {
	case Unknown:
		return "unknown"
	case AllowedForwarded:
		fallthrough
	case AllowedCache:
		return "allowed"
	default:
		return "blocked"
	}
}

type ReplyType uint

const (
	Waiting ReplyType = iota
	NODATA
	NXDOMAIN
	CNAME
	IP
	DOMAIN
	RRNAME
	SERVFAIL
	REFUSED
	NOTIMP
	OTHER
	DNSSEC
	NONE // dropped intentinally
	BLOB
)

func (r ReplyType) String() string {
	switch r {
	case Waiting:
		return "waiting"
	case NODATA:
		return "NODATA"
	case NXDOMAIN:
		return "NXDOMAIN"
	case IP:
		return "IP"
	case CNAME:
		return "CNAME"
	default:
		return "unknown"
	}
}

var ErrParseQuery = errors.New("failed to parse query")

func parseQueryLine(line string) (Query, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 12 {
		println(line)
		return Query{}, ErrParseQuery
	}

	unixSeconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Query{}, ErrParseQuery
	}
	timestamp := time.Unix(unixSeconds, 0)

	statusType, err := strconv.Atoi(parts[4])
	if err != nil {
		return Query{}, ErrParseQuery
	}

	dnsSec, err := strconv.Atoi(parts[5])
	if err != nil {
		return Query{}, ErrParseQuery
	}

	replyType, err := strconv.Atoi(parts[6])
	if err != nil {
		return Query{}, ErrParseQuery
	}

	delayMsTenth, err := strconv.Atoi(parts[7])
	if err != nil {
		return Query{}, ErrParseQuery
	}
	delayMs := float64(delayMsTenth) / 10

	blockListID, err := strconv.Atoi(parts[9])
	if err != nil {
		return Query{}, ErrParseQuery
	}

	return Query{
		Timestamp:    timestamp,
		QueryType:    parts[1],
		Domain:       parts[2],
		ClientIP:     parts[3],
		Status:       StatusType(statusType),
		DNSSEC:       dnsSec,
		ReplyType:    ReplyType(replyType),
		DelayMs:      delayMs,
		CNAMEDomain:  parts[8],
		RegexMatchID: blockListID,
		Upstream:     parts[10],
		QueryEDE:     "",
	}, nil
}
