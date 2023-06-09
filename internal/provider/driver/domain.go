package driver

import (
	"context"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/digitalocean/go-libvirt"
	"go.uber.org/zap"
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

func (d *Driver) DiskAttachedToNodes(ctx context.Context, lv *libvirt.Libvirt, file string) ([]string, error) {
	domains, _, err := lv.ConnectListAllDomains(1, libvirt.ConnectListDomainsActive|libvirt.ConnectListDomainsInactive)
	if err != nil {
		d.Logger.Warn("unable to list domains", zap.Error(err))
		return nil, err
	}
	allDisks := map[string][]Disk{}
	for _, domain := range domains {
		xml, err := lv.DomainGetXMLDesc(domain, 0)
		if err != nil {
			d.Logger.Warn("unable to get domain xml", zap.Error(err))
			return nil, err
		}
		disks, err := d.LookupDomainDisks(xml)
		if err != nil {
			d.Logger.Warn("unable to get domain disks", zap.Error(err))
			return nil, err
		}
		allDisks[hex.EncodeToString(domain.UUID[:])] = disks
	}
	nodeIds := []string{}
	for domainUUID, disks := range allDisks {
		for _, disk := range disks {
			if strings.EqualFold(disk.Source.File, file) {
				nodeIds = append(nodeIds, domainUUID)
			}
		}
	}
	return nodeIds, nil
}

func (d *Driver) AttachDisk(lv *libvirt.Libvirt, domainXml, disk, serial string) error {
	domainDoc := etree.NewDocument()
	if err := domainDoc.ReadFromString(domainXml); err != nil {
		return err
	}
	if domainDoc.FindElement(fmt.Sprintf("//domain/devices/disk[serial='%s']", serial)) == nil {
		diskDoc := etree.NewDocument()
		if err := diskDoc.ReadFromString(disk); err != nil {
			return err
		}
		devices := domainDoc.FindElement("//domain/devices")
		devices.AddChild(diskDoc.Root().Copy())
		newDomainXml, err := domainDoc.WriteToString()
		if err != nil {
			return err
		}
		domain, err := lv.DomainDefineXML(newDomainXml)
		if err != nil {
			return err
		}
		return lv.DomainAttachDevice(domain, disk)
	}
	return nil
}

func (d *Driver) DettachDisk(lv *libvirt.Libvirt, domainXml, serial string) error {
	domainDoc := etree.NewDocument()
	if err := domainDoc.ReadFromString(domainXml); err != nil {
		return err
	}
	disk := domainDoc.FindElement(fmt.Sprintf("//domain/devices/disk[serial='%s']", serial))
	if disk != nil {
		disk.Parent().RemoveChild(disk)
		newDomainXml, err := domainDoc.WriteToString()
		if err != nil {
			return err
		}
		domain, err := lv.DomainDefineXML(newDomainXml)
		if err != nil {
			return err
		}
		return lv.DomainDetachDeviceAlias(domain, disk.FindElement("alias").SelectAttrValue("name", ""), 0)
	}
	return nil
}
