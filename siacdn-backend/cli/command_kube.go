package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/NebulousLabs/Sia/api"
	"github.com/google/uuid"
	"github.com/thegreatdb/siacdn/siacdn-backend/kube"
	"github.com/thegreatdb/siacdn/siacdn-backend/models"
	urfavecli "github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
)

const kubeNamespace = "sia"

//var kubeStorageClass = "standard"
var kubeStorageClass = "fast"
var kubeDefaultStorage = resource.MustParse("30Gi")

var kubeMu sync.Mutex
var kubeInFlight map[uuid.UUID]bool = map[uuid.UUID]bool{}

func StartFlight(siaNode *models.SiaNode) {
	kubeMu.Lock()
	kubeInFlight[siaNode.ID] = true
	kubeMu.Unlock()
}

func StopFlight(siaNode *models.SiaNode) {
	kubeMu.Lock()
	delete(kubeInFlight, siaNode.ID)
	kubeMu.Unlock()
}

func KubeCommand() urfavecli.Command {
	return urfavecli.Command{
		Name:    "kube",
		Aliases: []string{"k"},
		Usage:   "Communicate with a local SiaCDN backend and coordinate changes with a kube server",
		Action:  kubeCommand,
	}
}

func kubeCommand(c *urfavecli.Context) error {
	clientset, _, err := kube.KubeClient()
	for {
		if err = PerformKubeRun(clientset); err != nil {
			return err
		}
		time.Sleep(time.Second / 2)
	}
	return nil
}

func PerformKubeRun(clientset *kubernetes.Clientset) error {
	siaNodes, err := getPendingSiaNodes()
	if err != nil {
		return err
	}
	for _, siaNode := range siaNodes {
		if err = pollKube(clientset, siaNode); err != nil {
			return err
		}
	}
	return nil
}

func pollKube(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	kubeMu.Lock()
	_, inFlight := kubeInFlight[siaNode.ID]
	kubeMu.Unlock()
	if inFlight {
		log.Println("Skipping " + siaNode.ID.String() + " because it is in-flight")
		return nil
	}
	// TODO: Also check length and skip if kubeInFlight has more than 10 going too
	switch siaNode.Status {
	case models.SIANODE_STATUS_CREATED:
		go pollKubeCreated(clientset, siaNode)
	case models.SIANODE_STATUS_DEPLOYED:
		go pollKubeDeployed(clientset, siaNode)
	case models.SIANODE_STATUS_INSTANCED:
		go pollKubeInstanced(clientset, siaNode)
	case models.SIANODE_STATUS_SNAPSHOTTED:
		go pollKubeSnapshotted(clientset, siaNode)
	case models.SIANODE_STATUS_SYNCHRONIZED:
		go pollKubeSynchronized(clientset, siaNode)
	case models.SIANODE_STATUS_INITIALIZED:
		go pollKubeInitialized(clientset, siaNode)
	case models.SIANODE_STATUS_UNLOCKED:
		go pollKubeUnlocked(clientset, siaNode)
	case models.SIANODE_STATUS_FUNDED:
		go pollKubeFunded(clientset, siaNode)
	case models.SIANODE_STATUS_CONFIGURED:
		go pollKubeConfigured(clientset, siaNode)
	default:
		log.Println("Unknown status: " + siaNode.Status)
	}
	return nil
}

