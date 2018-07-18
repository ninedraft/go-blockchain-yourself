package main

import (
	"fmt"
	"strconv"
	"time"

	"net/http"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// Chaincode implements a simple chaincode to manage an asset
type Chaincode struct {
	*BPI
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (chaincode *Chaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	chaincode.BPI = NewBPI("https://api.coindesk.com/v1/bpi/currentprice.json", 10*time.Second)

	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	go func() {
		var server = http.NewServeMux()
		server.HandleFunc("/bpi", func(writer http.ResponseWriter, request *http.Request) {
			var price, err = chaincode.GetPrice()
			if err != nil {
				fmt.Fprint(writer, err.Error())
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(writer, "%v", price.Bpi)
		})
		http.ListenAndServe(":8090", server)
	}()

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (chaincode *Chaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	switch fn {
	case "exchange": // echange addr currency coins
		if len(args) != 3 {
			return shim.Error(fmt.Sprintf("expected 3 args (addr currency quantity), got %d", len(args)))
		}

		var addr, currencyName, quatityStr = args[0], args[1], args[2]
		var quantity, errParseQuantity = strconv.ParseUint(quatityStr, 10, 64)

		if errParseQuantity == nil {
			return shim.Error(errParseQuantity.Error())
		}

		var price, errGetPrice = chaincode.GetPrice()
		if errGetPrice != nil {
			return shim.Error(errGetPrice.Error())
		}

		var marshalledWallet, errGetWallet = stub.GetState(addr)
		if errGetWallet != nil {
			return shim.Error(errGetWallet.Error())
		}

		var wallet, errParseWallet = strconv.ParseUint(string(marshalledWallet), 10, 64)
		if errParseWallet != nil {
			return shim.Error(errGetWallet.Error())
		}

		var currency, exists = price.Bpi[currencyName]
		if !exists {
			return shim.Error(fmt.Sprintf("currency %q not found", currencyName))
		}

		wallet = uint64(currency.RateFloat) * quantity

		var errPutState = stub.PutState(addr, []byte(strconv.FormatUint(wallet, 10)))
		if errPutState != nil {
			return shim.Error(errPutState.Error())
		}
		return shim.Success([]byte(fmt.Sprintf("wallet: %d", wallet)))
	case "get":

		if len(args) != 1 {
			return shim.Error(fmt.Sprintf("expected 1 args (addr), got %d", len(args)))
		}

		var addr = args[0]

		var marshalledWallet, errGetWallet = stub.GetState(addr)
		if errGetWallet != nil {
			return shim.Error(errGetWallet.Error())
		}

		var wallet, errParseWallet = strconv.ParseUint(string(marshalledWallet), 10, 64)
		if errParseWallet != nil {
			return shim.Error(errGetWallet.Error())
		}

		return shim.Success([]byte(fmt.Sprintf("wallet: %d", wallet)))
	default:
		return shim.Error(fmt.Sprintf("unknown command %q", fn))
	}
	// Return the result as success payload

}

// main function starts up the chaincode in the container during instantiate
func main() {
	shim.SetLoggingLevel(shim.LogDebug)
	if err := shim.Start(new(Chaincode)); err != nil {
		fmt.Printf("Error starting Chaincode chaincode: %s", err)
	}
}
