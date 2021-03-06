package catalog

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	apiVersion = "odo.dev/v1alpha1"
)

// DevfileRegistries contains the links of all devfile registries
var DevfileRegistries = []string{
	"https://raw.githubusercontent.com/elsony/devfile-registry/master",
	"https://che-devfile-registry.openshift.io/",
}

// GetDevfileRegistries gets devfile registries from preference file,
// if registry name is specified return the specific registry, otherwise return all registries
func GetDevfileRegistries(registryName string) (map[string]string, error) {
	devfileRegistries := make(map[string]string)

	cfg, err := preference.New()
	if err != nil {
		return nil, err
	}

	if cfg.OdoSettings.RegistryList != nil {
		for _, registry := range *cfg.OdoSettings.RegistryList {
			if len(registryName) != 0 {
				if registryName == registry.Name {
					devfileRegistries[registry.Name] = registry.URL
					return devfileRegistries, nil
				}
			} else {
				devfileRegistries[registry.Name] = registry.URL
			}
		}
	} else {
		return nil, nil
	}

	return devfileRegistries, nil
}

// GetDevfileIndex loads the devfile registry index.json
func GetDevfileIndex(devfileIndexLink string) ([]DevfileIndexEntry, error) {
	var devfileIndex []DevfileIndexEntry

	jsonBytes, err := util.HTTPGetRequest(devfileIndexLink)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to download the devfile index.json from %s", devfileIndexLink)
	}

	err = json.Unmarshal(jsonBytes, &devfileIndex)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to unmarshal the devfile index.json from %s", devfileIndexLink)
	}

	return devfileIndex, nil
}

// GetDevfile loads the devfile
func GetDevfile(devfileLink string) (Devfile, error) {
	var devfile Devfile

	yamlBytes, err := util.HTTPGetRequest(devfileLink)
	if err != nil {
		return Devfile{}, errors.Wrapf(err, "Unable to download the devfile from %s", devfileLink)
	}

	err = yaml.Unmarshal(yamlBytes, &devfile)
	if err != nil {
		return Devfile{}, errors.Wrapf(err, "Unable to unmarshal the devfile from %s", devfileLink)
	}

	return devfile, nil
}

// IsDevfileComponentSupported checks if the devfile is supported
// The supported devfile should satisfy the following conditions:
// 1. Devfile has dockerimage as component type
// 2. Devfile has alias
// 3. Devfile has run command
// 4. Devfile has build command
func IsDevfileComponentSupported(devfile Devfile) bool {
	hasDockerImage := false
	hasAlias := false
	hasRunCommand := false
	hasBuildCommand := false

	for _, component := range devfile.Components {
		if hasDockerImage && hasAlias {
			break
		}

		if !hasDockerImage {
			hasDockerImage = strings.Contains(component.Type, "dockerimage")
		}

		if !hasAlias {
			hasAlias = len(component.Alias) > 0
		}
	}

	for _, command := range devfile.Commands {
		if hasRunCommand && hasBuildCommand {
			break
		}

		if !hasRunCommand {
			hasRunCommand = strings.Contains(command.Name, "devRun")
		}

		if !hasBuildCommand {
			hasBuildCommand = strings.Contains(command.Name, "devBuild")
		}
	}

	if hasDockerImage && hasAlias && hasBuildCommand && hasRunCommand {
		return true
	}

	return false
}

