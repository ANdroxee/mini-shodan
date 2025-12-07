package scanner

import (
	"context"
	"github.com/Ullaakut/nmap/v3"
	"mini-shodan-backend/internal/models"
)


func ScanTarget(target string) (*models.Host, error) {

    scanner, _ := nmap.NewScanner(
        context.Background(),
        nmap.WithTargets(target),
        nmap.WithPorts("1-1000"), 
        nmap.WithServiceInfo(),
    )

    result, _, err := scanner.Run()
    if err != nil {
        return nil, err
    }

    hostData := &models.Host{
        IP: target,
        // ... on remplit le reste ici
    }
    
    return hostData, nil
}