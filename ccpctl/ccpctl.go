package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rob-moss/ccp-clientlibrary-go/ccp"
	// fork this github repo in to your ~/git/src dir
	// go get -u github.com/rob-moss/ccp-clientlibrary-go
)

// user.Current().HomeDir
var myself, _ = user.Current()                      // Get curent user and HOMEDIR
var defaultsFile = myself.HomeDir + "/.ccpctl.json" // also make this avaialble in the ENV var CCPCTLCONF

// debug levels:
//		0 for off
//		1 for functions and basic data
//		2 for warnings
//		3 for JSON output
var debuglvl = 1

// Debug function
func Debug(level int, errmsg string) {
	if level <= debuglvl {
		fmt.Println("Debug: " + errmsg)
	}
}

// Defaults file
type Defaults struct {
	CPName            string    `json:"cpnamedfl"`
	SSHUser           string    `json:"sshuser"`
	SSHKey            string    `json:"sshkey"`
	CPUser            string    `json:"cpuser"`
	CPPass            string    `json:"cppass"`
	CPURL             string    `json:"cpurl"`
	CPClusterDfl      string    `json:"cpclusterdfl"`      // default CCP cluster to work on
	CPToken           string    `json:"cptoken"`           // API token
	CPTokenTime       time.Time `json:"cptokentime"`       // API token expiry
	CPDatastoreDfl    string    `json:"cpdatastoredfl"`    // Default infra DS
	CPDatacenterDfl   string    `json:"cpdatacenterdfl"`   // Default infra DC
	CPImageDfl        string    `json:"cpimagedfl"`        // Default CCP image name
	CPNetworkDfl      string    `json:"cpnetworkdfl"`      // Default Network/Portgroup
	CPProviderDfl     string    `json:"cpproviderdfl"`     // Default Provider name
	CPProviderDflUUID string    `json:"cpproviderdflUUID"` // Default Provider UUID
	CPSubnetDfl       string    `json:"cpsubnetdfl`        // Default Subnet name
	CPSubnetDflUUID   string    `json:"cpsubnetdflUUID"`   // Default Subnet UUID
	CPVSClusterDfl    string    `json:"cpvsclusterdfl"`    // Default vSphere Cluster
}

// todo:
/*
- setcp: done
- setcp clusterdfl: done
- getcp: done
- getcluster: done
- getclusters: done
- getprovider: done
- getproviders: done
- getsubnet: done
- getsubnets: done
- addcluster:
- scalecluster:
- installaddon:
- deladdon:
- delcluster:


- stretch goals:
- base64 encode/decode password in JSON file
-- read in: decode
-- write out: encode

*/

// readDefaults reads the defaults JSON file in to a struct
func readDefaults() (*Defaults, error) {
	// empty
	var defaults Defaults

	jsonBody, err := ioutil.ReadFile(defaultsFile)
	if err != nil {
		fmt.Println(err)
		// do not bail if file not found
		// return nil, err
		return &defaults, nil
	}

	err = json.Unmarshal([]byte(jsonBody), &defaults)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	// fmt.Println("JSON Unmarshal Success")
	// fmt.Printf("Struct: %+v\n", defaults)

	return &defaults, nil
}

// writeDefaults write defaults out to a file
func writeDefaults(defaults *Defaults) error {

	jsonBody, err := json.Marshal(&defaults)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return err
	}
	// fmt.Println("JSON MarshalSuccess")
	// fmt.Printf("Struct: %+v\n", defaults)

	err = ioutil.WriteFile(defaultsFile, jsonBody, 0600)
	if err != nil {
		fmt.Println("Error writing defaults file to " + defaultsFile)
		fmt.Println(err)
		return err
	}

	return nil
}

func menuHelp() {
	fmt.Println(`
	ccpctl help
	-----------
	defaults
		setdefault 				// asks all the defaults
		setdefault cpName
		setdefault cpCluster
		setdefault cpSSHUser 	// Username ie ccpadmin
		setdefault ccpSSHKey 	// SSH key
		setdefault cpUser 		// CP username ie Admin
		setdefault cpPass 		// CP password ie C1sc0123

	add Control Plane info
		setcp <asks interactive>
		setcp cpname=cpname clusterdfl=clustername providerdfl=providername subnetdfl=subnetname datastoredfl=datastore datacenterdfl=dc
		setcp cpNetworkProviderUUID // looks up name, sets default
		setcp cpCloudProviderUUID // looks up name, sets default
		setcp cpNetworkVLAN // vSphere PortGroup
		setcp cpDatacenter	// vSphere DC
		setcp cpDatastore // vSphere DS
		delcp cpName // deletes Control Plane info
		getcp // lists CPs if no args
		getcps // lists CPs if no args
		getcp cpName // lists CP details

	cluster operations
		addcluster	<clustername> [provider=providername] [subnet=subnetname] [datastore=datastore] [datacenter=dc]
					uses defaults for provider, subnet, datastore, datacenter if not provided
		setcluster	<clustername> [provider=providername] [subnet=subnetname] [datastore=datastore] [datacenter=dc]
					uses defaults for provider, subnet, datastore, datacenter if not provided
		delcluster <clustername>
		getcluster <clustername> // pulls cluster info - master node IP(s), addons, # worker nodes
		getcluster <clustername> kubeconfig // gets and outputs kubeconfig
		getcluster <clustername> addons // lists Addons installed to cluster
		getcluster <clustername> masters // lists Master nodes installed to cluster
		getcluster <clustername> workers // lists Worker nodes installed to cluster
		scalecluster <clustername> workers=# [pool=poolname] // scale to this many worker nodes in a cluster

	cluster addons
		addclusteraddon <clustername> <addon> // install an addon
		delclusteraddon <clustername> <addon> // install an addon
		getclusteraddon <clustername> <addon> // install an addon

	kubectl config
		setkubeconf <clustername> [cpname] // sets ~/.kube/config to kubeconf

	control plane cluster install (API V2)
		installcp [subnet=subnetname] [datastore=datastore] [datacenter=dc] [iprange=1.2.3.4]
			[ipstart=1.2.3.4] [ipend=1.2.3.4] [cpuser=cpUser|use defaults] [cppass=cpPass|use defaults]
			all CP things
		`)
}