// ListDevfileComponents lists all the available devfile components
func ListDevfileComponents(registryName string) (DevfileComponentTypeList, error) {
	var catalogDevfileList DevfileComponentTypeList
	var err error

	// Get devfile registries
	catalogDevfileList.DevfileRegistries, err = GetDevfileRegistries(registryName)
	if err != nil {
		return catalogDevfileList, err
	}
	if catalogDevfileList.DevfileRegistries == nil {
		return catalogDevfileList, nil
	}

	for registryName, registryURL := range catalogDevfileList.DevfileRegistries {
		// Load the devfile registry index.json
		devfileIndexLink := registryURL + "/devfiles/index.json"
		devfileIndex, err := GetDevfileIndex(devfileIndexLink)
		if err != nil {
			log.Warningf("Registry %s is not set up properly with error: %v", registryName, err)
			break
		}

		// 1. Load devfiles that indexed in devfile registry index.json
		// 2. Populate devfile components with devfile data
		// 3. Form devfile component list
		for _, devfileIndexEntry := range devfileIndex {
			devfileIndexEntryLink := devfileIndexEntry.Links.Link

			// Load the devfile
			devfileLink := registryURL + devfileIndexEntryLink
			// TODO: We send http get resquest in this function multiple times
			// since devfile registry uses different links to host different devfiles,
			// this can reduce the performance especially when we load devfiles from
			// big registry. We may need to rethink and optimize this in the future
			devfile, err := GetDevfile(devfileLink)
			if err != nil {
				log.Warningf("Registry %s is not set up properly with error: %v", registryName, err)
				break
			}

			// Populate devfile component with devfile data and form devfile component list
			catalogDevfile := DevfileComponentType{
				Name:        strings.TrimSuffix(devfile.MetaData.GenerateName, "-"),
				DisplayName: devfileIndexEntry.DisplayName,
				Description: devfileIndexEntry.Description,
				Link:        devfileIndexEntryLink,
				Support:     IsDevfileComponentSupported(devfile),
				Registry: Registry{
					Name: registryName,
					URL:  registryURL,
				},
			}

			catalogDevfileList.Items = append(catalogDevfileList.Items, catalogDevfile)
		}
	}

	return catalogDevfileList, nil
}

// ListComponents lists all the available component types
func ListComponents(client *occlient.Client) (ComponentTypeList, error) {

	catalogList, err := getDefaultBuilderImages(client)
	if err != nil {
		return ComponentTypeList{}, errors.Wrap(err, "unable to get image streams")
	}

	if len(catalogList) == 0 {
		return ComponentTypeList{}, errors.New("unable to retrieve any catalog images from the OpenShift cluster")
	}

	return ComponentTypeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		Items: catalogList,
	}, nil
}

// SearchComponent searches for the component
func SearchComponent(client *occlient.Client, name string) ([]string, error) {
	var result []string
	componentList, err := ListComponents(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list components")
	}

	// do a partial search in all the components
	for _, component := range componentList.Items {
		// we only show components that contain the search term and that have at least non-hidden tag
		// since a component with all hidden tags is not shown in the odo catalog list components either
		if strings.Contains(component.ObjectMeta.Name, name) && len(component.Spec.NonHiddenTags) > 0 {
			result = append(result, component.ObjectMeta.Name)
		}
	}

	return result, nil
}

// ComponentExists returns true if the given component type and the version are valid, false if not
func ComponentExists(client *occlient.Client, componentType string, componentVersion string) (bool, error) {
	imageStream, err := client.GetImageStream("", componentType, componentVersion)
	if err != nil {
		return false, errors.Wrapf(err, "unable to get from catalog")
	}
	if imageStream == nil {
		return false, nil
	}
	return true, nil
}

// ListServices lists all the available service types
func ListServices(client *occlient.Client) (ServiceTypeList, error) {

	clusterServiceClasses, err := getClusterCatalogServices(client)
	if err != nil {
		return ServiceTypeList{}, errors.Wrapf(err, "unable to get cluster serviceClassExternalName")
	}

	// Sorting service classes alphabetically
	// Reference: https://golang.org/pkg/sort/#example_Slice
	sort.Slice(clusterServiceClasses, func(i, j int) bool {
		return clusterServiceClasses[i].Name < clusterServiceClasses[j].Name
	})

	return ServiceTypeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		Items: clusterServiceClasses,
	}, nil
}

