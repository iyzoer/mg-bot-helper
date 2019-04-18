package main

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// Connection model
type Connection struct {
	ID           int    `gorm:"primary_key"`
	ClientID     string `gorm:"client_id type:varchar(70);not null;unique" json:"clientId,omitempty"`
	APIKEY       string `gorm:"api_key type:varchar(100);not null" json:"api_key,omitempty" binding:"required"`
	APIURL       string `gorm:"api_url type:varchar(255);not null" json:"api_url,omitempty" binding:"required,validatecrmurl"`
	MGURL        string `gorm:"mg_url type:varchar(255);not null;" json:"mg_url,omitempty"`
	MGToken      string `gorm:"mg_token type:varchar(100);not null;unique" json:"mg_token,omitempty"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Active       bool           `json:"active,omitempty"`
	Commands     postgres.Jsonb `gorm:"commands type:jsonb;" json:"commands,omitempty"`
	Lang         string         `gorm:"lang type:varchar(2)" json:"lang,omitempty"`
	Currency     string         `gorm:"currency type:varchar(12)" json:"currency,omitempty"`
	CommandsHash string         `gorm:"commands_hash type:varchar(70)" json:"commands_hash,omitempty"`
	Timezone     string         `gorm:"timezone type:varchar(50)" json:"timezone,omitempty"`
}
