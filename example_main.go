package main

import (
	"fmt"

	// fork this github repo in to your ~/git/src dir
	// go get github.com/rob-moss/ccp-clientlibrary-go
	"github.com/rob-moss/ccp-clientlibrary-go/ccp"
)

var cpUser = "admin"               // user for CCP Control Plane
var cpPass = "password"            // Password for CCP Control Plane
var cpURL = "https://10.100.10.10" // URL of Control Plane

// var cpUser = os.GetEnv("CCPPUSER") // user for CCP Control Plane
// var cpPass = os.GetEnv("CCPPPASS") // Password for CCP Control Plane
// var cpURL = os.GetEnv("CCPURL")

// var jar, err = cookiejar.New(nil)

func main() {

	fmt.Println("* Entered main")

	client := ccp.NewClient(cpUser, cpPass, cpURL)

	err := client.Login(client)
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

	// if cluster == nil {
	// 	fmt.Println(err)
	// 	return
	// } else {
	// 	fmt.Printf("* Got cluster UUID %s\n", *cluster.UUID)
	// }

	// // GetClusterByUUID gets cluster by UUID, returns *Cluster struct
	// // clusterUUID = *cluster.UUID
	// // clustername = "foobar"
	// cluster, err = client.GetClusterByUUID(*cluster.UUID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// if cluster == nil {
	// 	fmt.Println(err)
	// 	return
	// } else {
	// 	fmt.Printf("* Got cluster %s by UUID\n", *cluster.Name)
	// }

	// providerClientConfigs, err := client.GetInfraProviders()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // Print out the providerClientConfig details
	// fmt.Println("* Provider Config name: " + *providerClientConfigs[0].Name + " hostname: " + *providerClientConfigs[0].Address + " UUID: " + *providerClientConfigs[0].UUID)

	infraProvider, err := client.GetInfraProviderByName("vsphere")
	if err != nil {
		fmt.Println(err)
		return
	}
	// Print out the providerClientConfig details
	fmt.Println("* Provider Config name: " + *infraProvider.Name + " hostname: " + *infraProvider.Address + " UUID: " + *infraProvider.UUID)

	// /// ---

	// // Get network provider and Subnet
	// networkProviderSubnets, err := client.GetNetworkProviderSubnets()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // Print out the providerClientConfig details
	// fmt.Println("* Network Provider  name: " + *networkProviderSubnets[0].Name + " CIDR: " + *networkProviderSubnets[0].CIDR + " UUID: " + *networkProviderSubnets[0].UUID)

	// // Get network provider by name and return single entry
	networkProviderSubnet, err := client.GetNetworkProviderSubnetByName("default-network-subnet")
	if err != nil {
		fmt.Println(err)
		return
	}
	// Print out the providerClientConfig details
	fmt.Println("* Network Provider  name: " + *networkProviderSubnet.Name + " CIDR: " + *networkProviderSubnet.CIDR + " UUID: " + *networkProviderSubnet.UUID)

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

	// -- work in progress
	// --- AddClusterBasic
	// Add a cluster
	// newCluster := ccp.Cluster{
	// 	Name:     ccp.String("romoss-testcp01-tenant02"),
	// 	Type:     ccp.String("vsphere"),
	// 	Template: string("hx1-ccp-tenant-image-1.16.3-ubuntu18-6.1.1-pre"),
	// 	SSHUser:  string("ccpadmin"),
	// 	SSHKey:   string("ecdsa-sha2-nistp521 AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAFXQk0bZlFiFV6FD5DT0HdVJ2TsL9wlciD3UkcFI+/kpIj2AfOqqoQjt0BYZKzNJ6z4a25nkIueQJFog04S0/+PkQGX/Hc2DVccatAOWMRCedwukdgfoURLHyEdgl9EeCmiyqnUe6XVxiqcX9dkqXuI1KsP/oRir8ZAui3nXvdyUm8TGA== ccpadmin@galaxy.cisco.com"),
	// }

	// -- work in progress
	// var newCluster ccp.Cluster
	// newCluster.Name = ccp.String("romoss-testcp01-tenant02")
	// newCluster.Type = ccp.String("vsphere")
	// newCluster.InfraProviderUUID = infraProvider.UUID
	// newCluster.SubnetUUID = networkProviderSubnet.UUID
	// newCluster.KubernetesVersion = ccp.String("1.16.3")
	// newCluster.MasterNodePool.Template = ccp.String("hx1-ccp-tenant-image-1.16.3-ubuntu18-6.1.1-pre")

	// -- work in progress
	// ccpsshuser := "ccpadmin"
	// ccpsshkey := "ecdsa-sha2-nistp521 AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAFXQk0bZlFiFV6FD5DT0HdVJ2TsL9wlciD3UkcFI+/kpIj2AfOqqoQjt0BYZKzNJ6z4a25nkIueQJFog04S0/+PkQGX/Hc2DVccatAOWMRCedwukdgfoURLHyEdgl9EeCmiyqnUe6XVxiqcX9dkqXuI1KsP/oRir8ZAui3nXvdyUm8TGA== ccpadmin@galaxy.cisco.com"
	// ccptemplateimg := string("hx1-ccp-tenant-image-1.16.3-ubuntu18-6.1.1-pre")
	// newCluster := ccp.Cluster{
	// 	Name: "romoss-testcp01-tenant02",
	// 	Type: "vsphere",
	// 	WorkerNodePool: WorkerNodePool{
	// 		Size:     int64(1),
	// 		Template: ccptemplateimg,
	// 		SSHUser:  ccpsshuser,
	// 		SSHKey:   ccpsshkey,
	// 	},
	// 	MasterNodePool: MasterNodePool{
	// 		Size:     int64(1),
	// 		Template: ccptemplateimg,
	// 		SSHUser:  ccpsshuser,
	// 		SSHKey:   ccpsshkey,
	// 	},
	// 	Infra: Infra{
	// 		Datastore:  "GFFA-HX1-CCPInstallTest01",
	// 		Datacentre: "GFFA-DC",
	// 		Networks:   "DV_VLAN1060",
	// 		Cluster:    "GFFA-HX1-Cluster",
	// 	},
	// }
	// fmt.Println(newCluster)
	fmt.Printf("* Closed\n")

}

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
