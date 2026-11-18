// Package risk centralise les ports considérés comme risqués et leur gravité.
package risk

const (
	Critical = "critical"
	High     = "high"
	Medium   = "medium"
)

type portInfo struct {
	Service  string
	Severity string
}

var ports = map[int]portInfo{
	21:   {"FTP", Critical},
	23:   {"Telnet", Critical},
	445:  {"SMB", High},
	3389: {"RDP", High},
	1433: {"MSSQL", High},
	3306: {"MySQL", High},
	5432: {"PostgreSQL", Medium},
	6379: {"Redis", Medium},
}

// Port renvoie le service et la gravité d'un port risqué ; ok=false si le port n'est pas surveillé.
func Port(number int) (service, severity string, ok bool) {
	p, ok := ports[number]
	return p.Service, p.Severity, ok
}

// Numbers liste les ports surveillés.
func Numbers() []int {
	out := make([]int, 0, len(ports))
	for n := range ports {
		out = append(out, n)
	}
	return out
}

// Weight est le nombre de points retirés au score de sécurité pour un problème de cette gravité.
func Weight(severity string) int {
	switch severity {
	case Critical:
		return 30
	case High:
		return 12
	case Medium:
		return 6
	}
	return 0
}
