// Copyright 2021 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teocdbcli

import (
	"testing"
)

func TestPluginBinary(t *testing.T) {
	t.Run("MarshalUnmarshalBinary", func(t *testing.T) {
		id := uint32(2)
		name := "plugin"
		fnc := "F"
		params := []string{"Param1", "Param2"}

		pgInput := &Plugin{id, name, fnc, params, false}
		pgOutput := &Plugin{}

		// Marshal
		data, err := pgInput.MarshalBinary()
		if err != nil {
			t.Error(err)
			return
		}

		// Unmarshal
		if err = pgOutput.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}

		// Compare after Marshal/Unmarshal
		if pgOutput.ID != id ||
			pgOutput.Name != name ||
			pgOutput.Func != fnc ||
			len(pgOutput.Params) == 2 &&
				(pgOutput.Params[0] != params[0] ||
					pgOutput.Params[1] != params[1]) {
			t.Errorf("unmarshalled structure fields values"+
				" not equal to input values:\n%d, '%s', '%s'\n"+
				"params: %v",
				pgOutput.ID, pgOutput.Name, pgOutput.Func, pgOutput.Params,
			)
			return
		}
	})
}