func pollKubeCreated(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeCreated: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)

	volumeClaims := clientset.PersistentVolumeClaims(kubeNamespace)
	deployments := clientset.AppsV1beta1Client.Deployments(kubeNamespace)
	services := clientset.Services(kubeNamespace)

	// First check for volume claim
	claim, err := volumeClaims.Get(siaNode.KubeNameVol(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Println("Error getting volume claim from kubernetes: " + err.Error())
		return err
	}
	// If it doesn't exist, create it
	if claim == nil || errors.IsNotFound(err) {
		spec := v1.PersistentVolumeClaimSpec{
			StorageClassName: &kubeStorageClass,
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName("storage"): kubeDefaultStorage,
				},
			},
		}

		claim = &v1.PersistentVolumeClaim{}
		claim.Name = siaNode.KubeNameVol()
		claim.Namespace = kubeNamespace
		claim.Spec = spec

		log.Println("Creating volume claim " + siaNode.KubeNameVol())
		claim, err = volumeClaims.Create(claim)
		if err != nil {
			log.Println("Error creating volume claim: " + err.Error())
			return err
		}
	} else {
		log.Println("Found volume claim " + siaNode.KubeNameVol())
	}

	// Then check for service
	service, err := services.Get(siaNode.KubeNameSer(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Println("Error getting service from kubernetes: " + err.Error())
		return err
	}
	// If it doesn't exist, create it
	if service == nil || errors.IsNotFound(err) {
		service = &v1.Service{}
		service.Name = siaNode.KubeNameSer()
		service.Namespace = kubeNamespace
		service.Spec = v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				v1.ServicePort{Name: "p1", Port: 9980, TargetPort: intstr.FromInt(9980), Protocol: v1.ProtocolTCP},
				v1.ServicePort{Name: "p2", Port: 9980, TargetPort: intstr.FromInt(9980), Protocol: v1.ProtocolUDP},
				v1.ServicePort{Name: "p3", Port: 9981, TargetPort: intstr.FromInt(9981), Protocol: v1.ProtocolTCP},
				v1.ServicePort{Name: "p4", Port: 9981, TargetPort: intstr.FromInt(9981), Protocol: v1.ProtocolUDP},
				v1.ServicePort{Name: "p5", Port: 9982, TargetPort: intstr.FromInt(9982), Protocol: v1.ProtocolTCP},
				v1.ServicePort{Name: "p6", Port: 9982, TargetPort: intstr.FromInt(9982), Protocol: v1.ProtocolUDP},
			},
			Selector: map[string]string{"app": siaNode.KubeNameApp()},
		}
		log.Println("Creating service " + siaNode.KubeNameSer())
		service, err = services.Create(service)
		if err != nil {
			log.Println("Error creating service: " + err.Error())
			return err
		}
	} else {
		log.Println("Found service " + siaNode.KubeNameSer())
	}

	// Finally, check for deployment
	deployment, err := deployments.Get(siaNode.KubeNameDep(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Println("Error getting deployment from kubernetes: " + err.Error())
		return err
	}
	// If deployment doesn't exist, create it
	if deployment == nil || errors.IsNotFound(err) {
		deployment := &v1beta1.Deployment{}
		deployment.Name = siaNode.KubeNameDep()
		deployment.Namespace = kubeNamespace
		deployment.Spec = v1beta1.DeploymentSpec{Template: v1.PodTemplateSpec{}}
		deployment.Spec.Template.Labels = map[string]string{"app": siaNode.KubeNameApp()}
		deployment.Spec.Template.Spec = v1.PodSpec{
			Volumes: []v1.Volume{
				v1.Volume{
					Name: "sia-volume",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: siaNode.KubeNameVol(),
						},
					},
				},
			},
			Containers: []v1.Container{
				v1.Container{
					Name:            "sia-docker",
					Image:           "gcr.io/gradientzoo-1233/sia-docker:latest",
					ImagePullPolicy: v1.PullAlways,
					Ports: []v1.ContainerPort{
						v1.ContainerPort{ContainerPort: 9980},
						v1.ContainerPort{ContainerPort: 9981},
						v1.ContainerPort{ContainerPort: 9982},
					},
					VolumeMounts: []v1.VolumeMount{
						v1.VolumeMount{Name: "sia-volume", MountPath: "/sia"},
					},
					Env: []v1.EnvVar{
						v1.EnvVar{
							Name: "SIA_API_PASSWORD",
							ValueFrom: &v1.EnvVarSource{
								SecretKeyRef: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "sia-secret",
									},
									Key: "siaapipassword",
								},
							},
						},
					},
				},
			},
		}
		log.Println("Creating deployment " + siaNode.KubeNameDep())
		deployment, err = deployments.Create(deployment)
		if err != nil {
			log.Println("Error creating deployment: " + err.Error())
			return err
		}
	} else {
		log.Println("Found deployment " + siaNode.KubeNameDep())
	}

	_, err = updateSiaNodeStatus(siaNode.ID, models.SIANODE_STATUS_DEPLOYED)
	return err
}

func pollKubeDeployed(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeDeployed: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)
	pod, _ := getPod(clientset, siaNode)
	if pod == nil || pod.Name == "" {
		return nil
	}
	log.Println("Found pod " + siaNode.KubeNameDep())
	_, err := updateSiaNodeStatus(siaNode.ID, models.SIANODE_STATUS_INSTANCED)
	return err
}

