// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Price is the golang structure of table price for DAO operations like Where/Data.
type Price struct {
	g.Meta    `orm:"table:price, do:true"`
	Id        interface{} //
	CoinId    interface{} //
	Price     interface{} //
	Timestamp interface{} //
}
