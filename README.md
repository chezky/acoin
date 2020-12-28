# ACOINS
### A cryptocurrency written in golang

Acoins is an open source cryptocurrency written purely in golang. All help is greatly appreciated.

### Documentation / Crypto Explanation
 Check out the [docs](https://godoc.org/github.com/chezky/acoin) for a full crypto explanation.


## Installation

### Requirements

 - Download [Golang](https://golang.org/dl/) that's suitable for your OS

### Running
    go build main.go
    main.exe [argument]
    - Create chain
        * main.exe createchain -address {addresss}
            i.e: main.exe createchain -address kevin
    - Print chain
        * main.exe printchain
    - Get Balance
        * main.exe getbalance -address {address}
            i.e: main.exe getbalance -address kevin
    - Send / Transfer
        * main.exe send -to {to} -from {from} -amount {amount}
            i.e: main.exe send -to kevin -from dave -amount 5 

## License
[MIT](https://choosealicense.com/licenses/mit/)
