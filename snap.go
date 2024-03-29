package segmago

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

const currentSnapshotVersion = 1

const infoString = "segmago snapshot"

type snapshot struct {
	Version     int
	Info        string
	State       json.RawMessage
	SelectedMem int
}

func (emu *emuState) loadSnapshot(snapBytes []byte) (*emuState, error) {
	var err error
	var reader io.Reader
	var unpackedBytes []byte
	var snap snapshot
	if reader, err = gzip.NewReader(bytes.NewReader(snapBytes)); err != nil {
		return nil, err
	} else if unpackedBytes, err = ioutil.ReadAll(reader); err != nil {
		return nil, err
	} else if err = json.Unmarshal(unpackedBytes, &snap); err != nil {
		return nil, err
	} else if snap.Version < currentSnapshotVersion {
		return emu.convertOldSnapshot(&snap)
	} else if snap.Version > currentSnapshotVersion {
		return nil, fmt.Errorf("this version of segmago is too old to open this snapshot")
	}

	return emu.convertLatestSnapshot(&snap)
}

func (emu *emuState) convertLatestSnapshot(snap *snapshot) (*emuState, error) {

	var err error
	var newState emuState

	if err = json.Unmarshal(snap.State, &newState); err != nil {
		return nil, err
	}

	newState.Mem.unmarshallSelectedMem(snap.SelectedMem)

	newState.Mem.CartStorage.rom = emu.Mem.CartStorage.rom
	newState.Mem.BIOSStorage.rom = emu.Mem.BIOSStorage.rom
	newState.Mem.NullStorage.rom = emu.Mem.NullStorage.rom

	newState.CPU.Write = newState.write
	newState.CPU.Read = newState.read
	newState.CPU.In = newState.in
	newState.CPU.Out = newState.out
	newState.CPU.RunCycles = newState.runCycles

	newState.devMode = emu.devMode

	return &newState, nil
}

var snapshotConverters = map[int]func(map[string]interface{}) error{

	// If new field can be zero, no need for converter.

	// Converters should look like this (including comment):
	// added 2017-XX-XX
	// 1: convertSnap0To1,
}

func (emu *emuState) convertOldSnapshot(snap *snapshot) (*emuState, error) {

	var state map[string]interface{}
	if err := json.Unmarshal(snap.State, &state); err != nil {
		return nil, fmt.Errorf("json unpack err: %v", err)
	}

	for i := snap.Version; i < currentSnapshotVersion; i++ {
		if converterFn, ok := snapshotConverters[snap.Version]; !ok {
			return nil, fmt.Errorf("unknown snapshot version: %v", snap.Version)
		} else if err := converterFn(state); err != nil {
			return nil, fmt.Errorf("error converting snapshot version %v: %v", i, err)
		}
	}

	var err error
	if snap.State, err = json.Marshal(state); err != nil {
		return nil, fmt.Errorf("json pack err: %v", err)
	}

	return emu.convertLatestSnapshot(snap)
}

func (emu *emuState) makeSnapshot() []byte {
	var err error
	var emuJSON []byte
	var snapJSON []byte
	if emuJSON, err = json.Marshal(emu); err != nil {
		panic(err)
	}
	snap := snapshot{
		Version:     currentSnapshotVersion,
		Info:        infoString,
		State:       json.RawMessage(emuJSON),
		SelectedMem: emu.Mem.marshallSelectedMem(),
	}
	if snapJSON, err = json.Marshal(&snap); err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	writer := gzip.NewWriter(buf)
	writer.Write(snapJSON)
	writer.Close()
	return buf.Bytes()
}
