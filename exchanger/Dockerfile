FROM hyperledger/fabric-tools:1.2.0 as fabric-tools
CMD peer chaincode install -p chaincodedev/exchanger/cmd/exchanger -n exchanger -v 0 && \
    peer chaincode instantiate -n exchanger -v 0 -c '{"Args":["set", "a","10"]}' -C myc