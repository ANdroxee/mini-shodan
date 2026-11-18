-- Shodan@Home — schéma de la base (appliqué automatiquement au démarrage, idempotent)

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
-- Nom donné par l'utilisateur, et « je connais cet appareil »
ALTER TABLE devices ADD COLUMN IF NOT EXISTS name VARCHAR(100);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS known BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_devices_ip ON devices(ip);
CREATE INDEX IF NOT EXISTS idx_devices_last_seen ON devices(last_seen);
CREATE INDEX IF NOT EXISTS idx_devices_vendor ON devices(vendor);

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
CREATE INDEX IF NOT EXISTS idx_ports_device_id ON ports(device_id);
CREATE INDEX IF NOT EXISTS idx_ports_port ON ports(port);
CREATE INDEX IF NOT EXISTS idx_ports_service ON ports(service);

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
CREATE INDEX IF NOT EXISTS idx_scans_start_time ON scans(start_time);

-- Demandes de scan à la demande (bouton « Vérifier maintenant »), traitées par le scheduler
CREATE TABLE IF NOT EXISTS scan_requests (
    id SERIAL PRIMARY KEY,
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    picked_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    device_id INTEGER REFERENCES devices(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    read BOOLEAN DEFAULT FALSE
);
-- Port concerné, détails techniques (preuve), et résolution automatique quand le problème disparaît
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS port INTEGER;
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS details TEXT;
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS resolved BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_alerts_read ON alerts(read);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_alerts_device ON alerts(device_id);