func pollKubeInstanced(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeInstanced: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)

	client, err := siaNode.SiaClient()
	if err != nil {
		return err
	}

	var resp api.ConsensusGET
	if err = client.Get("/consensus", &resp); err != nil {
		log.Println("Got error checking for instance: " + err.Error())
		return nil
	}

	log.Println("Finished snapshotting Sia node " + siaNode.Shortcode)

	_, err = updateSiaNodeStatus(siaNode.ID, models.SIANODE_STATUS_SNAPSHOTTED)

	return err
}

func pollKubeSnapshotted(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeSnapshotted: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)

	client, err := siaNode.SiaClient()
	if err != nil {
		return err
	}

	var resp api.ConsensusGET
	if err = client.Get("/consensus", &resp); err != nil {
		log.Println("Got error checking for consensus: " + err.Error())
		return nil
	}

	if resp.Synced {
		log.Println("Finished syncing blockchain on Sia node " + siaNode.Shortcode)
		_, err = updateSiaNodeStatus(siaNode.ID, models.SIANODE_STATUS_SYNCHRONIZED)
		return err
	}

	return nil
}

func pollKubeSynchronized(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeSynchronized: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)

	client, err := siaNode.SiaClient()
	if err != nil {
		return err
	}

	var resp api.WalletInitPOST
	if err = client.Post("/wallet/init", "", &resp); err != nil {
		log.Println("Got error initializing wallet: " + err.Error())
		return nil
	}

	if resp.PrimarySeed == "" {
		log.Println("Could not initialize wallet")
		return fmt.Errorf("Could not initialize wallet")
	}

	_, err = createWalletSeed(siaNode.ID, resp.PrimarySeed)
	if err != nil {
		log.Println("Got error saving wallet seed: " + err.Error())
		return err
	}

	log.Println("Initialized wallet on Sia node " + siaNode.Shortcode)
	_, err = updateSiaNodeStatus(siaNode.ID, models.SIANODE_STATUS_INITIALIZED)
	return err
}

func pollKubeInitialized(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeInitialized: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)

	seed, err := getWalletSeed(siaNode.ID)
	if err != nil {
		return err
	}
	if seed == nil {
		return fmt.Errorf("Could get wallet seed for unlock")
	}

	client, err := siaNode.SiaClient()
	if err != nil {
		return err
	}

	log.Println("Unlocking wallet " + siaNode.Shortcode)
	if err != nil {
		log.Println("Got error encoding password: " + err.Error())
		return nil
	}
	vals := url.Values{}
	vals.Set("encryptionpassword", seed.Words)
	var resp struct {
		Message string `json:"message"`
	}
	if err = client.Post("/wallet/unlock", vals.Encode(), &resp); err != nil {
		log.Println("Got error unlocking wallet: " + err.Error())
		return nil
	}

	log.Println(fmt.Sprintf("Got response: %s", resp.Message))
	_, err = updateSiaNodeStatus(siaNode.ID, models.SIANODE_STATUS_UNLOCKED)
	return err
	/*


		//req := fmt.Sprintf("{\"encryptionpassword\":\"%s\"}", siaNode.)

		var resp struct {
			EncryptionPassword string `json:"encryptionpassword"`
		}
		if err = client.Post("/wallet/init", "{}", &resp); err != nil {
			log.Println("Got error initializing wallet: " + err.Error())
			return nil
		}

		if resp.PrimarySeed == "" {
			log.Println("Could not initialize wallet")
			return fmt.Errorf("Could not initialize wallet")
		}

		_, err = createWalletSeed(siaNode.ID, resp.PrimarySeed)
		if err != nil {
			log.Println("Got error saving wallet seed: " + err.Error())
			return err
		}

		log.Println("Initialized wallet on Sia node " + siaNode.Shortcode)
		_, err = updateSiaNodeStatus(siaNode.ID, models.SIANODE_STATUS_INITIALIZED)
		return err
	*/
}

func pollKubeUnlocked(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeInitialized: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)
	return nil
}

func pollKubeFunded(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeFunded: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)
	return nil
}

