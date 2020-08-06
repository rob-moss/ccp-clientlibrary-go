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
	"io/ioutil"
	"net/http"
	"strconv"

	validator "gopkg.in/validator.v2"
)

/* toDo
- Create JSON config
- Make connection to CCP CP via Proxy (optional)
- Set defaults: image, sshkey, sshuser, provider, network
- Log in to CCP using X-Auth-Token
- Create functions to:
-- Get kubernetes version for deployments
-- Fetch provider by name -> uuid
-- Fetch subnet by name -> uuid
-- Create Cluster (Calico, vSphere)
-- Scale Cluster (Worker nodes)
-- Delete Cluster

v2 todo
- Create functions to:
-- Install Add-Ons
--- Istio
--- Harbor
--- HX-CSI
--- Monitoring
--- Logging
*/

// Cluster v3 cluster
type Cluster struct {
	//  Cluster Variable Name in Struct
	//								Go Type			Reference in JSON
	UUID                 *string               `json:"id,omitempty"` //
	Type                 *string               `json:"type,omitempty"  `
	Name                 *string               `json:"name,omitempty"  validate:"nonzero"`
	InfraProviderUUID    *string               `json:"provider,omitempty" `
	Status               *string               `json:"status,omitempty" `
	KubernetesVersion    *string               `json:"kubernetes_version,omitempty" validate:"nonzero"`
	KubeConfig           *string               `json:"kubeconfig,omitempty"`
	IPAllocationMethod   *string               `json:"ip_allocation_method,omitempty" validate:"nonzero"`
	MasterVIP            *string               `json:"master_vip,omitempty"`
	LoadBalancerIPNum    *int64                `json:"load_balancer_num,omitempty"  validate:"nonzero"`
	SubnetUUID           *string               `json:"subnet_id,omitempty"`
	NTPPools             *[]string             `json:"ntp_pools,omitempty"`
	NTPServers           *[]string             `json:"ntp_servers,omitempty"`
	RegistriesRootCA     *[]string             `json:"root_ca_registries,omitempty"`
	RegistriesSelfSigned *RegistriesSelfSigned `json:"self_signed_registries,omitempty"`
	RegistriesInsecure   *[]string             `json:"insecure_registries,omitempty"`
	DockerProxyHTTP      *string               `json:"docker_http_proxy,omitempty"`
	DockerProxyHTTPS     *string               `json:"docker_https_proxy,omitempty"`
	DockerBIP            *string               `json:"docker_bip,omitempty"`
	Infra                *Infra                `json:"vsphere_infra,omitempty"  validate:"nonzero"`
	MasterNodePool       *MasterNodePool       `json:"master_group,omitempty"  validate:"nonzero" `
	WorkerNodePool       *[]WorkerNodePool     `json:"node_groups,omitempty"  validate:"nonzero" `
	NetworkPlugin        *NetworkPlugin        `json:"network_plugin_profile,omitempty" validate:"nonzero"`
	IngressAsLB          *bool                 `json:"ingress_as_lb,omitempty"`
	NginxIngressClass    *string               `json:"nginx_ingress_class,omitempty"`
	ETCDEncrypted        *bool                 `json:"etcd_encrypted,omitempty"`
	SkipManagement       *bool                 `json:"skip_management,omitempty"`
	DockerNoProxy        *[]string             `json:"docker_no_proxy,omitempty"`
	RoutableCIDR         *string               `json:"routable_cidr,omitempty"`
	ImagePrefix          *string               `json:"image_prefix,omitempty"`
	ACIProfileUUID       *string               `json:"aci_profile,omitempty"`
	Description          *string               `json:"description,omitempty"`
	AWSIamEnabled        *bool                 `json:"aws_iam_enabled,omitempty"`
}

