-- Shodan@Home Database Schema
-- Version: 1.0

-- Table des devices
CREATE TABLE IF NOT EXISTS devices (
    id SERIAL PRIMARY KEY,
    ip VARCHAR(45) NOT NULL UNIQUE,
    hostname VARCHAR(255),
    mac VARCHAR(17),
    vendor VARCHAR(255),
    os VARCHAR(255),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index pour les recherches rapides
CREATE INDEX idx_devices_ip ON devices(ip);
CREATE INDEX idx_devices_last_seen ON devices(last_seen);
CREATE INDEX idx_devices_vendor ON devices(vendor);

-- Table des ports
CREATE TABLE IF NOT EXISTS ports (
    id SERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    port INTEGER NOT NULL,
    protocol VARCHAR(10) NOT NULL DEFAULT 'tcp',
    service VARCHAR(100),
    version VARCHAR(255),
    banner TEXT,
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(device_id, port, protocol)
);

-- Index pour les ports
CREATE INDEX idx_ports_device_id ON ports(device_id);
CREATE INDEX idx_ports_port ON ports(port);
CREATE INDEX idx_ports_service ON ports(service);

-- Table des scans (historique)
CREATE TABLE IF NOT EXISTS scans (
    id SERIAL PRIMARY KEY,
    target VARCHAR(100) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_seconds INTEGER,
    hosts_found INTEGER DEFAULT 0,
    ports_found INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'running'
);

CREATE INDEX idx_scans_start_time ON scans(start_time);

-- Table des alertes
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    device_id INTEGER REFERENCES devices(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    read BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_alerts_read ON alerts(read);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_created_at ON alerts(created_at);

-- Vue pour les devices actifs (vus dans les dernières 24h)
CREATE OR REPLACE VIEW active_devices AS
SELECT *
FROM devices
WHERE last_seen > NOW() - INTERVAL '24 hours';

-- Vue pour les ports risqués
CREATE OR REPLACE VIEW risky_ports AS
SELECT 
    d.ip,
    d.hostname,
    p.port,
    p.service,
    p.version,
    CASE 
        WHEN p.port IN (21, 23) THEN 'critical'
        WHEN p.port IN (445, 3389, 1433, 3306) THEN 'high'
        ELSE 'medium'
    END as risk_level
FROM devices d
JOIN ports p ON d.id = p.device_id
WHERE p.port IN (21, 23, 445, 3389, 1433, 3306, 5432, 6379);

-- Fonction pour nettoyer les vieux scans
CREATE OR REPLACE FUNCTION cleanup_old_scans()
RETURNS void AS $$
BEGIN
    DELETE FROM scans
    WHERE start_time < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

-- Fonction pour obtenir les stats
CREATE OR REPLACE FUNCTION get_network_stats()
RETURNS TABLE (
    total_devices INTEGER,
    active_devices INTEGER,
    total_ports INTEGER,
    unique_services INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*)::INTEGER FROM devices) as total_devices,
        (SELECT COUNT(*)::INTEGER FROM active_devices) as active_devices,
        (SELECT COUNT(*)::INTEGER FROM ports) as total_ports,
        (SELECT COUNT(DISTINCT service)::INTEGER FROM ports WHERE service IS NOT NULL) as unique_services;
END;
$$ LANGUAGE plpgsql;