package main

import (
	"regexp"
)

var rx = regexp.MustCompile(`/+$`)

func getConnection(uid string) *Connection {
	var connection Connection
	orm.DB.First(&connection, "client_id = ?", uid)

	return &connection
}

func getConnectionByURL(urlCrm string) *Connection {
	var connection Connection
	orm.DB.First(&connection, "api_url = ?", urlCrm)

	return &connection
}

func getActiveConnection() []*Connection {
	var connection []*Connection
	orm.DB.Find(&connection, "active = ?", true)

	return connection
}

func getConnections() []*Connection {
	var connection []*Connection
	orm.DB.Find(&connection)

	return connection
}

func (c *Connection) setConnectionActivity() error {
	return orm.DB.Model(c).Where("client_id = ?", c.ClientID).Updates(map[string]interface{}{"active": c.Active, "api_url": c.APIURL}).Error
}

func (c *Connection) createConnection() error {
	return orm.DB.Create(c).Error
}

func (c *Connection) saveConnection() error {
	return orm.DB.Model(c).Where("client_id = ?", c.ClientID).Update(c).Error
}

func (c *Connection) NormalizeApiUrl() {
	c.APIURL = rx.ReplaceAllString(c.APIURL, ``)
}