// WorkerNodePool are the worker nodes - updated for v3
type WorkerNodePool struct {
	Name              *string   `json:"name,omitempty" validate:"nonzero"`      //v3
	Size              *int64    `json:"size,omitempty" validate:"nonzero"`      //v3
	Template          *string   `json:"template,omitempty" validate:"nonzero"`  //v2
	VCPUs             *int64    `json:"vcpus,omitempty" validate:"nonzero"`     //v2
	Memory            *int64    `json:"memory_mb,omitempty" validate:"nonzero"` //v2
	GPUs              *[]string `json:"gpus,omitempty"`                         //v3
	SSHUser           *string   `json:"ssh_user,omitempty"`                     //v3
	SSHKey            *string   `json:"ssh_key,omitempty"`                      //v3
	Node              *[]Node   `json:"nodes,omitempty"`                        //v3
	KubernetesVersion *string   `json:"kubernetes_version,omitempty"`           //v3
}

// RegistriesSelfSigned v3
type RegistriesSelfSigned struct {
	Cert *string `json:"selfsignedca,omitempty" `
}

// Infra updated for v3
type Infra struct { // checked for v3
	Datacenter   *string   `json:"datacenter,omitempty"  validate:"nonzero"`
	Datastore    *string   `json:"datastore,omitempty"  validate:"nonzero"`
	Cluster      *string   `json:"cluster,omitempty" validate:"nonzero"`
	Networks     *[]string `json:"networks,omitempty"  validate:"nonzero"`
	ResourcePool *string   `json:"resource_pool,omitempty"`
}

// MasterNodePool updated for v3
type MasterNodePool struct {
	Name              *string   `json:"name,omitempty"`                         // v2
	Size              *int64    `json:"size,omitempty"`                         // v2
	Template          *string   `json:"template,omitempty" validate:"nonzero"`  //v3
	VCPUs             *int64    `json:"vcpus,omitempty" validate:"nonzero"`     //v3
	Memory            *int64    `json:"memory_mb,omitempty" validate:"nonzero"` //v3
	GPUs              *[]string `json:"gpus,omitempty"`                         //v3
	SSHUser           *string   `json:"ssh_user,omitempty"`                     //v3
	SSHKey            *string   `json:"ssh_key,omitempty"`                      //v3
	Node              *[]Node   `json:"nodes,omitempty"`                        //v3
	KubernetesVersion *string   `json:"kubernetes_version,omitempty"`           //v3
}

// Node updated for v3
type Node struct {
	// v3 clusters
	Name         *string `json:"name,omitempty"`
	Status       *string `json:"status,omitempty"`
	StatusDetail *string `json:"status_detail,omitempty" validate:"nonzero"`
	StatusReason *string `json:"status_reason,omitempty" validate:"nonzero"`
	PublicIP     *string `json:"public_ip,omitempty"`
	PrivateIP    *string `json:"private_ip,omitempty"`
	Phase        *string `json:"phase,omitempty"`
	//	State        *string `json:"status,omitempty"`

}

// Label updated for v3
type Label struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// Deployer updated for v3
type Deployer struct {
	ProxyCMD     *string   `json:"proxy_cmd,omitempty"`
	ProviderType *string   `json:"provider_type,omitempty" validate:"nonzero"`
	Provider     *Provider `json:"provider,omitempty" validate:"nonzero"`
}

// NetworkPlugin now caters for PluginDetails
type NetworkPlugin struct {
	Name    *string               `json:"name,omitempty"`
	Details *NetworkPluginDetails `json:"details,omitempty"`
}

// NetworkPluginDetails updated for v3
type NetworkPluginDetails struct {
	PodCIDR *string `json:"pod_cidr,omitempty"`
}

// type HelmChart struct {
// 	HelmChartUUID *string `json:"helmchart_uuid,omitempty"`
// 	ClusterUUID   *string `json:"cluster_UUID,omitempty"`
// 	ChartURL      *string `json:"chart_url,omitempty"`
// 	Name          *string `json:"name,omitempty"`
// 	Options       *string `json:"options,omitempty"`
// }

// Provider vsphere provider for v2
type Provider struct {
	VsphereDataCenter         *string              `json:"vsphere_datacenter,omitempty"`
	VsphereDatastore          *string              `json:"vsphere_datastore,omitempty"`
	VsphereSCSIControllerType *string              `json:"vsphere_scsi_controller_type,omitempty"`
	VsphereWorkingDir         *string              `json:"vsphere_working_dir,omitempty"`
	VsphereClientConfigUUID   *string              `json:"vsphere_client_config_uuid,omitempty" validate:"nonzero"`
	ClientConfig              *VsphereClientConfig `json:"client_config,omitempty"`
}

