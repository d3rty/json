package dirty

// UnmarshalOld is the main Unmarshal function
// Legacy: is kept for reference
// func UnmarshalOld(data []byte, clean any) error {
// 	// Green phase: we could convert directly to clean
// 	var err error
// 	if err = json.Unmarshal(data, clean); err == nil {
// 		return nil
// 	}

// 	// Before starting dirty unmarshalling let's ensure user manually didn't disable dirtying
// 	// If yes - simply return original json error.
// 	if _, ok := clean.(interface{ isDisabled() }); ok {
// 		return err
// 	}

// 	// Let's ensure user configured everything correctly for dirty unmarshalling
// 	// If not - that was a regular unmarshalling, simply return original json error.

// 	schemer, ok := clean.(Dirtyable)
// 	if !ok {
// 		return err
// 	}
// 	container, ok := clean.(d3rtyContainer)
// 	if !ok {
// 		return err
// 	}

// 	// Yellow phase: try unmarshal into dirty model

// 	scheme := schemer.Dirty()
// 	container.init(scheme)

// 	if err := container.(d3rtyMarker).unmarshal(data); err != nil {
// 		// RED Phase: we couldn't unmarshal even in dirty model
// 		return errors.New("red:to be implemented 2")
// 	}

// 	// Here comes the slow part, refactor it so it's fast
// 	// Currently for idea purpose that's OK
// 	buffer, err := json.Marshal(scheme)
// 	if err != nil {
// 		return errors.New("err1")
// 	}
// 	if err := json.Unmarshal(buffer, clean); err != nil {
// 		return errors.New("err2")
// 	}

// 	// Yellow Phase: OK: converting from dirty into clean model
// 	return nil
// }
