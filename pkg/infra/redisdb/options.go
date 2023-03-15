package redisdb

// Options provide options for the redis client
type Options struct {
	Address     string
	Password    string
	ServiceName string
	TLS         bool
	Database    int
}