// VsphereClientConfig for provider
type VsphereClientConfig struct {
	IP       *string `json:"ip,omitempty"`
	Port     *int64  `json:"port,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

// GetClusters function for v3
func (s *Client) GetClusters() ([]Cluster, error) {
	Debug(3, "GetClusters")

	url := s.BaseURL + "/v3/clusters"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Print out the Println of bytes
	// to debug: uncomment below. Prints JSON payload
	// fmt.Println(string(bytes))
	Debug(3, "Cluster JSON Payload:\n"+string(bytes))

	// Create an Array of Clusters
	var data []Cluster

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	// Print out list of Clusters and their index
	Debug(2, "Found "+strconv.Itoa(len(data))+" clusters")
	for i, cl := range data {
		Debug(2, "Found cluster "+strconv.Itoa(i)+" named "+*cl.Name+" with UUID "+*cl.UUID)
	}

	return data, nil
}

// GetClusterByName get all clusters, iterate through to find slice matching clusterName
func (s *Client) GetClusterByName(clusterName string) (*Cluster, error) {

	Debug(3, "GetClusterByName")
	clusters, err := s.GetClusters()
	if err != nil {
		return nil, err
	}

	for i, x := range clusters {
		Debug(3, "Iteration "+strconv.Itoa(i)+"Cluster found: "+string(*x.Name)+"\n")
		if string(clusterName) == string(*x.Name) {
			Debug(2, "Found matching cluster "+clusterName+" = "+*x.Name)
			return &x, nil
		}
	}
	return nil, errors.New("Cannot find cluster " + clusterName)
}

// // GetClusterByUUID alias for GetCluster
// func (s *Client) GetClusterByUUID(clusterUUID string) (*Cluster, error) {
// 	return GetCluster(clusterUUID)
// }

// GetClusterByUUID v3 cluster by UUID
func (s *Client) GetClusterByUUID(clusterUUID string) (*Cluster, error) {
	Debug(3, "GetClusterByUUID")

	url := fmt.Sprintf(s.BaseURL + "/v3/clusters/" + clusterUUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}
	var data *Cluster

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// PatchCluster spec for JSON scale
type PatchCluster struct {
	Name *string `json:"name" validate:"nonzero"`
	Size *int64  `json:"size" validate:"nonzero"`
}

// ScaleCluster scales an existing cluster
func (s *Client) ScaleCluster(clusterUUID, workerPoolName string, size int64) (*PatchCluster, error) {

	Debug(1, "Func: ScaleCluster")
	url := s.BaseURL + "/v3/clusters/" + clusterUUID + "/node-pools/" + workerPoolName + "/"
	Debug(2, "PATCH URL: "+url)

	cluserScale := PatchCluster{
		Name: String(workerPoolName),
		Size: Int64(size),
	}
	j, err := json.Marshal(cluserScale)
	if err != nil {
		return nil, err
	}

	Debug(3, "Sending JSON patch: "+string(j))

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}

	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	var data PatchCluster
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// --- below do not work, need fixin
//

// ConvertJSONToCluster convers JSON
func (s *Client) ConvertJSONToCluster(jsonFile string) (*Cluster, error) {
	Debug(1, "Entered ConvertJSONToCluster")

	// Debug(2, "Cluster Struct for cluster named "+string(*cluster.Name))
	jsonBody, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var newCluster Cluster
	err = json.Unmarshal([]byte(jsonBody), &newCluster)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("Success")
	}
	fmt.Printf("Struct: %+v\n", newCluster)

	return &newCluster, nil
}

// AddCluster creates a new cluster
func (s *Client) AddCluster(cluster *Cluster) (*Cluster, error) {
	Debug(1, "AddCluster for "+string(*cluster.Name))

	Debug(2, "Start validating Cluster struct")
	errs := validator.Validate(cluster)
	if errs != nil {
		Debug(1, "Errors validating Cluster struct with validator.Validate(): "+string(errs.Error()))
		return nil, errs
	} else {
		Debug(3, "No Errors validating Cluster struct")
	}

	url := s.BaseURL + "/v3/clusters/"

	j, err := json.Marshal(cluster)
	if err != nil {
		Debug(1, "Errors marshaling with json.Marshal(): "+string(err.Error()))
		return nil, err
	} else {
		Debug(3, "No errors Marshaling JSON")
	}

	Debug(2, "About to POST to url "+url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	if err != nil {
		Debug(1, "Errors POSTing with http.NewRequest: "+string(err.Error()))
		return nil, err
	}

	bytes, err := s.doRequest(req)
	if err != nil {
		Debug(1, "Errors POSTing with s.doRequest: "+string(err.Error()))
		Debug(1, "POST response: "+string(bytes))
		return nil, err
	}
	Debug(3, "POST response: "+string(bytes))

	var data Cluster

	// err = json.Unmarshal(bytes, &data)
	Debug(2, "Unmarshaling response")
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		Debug(1, "Errors unmarshaling with json.Unmarshal: "+string(err.Error()))
		return nil, err
	} else {
		Debug(2, "Unmarshaled response successfully")
	}

	Debug(2, "CCP API responded with JSON payload for cluster named "+*data.Name+" with UUID "+*data.UUID)
	if *data.UUID == "" {
		Debug(1, "CCP API created cluster named "+*data.Name+" with UUID "+*data.UUID)
	}
	return &data, nil
}

// DeleteCluster deletes a cluster
func (s *Client) DeleteCluster(clusterUUID string) error {
	Debug(2, "Entered DeleteCluster for UUID "+clusterUUID)

	if clusterUUID == "" {
		return errors.New("Cluster UUID to delete is required")
	}

	url := s.BaseURL + "/v3/clusters/" + clusterUUID + "/"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = s.doRequest(req)
	if err != nil {
		return err
	}

	Debug(2, "Request sent to API with no error response")
	return nil
}

// Not working?

// AddClusterBasic add a v3 cluster the easy way
func (s *Client) AddClusterBasic(cluster *Cluster) (*Cluster, error) {

	/*

		This function was added in order to provide users a better experience with adding clusters. The list of required
		fields has been shortend with all defaults and computed values such as UUIDs to be automatically configured on behalf of the user.

		The following fields and values will be configured. The remainder to be specified by the user

		ProviderClientConfigUUID
		KubernetesVersion - default will be set to 1.10.1
		Type - default will be set to 1
		Deployer
			ProviderType will be set to "vsphere"
			Provider
				VsphereDataCenter - already specified as part of Cluster struct so will use this same value
				VsphereClientConfigUUID
				VsphereDatastore - already specified as part of Cluster struct so will use this same value
				VsphereWorkingDir - default will be set to /VsphereDataCenter/vm
		NetworkPlugin
			Name - default will be set to contiv-vpp
			Status - default will be set to ""
			Details - default will be set to "{\"pod_cidr\":\"192.168.0.0/16\"}"
		WorkerNodePool
			VCPUs - default will be set to 2
			Memory - default will be set to 16384
		MasterNodePool
			VCPUs - default will be set to 2
			Memory - default will be set to 8192

	*/

	var data Cluster

	// The following will configured the defaults for the cluster as specified above as well as check that the minimum
	// fields are provided

	if nonzero(cluster.Name) {
		return nil, errors.New("Cluster.Name is missing")
	}
	if nonzero(cluster.Infra.Datacenter) {
		return nil, errors.New("Cluster.Infra.Datacenter is missing")
	}

	// if nonzero(cluster.ResourcePool) {
	// 	return nil, errors.New("Cluster.ResourcePool is missing")
	// }
	if nonzero(cluster.MasterNodePool.SSHUser) {
		return nil, errors.New("cluster.MasterNodePool.SSHUser is missing")
	}
	if nonzero(cluster.MasterNodePool.SSHKey) {
		return nil, errors.New("cluster.MasterNodePool.SSHKey is missing")
	}

	// loop over array of WorkerNodePool
	for k, v := range *cluster.WorkerNodePool {
		fmt.Printf("k=%s, v=%+v", k, v)

		if nonzero(v.SSHUser) {
			return nil, errors.New("v.SSHUser is missing")
		}

		if nonzero(v.SSHKey) {
			return nil, errors.New("v.SSHKey is missing")
		}
		if nonzero(v.Size) {
			return nil, errors.New("v.Size is missing")
		}
		if nonzero(v.Template) {
			return nil, errors.New("v.Template is missing")
		}
	}

	if nonzero(cluster.MasterNodePool.Size) {
		return nil, errors.New("cluster.MasterNodePool.Size is missing")
	}

	if nonzero(cluster.MasterNodePool.Template) {
		return nil, errors.New("cluster.MasterNodePool.Template is missing")
	}

	// check that cluster.MasterNodePool.Template and cluster.WorkerNodePool.Template are the same

	// Retrieve the provider client config UUID rather than have the user need to provide this themselves.
	// This is also built for a single provider client config and as of CCP 1.5 this wll be Vsphere
	providerClientConfigs, err := s.GetInfraProviderByName("vsphere")
	if err != nil {
		return nil, err
	}

	networkPlugin := NetworkPlugin{
		Name: String("calico"),
		// Details: String("{\"pod_cidr\":\"192.168.0.0/16\"}"),
		Details: &NetworkPluginDetails{
			PodCIDR: String("192.168.0.0/16"),
		},
	}

	// provider := Provider{
	// 	VsphereDataCenter:       String(*cluster.Infra.Datacenter),
	// 	VsphereDatastore:        String(*cluster.Infra.Datastore),
	// 	VsphereClientConfigUUID: String(*providerClientConfigs[0].UUID),
	// 	//	VsphereWorkingDir:       String("/" + *cluster.Infra.Datacenter + "/vm"),
	// }

	// deployer := Deployer{
	// 	ProviderType: String("vsphere"),
	// 	Provider:     &provider,
	// }

	workerNodePool := WorkerNodePool{
		VCPUs:    Int64(2),
		Memory:   Int64(16384),
		Template: String(*cluster.MasterNodePool.Template), // use same template as master
	}

	masterNodePool := MasterNodePool{
		VCPUs:    Int64(2),
		Memory:   Int64(16384),
		Template: String(*cluster.MasterNodePool.Template),
	}

	// Since it returns a list we will use the UUID from the first element
	cluster.InfraProviderUUID = String(*providerClientConfigs.UUID)
	cluster.KubernetesVersion = String("1.16.3") // todo: fetch this somehow
	// cluster.Type = Int64(1)
	cluster.NetworkPlugin = &networkPlugin
	// cluster.Deployer = &deployer

	//	cluster.WorkerNodePool = &workerNodePool
	cluster.WorkerNodePool = &[]WorkerNodePool{workerNodePool}

	cluster.MasterNodePool = &masterNodePool

	// Need to reset the cluster level template to nil otherwise we receive the following error
	// "Cluster level template cannot be provided when master_node_pool and worker_node_pool are provided"
	//	cluster.Template = nil

	url := fmt.Sprintf(s.BaseURL + "/v3/clusters")

	j, err := json.Marshal(cluster)

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

	err = json.Unmarshal(bytes, &data)

	if err != nil {
		return nil, err
	}

	cluster = &data

	return cluster, nil
}

// // PatchCluster does the things
// func (s *Client) PatchCluster(cluster *Cluster) (*Cluster, error) {

// 	var data Cluster

// 	if nonzero(cluster.UUID) {
// 		return nil, errors.New("Cluster.UUID is missing")
// 	}

// 	clusterUUID := *cluster.UUID

// 	url := fmt.Sprintf(s.BaseURL + "/v3/clusters/" + clusterUUID)

// 	j, err := json.Marshal(cluster)

// 	if err != nil {

// 		return nil, err
// 	}

// 	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(j))
// 	if err != nil {
// 		return nil, err
// 	}

// 	bytes, err := s.doRequest(req)

// 	if err != nil {
// 		return nil, err
// 	}

// 	err = json.Unmarshal(bytes, &data)

// 	if err != nil {
// 		return nil, err
// 	}

// 	cluster = &data

// 	return cluster, nil
// }
