package model

import "fmt"

type NodeID string

type NodeInfo struct {
	ID      NodeID
	Address string
}

func (n NodeInfo) String() string {
	return fmt.Sprintf("%s(%s)", n.ID, n.Address)
}

func ParseNodeInfo(addr string) NodeInfo {
	return NodeInfo{Address: addr}
}
