package vmware

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"

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

// LoadConfig loads YAML and .env
func LoadConfig(path string) *Config {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, reading from env")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config.yaml: %v", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config.yaml: %v", err)
	}

	// Replace password_env with actual password
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
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"name", "summary"}, &hosts)
	if err != nil {
		log.Fatalf("Failed to retrieve hosts: %v", err)
	}

	h := hosts[0] // single host
	cpuGHz := h.Summary.Hardware.CpuMhz * int32(h.Summary.Hardware.NumCpuCores) / 1000
	memGB := h.Summary.Hardware.MemorySize / (1024 * 1024 * 1024)

	return HostInfo{
		Name:     h.Summary.Config.Name,
		CPUGHz:   cpuGHz,
		RAMGB:    memGB,
		CPUUsage: h.Summary.QuickStats.OverallCpuUsage,
		MemUsage: h.Summary.QuickStats.OverallMemoryUsage / 1024,
	}
}