// SearchService searches for the services
func SearchService(client *occlient.Client, name string) (ServiceTypeList, error) {
	var result []ServiceType
	serviceList, err := ListServices(client)
	if err != nil {
		return ServiceTypeList{}, errors.Wrap(err, "unable to list services")
	}

	// do a partial search in all the services
	for _, service := range serviceList.Items {
		if strings.Contains(service.ObjectMeta.Name, name) {
			result = append(result, service)
		}
	}

	return ServiceTypeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		Items: result,
	}, nil
}

// getClusterCatalogServices returns the names of all the cluster service
// classes in the cluster
func getClusterCatalogServices(client *occlient.Client) ([]ServiceType, error) {
	var classNames []ServiceType

	classes, err := client.GetClusterServiceClasses()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster service classes")
	}

	planListItems, err := client.GetAllClusterServicePlans()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get service plans")
	}
	for _, class := range classes {

		var planList []string
		for _, plan := range planListItems {
			if plan.Spec.ClusterServiceClassRef.Name == class.Spec.ExternalID {
				planList = append(planList, plan.Spec.ExternalName)
			}
		}
		classNames = append(classNames, ServiceType{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceType",
				APIVersion: apiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: class.Spec.ExternalName,
			},
			Spec: ServiceSpec{
				Hidden:   occlient.HasTag(class.Spec.Tags, "hidden"),
				PlanList: planList,
			},
		})
	}
	return classNames, nil
}

// getDefaultBuilderImages returns the default builder images available in the
// openshift and the current namespaces
func getDefaultBuilderImages(client *occlient.Client) ([]ComponentType, error) {

	var imageStreams []imagev1.ImageStream
	currentNamespace := client.GetCurrentProjectName()

	// Fetch imagestreams from default openshift namespace
	openshiftNSImageStreams, openshiftNSISFetchError := client.GetImageStreams(occlient.OpenShiftNameSpace)
	if openshiftNSISFetchError != nil {
		// Tolerate the error as it might only be a partial failure
		// We may get the imagestreams from other Namespaces
		//err = errors.Wrapf(openshiftNSISFetchError, "unable to get Image Streams from namespace %s", occlient.OpenShiftNameSpace)
		// log it for debugging purposes
		klog.V(4).Infof("Unable to get Image Streams from namespace %s. Error %s", occlient.OpenShiftNameSpace, openshiftNSISFetchError.Error())
	}

	// Fetch imagestreams from current namespace
	currentNSImageStreams, currentNSISFetchError := client.GetImageStreams(currentNamespace)
	// If failure to fetch imagestreams from current namespace, log the failure for debugging purposes
	if currentNSISFetchError != nil {
		// Tolerate the error as it is totally a valid scenario to not have any imagestreams in current namespace
		// log it for debugging purposes
		klog.V(4).Infof("Unable to get Image Streams from namespace %s. Error %s", currentNamespace, currentNSISFetchError.Error())
	}

	// If failure fetching imagestreams from both namespaces, error out
	if openshiftNSISFetchError != nil && currentNSISFetchError != nil {
		return nil, errors.Wrapf(
			fmt.Errorf("%s.\n%s", openshiftNSISFetchError, currentNSISFetchError),
			"Failed to fetch imagestreams from both openshift and %s namespaces.\nPlease ensure that a builder imagestream of required version for the component exists in either openshift or %s namespaces",
			currentNamespace,
			currentNamespace,
		)
	}

	// Resultant imagestreams is list of imagestreams from current and openshift namespaces
	imageStreams = append(imageStreams, openshiftNSImageStreams...)
	imageStreams = append(imageStreams, currentNSImageStreams...)

	// create a map from name (builder image name + tag) to the ImageStreamTag
	// we need this in order to filter out hidden tags
	imageStreamTagMap := make(map[string]imagev1.ImageStreamTag)

	currentNSImageStreamTags, currentNSImageStreamTagsErr := client.GetImageStreamTags(currentNamespace)
	openshiftNSImageStreamTags, openshiftNSImageStreamTagsErr := client.GetImageStreamTags(occlient.OpenShiftNameSpace)

	// If failure fetching imagestreamtags from both namespaces, error out
	if currentNSImageStreamTagsErr != nil && openshiftNSImageStreamTagsErr != nil {
		return nil, errors.Wrapf(
			fmt.Errorf("%s.\n%s", currentNSImageStreamTagsErr, openshiftNSImageStreamTagsErr),
			"Failed to fetch imagestreamtags from both openshift and %s namespaces.\nPlease ensure that a builder imagestream of required version for the component exists in either openshift or %s namespaces",
			currentNamespace,
			currentNamespace,
		)
	}

	// create a map from name to ImageStreamTag out of all the ImageStreamTag objects we collect
	var imageStreamTags []imagev1.ImageStreamTag
	imageStreamTags = append(imageStreamTags, currentNSImageStreamTags...)
	imageStreamTags = append(imageStreamTags, openshiftNSImageStreamTags...)
	for _, imageStreamTag := range imageStreamTags {
		imageStreamTagMap[imageStreamTag.Name] = imageStreamTag
	}

	builderImages := getBuildersFromImageStreams(imageStreams, imageStreamTagMap)

	return builderImages, nil
}

