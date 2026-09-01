package docker

type ContainerInspect struct {
	Id              string                 `json:"Id"`
	Created         string                 `json:"Created"`
	Path            string                 `json:"Path"`
	Args            []string               `json:"Args"`
	State           ContainerState         `json:"State"`
	Image           string                 `json:"Image"`
	Name            string                 `json:"Name"`
	RestartCount    int                    `json:"RestartCount"`
	Driver          string                 `json:"Driver"`
	Platform        string                 `json:"Platform"`
	HostConfig      ContainerHostConfig    `json:"HostConfig"`
	Config          ContainerConfig        `json:"Config"`
	NetworkSettings ContainerNetworkSettings `json:"NetworkSettings"`
}

type ContainerState struct {
	Status     string `json:"Status"`
	Running    bool   `json:"Running"`
	Paused     bool   `json:"Paused"`
	Restarting bool   `json:"Restarting"`
	OOMKilled  bool   `json:"OOMKilled"`
	Dead       bool   `json:"Dead"`
	Pid        int    `json:"Pid"`
	ExitCode   int    `json:"ExitCode"`
	Error      string `json:"Error"`
	StartedAt  string `json:"StartedAt"`
	FinishedAt string `json:"FinishedAt"`
}

type ContainerHostConfig struct {
	NetworkMode   string                   `json:"NetworkMode"`
	PortBindings  map[string][]PortBinding `json:"PortBindings"`
	RestartPolicy ContainerRestartPolicy   `json:"RestartPolicy"`
	AutoRemove    bool                     `json:"AutoRemove"`
}

type ContainerRestartPolicy struct {
	Name              string `json:"Name"`
	MaximumRetryCount int    `json:"MaximumRetryCount"`
}

type ContainerConfig struct {
	Hostname   string            `json:"Hostname"`
	Env        []string          `json:"Env"`
	Cmd        []string          `json:"Cmd"`
	Image      string            `json:"Image"`
	WorkingDir string            `json:"WorkingDir"`
	Entrypoint []string          `json:"Entrypoint"`
	Labels     map[string]string `json:"Labels"`
}

type ContainerNetworkSettings struct {
	SandboxID  string                      `json:"SandboxID"`
	SandboxKey string                      `json:"SandboxKey"`
	Ports      map[string][]PortBinding    `json:"Ports"`
	Networks   map[string]ContainerNetwork `json:"Networks"`
}

type PortBinding struct {
	HostIp   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type ContainerNetwork struct {
	NetworkID  string `json:"NetworkID"`
	EndpointID string `json:"EndpointID"`
	Gateway    string `json:"Gateway"`
	IPAddress  string `json:"IPAddress"`
	MacAddress string `json:"MacAddress"`
}
