package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"reflect"
	"strings"
	"time"

	// fork this github repo in to your ~/git/src dir
	// go get -u github.com/rob-moss/ccp-clientlibrary-go
	"github.com/rob-moss/ccp-clientlibrary-go/ccp"
)

var cpUser = "admin"           // user for CCP Control Plane
var cpPass = "password"        // Password for CCP Control Plane
var cpURL = "https://10.1.1.1" // IP or Hostname of CCP Control Plane

var myself, _ = user.Current()                      // Get curent user and HOMEDIR
var defaultsFile = myself.HomeDir + "/.ccpctl.json" // also make this avaialble in the ENV var CCPCTLCONF

// var cpUser = os.GetEnv("CCPPUSER") // user for CCP Control Plane
// var cpPass = os.GetEnv("CCPPPASS") // Password for CCP Control Plane
// var cpURL = os.GetEnv("CCPURL")  // IP or Hostname of CCP Control Plane

// debug levels:
//		0 for off
//		1 for functions and basic data
//		2 for warnings
//		3 for JSON output
var debuglvl = 3

// Debug function
func Debug(level int, errmsg string) {
	if level <= debuglvl {
		fmt.Println("Debug: " + errmsg)
	}
}

// // Settings the default settings
// var Settings Defaults

// Defaults file
type Defaults struct {
	CPName    string         `json:"cpname"`
	CPCluster string         `json:"cpcluster"`
	SSHUser   string         `json:"sshuser"`
	SSHKey    string         `json:"sshkey"`
	CPData    []ControlPlane `json:"controlplanes"`
}

// Settings global settings from JSON file
var Settings *Defaults

// ControlPlane config file
type ControlPlane struct {
	CPName          string    `json:"cpname"`
	CPUser          string    `json:"cpuser"`
	CPPass          string    `json:"cppass"`
	CPURL           string    `json:"cpurl"`
	CPToken         string    `json:"cptoken"`
	CPTokenTime     time.Time `json:"cptokentime"`
	CPProviderDfl   string    `json:"cpproviderUUID"`
	CPSubnetDfl     string    `json:"cpsubnetUUID"`
	CPDatastoreDfl  string    `json:"cpdatastore`
	CPDatacenterDfl string    `json:"cpdatacenter"`
}

// readDefaults reads the defaults JSON file in to a struct
func readDefaults() (*Defaults, error) {
	jsonBody, err := ioutil.ReadFile(defaultsFile)
	if err != nil {
		fmt.Println(err)
		// do not bail if file not found
		// return nil, err
	}
	var defaults Defaults
	err = json.Unmarshal([]byte(jsonBody), &defaults)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("JSON Unmarshal Success")
	fmt.Printf("Struct: %+v\n", defaults)

	return &defaults, nil
}

// writeDefaults write defaults out to a file
func writeDefaults(defaults *Defaults) error {

	jsonBody, err := json.Marshal(&defaults)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return err
	}
	fmt.Println("JSON MarshalSuccess")
	fmt.Printf("Struct: %+v\n", defaults)

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
	setcp cpname=cpname provider=providername subnet=subnetname datastore=datastore datacenter=dc
	setcp cpName=cpname [provider=providername] [subnet=subnetname] [datastore=datastore] [datacenter=dc]
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

// menuSetCP sets Control Plane data
func (s *Settings) menuSetCP(args []string) (*ControlPlane, error) {
	fmt.Println(args)
	var newCP ControlPlane // new ControlPlane struct - populate this and return it
	// get all args and loop over case statement to collect all vars
	for _, arg := range args {
		fmt.Println("arg = " + arg)

		param, value := splitparam(arg)
		fmt.Println("param: " + param + " value: " + value)

		//	cpname=cpname provider=providername subnet=subnetname datastore=datastore datacenter=dc
		switch param {
		case "cpname":
			fmt.Println("cpname " + value)
			newCP.CPName = value
		case "cpuser":
			fmt.Println("cpuser " + value)
			newCP.CPUser = value
		case "cppass":
			fmt.Println("cppass " + value)
			newCP.CPPass = value
		case "provider":
			fmt.Println("provider " + value)
			newCP.CPProviderDfl = value
		case "subnet":
			fmt.Println("subnet " + value)
			newCP.CPSubnetDfl = value
		case "datastore":
			fmt.Println("datastore " + value)
			newCP.CPDatastoreDfl = value
		case "datacenter":
			fmt.Println("cpname " + value)
			newCP.CPDatastoreDfl = value
		case "default":
			fmt.Println("Not understood: param=" + param + " value=" + value)
		}
	}

	// check if no args - then do wizard or help

	// if args are good, populate CP struct
	if nonzero(newCP.CPName) {
		return nil, errors.New("CPName is missing")
	}
	return &newCP, nil
}

// main
func main() {
	fmt.Println("* Entered main")

	fmt.Println("Commandline args: ")
	fmt.Println(os.Args[1:])

	Settings, err := readDefaults()
	if err != nil {
		fmt.Println("Read Defults error:", err)
		return
	}
	fmt.Println("Read defaults Success")
	fmt.Println(Settings)

	for pos, arg := range os.Args[1:] {
		switch arg {
		case "setdefault":
			fmt.Println("setdefault " + string(pos))
		case "setcp":
			fmt.Println("setcp")
			newCP := menuSetCP(os.Args[2:])
			Settings += newCP
			// return
		case "delcp":
			fmt.Println("bar")
		case "getcp":
			fmt.Println("bar")
		case "addcluster":
			fmt.Println("bar")
		case "setcluster":
			fmt.Println("bar")
		case "delcluster":
			fmt.Println("bar")
		case "getcluster":
			fmt.Println("bar")
		case "scalecluster":
			fmt.Println("bar")
		case "installaddon":
			fmt.Println("bar")
		case "deladdon":
			fmt.Println("bar")
		case "help":
			fmt.Println("help")
			menuHelp()
			return
		default:
			fmt.Printf("%s.\n", arg)
		}
	}

	err = writeDefaults(Settings)
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

	client := ccp.NewClient(cpUser, cpPass, cpURL)

	err = client.Login(client)
	if err != nil {
		fmt.Println(err)
	}

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

	cluster, err := client.GetClusterByName("romoss-testcp01-tenant02")
	if err != nil {
		fmt.Println(err)
		return
	}
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

	err = client.InstallAddonKubeflow(*cluster.UUID)
	if err != nil {
		ccp.Debug(1, "InstallAddonKubeflow Error:")
		fmt.Println(err)
		return
	}

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

	err = client.InstallAddonIstio(*cluster.UUID)
	if err != nil {
		ccp.Debug(1, "Error:")
		fmt.Println(err)
		return
	}
	// time.Sleep(2 * time.Seconds)
	time.Sleep(2 * time.Second)

	err = client.InstallAddonMonitoring(*cluster.UUID)
	if err != nil {
		ccp.Debug(1, "Error:")
		fmt.Println(err)
		return
	}
	time.Sleep(2 * time.Second)

	err = client.InstallAddonLogging(*cluster.UUID)
	if err != nil {
		ccp.Debug(1, "Error:")
		fmt.Println(err)
		return
	}
	time.Sleep(2 * time.Second)

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

	err = client.InstallAddonHarbor(*cluster.UUID)
	if err != nil {
		ccp.Debug(1, "Error:")
		fmt.Println(err)
		return
	}
	time.Sleep(2 * time.Second)

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
