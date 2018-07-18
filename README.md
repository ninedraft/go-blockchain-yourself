# go-blockchain-yourself


Changes:
- devnetwork:
    1. [docker-compose-simple.yaml](devnetwork/chaincode-docker-devmode/docker-compose-simple.yaml):
        + **peer** service: [fix port from 7051 to 7052](https://stackoverflow.com/questions/48007519/unimplemented-desc-unknown-service-protos-chaincodesupport/48020091#48020091)
        + **cli** service: change volume mount  ```./../chaincode:/opt/gopath/src/chaincodedev/chaincode -> ./../../exchanger:/opt/gopath/src/chaincodedev/exchanger```
        + **chaincode** service: change volume mount: ```./../chaincode:/opt/gopath/src/chaincode -> ./../../exchanger/build/:/opt/gopath/bin/chaincode ```
