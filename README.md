
# 🥪 MEV Snipping Copytrader Go

Snipe new tokens bought by the most powerful MEV bots (Sandwitch, Arbitrage, Snipping) in Go on EVM networks.

This bot scans contracts such as MEV bots (Sandwitch, Arbitrage, Snipping) and searches for new ERC-20 tokens purchased, this bot allows you to take advantage of the powerful checks of MEV bots (Poison tokens, Salmonella tokens etc) to discover safer tokens.

### Config
- WssRpcNode (string) - Url of your WSS RPC Node. 
- SandwichBotEOAAddress (string) - Address of MEV Bot caller.
- SandwitchBotContractAddress (string) - Address of MEV Contract.

### Install
Requirements : Go.
- `go get`
- `go run .`

### Twitter
@deltartificial
