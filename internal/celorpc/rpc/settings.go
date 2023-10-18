package rpc

// Your Infura API key and the Celo RPC URL
const ApiKey = "76891b8517e248fe9a49473d68f8f7f7"
const RpcURL = "https://mainnet.infura.io/v3/" + ApiKey

// Celo RPC config
var CeloRpcConfig = &RpcConfig{
	HttpUrl: RpcURL,
	User:    "",
	Pass:    "",
}
