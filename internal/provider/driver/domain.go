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
	"golang.org/x/sync/errgroup"
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

func (d *Driver) DiskAttachedToNodes(ctx context.Context, file string) ([]string, error) {
	domains, _, err := d.Libvirt.ConnectListAllDomains(1, libvirt.ConnectListDomainsActive|libvirt.ConnectListDomainsInactive)
	if err != nil {
		d.Logger.Warn("unable to list domains", zap.Error(err))
		return nil, err
	}
	g, _ := errgroup.WithContext(ctx)
	allDisks := map[string][]Disk{}
	for _, domain := range domains {
		domain := domain
		g.Go(func() error {
			xml, err := d.Libvirt.DomainGetXMLDesc(domain, 0)
			if err != nil {
				return err
			}
			disks, err := d.LookupDomainDisks(xml)
			if err != nil {
				return err
			}
			allDisks[hex.EncodeToString(domain.UUID[:])] = disks
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		d.Logger.Warn("unable to list domains disks", zap.Error(err))
		return nil, err
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

func (d *Driver) AttachDisk(domain libvirt.Domain, domainXml, disk, serial string) error {
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
		return d.Libvirt.DomainUpdateDeviceFlags(domain, newDomainXml, libvirt.DomainDeviceModifyConfig)
	}
	return nil
}

func (d *Driver) DettachDisk(domain libvirt.Domain, domainXml, serial string) error {
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
		return d.Libvirt.DomainUpdateDeviceFlags(domain, newDomainXml, libvirt.DomainDeviceModifyConfig)
	}
	return nil
}
