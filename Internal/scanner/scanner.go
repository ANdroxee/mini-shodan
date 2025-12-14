package scanner

import (
	"context"
	"encoding/xml"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"shodan-home/internal/database"
)

// Scanner gère les scans réseau
type Scanner struct {
	workers  int
	timeout  time.Duration
	nmapPath string
}

// NmapRun structure pour parser le XML Nmap
type NmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []NmapHost `xml:"host"`
}

type NmapHost struct {
	Status struct {
		State string `xml:"state,attr"`
	} `xml:"status"`
	Addresses []struct {
		Addr     string `xml:"addr,attr"`
		AddrType string `xml:"addrtype,attr"`
		Vendor   string `xml:"vendor,attr"`
	} `xml:"address"`
	Hostnames struct {
		Hostname []struct {
			Name string `xml:"name,attr"`
			Type string `xml:"type,attr"`
		} `xml:"hostname"`
	} `xml:"hostnames"`
	OS struct {
		OSMatch []struct {
			Name     string `xml:"name,attr"`
			Accuracy int    `xml:"accuracy,attr"`
		} `xml:"osmatch"`
	} `xml:"os"`
	Ports struct {
		Port []struct {
			PortID   int    `xml:"portid,attr"`
			Protocol string `xml:"protocol,attr"`
			State    struct {
				State string `xml:"state,attr"`
			} `xml:"state"`
			Service struct {
				Name    string `xml:"name,attr"`
				Product string `xml:"product,attr"`
				Version string `xml:"version,attr"`
			} `xml:"service"`
			Script []struct {
				ID     string `xml:"id,attr"`
				Output string `xml:"output,attr"`
			} `xml:"script"`
		} `xml:"port"`
	} `xml:"ports"`
}

// NewScanner crée un nouveau scanner
func NewScanner(workers int) *Scanner {
	return &Scanner{
		workers:  workers,
		timeout:  60 * time.Second,
		nmapPath: "nmap",
	}
}

// ScanNetwork scanne un réseau complet
func (s *Scanner) ScanNetwork(ctx context.Context, target string) ([]database.Device, error) {
	fmt.Printf("🔍 Discovering hosts on %s...\n", target)

	// Étape 1: Découverte rapide des hôtes
	activeHosts, err := s.discoverHosts(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("host discovery failed: %w", err)
	}

	fmt.Printf("✅ Found %d active hosts\n", len(activeHosts))

	// Étape 2: Scan détaillé avec workers
	return s.scanHostsConcurrent(ctx, activeHosts)
}

// discoverHosts fait un ping scan rapide
func (s *Scanner) discoverHosts(ctx context.Context, target string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.nmapPath, "-sn", "-T4", target, "-oX", "-")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nmap discovery failed: %w", err)
	}

	var nmapRun NmapRun
	if err := xml.Unmarshal(output, &nmapRun); err != nil {
		return nil, fmt.Errorf("failed to parse nmap output: %w", err)
	}

	var hosts []string
	for _, host := range nmapRun.Hosts {
		if host.Status.State == "up" {
			for _, addr := range host.Addresses {
				if addr.AddrType == "ipv4" {
					hosts = append(hosts, addr.Addr)
				}
			}
		}
	}

	return hosts, nil
}

// scanHostsConcurrent scanne plusieurs hôtes en parallèle
func (s *Scanner) scanHostsConcurrent(ctx context.Context, hosts []string) ([]database.Device, error) {
	results := make(chan database.Device, len(hosts))
	errors := make(chan error, len(hosts))
	var wg sync.WaitGroup

	// Canal de jobs
	jobs := make(chan string, len(hosts))

	// Lancer les workers
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					device, err := s.scanHost(ctx, ip)
					if err != nil {
						errors <- err
						continue
					}
					results <- device
				}
			}
		}()
	}

	// Envoyer les jobs
	for _, host := range hosts {
		jobs <- host
	}
	close(jobs)

	// Attendre la fin
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collecter les résultats
	var devices []database.Device
	for device := range results {
		devices = append(devices, device)
		fmt.Printf("  ✓ Scanned %s (%s) - %d ports\n", device.IP, device.Hostname, len(device.Ports))
	}

	return devices, nil
}

