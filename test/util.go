package test

import (
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"

	_ "embed"
)

//go:embed xml/domain.xml
var testDomainXml string

func CreateTestDomain(t *testing.T, libvirt *libvirt.Libvirt) libvirt.Domain {
	domain, err := libvirt.DomainCreateXML(testDomainXml, 0)
	assert.NoError(t, err)
	return domain
}
