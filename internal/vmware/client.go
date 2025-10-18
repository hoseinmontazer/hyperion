package vmware

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/joho/godotenv"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

type HostConfig struct {
	Host        string `yaml:"host"`
	Username    string `yaml:"username"`
	PasswordEnv string `yaml:"password_env"`
	Insecure    bool   `yaml:"insecure"`
}

type Config struct {
	ESXi []HostConfig `yaml:"esxi"`
}

func LoadConfig() *Config {

	// Load .env first
	_ = godotenv.Load() // optional, if you want .env file support

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Cannot get home directory: %v", err)

	}
	configDir := filepath.Join(home, ".hyperion")
	configPath := filepath.Join(configDir, "config.yaml")

	// If config folder doesn't exist, create it
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0700)
		if err != nil {
			log.Fatalf("Failed to create config directory: %v", err)
		}
	}

	// If config file doesn't exist, create sample
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		sample := Config{
			ESXi: []HostConfig{
				{
					Host:        "192.168.100.10",
					Username:    "root",
					PasswordEnv: "ESXI_PASSWORD_1",
					Insecure:    true,
				},
				{
					Host:        "192.168.100.11",
					Username:    "root",
					PasswordEnv: "ESXI_PASSWORD_2",
					Insecure:    true,
				},
			},
		}
		data, _ := yaml.Marshal(&sample)
		err := ioutil.WriteFile(configPath, data, 0600)
		if err != nil {
			log.Fatalf("Failed to write sample config: %v", err)
		}
		fmt.Printf("Sample config created at %s\n", configPath)
		fmt.Println("Please edit it and set your ESXi passwords in environment variables")
		os.Exit(0)
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config.yaml: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config.yaml: %v", err)
	}

	// Replace password_env with actual password from ENV
	for i := range cfg.ESXi {
		pw := os.Getenv(cfg.ESXi[i].PasswordEnv)
		if pw == "" {
			log.Fatalf("Environment variable %s not set", cfg.ESXi[i].PasswordEnv)
		}
		cfg.ESXi[i].PasswordEnv = pw
	}

	return &cfg
}

func ConnectESXI(ctx context.Context, host HostConfig) *govmomi.Client {
	u, err := url.Parse(fmt.Sprintf("https://%s/sdk", host.Host))
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)

	}

	u.User = url.UserPassword(host.Username, host.PasswordEnv)
	c, err := govmomi.NewClient(ctx, u, host.Insecure)

	if err != nil {
		log.Fatalf("Failed to connect to ESXi: %v", err)
	}

	return c

}

// ListVMs lists VMs with name, power state, IP
func ListVMs(ctx context.Context, c *govmomi.Client) {
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		log.Fatalf("Failed to create container view: %v", err)
	}
	defer v.Destroy(ctx)

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name", "summary"}, &vms)
	if err != nil {
		log.Fatalf("Failed to retrieve VMs: %v", err)
	}

	fmt.Println("VMs:")
	for _, vm := range vms {
		name := vm.Name
		power := vm.Summary.Runtime.PowerState
		ip := vm.Summary.Guest.IpAddress
		fmt.Printf(" - %-20s | Power: %-8s | IP: %s\n", name, power, ip)
	}
}

// ListHostResources prints CPU, RAM, and datastore info
func ListHostResources(ctx context.Context, c *govmomi.Client) {
	// Create view for hosts
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		log.Fatalf("Failed to create host container view: %v", err)
	}
	defer v.Destroy(ctx)

	var hosts []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"name", "summary"}, &hosts)
	if err != nil {
		log.Fatalf("Failed to retrieve hosts: %v", err)
	}

	for _, h := range hosts {
		name := h.Summary.Config.Name
		cpu := h.Summary.Hardware.CpuMhz * int32(h.Summary.Hardware.NumCpuCores) / 1000 // GHz approx
		memGB := h.Summary.Hardware.MemorySize / (1024 * 1024 * 1024)
		fmt.Printf(" - %-15s | CPU: %-3d GHz | RAM: %-3d GB\n", name, cpu, memGB)
	}
}

