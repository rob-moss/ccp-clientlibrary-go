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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ACIProfile struct
type ACIProfile struct {
	UUID                     *string   `json:"id,omitempty"`
	Name                     *string   `json:"name,omitempty"`
	APICHosts                *string   `json:"apic_hosts,omitempty"`
	APICUsername             *string   `json:"apic_username,omitempty"`
	APICPassword             *string   `json:"apic_password,omitempty"`
	ACIVMMDomainName         *string   `json:"aci_vmm_domain_name,omitempty" `
	ACIInfraVLANID           *int      `json:"aci_infra_vlan_id,omitempty" `
	VRFName                  *string   `json:"vrf_name,omitempty"`
	L3OutsidePolicyName      *string   `json:"l3_outside_policy_name,omitempty" `
	L3OutsideNetworkName     *string   `json:"l3_outside_network_name,omitempty" `
	AAEPName                 *string   `json:"aaep_name,omitempty" `
	Nameservers              *[]string `json:"nameservers,omitempty" `
	ControlPlaneContractName *string   `json:"control_plane_contract_name,omitempty" `
	NodeVLANStart            *int      `json:"node_vlan_start,omitempty"`
	NodeVLANEnd              *int      `json:"node_vlan_end,omitempty" `
	PodSubnetStart           *string   `json:"pod_subnet_start,omitempty" `
	ServiceSubnetStart       *string   `json:"service_subnet_start,omitempty" `
	MulticastRange           *string   `json:"multicast_range,omitempty" `
	ACITenant                *string   `json:"aci_tenant,omitempty" `
}

// GetACIProfiles gets
func (s *Client) GetACIProfiles() ([]ACIProfile, error) {

	url := s.BaseURL + "/v3/aci-profiles"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	fmt.Print(string(bytes))
	// Create an Array of ACI Profiles
	var data []ACIProfile

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetACIProfileByName gets
func (s *Client) GetACIProfileByName(profileName string) (*ACIProfile, error) {

	aciProfiles, err := s.GetACIProfiles()
	if err != nil {
		return nil, err
	}

	for _, x := range aciProfiles {

		if string(profileName) == string(*x.Name) {
			return &x, nil
		}
	}
	return nil, errors.New("Cannot find ACI Profile " + profileName)
}

// AddACIProfile adds
func (s *Client) AddACIProfile(aciProfile *ACIProfile) (*ACIProfile, error) {

	url := s.BaseURL + "/v3/aci-profiles/"

	j, err := json.Marshal(&aciProfile)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))

	if err != nil {
		return nil, err
	}

	bytes, err := s.doRequest(req)

	if err != nil {
		return nil, err
	}

	var data ACIProfile

	err = json.Unmarshal(bytes, &data)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// DeleteACIProfile delete a profile
func (s *Client) DeleteACIProfile(profileUUID string) error {

	if profileUUID == "" {
		return errors.New("Cluster UUID to delete is required")
	}

	url := s.BaseURL + "/v3/aci-profiles/" + profileUUID + "/"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = s.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// PatchACIProfile patch an ACI profile
func (s *Client) PatchACIProfile(profile *ACIProfile, profileUUID string) (*ACIProfile, error) {

	var data ACIProfile

	url := fmt.Sprintf(s.BaseURL + "/v3/aci-profiles/" + profileUUID + "/")

	j, err := json.Marshal(profile)

	if err != nil {

		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}

	bytes, err := s.doRequest(req)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &data)

	if err != nil {
		return nil, err
	}

	profile = &data

	return profile, nil
}