func splitparam(arg string) (param, value string) {
	if arg == "" {
		return "", ""
	}
	matched, _ := regexp.Match("=", []byte(arg))
	if !matched {
		return "", ""
	}
	s := strings.Split(arg, "=")
	first := s[0]
	rest := strings.Join(s[1:], "=")
	// return s[0], s[1]
	return first, rest

}

//modified from unexported nonzero function in the validtor package
//https://github.com/go-validator/validator/blob/v2/builtins.go
// nonzero
func nonzero(v interface{}) bool {
	st := reflect.ValueOf(v)
	nonZeroValue := false
	switch st.Kind() {
	case reflect.Ptr, reflect.Interface:
		nonZeroValue = st.IsNil()
	case reflect.Invalid:
		nonZeroValue = true // always invalid
	case reflect.Struct:
		nonZeroValue = false // always valid since only nil pointers are empty
	default:
		return true
	}

	if nonZeroValue {
		return true
	}
	return false
}

func menuHelpCP() {
	fmt.Println(`
	add Control Plane info
		setcp 
			sshuser=ccpadmin				// must have
			sshkey=sshkey					// must have
			cpuser=admin					// must have
			cppass=password					// must have
			cpurl=https://10.100.100.1  	// must have
			// below defaults are optional, can be specified on the commandline, or these defults will be used
			clusterdfl=clustername	
			providerdfl=providername 
			subnetdfl=subnetname 
			datastoredfl=datastore 
			datacenterdfl=dc 
			vsclusterdfl=vsphereclustername
			imagedfl=ccp-tenant-image-1.16.3-ubuntu18-6.1.1
	`)
}

// MenuSetCP sets Control Plane data
// func (s *Defaults) menusetCP(args []string) (*ControlPlane, error) {
func menuSetCP(args []string, Settings Defaults, client *ccp.Client) (*Defaults, error) {
	if len(args) < 1 {
		menuHelpCP()
		// fmt.Println("Not enough args")
		return nil, errors.New("Not enough args")
	}
	// fmt.Println(args)
	// var newCP ControlPlane // new ControlPlane struct - populate this and return it
	// get all args and loop over case statement to collect all vars
	for _, arg := range args {
		// fmt.Println("arg = " + arg)

		param, value := splitparam(arg)
		// fmt.Println("param: " + param + " value: " + value)

		//	cpname=cpname provider=providername subnet=subnetname datastore=datastore datacenter=dc
		switch param {
		case "cpname":
			fmt.Println("cpname updated with: " + value)
			Settings.CPName = value
		case "cpuser":
			fmt.Println("cpuser updated with: " + value)
			Settings.CPUser = value
		case "cppass":
			fmt.Println("cppass updated with: " + value)
			Settings.CPPass = value
		case "cpurl":
			fmt.Println("cpurl updated with: " + value)
			Settings.CPURL = value
		case "providerdfl":
			provider, err := client.GetInfraProviderByName(value)
			if err != nil {
				fmt.Println("* Error getting provider: ", err)
				return nil, err
			}
			Settings.CPProviderDfl = value              // set provider name
			Settings.CPProviderDflUUID = *provider.UUID // set provider UUID
			fmt.Println("* Setting providerdfl to ", value, " and UUID ", *provider.UUID)
		case "subnetdfl":
			subnet, err := client.GetNetworkProviderSubnetByName(value)
			if err != nil {
				fmt.Println("* Error getting subnet: ", err)
				return nil, err
			}
			Settings.CPSubnetDfl = value            // set Subnet name
			Settings.CPSubnetDflUUID = *subnet.UUID // set provider UUID
			fmt.Println("* Setting subnetdfl to ", value, " and UUID ", *subnet.UUID)
		case "datastoredfl":
			fmt.Println("datastoredfl updated with: " + value)
			Settings.CPDatastoreDfl = value
		case "datacenterdfl":
			fmt.Println("datacenterdfl updated with: " + value)
			Settings.CPDatacenterDfl = value
		case "vsclusterdfl": // vSphere cluster
			fmt.Println("vscluster (vSphere cluster) updated with: " + value)
			Settings.CPVSClusterDfl = value
		case "clusterdfl": // CCP cluster
			fmt.Println("clusterdfl (CCP Cluster) updated with: " + value)
			Settings.CPClusterDfl = value
		case "sshuser":
			fmt.Println("sshuser updated with: " + value)
			Settings.SSHUser = value
		case "sshkey":
			fmt.Println("sshkey updated with: " + value)
			Settings.SSHKey = value
		case "imagedfl":
			fmt.Println("image updated with: " + value)
			Settings.CPImageDfl = value
		case "networkdfl":
			fmt.Println("network updated with: " + value)
			Settings.CPNetworkDfl = value
		default:
			fmt.Println("Not understood: param=" + param + " value=" + value)
			menuHelpCP()
			return nil, errors.New("Not understood param " + param + " value=" + value)
		}
	}

	// check if no args - then do wizard or help

	// if args are good, populate CP struct
	if !nonzero(Settings.CPName) {
		menuHelpCP()
		return nil, errors.New("CPName is missing")
	}
	if !nonzero(Settings.CPURL) {
		menuHelpCP()
		return nil, errors.New("CPURL is missing")
	}
	if !nonzero(Settings.CPUser) {
		menuHelpCP()
		return nil, errors.New("CPUser is missing")
	}
	if !nonzero(Settings.CPPass) {
		menuHelpCP()
		return nil, errors.New("CPName is missing")
	}
	// write the settings
	err := writeDefaults(&Settings)
	if err != nil {
		fmt.Println("writeDefaults error:", err)
		return nil, err
	}

	return &Settings, nil

}

