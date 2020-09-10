# timed-transaction-broadcast
Gas prices are crazy high on ethereum, however if you try to submit a transaction with a gas price lower than the considered _safe__low_ it is likely to be rejected by the node and never propogate.  This application lets you sign a transaction with a specific gas price and then transmit that pre-signed transaction when the gas prices lower, automatically.  This allows you to take advantage of things like overnight drops in gas prices to get transaction pushed through.  You can even sign multiple transaction and batch them (remember to increase your nonce for each of them).

# Installation
```
install golang
clone repo
cd repo directory
go get
go build
```
