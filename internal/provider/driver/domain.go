package driver

import (
	"encoding/xml"
)

type Disk struct {
	XMLName xml.Name `xml:"disk"`
	Type    string   `xml:"type,attr"`
	Device  string   `xml:"device,attr"`
	Alias   struct {
		Name string `xml:"name,attr"`
	} `xml:"alias"`
	Driver struct {
		Name string `xml:"name,attr"`
		Type string `xml:"type,attr"`
	} `xml:"driver"`
	Source struct {
		File string `xml:"file,attr"`
	} `xml:"source"`
	Target struct {
		Dev string `xml:"dev,attr"`
		Bus string `xml:"bus,attr"`
	} `xml:"target"`
	Serial string `xml:"serial"`
}

type Domain struct {
	XMLName xml.Name `xml:"domain"`
	Name    string   `xml:"name"`
	UUID    string   `xml:"uuid"`
	Devices struct {
		Disks []Disk `xml:"disk"`
	} `xml:"devices"`
}

func (d *Driver) LookupDomainDisks(xmlDesc string) ([]Disk, error) {
	var domainXML Domain
	err := xml.Unmarshal([]byte(xmlDesc), &domainXML)
	if err != nil {
		return nil, err
	}
	return domainXML.Devices.Disks, nil
}
