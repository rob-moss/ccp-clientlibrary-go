/*Copyright (c) 2019 Cisco and/or its affiliates.

This software is licensed to you under the terms of the Cisco Sample
Code License, Version 1.0 (the "License"). You may obtain a copy of the
License at

               https://developer.cisco.com/docs/licenses

All use of the material herein must be in accordance with the terms of
the License. All rights not expressly granted by the License are
reserved. Unless required by applicable law or agreed to separately in
writing, software distributed under the License is distributed on an "AS
IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
or implied.*/

package ccp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ProviderClientConfig struct for vSphere. AWS, GKE, AKS not yet made
type ProviderClientConfig struct {
	UUID               *string `json:"id,omitempty"`
	Type               *string `json:"type,omitempty"`
	Name               *string `json:"name,omitempty" `
	Address            *string `json:"address,omitempty" `
	Port               *int64  `json:"port,omitempty" `
	Username           *string `json:"username,omitempty" `
	InsecureSkipVerify *bool   `json:"insecure_skip_verify,omitempty" `
	// Password may be needed to set up a new Provider
	// Config *Config `json:"config,omitempty"`
}

// Vsphere struct
type Vsphere struct {
	Datacenters *[]string `json:"Datacenters,omitempty"`
	Clusters    *[]string `json:"Clusters,omitempty"`
	VMs         *[]string `json:"VMs,omitempty"`
	Networks    *[]string `json:"Networks,omitempty"`
	Datastores  *[]string `json:"Datastores,omitempty"`
	Pools       *[]string `json:"Pools,omitempty"`
}

// NetworkProviderSubnet struct
type NetworkProviderSubnet struct {
	UUID        *string   `json:"uuid,omitempty"`
	IPVersion   *int64    `json:"ip_version,omitempty"`
	GatewayIP   *string   `json:"gateway_ip,omitempty"`
	CIDR        *string   `json:"cidr,omitempty"`
	Pools       *[]string `json:"pools,omitempty"`
	Network     *string   `json:"network,omitempty"`
	Nameservers *[]string `json:"nameservers,omitempty"`
	Name        *string   `json:"name,omitempty"`
	TotalIPs    *int64    `json:"total_ips,omitempty"`
	FreeIPs     *int64    `json:"free_ips,omitempty"`
}

// GetNetworkProviderSubnetByName Get and return named Network Provider
func (s *Client) GetNetworkProviderSubnetByName(networkProviderName string) (*NetworkProviderSubnet, error) {

	networkProviderSubnets, err := s.GetNetworkProviderSubnets()
	if err != nil {
		return nil, err
	}
	// iterate over array networkProviderSubnets
	// var _ is discarding the iterator int
	// var x is each singular networkProviderSubnets struct
	for _, x := range networkProviderSubnets {
		if networkProviderName == string(*x.Name) {
			fmt.Println("Found matching network provider " + *x.Name)
			return &x, nil
		}
	}

	return nil, errors.New("Network provider " + networkProviderName + " not found")
}

// GetNetworkProviderSubnets Get and return All Providers
func (s *Client) GetNetworkProviderSubnets() ([]NetworkProviderSubnet, error) {

	url := s.BaseURL + "/2/network_service/subnets/"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}
	var data []NetworkProviderSubnet

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ----- working

// GetInfraProviders Get and return All Infra Providers
func (s *Client) GetInfraProviders() ([]ProviderClientConfig, error) {

	url := fmt.Sprintf(s.BaseURL + "/v3/providers")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}
	var data []ProviderClientConfig

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetInfraProviderByUUID by UUID
func (s *Client) GetInfraProviderByUUID(providerUUID string) (*ProviderClientConfig, error) {

	url := fmt.Sprintf(s.BaseURL + "/v3/providers/" + providerUUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}
	var data *ProviderClientConfig

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetInfraProviderByName by Name
func (s *Client) GetInfraProviderByName(providerName string) (*ProviderClientConfig, error) {

	providers, err := s.GetInfraProviders()
	if err != nil {
		return nil, err
	}
	// iterate over array networkProviderSubnets
	// var _ is discarding the iterator int
	// var x is each singular networkProviderSubnets struct
	for _, x := range providers {
		if providerName == string(*x.Name) {
			fmt.Println("Found matching network provider " + *x.Name)
			return &x, nil
		}
	}

	return nil, errors.New("Network provider " + providerName + " not found")
}
