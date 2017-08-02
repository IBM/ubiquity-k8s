package resources

const KubernetesVersion_1_5 = "1.5"
const KubernetesVersion_1_6OrLater = "atLeast1.6"
const UbiquityPluginDirName = "ibm~ubiquity"

type FlexVolumeResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	Device     string `json:"device"`
	VolumeName string `json:"volumeName"`
	Attached   bool   `json:"attached"`
}

type FlexVolumeMountRequest struct {
	MountPath   string            `json:"mountPath"`
	MountDevice string            `json:"name"`
	Opts        map[string]string `json:"opts"`
}

type FlexVolumeUnmountRequest struct {
	MountPath string `json:"mountPath"`
}

type FlexVolumeAttachRequest struct {
	Name    string            `json:"name"`
	Host    string            `json:"host"`
	Opts    map[string]string `json:"opts"`
	Version string            `json:"version"`
}
type FlexVolumeWaitForAttachRequest struct {
	Name string            `json:"name"`
	Opts map[string]string `json:"opts"`
}

type FlexVolumeDetachRequest struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	Version string `json:"version"`
}

type FlexVolumeIsAttachedRequest struct {
	Name string            `json:"name"`
	Host string            `json:"host"`
	Opts map[string]string `json:"opts"`
}

type FlexVolumeMountDeviceRequest struct {
	Name string            `json:"name"`
	Path string            `json:"path"`
	Opts map[string]string `json:"opts"`
}

type FlexVolumeUnmountDeviceRequest struct {
	Name string `json:"name"`
}

type FlexVolumeGetVolumeNameRequest struct {
	Opts map[string]string `json:"opts"`
}
