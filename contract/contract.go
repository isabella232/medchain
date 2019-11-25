package contract

import (
	"errors"

	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

// The query contract can simply store a query in an instance.

// MedchainContractID denotes a contract that can store and update
// key/value pairs corresponding to queries. Key is the query ID
// and value is the query itself (i.e., it is the concatenation
// of query/status/user)
var MedchainContractID = "queryContract"

type medchainContract struct {
	byzcoin.BasicContract
	QueryData
}

func contractValueFromBytes(in []byte) (byzcoin.Contract, error) {
	cv := &medchainContract{}
	err := protobuf.Decode(in, &cv.QueryData)
	if err != nil {
		return nil, err
	}
	return cv, nil
}

// medchianContract implments the main logic of medchian node
// It is a key/value store type contract that manipulates queries
// received from the client (e.g., medco-connector) and writes to
// Byzcoin "instances".
// This contract implements 2 main functionalities:
// (1) Spawn new key-value instances of queries and store all the arguments in the data field.
// (2) Update existing key-value instances.
func (c *medchainContract) Spawn(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins

	// Find the darcID for this instance
	var darcID darc.ID
	_, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to get the darc ID: %v", err)
	}

	// Put the data received from the inst.Spawn.Args into our QueryData structure.
	cs := &c.QueryData
	for _, kv := range inst.Spawn.Args {
		cs.Storage = append(cs.Storage, Query{kv.Name, kv.Value})
	}

	// Encode the data into our QueryDataStorage structure that holds all the key-value pairs
	csBuf, err := protobuf.Encode(&c.QueryData)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to encode QueryDataStorage: %v", err)
	}

	// Then create a StateChange request with the data of the instance. The
	// InstanceID is given by the DeriveID method of the instruction that allows
	// to create multiple instanceIDs out of a given instruction in a pseudo-
	// random way that will be the same for all nodes.
	sc = []byzcoin.StateChange{
		byzcoin.NewStateChange(byzcoin.Create, inst.DeriveID(""), MedchainContractID, csBuf, darcID),
	}
	return sc, cout, nil
}

func (c *medchainContract) Invoke(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins
	var darcID darc.ID
	_, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}

	if inst.Invoke.Command != "update" {
		return nil, nil, errors.New("MedChain contract only supports spwan/update requests")
	}
	// The only command we can invoke is 'update' which will store the new values
	// given in the arguments in the data.
	//  1. decode the existing data
	//  2. update the data
	//  3. encode the data into protobuf again

	kvd := &c.QueryData
	kvd.Update(inst.Invoke.Args)
	var buf []byte
	buf, err = protobuf.Encode(kvd)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to encode data with error : %v", err)
	}
	sc = []byzcoin.StateChange{
		byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
			MedchainContractID, buf, darcID),
	}
	return
}

// Update goes through all the arguments and:
//  - updates the value if the key already exists
//  - deletes the key-value pair if the value is empty (??)
//  - adds a new key-value pair if the key does not exist yet
func (cs *QueryData) Update(args byzcoin.Arguments) {
	for _, kv := range args {
		var updated bool
		for i, stored := range cs.Storage {
			if stored.ID == kv.Name {
				updated = true
				if kv.Value == nil || len(kv.Value) == 0 {
					cs.Storage = append(cs.Storage[0:i], cs.Storage[i+1:]...)
					break
				}
				cs.Storage[i].Value = kv.Value
			}

		}
		if !updated {
			cs.Storage = append(cs.Storage, Query{kv.Name, kv.Value})
		}
	}
}