// SliceSupportedTags splits the tags in to fully supported and unsupported tags
func SliceSupportedTags(component ComponentType) ([]string, []string) {

	// this makes sure that json marshal shows these lists as [] instead of null
	supTag, unSupTag := []string{}, []string{}
	tagMap := createImageTagMap(component.Spec.ImageStreamRef.Spec.Tags)

	for _, tag := range component.Spec.NonHiddenTags {
		imageName := tagMap[tag]
		if isSupportedImage(imageName) {
			supTag = append(supTag, tag)
		} else {
			unSupTag = append(unSupTag, tag)
		}
	}
	return supTag, unSupTag
}

// IsComponentTypeSupported takes the componentType e.g. java:8 and return true if
// it is fully supported i.e. debug mode and more.
func IsComponentTypeSupported(client *occlient.Client, componentType string) (bool, error) {
	_, componentType, _, componentVersion := util.ParseComponentImageName(componentType)

	imageStream, err := client.GetImageStream("", componentType, componentVersion)
	if err != nil {
		return false, err
	}
	tagMap := createImageTagMap(imageStream.Spec.Tags)

	return isSupportedImage(tagMap[componentVersion]), nil
}

// createImageTagMap takes a list of image TagReferences and creates a map of type tag name => image name e.g. 1.11 => openshift/nodejs-11
func createImageTagMap(tagRefs []imagev1.TagReference) map[string]string {
	tagMap := make(map[string]string)
	for _, tagRef := range tagRefs {
		imageName := tagRef.From.Name
		if tagRef.From.Kind == "DockerImage" {
			// we get the image name from the repo url e.g. registry.redhat.com/openshift/nodejs:10 will give openshift/nodejs:10
			imageNameParts := strings.SplitN(imageName, "/", 2)

			var urlImageName string
			// this means the docker image url might just be something like nodejs:10, no namespace or registry info
			if len(imageNameParts) == 1 {
				urlImageName = imageNameParts[0]
				// else block executes when there is a registry information attached in the docker image url
			} else {
				// we dont want the registry url portion
				urlImageName = imageNameParts[1]
			}
			// here we remove the tag and digest
			ns, img, tag, _, _ := occlient.ParseImageName(urlImageName)
			imageName = ns + "/" + img + ":" + tag
		} else if tagRef.From.Kind == "ImageStreamTag" {
			tagList := strings.Split(imageName, ":")
			tag := tagList[len(tagList)-1]
			// if the kind is a image stream tag that means its pointing to an existing dockerImage or image stream image
			// we just look it up from the tapMap we already have
			imageName = tagMap[tag]
		}
		tagMap[tagRef.Name] = imageName
	}
	return tagMap
}

