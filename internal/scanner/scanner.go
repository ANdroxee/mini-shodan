package scanner

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"shodan-home/internal/database"
)

// Scanner gère les scans réseau avec nmap
type Scanner struct {
	workers     int
	hostTimeout time.Duration
	nmapPath    string
	privileged  bool // nmap peut utiliser les sockets brutes (root ou capacités NET_RAW/NET_ADMIN)
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
	if workers < 1 {
		workers = 1
	}
	// En conteneur, nmap reçoit les capacités réseau (setcap) : NMAP_PRIVILEGED=true l'en informe.
	privileged := os.Geteuid() == 0 || os.Getenv("NMAP_PRIVILEGED") == "true"
	return &Scanner{
		workers:     workers,
		hostTimeout: 3 * time.Minute,
		nmapPath:    "nmap",
		privileged:  privileged,
	}
}

func (s *Scanner) baseArgs() []string {
	if s.privileged && os.Geteuid() != 0 {
		return []string{"--privileged"}
	}
	return nil
}

// ScanNetwork découvre les hôtes actifs puis les scanne en détail
func (s *Scanner) ScanNetwork(ctx context.Context, target string) ([]database.Device, error) {
	fmt.Printf("🔍 Discovering hosts on %s...\n", target)

	discovered, err := s.discoverHosts(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("host discovery failed: %w", err)
	}
	fmt.Printf("✅ Found %d active hosts\n", len(discovered))

	return s.scanHostsConcurrent(ctx, discovered), nil
}

// discoverHosts fait un ping scan rapide ; garde ce qu'on y apprend (MAC, fabricant, nom)
func (s *Scanner) discoverHosts(ctx context.Context, target string) ([]database.Device, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	args := append(s.baseArgs(), "-sn", "-T4", target, "-oX", "-")
	output, err := exec.CommandContext(ctx, s.nmapPath, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("nmap discovery failed: %w", err)
	}

	var run NmapRun
	if err := xml.Unmarshal(output, &run); err != nil {
		return nil, fmt.Errorf("failed to parse nmap output: %w", err)
	}

	var hosts []database.Device
	for _, h := range run.Hosts {
		if h.Status.State != "up" {
			continue
		}
		d := convertHost(h)
		if d.IP != "" {
			hosts = append(hosts, d)
		}
	}
	return hosts, nil
}

// scanHostsConcurrent scanne les hôtes en parallèle. Un hôte dont le scan détaillé échoue
// reste dans les résultats (marqué Partial) avec les infos de découverte.
func (s *Scanner) scanHostsConcurrent(ctx context.Context, hosts []database.Device) []database.Device {
	jobs := make(chan database.Device)
	results := make(chan database.Device, len(hosts))
	var wg sync.WaitGroup

	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for base := range jobs {
				device, err := s.scanHost(ctx, base.IP)
				if err != nil {
					fmt.Printf("  ⚠️  %s: %v\n", base.IP, err)
					base.Partial = true
					results <- base
					continue
				}
				// Compléter avec la découverte (la MAC n'est visible que sur le même réseau local)
				if device.MAC == "" {
					device.MAC, device.Vendor = base.MAC, base.Vendor
				}
				if device.Hostname == "" {
					device.Hostname = base.Hostname
				}
				results <- device
			}
		}()
	}

	go func() {
		for _, h := range hosts {
			select {
			case jobs <- h:
			case <-ctx.Done():
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var devices []database.Device
	for d := range results {
		devices = append(devices, d)
		fmt.Printf("  ✓ Scanned %s (%s) - %d ports\n", d.IP, d.Hostname, len(d.Ports))
	}
	return devices
}

// scanHost scanne un hôte en détail (ports, versions, OS si possible, bannières)
func (s *Scanner) scanHost(ctx context.Context, ip string) (database.Device, error) {
	ctx, cancel := context.WithTimeout(ctx, s.hostTimeout)
	defer cancel()

	args := append(s.baseArgs(), "-sV", "--script=banner", "-T4")
	if s.privileged {
		args = append(args, "-O", "--osscan-limit") // la détection d'OS exige des droits réseau
	}
	args = append(args, ip, "-oX", "-")

	output, err := exec.CommandContext(ctx, s.nmapPath, args...).Output()
	if err != nil {
		return database.Device{}, fmt.Errorf("scan failed: %w", err)
	}

	var run NmapRun
	if err := xml.Unmarshal(output, &run); err != nil {
		return database.Device{}, fmt.Errorf("failed to parse XML: %w", err)
	}
	if len(run.Hosts) == 0 {
		return database.Device{}, fmt.Errorf("host did not answer")
	}
	return convertHost(run.Hosts[0]), nil
}

// convertHost transforme un hôte nmap en appareil
func convertHost(h NmapHost) database.Device {
	d := database.Device{LastSeen: time.Now()}

	for _, addr := range h.Addresses {
		switch addr.AddrType {
		case "ipv4":
			d.IP = addr.Addr
		case "mac":
			d.MAC = strings.ToLower(addr.Addr)
			d.Vendor = addr.Vendor
		}
	}
	if len(h.Hostnames.Hostname) > 0 {
		d.Hostname = h.Hostnames.Hostname[0].Name
	}
	if len(h.OS.OSMatch) > 0 {
		d.OS = h.OS.OSMatch[0].Name
	}

	for _, p := range h.Ports.Port {
		if p.State.State != "open" {
			continue
		}
		port := database.Port{
			Number:   p.PortID,
			Protocol: p.Protocol,
			Service:  p.Service.Name,
			Version:  strings.TrimSpace(p.Service.Product + " " + p.Service.Version),
			LastSeen: time.Now(),
		}
		for _, script := range p.Script {
			if script.ID == "banner" {
				port.Banner = script.Output
				break
			}
		}
		d.Ports = append(d.Ports, port)
	}
	return d
}