func pollKubeConfigured(clientset *kubernetes.Clientset, siaNode *models.SiaNode) error {
	log.Println("PollKubeConfigured: " + siaNode.Shortcode)
	StartFlight(siaNode)
	defer StopFlight(siaNode)
	return nil
}

// Utils

func getPod(clientset *kubernetes.Clientset, siaNode *models.SiaNode) (*v1.Pod, error) {
	opts := metav1.ListOptions{LabelSelector: "app=" + siaNode.KubeNameApp()}
	pods, err := clientset.Pods(kubeNamespace).List(opts)
	if err != nil && !errors.IsNotFound(err) {
		log.Println("Error getting pod from kubernetes: " + err.Error())
		return nil, err
	}

	if pods == nil || errors.IsNotFound(err) || pods.Items == nil || len(pods.Items) == 0 {
		log.Println("Not found yet: " + siaNode.KubeNameApp())
		return nil, err
	}

	return &pods.Items[0], nil
}

// Administrative API Calls

func getPendingSiaNodes() ([]*models.SiaNode, error) {
	url := fmt.Sprintf("%s/sianodes/pending/all?secret=%s", URLRoot, SiaCDNSecretKey)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("Could not create request GET " + url)
		return nil, err
	}

	res, err := cliClient.Do(req)
	if err != nil {
		log.Println("Error making getPendingSiaNodes request: " + err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Could not read getPendingSiaNodes response: " + err.Error())
		return nil, err
	}

	var nodes struct {
		SiaNodes []*models.SiaNode `json:"sianodes"`
	}
	if err = json.Unmarshal(body, &nodes); err != nil {
		log.Println("Could not decode response: " + err.Error())
		return nil, err
	}

	return nodes.SiaNodes, nil
}

func updateSiaNodeStatus(id uuid.UUID, status string) (*models.SiaNode, error) {
	url := fmt.Sprintf("%s/sianodes/status?secret=%s", URLRoot, SiaCDNSecretKey)

	reqBodyData := struct {
		Id     uuid.UUID `json:"id"`
		Status string    `json:"status"`
	}{Id: id, Status: status}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(reqBodyData)
	if err != nil {
		log.Println("Could not encode sia node update json: " + err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		log.Println("Could not create request POST " + url)
		return nil, err
	}

	res, err := cliClient.Do(req)
	if err != nil {
		log.Println("Error making updateSiaNodeStatus request: " + err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Could not read updateSiaNodeStatus response: " + err.Error())
		return nil, err
	}

	var node struct {
		SiaNode *models.SiaNode `json:"sianode"`
	}
	if err = json.Unmarshal(body, &node); err != nil {
		log.Println("Could not decode response: " + err.Error())
		return nil, err
	}
	return node.SiaNode, nil
}

func createWalletSeed(siaNodeID uuid.UUID, words string) (*models.WalletSeed, error) {
	url := fmt.Sprintf("%s/wallets/%s/seed?secret=%s", URLRoot, siaNodeID, SiaCDNSecretKey)

	reqBodyData := struct {
		Words string `json:"words"`
	}{Words: words}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(reqBodyData)
	if err != nil {
		log.Println("Could not encode sia wallet seed create json: " + err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		log.Println("Could not create request POST " + url)
		return nil, err
	}

	res, err := cliClient.Do(req)
	if err != nil {
		log.Println("Error making createWalletSeed request: " + err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Could not read createWalletSeed response: " + err.Error())
		return nil, err
	}

	var resp struct {
		WalletSeed *models.WalletSeed `json:"wallet_seed"`
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("Could not decode response: " + err.Error())
		return nil, err
	}
	return resp.WalletSeed, nil
}

func getWalletSeed(siaNodeID uuid.UUID) (*models.WalletSeed, error) {
	url := fmt.Sprintf("%s/wallets/%s/seed?secret=%s",
		URLRoot, siaNodeID.String(), SiaCDNSecretKey)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("Could not create request GET " + url)
		return nil, err
	}

	res, err := cliClient.Do(req)
	if err != nil {
		log.Println("Error making getWalletSeed request: " + err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Could not read getWalletSeed response: " + err.Error())
		return nil, err
	}

	var resp struct {
		Seed *models.WalletSeed `json:"wallet_seed"`
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("Could not decode response: " + err.Error())
		return nil, err
	}

	return resp.Seed, nil
}
