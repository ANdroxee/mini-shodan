package models

import "gorm.io/gorm"


type Host struct {
	gorm.Model           edAt
	IP        string     `gorm:"uniqueIndex"` 
	Hostname  string
	OS        string
	MacAddr   string
	Vendor    string
	Ports     []Port     `gorm:"foreignKey:HostID"` 
}


type Port struct {
	gorm.Model
	HostID    uint
	Number    int
	Protocol  string    
	Service   string     
	Version   string    
	Banner    string     
}