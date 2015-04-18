package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	args := struct {
		inFile    string
		outFile   string
		debugFile string
		blacklist string
		whitelist string
	}{}
	flag.StringVar(&args.inFile, "in",
		"en.atm.co2e.kt_Indicator_en_xml_v2.xml",
		"The XML file of data to unmarshal")
	flag.StringVar(&args.debugFile, "debug",
		"",
		"The file to output marshalled JSON to.")
	flag.StringVar(&args.outFile, "out",
		"co2e.json",
		"A file for debugging intermediary state.")
	flag.StringVar(&args.blacklist, "blacklist", "",
		"A file of blacklist abbreviation ('key' in the source file)"+" values of records to exclude")
	flag.StringVar(&args.whitelist, "whitelist", "",
		"A file of whilelist abbreviation ('key' in the source file)"+" values of records to include")
	flag.Parse()

	in, err := decruft(args.inFile)
	if args.debugFile != "" {
		ioutil.WriteFile(args.debugFile, in, 0644)
	}
	if err != nil {
		log.Fatal(err)
	}

	var r Root
	err = xml.Unmarshal(in, &r)
	if err != nil {
		log.Fatal(err)
	}

	if args.blacklist != "" {
		b, err := ioutil.ReadFile(args.blacklist)
		if err != nil {
			log.Fatal(err)
		}
		bl := strings.Split(string(b), "\n")
		m := make(map[string]struct{}, len(bl))
		for _, s := range bl {
			m[s] = struct{}{}
		}
		r.Records.blacklist(m)
	}

	err = toJSON(r.Records, args.outFile)
	if err != nil {
		log.Fatal(err)
	}
}

// Types for unmarshalling XML.
type Root struct { // ???(aoeu): Why is this struct needed?
	Records `xml:"Data>Record"`
}

type Records []Record

type Record struct {
	Country string
	Abbr    string
	Year    int
	Value   float64
}

func (r *Records) blacklist(abbr map[string]struct{}) {
	s := make(Records, len(*r))
	for _, rr := range *r {
		if _, ok := abbr[rr.Abbr]; !ok {
			s = append(s, rr)
		}
	}
	*r = s
}

func (r *Records) whitelist(abbr map[string]struct{}) {
	s := make(Records, len(*r))
	for _, rr := range *r {
		if _, ok := abbr[rr.Abbr]; ok {
			s = append(s, rr)
		}
	}
	*r = s
}

func toJSON(rr []Record, outFile string) error {
	// Country -> Year -> C02 emission value
	t := make(map[string]map[string]float64)
	for _, r := range rr {
		if _, ok := t[r.Country]; !ok {
			t[r.Country] = make(map[string]float64)
		}
		// TODO(aoeu): What's the best way to filter out 0 Values?
		// TODO(aoeu): Do some countries need backfilled dummy values for some years?
		if r.Year < 2011 {
			t[r.Country][strconv.Itoa(r.Year)] = r.Value
		}
	}
	// The data is arranged as a series of JSON objects.
	// This is either to appease d3 or some intermediary javascript.
	j := make([]map[string]interface{}, 0)
	for c, y := range t {
		e := make(map[string]interface{})
		e["Country"] = c
		for k, v := range y {
			e[k] = v
		}
		j = append(j, e)
	}
	b, err := json.MarshalIndent(j, "", "	")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outFile, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func decruft(filename string) ([]byte, error) {
	// TODO(aoeu): Implement a less suboptimal solution.
	args := []string{
		`-e`, `s/xmlns.*>/>/`,
		`-e`, `s/data/Data/`,
		`-e`, `s/record/Record/`,
		`-e`, `s/Country or Area/Country/`,
		`-e`, `s/field name="//`,
		`-e`, `s/"//`,
		`-e`, `s/<\(\w*\)\(.*<\/\)field/<\1\2\1/`, // Case in point.
		`-e`, `s/<Value \/>/<Value>0.0<\/Value>/`,
		`-e`, `s/\(^.*Country\) key="\(.*\)"\(>.*$\)/\1\3\n\<Abbr\>\2\<\/Abbr\>/`,
		filename,
	}
	return exec.Command("sed", args...).Output()
}
