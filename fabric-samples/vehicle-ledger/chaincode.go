package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Malfunction struct {
	Description string
	RepairCost  float32
}

type VehicleAsset struct {
	ID           string
	Brand        string
	Model        string
	Year         int
	Color        string
	OwnerID      string
	Price        float32
	Malfunctions []Malfunction
}

type PersonAsset struct {
	ID            string
	FirstName     string
	LastName      string
	Email         string
	WalletBalance float32
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	vehicleAssets := []VehicleAsset{
		{ID: "vehicle1", Brand: "Daihatsu", Model: "Charade", Year: 1991, Color: "red", OwnerID: "person1", Price: 1500, Malfunctions: []Malfunction{
			{Description: "Clutch", RepairCost: 200},
			{Description: "Oil leaking", RepairCost: 30},
		}},
		{ID: "vehicle2", Brand: "Fiat", Model: "Panda", Year: 2005, Color: "yellow", OwnerID: "person1", Price: 200, Malfunctions: []Malfunction{
			{Description: "Broken alternator", RepairCost: 50},
			{Description: "Loose exhaust pipe", RepairCost: 10},
			{Description: "Overheating", RepairCost: 80},
		}},
		{ID: "vehicle3", Brand: "Audi", Model: "A4", Year: 2008, Color: "green", OwnerID: "person2", Price: 5000, Malfunctions: []Malfunction{
			{Description: "Flat tyre", RepairCost: 20},
		}},
		{ID: "vehicle4", Brand: "Toyota", Model: "Corolla", Year: 2014, Color: "blue", OwnerID: "person1", Price: 3400, Malfunctions: []Malfunction{
			{Description: "Cracked windscreen", RepairCost: 100},
			{Description: "Loose back wiper", RepairCost: 5},
		}},
		{ID: "vehicle5", Brand: "Mercedes-Benz", Model: "S-class", Year: 2018, Color: "black", OwnerID: "person3", Price: 10000, Malfunctions: []Malfunction{}},
		{ID: "vehicle6", Brand: "BMW", Model: "X5", Year: 2018, Color: "white", OwnerID: "person3", Price: 6000, Malfunctions: []Malfunction{
			{Description: "Cracked headlight", RepairCost: 30},
		}},
	}

	personAssets := []PersonAsset{
		{ID: "person1", FirstName: "Mihajlo", LastName: "Kisic", Email: "mkisic@gmail.com", WalletBalance: 6500.0},
		{ID: "person2", FirstName: "Nebojsa", LastName: "Horvat", Email: "nhorvat@gmail.com", WalletBalance: 7000.0},
		{ID: "person3", FirstName: "Pera", LastName: "Peric", Email: "pperic@gmail.com", WalletBalance: 430.0},
	}

	for _, vehicleAsset := range vehicleAssets {
		vehicleAssetJSON, err := json.Marshal(vehicleAsset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(vehicleAsset.ID, vehicleAssetJSON)
		if err != nil {
			return fmt.Errorf("failed to put vehicles to world state. %v", err)
		}

		indexName := "color~owner~ID"
		colorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{vehicleAsset.Color, vehicleAsset.OwnerID, vehicleAsset.ID})
		if err != nil {
			return err
		}

		value := []byte{0x00}
		err = ctx.GetStub().PutState(colorOwnerIndexKey, value)
		if err != nil {
			return err
		}
	}

	for _, personAsset := range personAssets {
		personAssetJSON, err := json.Marshal(personAsset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(personAsset.ID, personAssetJSON)
		if err != nil {
			return fmt.Errorf("failed to put persons to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartContract) ReadPersonAsset(ctx contractapi.TransactionContextInterface, id string) (*PersonAsset, error) {
	personAssetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read person from world state: %v", err)
	}
	if personAssetJSON == nil {
		return nil, fmt.Errorf("the person asset %s does not exist", id)
	}

	var personAsset PersonAsset
	err = json.Unmarshal(personAssetJSON, &personAsset)
	if err != nil {
		return nil, err
	}

	return &personAsset, nil
}

func (s *SmartContract) ReadVehicleAsset(ctx contractapi.TransactionContextInterface, id string) (*VehicleAsset, error) {
	vehicleAssetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read vehicle from world state: %v", err)
	}
	if vehicleAssetJSON == nil {
		return nil, fmt.Errorf("the vehicle asset %s does not exist", id)
	}

	var vehicleAsset VehicleAsset
	err = json.Unmarshal(vehicleAssetJSON, &vehicleAsset)
	if err != nil {
		return nil, err
	}

	return &vehicleAsset, nil
}

func (s *SmartContract) GetVehiclesByColor(ctx contractapi.TransactionContextInterface, color string) ([]*VehicleAsset, error) {
	coloredVehicleIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~ID", []string{color})
	if err != nil {
		return nil, err
	}

	defer coloredVehicleIter.Close()

	retList := make([]*VehicleAsset, 0)

	for i := 0; coloredVehicleIter.HasNext(); i++ {
		responseRange, err := coloredVehicleIter.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		retVehicleID := compositeKeyParts[2]

		vehicleAsset, err := s.ReadVehicleAsset(ctx, retVehicleID)
		if err != nil {
			return nil, err
		}

		retList = append(retList, vehicleAsset)
	}

	return retList, nil
}

func (s *SmartContract) GetVehiclesByColorAndOwner(ctx contractapi.TransactionContextInterface, color string, ownerID string) ([]*VehicleAsset, error) {
	exists, err := s.PersonAssetExists(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("the person %v does not exist", ownerID)
	}

	coloredVehicleByOwnerIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~ID", []string{color, ownerID})
	if err != nil {
		return nil, err
	}

	defer coloredVehicleByOwnerIter.Close()

	retList := make([]*VehicleAsset, 0)

	for i := 0; coloredVehicleByOwnerIter.HasNext(); i++ {
		responseRange, err := coloredVehicleByOwnerIter.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		retVehicleID := compositeKeyParts[2]

		vehicleAsset, err := s.ReadVehicleAsset(ctx, retVehicleID)
		if err != nil {
			return nil, err
		}

		retList = append(retList, vehicleAsset)
	}

	return retList, nil
}

func (s *SmartContract) TransferVehicleAsset(ctx contractapi.TransactionContextInterface, id string, newOwnerID string, acceptMalfunction bool) (bool, error) {
	vehicleAsset, err := s.ReadVehicleAsset(ctx, id)
	if err != nil {
		return false, err
	}

	if vehicleAsset.OwnerID == newOwnerID {
		return false, fmt.Errorf("person %s is already the owner of the vehicle", newOwnerID)
	}

	buyer, err := s.ReadPersonAsset(ctx, newOwnerID)
	if err != nil {
		return false, err
	}

	seller, err := s.ReadPersonAsset(ctx, vehicleAsset.OwnerID)
	if err != nil {
		return false, err
	}

	vehiclePrice := float32(0)

	if vehicleAsset.Malfunctions == nil || len(vehicleAsset.Malfunctions) == 0 {
		vehiclePrice = vehicleAsset.Price
	} else if acceptMalfunction {
		malfuctionPrice := float32(0)
		for _, malfunction := range vehicleAsset.Malfunctions {
			malfuctionPrice += malfunction.RepairCost
		}
		vehiclePrice = vehicleAsset.Price - malfuctionPrice
	} else {
		return false, fmt.Errorf("the buyer will not accept a malfunctioned vehicle")
	}

	oldOwnerID := vehicleAsset.OwnerID
	vehicleAsset.OwnerID = newOwnerID

	if buyer.WalletBalance >= vehiclePrice {
		buyer.WalletBalance -= vehiclePrice
		seller.WalletBalance += vehiclePrice
	} else {
		return false, fmt.Errorf("the buyer does not own enough money to purchase the vehicle")
	}

	vehicleAssetJSON, err := json.Marshal(vehicleAsset)
	if err != nil {
		return false, err
	}

	buyerJSON, err := json.Marshal(buyer)
	if err != nil {
		return false, err
	}

	sellerJSON, err := json.Marshal(seller)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(id, vehicleAssetJSON)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(buyer.ID, buyerJSON)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(seller.ID, sellerJSON)
	if err != nil {
		return false, err
	}

	indexName := "color~owner~ID"
	colorNewOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{vehicleAsset.Color, newOwnerID, vehicleAsset.ID})
	if err != nil {
		return false, err
	}

	value := []byte{0x00}
	err = ctx.GetStub().PutState(colorNewOwnerIndexKey, value)
	if err != nil {
		return false, err
	}

	colorOldOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{vehicleAsset.Color, oldOwnerID, vehicleAsset.ID})
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().DelState(colorOldOwnerIndexKey)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SmartContract) AddMalfunction(ctx contractapi.TransactionContextInterface, id string, description string, repairCost float32) error {
	vehicleAsset, err := s.ReadVehicleAsset(ctx, id)
	if err != nil {
		return err
	}

	newMalfunction := Malfunction{
		Description: description,
		RepairCost:  repairCost,
	}

	vehicleAsset.Malfunctions = append(vehicleAsset.Malfunctions, newMalfunction)

	totalRepairCost := float32(0)
	for _, malfunction := range vehicleAsset.Malfunctions {
		totalRepairCost += malfunction.RepairCost
	}

	if totalRepairCost > vehicleAsset.Price {
		return ctx.GetStub().DelState(id)
	}

	vehicleAssetJSON, err := json.Marshal(vehicleAsset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, vehicleAssetJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) ChangeVehicleColor(ctx contractapi.TransactionContextInterface, id string, newColor string) (string, error) {
	vehicleAsset, err := s.ReadVehicleAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldColor := vehicleAsset.Color
	vehicleAsset.Color = newColor

	vehicleAssetJSON, err := json.Marshal(vehicleAsset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, vehicleAssetJSON)
	if err != nil {
		return "", err
	}

	indexName := "color~owner~ID"
	newColorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{newColor, vehicleAsset.OwnerID, vehicleAsset.ID})
	if err != nil {
		return "", err
	}

	value := []byte{0x00}
	err = ctx.GetStub().PutState(newColorOwnerIndexKey, value)
	if err != nil {
		return "", err
	}

	oldColorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{oldColor, vehicleAsset.OwnerID, vehicleAsset.ID})
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().DelState(oldColorOwnerIndexKey)
	if err != nil {
		return "", err
	}

	return oldColor, nil
}

func (s *SmartContract) RepairVehicle(ctx contractapi.TransactionContextInterface, id string) error {
	vehicleAsset, err := s.ReadVehicleAsset(ctx, id)
	if err != nil {
		return err
	}

	personAsset, err := s.ReadPersonAsset(ctx, vehicleAsset.OwnerID)
	if err != nil {
		return err
	}

	repairCostSum := float32(0)
	for _, malfunction := range vehicleAsset.Malfunctions {
		repairCostSum += malfunction.RepairCost
		if repairCostSum > personAsset.WalletBalance {
			return fmt.Errorf("the owner of the vehicle cannot afford to pay the vehicle repair price")
		}
	}

	vehicleAsset.Malfunctions = []Malfunction{}
	personAsset.WalletBalance -= repairCostSum

	vehicleAssetJSON, err := json.Marshal(vehicleAsset)
	if err != nil {
		return err
	}

	personAssetJSON, err := json.Marshal(personAsset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, vehicleAssetJSON)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(personAsset.ID, personAssetJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) PersonAssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	personAssetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read person asset from world state: %v", err)
	}

	return personAssetJSON != nil, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating vehicles-and-persons chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting vehicles-and-persons chaincode: %v", err)
	}
}
