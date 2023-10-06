// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Balance is the golang structure for table balance.
type Balance struct {
	Id        int    `json:"id"        description:""`
	Timestamp int    `json:"timestamp" description:""`
	Address   string `json:"address"   description:""`
	TotalUsd  string `json:"totalUsd"  description:""`
}
