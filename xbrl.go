package xbrl

import (
	"encoding/xml"
	"fmt"
	"strings"

	strip "github.com/grokify/html-strip-tags-go"
)

// XBRL represents the parsed XBRL data.
type XBRL struct {
	Facts []Fact `json:"facts"`
}

// Fact represents a single fact in the XBRL data.
type Fact struct {
	Context  *Context `json:"context"`
	Concept  string   `json:"concept"`
	Value    string   `json:"value"`
	Decimals string   `json:"decimals,omitempty"`
	Unit     string   `json:"unit,omitempty"`
}

// Context represents the context of a fact in the XBRL data.
type Context struct {
	Entity   string    `json:"entity"`
	Segments []Segment `json:"segments"`
	Period   *Period   `json:"period"`
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

// Parse takes a string of XBRL data and returns a XBRL struct with the parsed data.
func Parse(xbrl string) (*XBRL, error) {
	var doc struct {
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

	err := xml.Unmarshal([]byte(xbrl), &doc)
	if err != nil {
		return nil, err
	}

	units := make(map[string]string)
	for _, unit := range doc.Units {
		if unit.Divide.Numerator != "" && unit.Divide.Denominator != "" {
			units[unit.ID] = fmt.Sprintf("%s/%s", unit.Divide.Numerator, unit.Divide.Denominator)
		} else {
			units[unit.ID] = unit.Measure
		}
	}

	contexts := make(map[string]*Context)
	for _, ctx := range doc.Contexts {
		segments := make([]Segment, 0)
		for _, seg := range ctx.Segments {
			for _, member := range seg.Members {
				dimensionParts := strings.Split(member.Dimension, ":")
				dimension := dimensionParts[len(dimensionParts)-1]

				memberParts := strings.Split(member.Value, ":")
				memberValue := memberParts[len(memberParts)-1]

				segments = append(segments, Segment{
					Dimension: dimension,
					Member:    memberValue,
				})
			}
		}

		contexts[ctx.ID] = &Context{
			Entity:   ctx.Entity,
			Segments: segments,
			Period: &Period{
				Instant:   ctx.Period.Instant,
				StartDate: ctx.Period.StartDate,
				EndDate:   ctx.Period.EndDate,
			},
		}
	}

	facts := make([]Fact, 0)
	for _, fact := range doc.Facts {
		context, ok := contexts[fact.ContextRef]
		if !ok {
			continue
		}
		unit := units[fact.UnitRef]
		facts = append(facts, Fact{
			Context:  context,
			Concept:  fact.XMLName.Local,
			Value:    cleanValue(fact.Value),
			Decimals: fact.Decimals,
			Unit:     cleanUnit(unit),
		})
	}

	return &XBRL{Facts: facts}, nil
}

func cleanValue(value string) string {
	stripHTML := strip.StripTags(value)
	stripSpaces := strings.ReplaceAll(stripHTML, "\n", "")
	words := strings.Fields(stripSpaces)
	cleanedValue := strings.Join(words, " ")
	return cleanedValue
}

func cleanUnit(unit string) string {
	unitParts := strings.Split(unit, ":")
	if len(unitParts) == 1 {
		return unitParts[0]
	}
	return unitParts[len(unitParts)-1]
}