// scanHost scanne un hôte en détail
func (s *Scanner) scanHost(ctx context.Context, ip string) (database.Device, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Scan avec version detection
	cmd := exec.CommandContext(
		ctx,
		s.nmapPath,
		"-sV",             // Version detection
		"-O",              // OS detection
		"--script=banner", // Grab banners
		"-T4",             // Aggressive timing
		ip,
		"-oX", "-",
	)

	output, err := cmd.Output()
	if err != nil {
		return database.Device{}, fmt.Errorf("scan failed for %s: %w", ip, err)
	}

	return s.parseNmapOutput(output)
}

// parseNmapOutput parse le XML Nmap
func (s *Scanner) parseNmapOutput(xmlData []byte) (database.Device, error) {
	var nmapRun NmapRun
	if err := xml.Unmarshal(xmlData, &nmapRun); err != nil {
		return database.Device{}, fmt.Errorf("failed to parse XML: %w", err)
	}

	if len(nmapRun.Hosts) == 0 {
		return database.Device{}, fmt.Errorf("no hosts found in output")
	}

	host := nmapRun.Hosts[0]
	device := database.Device{
		LastSeen: time.Now(),
	}

	// Extraire les adresses
	for _, addr := range host.Addresses {
		switch addr.AddrType {
		case "ipv4":
			device.IP = addr.Addr
		case "mac":
			device.MAC = addr.Addr
			device.Vendor = addr.Vendor
		}
	}

	// Hostname
	if len(host.Hostnames.Hostname) > 0 {
		device.Hostname = host.Hostnames.Hostname[0].Name
	}

	// OS
	if len(host.OS.OSMatch) > 0 {
		device.OS = host.OS.OSMatch[0].Name
	}

	// Ports
	for _, p := range host.Ports.Port {
		if p.State.State != "open" {
			continue
		}

		port := database.Port{
			Number:   p.PortID,
			Protocol: p.Protocol,
			Service:  p.Service.Name,
			Version:  fmt.Sprintf("%s %s", p.Service.Product, p.Service.Version),
			LastSeen: time.Now(),
		}

		// Banner
		for _, script := range p.Script {
			if script.ID == "banner" {
				port.Banner = script.Output
				break
			}
		}

		device.Ports = append(device.Ports, port)
	}

	return device, nil
}

// QuickScan fait un scan rapide sans version detection
func (s *Scanner) QuickScan(ctx context.Context, target string) ([]database.Device, error) {
	cmd := exec.CommandContext(ctx, s.nmapPath, "-sS", "-T4", target, "-oX", "-")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var nmapRun NmapRun
	if err := xml.Unmarshal(output, &nmapRun); err != nil {
		return nil, err
	}

	var devices []database.Device
	for _, host := range nmapRun.Hosts {
		device := s.convertNmapHost(host)
		devices = append(devices, device)
	}

	return devices, nil
}

// convertNmapHost convertit NmapHost en database.Device
func (s *Scanner) convertNmapHost(host NmapHost) database.Device {
	device := database.Device{LastSeen: time.Now()}

	for _, addr := range host.Addresses {
		switch addr.AddrType {
		case "ipv4":
			device.IP = addr.Addr
		case "mac":
			device.MAC = addr.Addr
			device.Vendor = addr.Vendor
		}
	}

	if len(host.Hostnames.Hostname) > 0 {
		device.Hostname = host.Hostnames.Hostname[0].Name
	}

	if len(host.OS.OSMatch) > 0 {
		device.OS = host.OS.OSMatch[0].Name
	}

	for _, p := range host.Ports.Port {
		if p.State.State == "open" {
			device.Ports = append(device.Ports, database.Port{
				Number:   p.PortID,
				Protocol: p.Protocol,
				Service:  p.Service.Name,
				Version:  fmt.Sprintf("%s %s", p.Service.Product, p.Service.Version),
				LastSeen: time.Now(),
			})
		}
	}

	return device
}
