// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Balance is the golang structure of table balance for DAO operations like Where/Data.
type Balance struct {
	g.Meta    `orm:"table:balance, do:true"`
	Id        interface{} //
	Timestamp interface{} //
	Address   interface{} //
	TotalUsd  interface{} //
}
