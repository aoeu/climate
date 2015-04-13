package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
)

var inputFilename string
var defaultPath = "en.atm.co2e.kt_Indicator_en_xml_v2.xml"

func main() {
	flag.StringVar(&inputFilename, "in",
		defaultPath, "The XML file of data to unmarshal")
	flag.Parse()
	in, err := decruft(inputFilename)
	ioutil.WriteFile("wat.xml", in, 0644)
	if err != nil {
		log.Fatal(err)
	}
	var r Root
	err = xml.Unmarshal(in, &r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", r)
}

type Root struct {
	Records []Record `xml:"Data>Record"`
}

type Record struct {
	Country string
	Year    int
	Value   float64
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
		filename,
	}
	return exec.Command("sed", args...).Output()
}
