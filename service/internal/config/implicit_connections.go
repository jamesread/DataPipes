package config

const (
	ImplicitDownloadCSVConnectionID = "download_csv"
)

// ResolveConnection returns a configured or implicit connection by id.
func (c *Config) ResolveConnection(id string) *Connection {
	if id == ImplicitDownloadCSVConnectionID {
		return &Connection{Type: ConnectionTypeDownloadCSV}
	}
	if c == nil || c.Connections == nil {
		return nil
	}
	return c.Connections[id]
}

func IsImplicitConnection(id string) bool {
	return id == ImplicitDownloadCSVConnectionID
}