// ShowHostUsage prints current CPU and memory usage
func ShowHostUsage(ctx context.Context, c *govmomi.Client) {
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		log.Fatalf("Failed to create host container view: %v", err)
	}
	defer v.Destroy(ctx)

	var hosts []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"name", "summary"}, &hosts)
	if err != nil {
		log.Fatalf("Failed to retrieve hosts: %v", err)
	}

	for _, h := range hosts {
		name := h.Summary.Config.Name
		cpuUsage := h.Summary.QuickStats.OverallCpuUsage           // MHz
		memUsage := h.Summary.QuickStats.OverallMemoryUsage / 1024 // MB
		fmt.Printf(" - %-15s | CPU Usage: %-5d MHz | Memory Usage: %-5d MB\n", name, cpuUsage, memUsage)
	}
}
func GetVMNames(ctx context.Context, c *govmomi.Client) []string {
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		log.Fatalf("Failed to create container view: %v", err)
	}
	defer v.Destroy(ctx)

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name"}, &vms)
	if err != nil {
		log.Fatalf("Failed to retrieve VMs: %v", err)
	}

	names := []string{}
	for _, vm := range vms {
		names = append(names, vm.Name)
	}
	return names
}

type HostInfo struct {
	Name     string
	IP       string
	CPUGHz   int32
	RAMGB    int64
	CPUUsage int32
	MemUsage int32
}

// GetHostInfo retrieves host summary
func GetHostInfo(ctx context.Context, c *govmomi.Client) HostInfo {
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		log.Fatalf("Failed to create host container view: %v", err)
	}
	defer v.Destroy(ctx)

	var hosts []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"name", "summary", "config.network.vnic"}, &hosts)
	if err != nil {
		log.Fatalf("Failed to retrieve hosts: %v", err)
	}

	h := hosts[0] // single host
	cpuGHz := h.Summary.Hardware.CpuMhz * int32(h.Summary.Hardware.NumCpuCores) / 1000
	memGB := h.Summary.Hardware.MemorySize / (1024 * 1024 * 1024)
	ip := ""
	if len(h.Config.Network.Vnic) > 0 {
		for _, vnic := range h.Config.Network.Vnic {
			if vnic.Spec.Ip != nil && vnic.Spec.Ip.IpAddress != "" {
				ip = vnic.Spec.Ip.IpAddress
				break
			}
		}
	}

	return HostInfo{
		Name:     h.Summary.Config.Name,
		IP:       ip,
		CPUGHz:   cpuGHz,
		RAMGB:    memGB,
		CPUUsage: h.Summary.QuickStats.OverallCpuUsage,
		MemUsage: h.Summary.QuickStats.OverallMemoryUsage / 1024,
	}
}

type VMInfo struct {
	Name      string
	Host      string
	Power     string
	IP        string
	CPU       int32
	RAMMB     int32
	GuestOS   string
	StorageGB int64
	Networks  []string
}

// GetVMInfos returns VM names and IPs
func GetVMInfos(ctx context.Context, c *govmomi.Client) []VMInfo {
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		log.Fatalf("Failed to create container view: %v", err)
	}
	defer v.Destroy(ctx)

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name", "summary", "guest", "config", "storage", "network"}, &vms)
	if err != nil {
		log.Fatalf("Failed to retrieve VMs: %v", err)
	}

	var vmInfos []VMInfo
	for _, vm := range vms {
		// Gather network IPs
		var ips []string
		if vm.Guest != nil && vm.Guest.Net != nil {
			for _, n := range vm.Guest.Net {
				for _, ip := range n.IpAddress {
					ips = append(ips, ip)
				}
			}
		}

		// Storage in GB (simple approximation)
		storageGB := int64(0)
		if vm.Summary.Storage != nil {
			storageGB = vm.Summary.Storage.Committed / (1024 * 1024 * 1024)
		}

		vmInfos = append(vmInfos, VMInfo{
			Name:      vm.Summary.Config.Name,
			Power:     string(vm.Summary.Runtime.PowerState),
			IP:        vm.Summary.Guest.IpAddress,
			CPU:       int32(vm.Summary.Config.NumCpu),
			RAMMB:     vm.Summary.Config.MemorySizeMB,
			GuestOS:   vm.Summary.Config.GuestId,
			StorageGB: storageGB,
			Networks:  ips,
		})
	}

	return vmInfos
}