func menuGetCP(Settings *Defaults) error {
	var prettyJSON bytes.Buffer

	jsonBody, err := json.Marshal(&Settings)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return err
	}

	err = json.Indent(&prettyJSON, jsonBody, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		return err
	}
	fmt.Println(&prettyJSON)
	return nil
}

func prettyPrintJSONCluster(cluster *ccp.Cluster) {
	var prettyJSON bytes.Buffer

	jsonBody, err := json.Marshal(cluster)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		// return err
	}

	err = json.Indent(&prettyJSON, jsonBody, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		// return err
	}
	fmt.Println(&prettyJSON)
}

func prettyPrintJSONClusters(clusters *[]ccp.Cluster) {
	var prettyJSON bytes.Buffer

	jsonBody, err := json.Marshal(clusters)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		// return err
	}

	err = json.Indent(&prettyJSON, jsonBody, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		// return err
	}
	fmt.Println(&prettyJSON)
}

func prettyPrintJSONProvider(provider *ccp.ProviderClientConfig) {
	var prettyJSON bytes.Buffer

	jsonBody, err := json.Marshal(provider)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		// return err
	}

	err = json.Indent(&prettyJSON, jsonBody, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		// return err
	}
	fmt.Println(&prettyJSON)
}

func prettyPrintJSONSubnet(provider *ccp.NetworkProviderSubnet) {
	var prettyJSON bytes.Buffer

	jsonBody, err := json.Marshal(provider)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		// return err
	}

	err = json.Indent(&prettyJSON, jsonBody, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		// return err
	}
	fmt.Println(&prettyJSON)
}

func prettyPrintJSONClusterAddOns(clusteraddons *ccp.ClusterInstalledAddons) {
	var prettyJSON bytes.Buffer

	jsonBody, err := json.Marshal(clusteraddons)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return
	}

	err = json.Indent(&prettyJSON, jsonBody, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		// return err
	}
	fmt.Println(&prettyJSON)
}
func prettyPrintJSONClusterAddOnsCatalogue(clusteraddons *ccp.AddOnsCatalogue) {
	var prettyJSON bytes.Buffer

	jsonBody, err := json.Marshal(clusteraddons)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return
	}

	err = json.Indent(&prettyJSON, jsonBody, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		// return err
	}
	fmt.Println(&prettyJSON)
}

func menuGetClusters(client *ccp.Client, jsonout bool) {
	clusters, err := client.GetClusters()
	if err != nil {
		fmt.Println("GetClusters error:", err)
		// return err
	}

	if jsonout {
		prettyPrintJSONClusters(&clusters)
	} else {
		for _, cluster := range clusters {
			fmt.Println("Clustername: ", *cluster.Name, " Status: ", *cluster.Status, " Cluster type: ", *cluster.Type, " Cluster UUID: ", *cluster.UUID)
		}
	}

	// return nil
}

