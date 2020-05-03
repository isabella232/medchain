package admin

import (
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

var ContractAccessRightID = "accessright"

type ContractAccessRight struct {
	byzcoin.BasicContract
	AccessRight
}

func contractAccessRightFromBytes(in []byte) (byzcoin.Contract, error) {
	cv := &ContractAccessRight{}
	err := protobuf.Decode(in, &cv.AccessRight)
	if err != nil {
		return nil, err
	}
	return cv, nil
}

// Spawn implements the byzcoin.Contract interface
func (c *ContractAccessRight) Spawn(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins
	// Find the darcID for this instance.
	var darcID darc.ID
	_, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}

	sc = []byzcoin.StateChange{
		byzcoin.NewStateChange(byzcoin.Create, inst.DeriveID(""),
			ContractAccessRightID, inst.Spawn.Args.Search("ar"), darcID),
	}
	return
}

// Invoke implements the byzcoin.Contract interface
func (c *ContractAccessRight) Invoke(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins

	// Find the darcID for this instance.
	var darcID darc.ID

	v, _, _, darcID, err := rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}

	switch inst.Invoke.Command {
	case "add":
		nid := string(inst.Invoke.Args.Search("id"))
		nar := string(inst.Invoke.Args.Search("ar"))
		ar := AccessRight{}
		err = protobuf.Decode(v, &ar)
		if err != nil {
			return nil, nil, xerrors.Errorf("Decoding %w", err)
		}
		idx, _ := Find(ar.Ids, nid)
		if idx != -1 {
			return nil, nil, xerrors.New("The id is already registered")
		}
		ar.Access = append(ar.Access, nar)
		ar.Ids = append(ar.Ids, nid)
		buf, err2 := protobuf.Encode(&ar)
		if err != nil {
			return nil, nil, xerrors.Errorf("Encoding the access right struct: %w", err2)
		}
		sc = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractAccessRightID, buf, darcID),
		}
		return

	case "update":
		nid := string(inst.Invoke.Args.Search("id"))
		nar := string(inst.Invoke.Args.Search("ar"))
		ar := AccessRight{}
		err = protobuf.Decode(v, &ar)
		if err != nil {
			return nil, nil, xerrors.Errorf("Decoding %w", err)
		}
		idx, _ := Find(ar.Ids, nid)
		if idx == -1 {
			return nil, nil, xerrors.New("There is no such value")
		}
		ar.Access[idx] = nar
		buf, err2 := protobuf.Encode(&ar)
		if err != nil {
			return nil, nil, xerrors.Errorf("Encoding the access right struct: %w", err2)
		}
		sc = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractAccessRightID, buf, darcID),
		}
		return

	case "delete":
		nid := string(inst.Invoke.Args.Search("id"))
		ar := AccessRight{}
		err = protobuf.Decode(v, &ar)
		if err != nil {
			return nil, nil, xerrors.Errorf("Decoding %w", err)
		}
		idx, _ := Find(ar.Ids, nid)
		if idx == -1 {
			return nil, nil, xerrors.New("There is no such value")
		}
		ar.Access = append(ar.Access[:idx], ar.Access[idx+1:]...)
		ar.Ids = append(ar.Ids[:idx], ar.Ids[idx+1:]...)
		buf, err2 := protobuf.Encode(&ar)
		if err != nil {
			return nil, nil, xerrors.Errorf("Encoding the access right struct: %w", err2)
		}
		sc = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractAccessRightID, buf, darcID),
		}
		return

	default:
		return nil, nil, xerrors.New("AccessRight contract can only update, delete or modify the access")
	}
}

// Delete implements the byzcoin.Contract interface
func (c *ContractAccessRight) Delete(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins

	// Find the darcID for this instance.
	var darcID darc.ID
	_, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}

	sc = byzcoin.StateChanges{
		byzcoin.NewStateChange(byzcoin.Remove, inst.InstanceID, ContractAccessRightID, nil, darcID),
	}
	return
}

// VerifyDeferredInstruction implements the byzcoin.Contract interface
func (c *ContractAccessRight) VerifyDeferredInstruction(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, ctxHash []byte) error {
	return inst.VerifyWithOption(rst, ctxHash, &byzcoin.VerificationOptions{IgnoreCounters: true})
}
