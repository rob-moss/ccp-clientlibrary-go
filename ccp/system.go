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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type LivenessHealth struct {
	CXVersion      *string `json:"CXVersion,omitempty"`
	TimeOnMgmtHost *string `json:"TimeOnMgmtHost,omitempty"`
}

type Health struct {
	TotalSystemHealth *string          `json:"TotalSystemHealth,omitempty"`
	CurrentNodes      *int64           `json:"CurrentNodes,omitempty"`
	ExpectedNodes     *int64           `json:"ExpectedNodes,omitempty"`
	NodesStatus       *[]NodeStatus    `json:"NodesStatus,omitempty"`
	PodStatusList     *[]PodStatusList `json:"PodStatusList,omitempty"`
}

type NodeStatus struct {
	NodeName           *string `json:"NodeName,omitempty"`
	NodeCondition      *string `json:"NodeCondition,omitempty"`
	NodeStatus         *string `json:"NodeStatus,omitempty"`
	LastTransitionTime *string `json:"LastTransitionTime,omitempty"`
}

type PodStatusList struct {
	PodName            *string `json:"PodName,omitempty"`
	PodCondition       *string `json:"PodCondition,omitempty"`
	PodStatus          *string `json:"PodStatus,omitempty"`
	LastTransitionTime *string `json:"LastTransitionTime,omitempty"`
}

// Login for v2
// func (s *Client) Login(client *Client) error {
//
// 	url := fmt.Sprintf(s.BaseURL + "/2/system/login?username=" + client.Username + "&password=" + client.Password)
//
// 	j, err := json.Marshal(client)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
// 	if err != nil {
// 		return err
// 	}
//
// 	_, err = s.doRequest(req)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }

// func (s *Client) GetLivenessHealth() (*LivenessHealth, error) {
//
// 	url := fmt.Sprintf(s.BaseURL + "/2/system/livenessHealth")
//
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	bytes, err := s.doRequest(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var data LivenessHealth
//
// 	err = json.Unmarshal(bytes, &data)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	health := &data
//
// 	return health, nil
// }

// LoginCreds for provider
type LoginCreds struct {
	Username *string `json:"username,omitempty" validate:"nonzero"`
	Password *string `json:"password,omitempty" validate:"nonzero"`
}

// How the heck does this work?
// func (s *Client) Login(client *Client) error {
//       ^ attach method: can take on functions
//                        ^ input arg
//                                        ^ return

// Login updated for v3
func (s *Client) Login(client *Client) error {

	url := s.BaseURL + "/v3/system/login"

	loginCreds := LoginCreds{
		Username: String(client.Username),
		Password: String(client.Password),
	}

	//var client *http.Client

	var PTransport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cl := http.Client{
		Transport: PTransport,
	}

	// Marshal the JSON payload to then send
	j, err := json.Marshal(loginCreds)
	if err != nil {
		return err
	}

	// print the JSON query
	//	fmt.Println(string(j))
	// Send the JSON payload
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	// this won't work for v3 - we need to capture some headers
	// _, err = s.doRequest(req)

	// req.Header.Add("Content-Type", "application/json")
	// // req.SetBasicAuth(s.Username, s.Password)
	// tr := &http.Transport{
	// 	// below needed to use Proxy from environment
	// 	Proxy: http.ProxyFromEnvironment,
	// 	// TLS checking disabled as most CCP instances use self-signed certs
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// s.Client{Transport: tr, Jar: jar}
	// // tr.Proxy{}
	// http = &http.Client{Transport: tr, Jar: jar}

	resp, err := cl.Do(req)
	// fmt.Printf("xauth: " + resp)
	fmt.Println("Response: \n")
	fmt.Println(resp)

	fmt.Println("Geting X-Auth-Token")

	// get the X-Auth-Token
	// X-Auth-Token

	var xauthtoken = resp.Header.Get("X-Auth-Token")
	fmt.Println("X-Auth-Token = " + xauthtoken)

	// set xauth
	s.XAuthToken = xauthtoken

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bodystring := string(body)
	fmt.Println("Body: \n")
	fmt.Println(bodystring)

	// fmt.Printf("xauth: " + xauthtoken + " csrf: " + csrftoken + "\n")
	if err != nil {
		return err
	}

	return nil
}

// func (s *Client) doRequest(req *http.Request) ([]byte, error) {
//
// 	var client *http.Client
//
// 	req.Header.Add("Content-Type", "application/json")
// 	//req.SetBasicAuth(s.Username, s.Password)
// 	tr := &http.Transport{
// 		// below needed to use Proxy from environment
// 		Proxy: http.ProxyFromEnvironment,
// 		// TLS checking disabled as most CCP instances use self-signed certs
// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
// 	}
// 	client = &http.Client{Transport: tr, Jar: jar}
//
// 	resp, err := client.Do(req)
//
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if 200 != resp.StatusCode && 201 != resp.StatusCode && 202 != resp.StatusCode && 204 != resp.StatusCode {
// 		return nil, fmt.Errorf("%s", body)
// 	}
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return body, nil
// }

// GetLivenessHealth foobar
func (s *Client) GetLivenessHealth() (*LivenessHealth, error) {

	url := fmt.Sprintf(s.BaseURL + "/2/system/livenessHealth")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}
	var data LivenessHealth

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	health := &data

	return health, nil
}

func (s *Client) GetHealth() (*Health, error) {

	url := fmt.Sprintf(s.BaseURL + "/2/system/health")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}
	var data Health

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	health := &data

	return health, nil
}
