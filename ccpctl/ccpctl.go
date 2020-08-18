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
	CPToken           string    `json:"cptoken"`
	CPTokenTime       time.Time `json:"cptokentime"`
	CPProviderDfl     string    `json:"cpproviderdfl"`
	CPProviderDflUUID string    `json:"cpproviderdflUUID"`
	CPSubnetDfl       string    `json:"cpsubnetdfl`
	CPSubnetDflUUID   string    `json:"cpsubnetdflUUID"`
	CPDatastoreDfl    string    `json:"cpdatastoredfl"`
	CPDatacenterDfl   string    `json:"cpdatacenterdfl"`
	CPClusterDfl      string    `json:"cpclusterdfl"`
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
	s := strings.Split(arg, "=")
	return s[0], s[1]

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
		setcp cpname=cpname clusterdfl=clustername providerdfl=providername subnetdfl=subnetname datastoredfl=datastore datacenterdfl=dc
	`)
}

// MenuSetCP sets Control Plane data
// func (s *Defaults) menusetCP(args []string) (*ControlPlane, error) {
func menuSetCP(args []string, Settings Defaults, client *ccp.Client) (*Defaults, error) {
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
			// fmt.Println("providerdfl updated with" + value)
			// look up name and also set UUID
			provider, err := client.GetInfraProviderByName(value)
			if err != nil {
				fmt.Println("* Error getting provider: ", err)
				return nil, err
			}
			Settings.CPProviderDfl = value              // set provider name
			Settings.CPProviderDflUUID = *provider.UUID // set provider UUID
			fmt.Println("* Setting providerdfl to ", value, " and UUID ", *provider.UUID)
		case "subnetdfl":
			// fmt.Println("subnetdfl " + value)
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
		case "clusterdfl":
			fmt.Println("clusterdfl updated with: " + value)
			Settings.CPClusterDfl = value
		case "sshuser":
			fmt.Println("sshuser updated with: " + value)
			Settings.SSHUser = value
		case "sshkey":
			fmt.Println("sshkey updated with: " + value)
			Settings.SSHKey = value
		case "default":
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
			fmt.Println("Clustername: ", *cluster.Name, " Cluster type: ", *cluster.Type, " Cluster UUID: ", *cluster.UUID)
		}
	}

	// return nil
}

func menuAddCluster(client *ccp.Client, args []string, Settings *Defaults) error {
	// check for defaults
	// range over args to cluster specifics:
	// name
	// subnet
	// provider
	// workers

	var newclname, newclimage, newclkubver, newcldstore, newclnet, newcldc, newclworkers string
	for _, arg := range os.Args[1:] {
		param, value := splitparam(arg)
		switch param {
		case "name":
			fmt.Println("cluster name: " + value)
			newclname = value
		case "image":
			newclimage = value
		case "kubver":
			newclkubver = value // can I get this from the image?
		case "workers":
			newclworkers = value
		case "datastore":
			newcldstore = value
		case "dc":
			newcldc = value
		case "network":
			newclnet = value
		}
	}

	if !nonzero(newclname) {
		fmt.Println("Error: cluster name is blank, exiting")
		return errors.New("Error: cluster name is blank, exiting")
	}
	if !nonzero(newclimage) {
		fmt.Println("Error: image is blank, exiting")
		return errors.New("Error: image is blank, exiting")
	}
	ccptemplateimg := ccp.String("hx1-ccp-tenant-image-1.16.3-ubuntu18-6.1.1-pre")
	kubernetesversion := ccp.String("1.16.3")
	newCluster := &ccp.Cluster{
		Name: ccp.String("romoss-testcp01-tenant03"),
		Type: ccp.String("vsphere"),
		// WorkerNodePool: newWorkers,
		WorkerNodePool: &[]ccp.WorkerNodePool{
			// first worker node pool
			ccp.WorkerNodePool{
				Name:              ccp.String("node-pool"), // default name
				Size:              ccp.Int64(1),
				VCPUs:             ccp.Int64(8),
				Memory:            ccp.Int64(32768),
				Template:          ccptemplateimg,
				SSHUser:           &Settings.SSHUser,
				SSHKey:            &Settings.SSHKey,
				KubernetesVersion: kubernetesversion,
			},
		},
		MasterNodePool: &ccp.MasterNodePool{
			Name:              ccp.String("master-group"),
			Size:              ccp.Int64(1),
			VCPUs:             ccp.Int64(2),
			Memory:            ccp.Int64(16384),
			Template:          ccptemplateimg,
			SSHUser:           &Settings.SSHUser,
			SSHKey:            &Settings.SSHKey,
			KubernetesVersion: kubernetesversion,
		},
		Infra: &ccp.Infra{
			Datastore:  ccp.String("GFFA-HX1-CCPInstallTest01"),
			Datacenter: ccp.String("GFFA-DC"),
			Networks:   &[]string{"DV_VLAN1060"},
			Cluster:    ccp.String("GFFA-HX1-Cluster"),
		},
		KubernetesVersion:  kubernetesversion,
		InfraProviderUUID:  infraProvider.UUID,
		SubnetUUID:         networkProviderSubnet.UUID,
		LoadBalancerIPNum:  ccp.Int64(2),
		IPAllocationMethod: ccp.String("ccpnet"),
		AWSIamEnabled:      ccp.Bool(false),
		NetworkPlugin: &ccp.NetworkPlugin{
			Name: ccp.String("calico"),
			Details: &ccp.NetworkPluginDetails{
				PodCIDR: ccp.String("192.168.0.0/16"),
			},
		},
	}
	fmt.Println(newCluster)

	// err = ccp.AddCluster(newCluster)
}

func menuGetCluster(client *ccp.Client, clusterName string, jsonout bool) error {
	cluster, err := client.GetClusterByName(clusterName)
	if err != nil {
		fmt.Println("DeleteCluster error:", err)
		return err
	}
	if jsonout {
		prettyPrintJSONCluster(cluster)
	} else {
		fmt.Println("Clustername: ", *cluster.Name, " Cluster type: ", *cluster.Type, " Cluster UUID: ", *cluster.UUID)
	}
	return nil
}

func menuDelCluster(client *ccp.Client, clusterUUID string) error {
	err := client.DeleteCluster(clusterUUID)
	if err != nil {
		fmt.Println("DeleteCluster error:", err)
		return err
	}
	fmt.Println("Cluster ", clusterUUID, " deleted")
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
	if timediff >= float64(180) {
		fmt.Println("* Login older than 180 minutes, logging in again")
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
	// fmt.Println("* JSON output set to ", jsonout)
	//
	// ---- Main code to check commandline params ---- //
	//

cli:
	for _, arg := range os.Args[1:] {
		switch arg {
		case "logincp":
			err := client.Login(client)
			if err != nil {
				fmt.Println(err)
			}
			Settings.CPToken = client.XAuthToken // keep the xauthtoken
			Settings.CPTokenTime = time.Now()    // timestamp now
		case "setdefault":
			// fmt.Println("setdefault " + string(pos))
			fmt.Println("Not implemented yet")
		case "setcp":
			// create new ControlPlane struct
			fmt.Println("* setcp Set Control Plane data")
			Settings, err = menuSetCP(os.Args[2:], *Settings, client)
			if err != nil {
				fmt.Println("menuSetCP error:", err)
				// exit without saving new json
				return
			}
			// drop through to default below writeDefaults(Settings)
			break cli
		case "delcp":
			fmt.Println("Not implemented yet")
			return
		case "getcp":
			menuGetCP(Settings)
			return
		case "addcluster":
			fmt.Println("addcluster <clustername>")
			err = menuAddCluster(client, os.Args[:1], Settings)
			if err != nil {
				fmt.Println("menuSetCP error:", err)
				return
			}

		case "setcluster":
			fmt.Println("Not implemented yet")
		case "delcluster":
			fmt.Println("delcluster [<clustername>]")
			// menuDelCluster(client, os.Args[2])
			return
		case "getcluster":
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
		case "installaddon":
			fmt.Println("installaddon [<clustername>] <addon>")
		case "deladdon":
			fmt.Println("deladdon [<clustername>] <addon>")
		case "getaddons":
			fmt.Println("getaddons [<clustername>]")
			// case "getclusterjson":
			// 	menuGetClusterJSON(client, os.Args[2], jsonout)
			// 	return
			// case "getclustersjson":
			// 	menuGetClustersJSON(client)
			// 	return
			// case "getprovidersjson":
			// 	menuGetProvidersJSON(client, jsonout)
			// 	return
		case "help":
			fmt.Println("help")
			menuHelp()
			return
		default:
			fmt.Printf("Unknown option:   %s.\n", arg)
		}
	}

	fmt.Println("* Exiting and saving settings")
	err = writeDefaults(Settings)
	if err != nil {
		fmt.Println("writeDefaults error:", err)
		return
	}

	return

	// read config file
	// load to struct
	// load default CP
	// check if CPTokenTime greater than X hours

	// ccpctl help
	// defaults
	//		setdefault // asks all the defaults
	//		setdefault cpName
	//		setdefault cpCluster
	//		setdefault cpSSHUser // Username ie ccpadmin
	//		setdefault ccpSSHKey // SSH key
	//		setdefault cpUser // CP username ie Admin
	//		setdefault cpPass // CP password ie C1sc0123
	// add Control Plane info
	//		setcp <asks interactive>
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

	//
	//
	//
	// old code //
	//
	//
	// client := ccp.NewClient(cpUser, cpPass, cpURL)

	// err = client.Login(client)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// --

	// clusters, err := client.GetClusters()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("* clusters = " + strconv.Itoa(len(clusters)))

	// ----
	// fmt.Println("* Get first cluster name")
	// fmt.Println(string(*clusters[0].Name))
	// clustername := string(*clusters[0].Name)

	// GetClusterByName gets all clusters, searches for matching cluster name, returns *Cluster struct
	// clustername := string("romoss-testcp01-tenant01")
	// cluster, err := client.GetClusterByName(clustername)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	//
	// if cluster == nil {
	// 	fmt.Println(err)
	// 	return
	// } else {
	// 	fmt.Printf("* Got cluster UUID %s\n", *cluster.UUID)
	// }

	// // GetClusterByUUID gets cluster by UUID, returns *Cluster struct
	// cluster, err = client.GetClusterByUUID(*cluster.UUID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	//
	// if cluster == nil {
	// 	fmt.Println(err)
	// 	return
	// } else {
	// 	fmt.Printf("* Got cluster %s by UUID\n", *cluster.Name)
	// }

	// ---- GetInfraProviders
	// providerClientConfigs, err := client.GetInfraProviders()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // Print out the providerClientConfig details
	// fmt.Println("* Provider Config name: " + *providerClientConfigs[0].Name + " hostname: " + *providerClientConfigs[0].Address + " UUID: " + *providerClientConfigs[0].UUID)

	// ---- GetInfraProviderByName
	// infraProvider, err := client.GetInfraProviderByName("vsphere")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // Print out the providerClientConfig details
	// fmt.Println("* Provider Config name: " + *infraProvider.Name + " hostname: " + *infraProvider.Address + " UUID: " + *infraProvider.UUID)

	// // Get network provider and Subnet
	// networkProviderSubnets, err := client.GetNetworkProviderSubnets()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // Print out the providerClientConfig details
	// fmt.Println("* Network Provider  name: " + *networkProviderSubnets[0].Name + " CIDR: " + *networkProviderSubnets[0].CIDR + " UUID: " + *networkProviderSubnets[0].UUID)

	// --- GetNetworkProviderSubnetByName
	// Get network provider by name and return single entry
	// networkProviderSubnet, err := client.GetNetworkProviderSubnetByName("default-network-subnet")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // Print out the providerClientConfig details
	// fmt.Println("* Network Provider  name: " + *networkProviderSubnet.Name + " CIDR: " + *networkProviderSubnet.CIDR + " UUID: " + *networkProviderSubnet.UUID)

	// --- scale a cluster
	// clusterUUID := string(*cluster.UUID)
	// clusterWorkerPoolName := string("node-group")
	// clusterSize := int64(3)
	// scaleCluster, err := client.ScaleCluster(clusterUUID, clusterWorkerPoolName, clusterSize)
	// if err != nil {
	// 	fmt.Println(err)
	// 	// Print out the Println of bytes
	// 	// to debug: uncomment below. Prints JSON payload
	// 	//fmt.Println(string(scaleCluster))
	// } else {
	// 	fmt.Println("Name: " + *scaleCluster.Name + " Size: " + string(*scaleCluster.Size))
	// }

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
