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
		TenantID:         os.Getenv("OS_PROJECTID"),  // projectID is tenantID
		DomainID:         os.Getenv("OS_DOMAINID"),
	}

	logger.Info("[+] authenticating")
	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		logger.Error("authentication failed:", err)
		return nil, err
	}

	logger.Info("[+] connecting to Nova")
	computeClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: "RegionOne"})

	if err != nil {
		logger.Warn("compute client creation failed:", err)
		return nil, errors.New("compute client is not initialized")
	}

	logger.Info("[+] configuring compute client")
	computeClient.Microversion = "2.6"
	createOpts := remoteconsoles.CreateOpts{
		Protocol: remoteconsoles.ConsoleProtocolVNC,
		Type:     remoteconsoles.ConsoleTypeNoVNC,
	}

	logger.Info("[+] creating remote console")
	logger.Info("[+] server ID:", serverID)
	logger.Info("[+] compute client:", computeClient)
	logger.Info("[+] options:", createOpts)


	remoteConsole, err := remoteconsoles.Create(computeClient, serverID, createOpts).Extract()
	if err != nil {
		logger.Warn("server url not found:", err)
		return nil, err
	}

	return remoteConsole, nil
}
