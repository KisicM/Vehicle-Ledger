# Vehicle-Ledger
Blockchain-based system for managing vehicle ownership and information.

This blockchain-based vehicle ownership management network is powered by Hyperledger Fabric and GO SDK. The network utilizes Hyperledger Fabric, an enterprise-grade blockchain platform, to provide a scalable and secure solution for managing vehicle ownership data.

With the GO SDK, the network has developed a user-friendly application for vehicle owners to register and transfer ownership of their vehicles, as well as for potential buyers to verify ownership status before making a purchase.

By leveraging the power of blockchain technology, the network offers a transparent and immutable record of ownership history, ensuring that all parties involved in a transaction can trust the authenticity of the data. The network provides a secure and efficient solution for managing vehicle ownership that is revolutionizing the industry.

## Setup

Navigate to test-network directory. Then execute the following commands:

```bash
./network.sh up
./network.sh createChannel
```

Configure environment variables:

```bash
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/configtx/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
```

Deploy and interact with chaincode:

```bash
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" --peerAddresses localhost:11051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt" --peerAddresses localhost:12051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org4.example.com/peers/peer0.org4.example.com/tls/ca.crt" -c '{"function":"InitLedger","Args":[]}'

peer chaincode query -C mychannel -n basic -c '{"Args":["GetAllAssets"]}'
```
