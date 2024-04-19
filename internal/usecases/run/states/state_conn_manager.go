package states

import "github.com/go-zookeeper/zk"

// ConnManager saves and retrieves active connection
type ConnManager interface {
	Set(*zk.Conn)
	Get() (*zk.Conn, error)
}
