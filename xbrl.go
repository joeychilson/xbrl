package xbrl

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	strip "github.com/grokify/html-strip-tags-go"
)

// XBRL represents the parsed XBRL data.
type XBRL struct {
	Facts []Fact `json:"facts"`
}

// Fact represents a single fact in the XBRL data.
type Fact struct {
	Context  Context `json:"context"`
	Concept  string  `json:"concept"`
	Value    any     `json:"value"`
	Decimals string  `json:"decimals,omitempty"`
	Unit     string  `json:"unit,omitempty"`
}

// Context represents the context of a fact in the XBRL data.
type Context struct {
	Entity   string    `json:"entity"`
	Segments []Segment `json:"segments"`
	Period   Period    `json:"period"`
}

// Segment represents a segment in the context of a fact in the XBRL data.
type Segment struct {
	Dimension string `json:"dimension"`
	Member    string `json:"member"`
}

// Period represents the period of a fact in the XBRL data.
type Period struct {
	Instant   string `json:"instant,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

// UnmarshalXML decodes the XML data into the XBRL struct.
func (x *XBRL) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw rawXBRL
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	units := make(map[string]string)
	for _, unit := range raw.Units {
		if unit.Divide.Numerator != "" && unit.Divide.Denominator != "" {
			units[unit.ID] = getLastPart(fmt.Sprintf("%s/%s", unit.Divide.Numerator, unit.Divide.Denominator), ':')
		} else {
			units[unit.ID] = getLastPart(unit.Measure, ':')
		}
	}

	contexts := make(map[string]Context)
	for _, ctx := range raw.Contexts {
		segments := make([]Segment, 0)
		for _, seg := range ctx.Segments {
			for _, member := range seg.Members {
				segments = append(segments, Segment{
					Dimension: strings.TrimSuffix(getLastPart(member.Dimension, ':'), "Axis"),
					Member:    strings.TrimSuffix(getLastPart(member.Value, ':'), "Member"),
				})
			}
		}
		contexts[ctx.ID] = Context{
			Entity:   ctx.Entity,
			Segments: segments,
			Period: Period{
				Instant:   ctx.Period.Instant,
				StartDate: ctx.Period.StartDate,
				EndDate:   ctx.Period.EndDate,
			},
		}
	}

	facts := make([]Fact, 0, len(raw.Facts))
	for _, fact := range raw.Facts {
		context, ok := contexts[fact.ContextRef]
		if !ok {
			continue
		}

		var value any

		unit, ok := units[fact.UnitRef]
		if !ok {
			value = parseText(fact.Value)
		} else {
			value = parseValue(fact.Value)
		}

		facts = append(facts, Fact{
			Context:  context,
			Concept:  fact.XMLName.Local,
			Value:    value,
			Decimals: fact.Decimals,
			Unit:     unit,
		})
	}
	x.Facts = facts
	return nil
}

// rawXBRL represents the raw XML structure of the XBRL data.
type rawXBRL struct {
	XMLName xml.Name `xml:"xbrl"`
	Units   []struct {
		ID      string `xml:"id,attr"`
		Measure string `xml:"measure"`
		Divide  struct {
			Numerator   string `xml:"unitNumerator>measure"`
			Denominator string `xml:"unitDenominator>measure"`
		} `xml:"divide"`
	} `xml:"unit"`
	Contexts []struct {
		ID       string `xml:"id,attr"`
		Entity   string `xml:"entity>identifier"`
		Segments []struct {
			XMLName xml.Name `xml:"segment"`
			Members []struct {
				XMLName   xml.Name `xml:"explicitMember"`
				Dimension string   `xml:"dimension,attr"`
				Value     string   `xml:",innerxml"`
			} `xml:"explicitMember"`
		} `xml:"entity>segment"`
		Period struct {
			Instant   string `xml:"instant"`
			StartDate string `xml:"startDate"`
			EndDate   string `xml:"endDate"`
		} `xml:"period"`
	} `xml:"context"`
	Facts []struct {
		XMLName    xml.Name `xml:""`
		ContextRef string   `xml:"contextRef,attr"`
		Value      string   `xml:",chardata"`
		Decimals   string   `xml:"decimals,attr"`
		UnitRef    string   `xml:"unitRef,attr"`
	} `xml:",any"`
}

// parseText removes HTML tags and newlines from the given value and returns just the text.
func parseText(value string) string {
	stripHTML := strip.StripTags(strings.ReplaceAll(value, "\n", ""))
	words := strings.Fields(stripHTML)
	return strings.Join(words, " ")
}

// parseValue converts the given value to a boolean, integer, float, or returns the original value.
func parseValue(value string) any {
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	}
	if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intValue
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	return value
}

// getLastPart returns the last part of the string after the given separator.
func getLastPart(s string, sep rune) string {
	pos := strings.LastIndexByte(s, byte(sep))
	if pos == -1 {
		return s
	}
	return s[pos+1:]
}