// isSupportedImages returns if the image is supported or not. the supported images have been provided here
// https://github.com/openshift/odo-init-image/blob/master/language-scripts/image-mappings.json
func isSupportedImage(imgName string) bool {
	supportedImages := []string{
		"redhat-openjdk-18/openjdk18-openshift:latest",
		"openjdk/openjdk-11-rhel8:latest",
		"openjdk/openjdk-11-rhel7:latest",
		"centos/nodejs-10-centos7:latest",
		"centos/nodejs-12-centos7:latest",
		"rhscl/nodejs-10-rhel7:latest",
		"rhscl/nodejs-12-rhel7:latest",
		"bucharestgold/centos7-s2i-nodejs:latest",
		"nodeshift/centos7-s2i-nodejs:latest",

		// older images which we should remove soon
		"rhoar-nodejs/nodejs-8:latest",
		"rhoar-nodejs/nodejs-10:latest",
		"bucharestgold/centos7-s2i-nodejs:8.x",
		"bucharestgold/centos7-s2i-nodejs:10.x",
		"centos/nodejs-8-centos7:latest",
		"rhscl/nodejs-8-rhel7:latest",
	}
	for _, supImage := range supportedImages {
		if supImage == imgName {
			return true
		}
	}
	return false
}

// getBuildersFromImageStreams returns all the builder Images from the image streams provided and also hides the builder images
// which have hidden annotation attached to it
func getBuildersFromImageStreams(imageStreams []imagev1.ImageStream, imageStreamTagMap map[string]imagev1.ImageStreamTag) []ComponentType {
	var builderImages []ComponentType
	// Get builder images from the available imagestreams
	for _, imageStream := range imageStreams {
		var allTags []string
		var hiddenTags []string
		buildImage := false

		for _, tagReference := range imageStream.Spec.Tags {
			allTags = append(allTags, tagReference.Name)
			// Check to see if it is a "builder" image
			if _, ok := tagReference.Annotations["tags"]; ok {
				for _, t := range strings.Split(tagReference.Annotations["tags"], ",") {
					// If the tagReference has "builder" then we will add the image to the list
					if t == "builder" {
						buildImage = true
					}
				}
			}

		}

		// Append to the list of images if a "builder" tag was found
		if buildImage {
			// We need to gauge the ImageStreamTag of each potential builder image, because it might contain
			// the 'hidden' tag. If so, this builder image is deprecated and should not be offered to the user
			// as candidate
			for _, tag := range allTags {
				imageStreamTag := imageStreamTagMap[imageStream.Name+":"+tag]
				if _, ok := imageStreamTag.Annotations["tags"]; ok {
					for _, t := range strings.Split(imageStreamTag.Annotations["tags"], ",") {
						// If the tagReference has "builder" then we will add the image to the list
						if t == "hidden" {
							klog.V(5).Infof("Tag: %v of builder: %v is marked as hidden and therefore will be excluded", tag, imageStream.Name)
							hiddenTags = append(hiddenTags, tag)
						}
					}
				}

			}

			catalogImage := ComponentType{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ComponentType",
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      imageStream.Name,
					Namespace: imageStream.Namespace,
				},
				Spec: ComponentSpec{
					AllTags:        allTags,
					NonHiddenTags:  getAllNonHiddenTags(allTags, hiddenTags),
					ImageStreamRef: imageStream,
				},
			}
			builderImages = append(builderImages, catalogImage)
			klog.V(5).Infof("Found builder image: %#v", catalogImage)
		}

	}
	return builderImages
}

func getAllNonHiddenTags(allTags []string, hiddenTags []string) []string {
	result := make([]string, 0, len(allTags))
	for _, t1 := range allTags {
		found := false
		for _, t2 := range hiddenTags {
			if t1 == t2 {
				found = true
				break
			}
		}

		if !found {
			result = append(result, t1)
		}
	}
	return result
}