func getKubeVerFromImage(value string, a string, b string) string {
	// https://www.dotnetperls.com/between-before-after-go
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func strtoint(string string) int {
	int, err := strconv.Atoi(string)
	if err != nil {
		return 0
	}
	return int
}

func strtoint64(string string) int64 {
	int, err := strconv.Atoi(string)
	if err != nil {
		return int64(0)
	}
	return int64(int)
}

func menuClusterHelp() {
	fmt.Println(`
	addcluster
		clustername 			// Must have this
		[image=ccpimage]		// Set here or read default from setcp
		[loadbalancers]			// Default to 2
		[workers=1]			// Default to 1
		[masters=2]			// Default to 1. Must be either 1 or 3
		[datastore]			// Set here or read default from setcp
		[dc]				// Set here or read default from setcp
		[network]			// Set here or read default from setcp
		[vscluster]			// Set here or read default from setcp
		[provideruuid]			// Set here or read default from setcp
		[subnetuuid]			// Set here or read default from setcp
		[podcidr]			// Default 192.168.0.0/16

	getcluster
		clustername 			// Must have this
	`)
}

func menuAddCluster(client *ccp.Client, args []string, Settings *Defaults) (*ccp.Cluster, error) {

	// fmt.Println("Args: ", args)
	// Check if enough args
	if len(args) < 1 {
		menuClusterHelp()
		return nil, errors.New("Error: cluster name is blank, exiting")
	}
	// set vars
	var newclname, newclimage, newcldstore, newcldc, newclvscluster string
	var newclprovideruuid, newclsubnetuuid, newclpodcidr string
	var newclnet []string
	var newcllbipnum, newclworkers, newclmasters int64

	newclname = args[0] // first item is clustername

	// check for settings
	for _, arg := range args[1:] { // range over args starting from index 1 (2nd arg)
		param, value := splitparam(arg)
		switch param {
		case "name":
			newclname = value // checked
		case "image":
			newclimage = value // checked
		case "workers":
			newclworkers = strtoint64(value) // checked
		case "masters":
			newclmasters = strtoint64(value) // checked
		case "datastore":
			newcldstore = value // checked
		case "loadbalancers":
			newcllbipnum = strtoint64(value) // checked
		case "dc":
			newcldc = value // checked
		case "network":
			// newclnet[0] = value // checked
			newclnet = append(newclnet, value) // checked
		case "vscluster":
			newclvscluster = value // checked
		case "provideruuid":
			newclprovideruuid = value // checked
		case "subnetuuid":
			newclsubnetuuid = value // checked
		case "podcidr":
			newclpodcidr = value
		default:
			fmt.Println("Error, flag ", arg, " unknown")
		}
	}

	// if not set then return with an error
	if newclname == "" {
		return nil, errors.New("Error: cluster name is blank, exiting")
	}
	fmt.Println("* Clustername: ", newclname)

	if newclimage == "" { // if blank
		if Settings.CPImageDfl == "" {
			return nil, errors.New("Error: image is blank, and defaults blank. Either setcp or specify")
		}
		newclimage = Settings.CPImageDfl
	}
	fmt.Println("* Image: ", newclimage)

	// if not defined then set with a useful default
	if newclworkers < 1 {
		fmt.Println("* Workers: workers is blank, setting to 1")
		newclworkers = 1
	}
	if newclmasters < 1 {
		fmt.Println("* Masters: masters is blank, setting to 1")
		newclmasters = 1
	} else {
		switch newclmasters {
		case 1:
			// ok
		case 3:
			// ok
		default:
			// not ok.
			return nil, errors.New("Error: newclmasters not set to 1 or 3, exiting")
		}
	}
	if newcllbipnum < 1 {
		fmt.Println("* Loadbalancers: Load Balancers is blank, setting to 2")
		newcllbipnum = 2
	}
	if newclpodcidr == "" {
		fmt.Println("* PodCIDR: podcidr is blank, setting to 192.168.0.0/16")
		newclpodcidr = "192.168.0.0/16"
	}

	// if not defined then pull from Defaults
	if Settings.SSHUser == "" {
		return nil, errors.New("* SSHUser: SSH User is blank, Set it in setcp. Exiting")
	}
	if Settings.SSHKey == "" {
		return nil, errors.New("* SSHKey: SSH Key is blank, Set it in setcp.  Exiting")
	}

	// check if these are blank, try to pull defaults or exit
	if newcldstore == "" { // if blank
		if Settings.CPDatastoreDfl == "" {
			return nil, errors.New("* Datastore: datastore is blank, and defaults blank. Either setcp or specify")
		}
		newcldstore = Settings.CPDatastoreDfl
	}
	if newcldc == "" { // if blank
		if Settings.CPDatacenterDfl == "" {
			return nil, errors.New("* Datacenter: datacenter is blank, and defaults blank. Either setcp or specify")
		}
		newcldc = Settings.CPDatacenterDfl
	}
	if newclprovideruuid == "" { // if blank
		if Settings.CPProviderDflUUID == "" {
			return nil, errors.New("* ProviderUUID: infra provider UUID is blank, and defaults blank. Either setcp or specify")
		}
		newclprovideruuid = Settings.CPProviderDflUUID
	}
	if newclsubnetuuid == "" { // if blank
		if Settings.CPSubnetDflUUID == "" {
			return nil, errors.New("* SubnetUUID: subnet UUID is blank, and defaults blank. Either setcp or specify")
		}
		newclsubnetuuid = Settings.CPSubnetDflUUID
	}

	if len(newclnet) == 0 { // if blank
		if Settings.CPNetworkDfl == "" {
			return nil, errors.New("* Network: network is blank, and defaults blank. Either setcp or specify")
		}
		newclnet = append(newclnet, Settings.CPNetworkDfl)
	}
	// fmt.Println("-=-=- postclnet -=-=-")

	if newclvscluster == "" { // if blank
		if Settings.CPClusterDfl == "" {
			return nil, errors.New("* vSphere Cluster: vscluster is blank, and defaults blank. Either setcp or specify")
		}
		newclvscluster = Settings.CPVSClusterDfl
	}

	// all settings checked, should have everything ready

	kubernetesversion := getKubeVerFromImage(newclimage, "image-", "-ubuntu18")
	newCluster := &ccp.Cluster{
		Name: &newclname,
		Type: ccp.String("vsphere"), // always vSphere
		// WorkerNodePool: newWorkers,
		WorkerNodePool: &[]ccp.WorkerNodePool{
			// first worker node pool
			ccp.WorkerNodePool{
				Name:              ccp.String("node-pool"), // default name
				Size:              ccp.Int64(newclworkers), // default 1 if not defined
				VCPUs:             ccp.Int64(8),            // Workers always 8
				Memory:            ccp.Int64(32768),        // Workers always 32 G
				Template:          &newclimage,             // CCP Image / Template
				SSHUser:           &Settings.SSHUser,       // from Defaults
				SSHKey:            &Settings.SSHKey,        // from Defaults
				KubernetesVersion: &kubernetesversion,      // from Image
			},
		},
		MasterNodePool: &ccp.MasterNodePool{
			Name:              ccp.String("master-group"),
			Size:              ccp.Int64(newclmasters),
			VCPUs:             ccp.Int64(2),
			Memory:            ccp.Int64(16384),
			Template:          &newclimage,
			SSHUser:           &Settings.SSHUser,
			SSHKey:            &Settings.SSHKey,
			KubernetesVersion: &kubernetesversion,
		},
		Infra: &ccp.Infra{
			Datastore:  &newcldstore,
			Datacenter: &newcldc,
			Networks:   &newclnet, // array of []strings
			Cluster:    &newclvscluster,
		},
		KubernetesVersion:  &kubernetesversion,
		InfraProviderUUID:  &newclprovideruuid,
		SubnetUUID:         &newclsubnetuuid,
		LoadBalancerIPNum:  ccp.Int64(newcllbipnum),
		IPAllocationMethod: ccp.String("ccpnet"),
		AWSIamEnabled:      ccp.Bool(false),
		NetworkPlugin: &ccp.NetworkPlugin{
			Name: ccp.String("calico"),
			Details: &ccp.NetworkPluginDetails{
				PodCIDR: ccp.String(newclpodcidr), // default 192.168.0.0/16
			},
		},
	}

	// if jsonout {
	// 	prettyPrintJSONCluster(newCluster)
	// }

	fmt.Println("* Sending new cluster to be created: ", newCluster.Name)

	// Create cluster
	cluster, err := client.AddCluster(newCluster)
	if err != nil {
		return nil, err
	}
	// Return created cluster struct
	return cluster, nil
}

func menuGetCluster(client *ccp.Client, clusterName string, jsonout bool) error {
	cluster, err := client.GetClusterByName(clusterName)
	if err != nil {
		fmt.Println("GetCluster error:", err)
		return err
	}
	if jsonout {
		prettyPrintJSONCluster(cluster)
	} else {
		fmt.Println("Clustername: ", *cluster.Name, " Status: ", *cluster.Status, " Cluster type: ", *cluster.Type, " Cluster UUID: ", *cluster.UUID)
	}
	return nil
}

func menuGetClusterAddons(client *ccp.Client, clusterName string, jsonout bool) error {
	cluster, err := client.GetClusterByName(clusterName)
	if err != nil {
		fmt.Println("GetCluster error:", err)
		return err
	}

	// list available addons
	catalogue, err := client.GetAddonsCatalogue(*cluster.UUID)
	if err != nil {
		fmt.Println("GetCluster error:", err)
		return err
	}

	if jsonout {
		prettyPrintJSONClusterAddOnsCatalogue(catalogue)
	}
	fmt.Println("Addons available:")
	// this is ugly, have to revisit this
	fmt.Println("* kubeflow: ", catalogue.CcpKubeflow.Name, "Description:", catalogue.CcpKubeflow.Description)
	fmt.Println("* dashboard: ", catalogue.CcpKubernetesDashboard.Name, "Description:", catalogue.CcpKubernetesDashboard.Description)
	fmt.Println("* harbor: ", catalogue.CcpHarborOperator.Name, "Description:", catalogue.CcpHarborOperator.Description)
	fmt.Println("* logging: ", catalogue.CcpEfk.Name, "", "Description:", catalogue.CcpEfk.Description)
	fmt.Println("* monitoring: ", catalogue.CcpMonitor.Name, "Description:", catalogue.CcpMonitor.Description)
	fmt.Println("* istio: ", catalogue.CcpIstioOperator.Name, "Description:", catalogue.CcpIstioOperator.Description)
	fmt.Println("* hx-csi: ", catalogue.CcpHxcsi.Name, "Description:", catalogue.CcpHxcsi.Description)
	fmt.Println("")
	// list installed addons
	installedaddons, err := client.GetClusterInstalledAddons(*cluster.UUID)
	if err != nil {
		fmt.Println("GetCluster error:", err)
		return err
	}

	fmt.Println("Cluster: ", *cluster.Name, "Installed Add-Ons: ", installedaddons.Count)
	// if no addons installed, exit clean
	if installedaddons.Count < 1 {
		return nil
	}

	if jsonout {
		prettyPrintJSONClusterAddOns(installedaddons)
	}
	for _, addon := range installedaddons.Results {
		fmt.Println("Installed Addon: ", addon.Name, "Status:", addon.AddonStatus.Status, "Helm Status:", addon.AddonStatus.HelmStatus, "Description:", addon.Description)
	}
	return nil
}

func menuGetClusterAddonsCatalogue(client *ccp.Client, clusterName string, jsonout bool) error {
	cluster, err := client.GetClusterByName(clusterName)
	if err != nil {
		fmt.Println("GetCluster error:", err)
		return err
	}

	addons, err := client.GetAddonsCatalogue(*cluster.UUID)
	if err != nil {
		fmt.Println("GetCluster error:", err)
		return err
	}

	if jsonout {
		prettyPrintJSONClusterAddOnsCatalogue(addons)
	}
	fmt.Println("Addons available:")
	fmt.Printf("%v\n", addons)
	return nil
}

func menuDelCluster(client *ccp.Client, clusterName string) error {
	cluster, err := client.GetClusterByName(clusterName)
	if err != nil {
		fmt.Println("DeleteCluster error:", err)
		return err
	}

	err = client.DeleteCluster(*cluster.UUID)
	if err != nil {
		fmt.Println("DeleteCluster error:", err)
		return err
	}
	fmt.Println("Cluster ", clusterName, " deleted")
	return nil
}

func menuGetProviders(client *ccp.Client, jsonout bool) {
	infraproviders, err := client.GetInfraProviders()
	if err != nil {
		fmt.Println("GetInfraProviders error:", err)
	}
	// loop over all
	for _, infraprovider := range infraproviders {
		if jsonout {
			prettyPrintJSONProvider(&infraprovider)
		} else {
			fmt.Println("Infra Provider: ", *infraprovider.Name, " Provider type: ", *infraprovider.Type, " Provider UUID: ", *infraprovider.UUID)
		}
	}
}

func menuGetProvider(client *ccp.Client, providerName string, jsonout bool) {
	infraprovider, err := client.GetInfraProviderByName(providerName)
	if err != nil {
		fmt.Println("GetInfraProviderByName error:", err)
	}
	if jsonout {
		prettyPrintJSONProvider(infraprovider)
	} else {
		fmt.Println("Infra Provider: ", *infraprovider.Name, " Provider type: ", *infraprovider.Type, " Provider UUID: ", *infraprovider.UUID)
	}
}

func menuGetSubnets(client *ccp.Client, jsonout bool) {
	subnets, err := client.GetNetworkProviderSubnets()
	if err != nil {
		fmt.Println("GetInfraProviders error:", err)
	}
	// loop over all
	for _, subnet := range subnets {
		if jsonout {
			prettyPrintJSONSubnet(&subnet)
		} else {
			fmt.Println("Subnet Name: ", *subnet.Name, "Subnet CIDR: ", *subnet.CIDR, " Subnet UUID: ", *subnet.UUID)
		}
	}
}

func menuGetSubnet(client *ccp.Client, subnetName string, jsonout bool) {
	subnet, err := client.GetNetworkProviderSubnetByName(subnetName)
	if err != nil {
		fmt.Println("GetNetworkProviderSubnetByName error:", err)
	}
	if jsonout {
		prettyPrintJSONSubnet(subnet)
	} else {
		fmt.Println("Subnet Name: ", *subnet.Name, "Subnet CIDR: ", *subnet.CIDR, " Subnet UUID: ", *subnet.UUID)
	}
}

// ccpctl help

// add Control Plane info
//		setcp cpname=cpname provider=providername subnet=subnetname datastore=datastore datacenter=dc
//		setcp cpName=cpname [provider=providername] [subnet=subnetname] [datastore=datastore] [datacenter=dc]
//		setcp cpNetworkProviderUUID // looks up name, sets default
//		setcp cpCloudProviderUUID // looks up name, sets default
//		setcp cpNetworkVLAN // vSphere PortGroup
//		setcp cpDatacenter	// vSphere DC
//		setcp cpDatastore // vSphere DS
//		delcp cpName // deletes Control Plane info
//		getcp // lists CPs if no args
//		getcps // lists CPs if no args
//		getcp cpName // lists CP details
// cluster operations
//		addcluster	<clustername> [provider=providername] [subnet=subnetname] [datastore=datastore] [datacenter=dc]
//					uses defaults for provider, subnet, datastore, datacenter if not provided
//		setcluster	<clustername> [provider=providername] [subnet=subnetname] [datastore=datastore] [datacenter=dc]
//					uses defaults for provider, subnet, datastore, datacenter if not provided
//		delcluster <clustername>
//		getcluster <clustername> // pulls cluster info - master node IP(s), addons, # worker nodes
//		getcluster <clustername> kubeconfig // gets and outputs kubeconfig
//		getcluster <clustername> addons // lists Addons installed to cluster
//		getcluster <clustername> masters // lists Master nodes installed to cluster
//		getcluster <clustername> workers // lists Worker nodes installed to cluster
//		scalecluster <clustername> workers=# [pool=poolname] // scale to this many worker nodes in a cluster
// cluster addons
// 		addclusteraddon <clustername> <addon> // install an addon
// 		delclusteraddon <clustername> <addon> // install an addon
// 		getclusteraddon <clustername> <addon> // install an addon
//
// kubectl config
//		setkubeconf <clustername> [cpname] // sets ~/.kube/config to kubeconf
//
// control plane cluster install (API V2)
//		installcp [subnet=subnetname] [datastore=datastore] [datacenter=dc] [iprange=1.2.3.4]
//			[ipstart=1.2.3.4] [ipend=1.2.3.4] [cpuser=cpUser|use defaults] [cppass=cpPass|use defaults]
//			all CP things

// main
func main() {
	fmt.Println("* Entered main")

	// read defaults file
	Settings, err := readDefaults()
	if err != nil {
		fmt.Println("* Read Defults error:", err)
		return
	}

	// create the CCP Client side struct
	client := ccp.NewClient(Settings.CPUser, Settings.CPPass, Settings.CPURL)
	client.XAuthToken = Settings.CPToken

	// ---------------------------------------------
	// check if CPTokenTime is older than 30 mins
	t1 := Settings.CPTokenTime
	t2 := time.Now()
	timediff := t2.Sub(t1).Minutes()
	// fmt.Println(int64(timediff), "Minutes elapsed since login")
	if timediff >= float64(180) || Settings.CPToken == "" {
		fmt.Println("* Logging in again")
		err := client.Login(client)
		if err != nil {
			fmt.Println(err)
		}
		Settings.CPToken = client.XAuthToken // keep the xauthtoken
		Settings.CPTokenTime = time.Now()    // timestamp now
		err = writeDefaults(Settings)
		if err != nil {
			fmt.Println("* Write Defults error: ", err)
			return
		}
	}

	// Check if JSON output is requested
	jsonout := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "json":
			jsonout = true
			fmt.Println("* Setting JSON output")
		}
	}

	//
	// ---- Main code to check commandline params ---- //
	//

	for _, arg := range os.Args[1:] {
		switch arg {
		case "logincp":
			err := client.Login(client)
			if err != nil {
				fmt.Println(err)
			}
			Settings.CPToken = client.XAuthToken // keep the xauthtoken
			Settings.CPTokenTime = time.Now()    // timestamp now
			return
		case "setdefault":
			// fmt.Println("setdefault " + string(pos))
			fmt.Println("Not implemented yet")
			return
		case "setcp":
			fmt.Println("* setcp Set Control Plane data")
			Settings, err = menuSetCP(os.Args[2:], *Settings, client)
			if err != nil {
				fmt.Println("menuSetCP error:", err)
			}
			return
		case "delcp":
			fmt.Println("Not implemented yet")
			return
		case "getcp":
			menuGetCP(Settings)
			return
		// Clusters
		case "addcluster":
			if len(os.Args[1:]) < 2 {
				menuClusterHelp()
				return
			}
			newcluster, err := menuAddCluster(client, os.Args[2:], Settings)
			if err != nil {
				fmt.Println("addcluster error:", err)
				return
			}
			fmt.Println("* New cluster created:", newcluster.Name)
			if jsonout {
				prettyPrintJSONCluster(newcluster)
			}
			return
		case "setcluster":
			// would patch a cluster
			fmt.Println("Not implemented yet")
			return
		case "delcluster":
			menuDelCluster(client, os.Args[2])
			return
		case "getcluster":
			if len(os.Args) < 3 {
				menuClusterHelp()
				fmt.Println("* Error: missing clustername")
				return
			}
			menuGetCluster(client, os.Args[2], jsonout)
			return
		case "getclusters":
			menuGetClusters(client, jsonout)
			return
		case "scalecluster":
			fmt.Println("scalecluster [<clustername>] workers=#")
			return
		// Infra providers
		case "getproviders":
			menuGetProviders(client, jsonout)
			return
		case "getprovider":
			menuGetProvider(client, os.Args[2], jsonout)
			return
		// Subnets
		case "getsubnets":
			menuGetSubnets(client, jsonout)
			return
		case "getsubnet":
			menuGetSubnet(client, os.Args[2], jsonout)
			return
		// addons todo
		case "getaddons":
			// if no cluster specified, get all available addons
			if len(os.Args[1:]) < 2 {
				// menuGetClusterAddonsCatalogue(client, os.Args[2], jsonout)
				fmt.Println("Need cluster name, exiting")
				return
			}
			// if cluster specified, get all addons installed to cluster
			// fmt.Println("getaddons [<clustername>]")
			menuGetClusterAddons(client, os.Args[2], jsonout)
			return
		case "installaddon":
			fmt.Println("installaddon [<clustername>] <addon>")
			return
		case "deladdon":
			fmt.Println("deladdon [<clustername>] <addon>")
			return
		// print help
		case "help":
			fmt.Println("help")
			menuHelp()
			return
		default:
			fmt.Printf("Unknown option:   %s.\n", arg)
		}
	}

	fmt.Println("* Exiting")

	return

	// --- AddCluster from JSON File
	// jsonFile := "./cluster.json"
	// // Convert a JSON file to a Cluster struct
	// newCluster, err := client.ConvertJSONToCluster(jsonFile)
	// if err != nil {
	// 	fmt.Println("error:", err)
	// } else {
	// 	fmt.Println("Success")
	// }

	// fmt.Println("* New cluster name to create: " + *newCluster.Name)
	// createdCluster, err := client.AddCluster(newCluster)
	// if err != nil {
	// 	fmt.Println("Error from AddCluster:")
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("* Cluster sent to API: " + *createdCluster.Name)

	// --- Delete Cluster
	// clustername := string("romoss-testcp01-tenant04")
	// cluster, err := client.GetClusterByName(clustername)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// err = client.DeleteCluster(*cluster.UUID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// --- create the Cluster struct, for later sending to the AddCluster function
	// // https://stackoverflow.com/questions/51916592/fill-a-struct-which-contains-slices
	// ccpsshuser := ccp.String("ccpadmin")
	// ccpsshkey := ccp.String("ecdsa-sha2-nistp521 AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAFXQk0bZlFiFV6FD5DT0HdVJ2TsL9wlciD3UkcFI+/kpIj2AfOqqoQjt0BYZKzNJ6z4a25nkIueQJFog04S0/+PkQGX/Hc2DVccatAOWMRCedwukdgfoURLHyEdgl9EeCmiyqnUe6XVxiqcX9dkqXuI1KsP/oRir8ZAui3nXvdyUm8TGA== ccpadmin@galaxy.cisco.com")
	// ccptemplateimg := ccp.String("hx1-ccp-tenant-image-1.16.3-ubuntu18-6.1.1-pre")
	// kubernetesversion := ccp.String("1.16.3")
	// newCluster := &ccp.Cluster{
	// 	Name: ccp.String("romoss-testcp01-tenant03"),
	// 	Type: ccp.String("vsphere"),
	// 	// WorkerNodePool: newWorkers,
	// 	WorkerNodePool: &[]ccp.WorkerNodePool{
	// 		// first worker node pool
	// 		ccp.WorkerNodePool{
	// 			Name:              ccp.String("node-pool"), // default name
	// 			Size:              ccp.Int64(1),
	// 			VCPUs:             ccp.Int64(8),
	// 			Memory:            ccp.Int64(32768),
	// 			Template:          ccptemplateimg,
	// 			SSHUser:           ccpsshuser,
	// 			SSHKey:            ccpsshkey,
	// 			KubernetesVersion: kubernetesversion,
	// 		},
	// 	},
	// 	MasterNodePool: &ccp.MasterNodePool{
	// 		Name:              ccp.String("master-group"),
	// 		Size:              ccp.Int64(1),
	// 		VCPUs:             ccp.Int64(2),
	// 		Memory:            ccp.Int64(16384),
	// 		Template:          ccptemplateimg,
	// 		SSHUser:           ccpsshuser,
	// 		SSHKey:            ccpsshkey,
	// 		KubernetesVersion: kubernetesversion,
	// 	},
	// 	Infra: &ccp.Infra{
	// 		Datastore:  ccp.String("GFFA-HX1-CCPInstallTest01"),
	// 		Datacenter: ccp.String("GFFA-DC"),
	// 		Networks:   &[]string{"DV_VLAN1060"},
	// 		Cluster:    ccp.String("GFFA-HX1-Cluster"),
	// 	},
	// 	KubernetesVersion:  kubernetesversion,
	// 	InfraProviderUUID:  infraProvider.UUID,
	// 	SubnetUUID:         networkProviderSubnet.UUID,
	// 	LoadBalancerIPNum:  ccp.Int64(2),
	// 	IPAllocationMethod: ccp.String("ccpnet"),
	// 	AWSIamEnabled:      ccp.Bool(false),
	// 	NetworkPlugin: &ccp.NetworkPlugin{
	// 		Name: ccp.String("calico"),
	// 		Details: &ccp.NetworkPluginDetails{
	// 			PodCIDR: ccp.String("192.168.0.0/16"),
	// 		},
	// 	},
	// }
	// fmt.Println(newCluster)

	// // now create the cluster
	// // fmt.Println("* New cluster name to create: " + *newCluster.Name)
	// createdCluster, err := client.AddCluster(newCluster)
	// if err != nil {
	// 	fmt.Println("Error from AddCluster:")
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("* Cluster sent to API: " + *createdCluster.Name)

	// // ---- GetInfraProviderByName
	// infraProvider, err := client.GetInfraProviderByName("vsphere")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // Print out the providerClientConfig details
	// fmt.Println("* Provider Config name: " + *infraProvider.Name + " hostname: " + *infraProvider.Address + " UUID: " + *infraProvider.UUID)

	// cluster, err := client.GetClusterByName("romoss-testcp01-tenant02")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// // GetAddons uses UUID
	// addons, err := client.GetAddonsCatalogue(*cluster.UUID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	//
	// j, err := json.Marshal(addons.CcpHxcsi)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("Raw JSON for HX-CSI:")
	// fmt.Println(string(j))

	// --- this works!
	// err = client.InstallAddonHXCSI(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "InstallAddonHXCSI Error:")
	// 	fmt.Println(err)
	// 	return
	// }

	// err = client.DeleteAddOnHXCSI(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "DeleteAddonHXCSI Error:")
	// 	fmt.Println(err)
	// 	return
	// }

	// err = client.InstallAddonKubeflow(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "InstallAddonKubeflow Error:")
	// 	fmt.Println(err)
	// 	return
	// }

	// err = client.InstallAddonIstioOp(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }
	// time.Sleep(2 * time.Seconds)

	// err = client.InstallAddonIstioInstance(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }
	// time.Sleep(2 * time.Seconds)

	// err = client.InstallAddonIstio(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }
	// // time.Sleep(2 * time.Seconds)
	// time.Sleep(2 * time.Second)

	// err = client.InstallAddonMonitoring(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }
	// time.Sleep(2 * time.Second)

	// err = client.InstallAddonLogging(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }
	// time.Sleep(2 * time.Second)

	// err = client.InstallAddonHarborOp(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }
	// time.Sleep(2 * time.Seconds)

	// err = client.InstallAddonHarborInstance(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }
	// time.Sleep(2 * time.Seconds)

	// err = client.InstallAddonHarbor(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }
	// time.Sleep(2 * time.Second)

	// ---- delete addons
	// err = client.DeleteAddOnLogging(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }

	// err = client.DeleteAddOnMonitor(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }

	// err = client.DeleteAddOnIstio(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }

	// err = client.DeleteAddOnHarbor(*cluster.UUID)
	// if err != nil {
	// 	ccp.Debug(1, "Error:")
	// 	fmt.Println(err)
	// 	return
	// }

	fmt.Printf("* Closed\n")

}
