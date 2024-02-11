package main

import (
	"errors"
	"os"

	"github.com/amitbet/vncproxy/logger"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/remoteconsoles"
)

func getRemoteConsole(serverID string) (*remoteconsoles.RemoteConsole, error) {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: os.Getenv("OS_IDENTITYENDPOINT"),
		Username:         os.Getenv("OS_USERNAME"),
		Password:         os.Getenv("OS_PASSWORD"),
		TenantID:         os.Getenv("OS_TENANTID"),
		DomainID:         os.Getenv("OS_FOMAINID"),
	}

	provider, _ := openstack.AuthenticatedClient(opts)
	compute, _ := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: "RegionOne"})
	compute.Microversion = "2.6"

	createOpts := remoteconsoles.CreateOpts{
		Protocol: remoteconsoles.ConsoleProtocolVNC,
		Type:     remoteconsoles.ConsoleTypeNoVNC,
	}

	remoteConsole, err := remoteconsoles.Create(compute, serverID, createOpts).Extract()
	if err != nil {
		logger.Warn("server url not found:", err)
		return nil, errors.New("failed to ")
	}

	return remoteConsole, nil
}
