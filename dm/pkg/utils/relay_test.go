// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"os"
	"strings"
	"testing"

	"github.com/go-mysql-org/go-mysql/replication"
	. "github.com/pingcap/check"
)

var _ = Suite(&testUtilsSuite{})

func TestSuite(t *testing.T) {
	TestingT(t)
}

type testUtilsSuite struct{}

func (t *testUtilsSuite) TestParseUUIDIndex(c *C) {
	f, err := os.CreateTemp("", "server-uuid.index")
	c.Assert(err, IsNil)
	defer os.Remove(f.Name())

	uuids := []string{
		"c65525fa-c7a3-11e8-a878-0242ac130005.000001",
		"c6ae5afe-c7a3-11e8-a19d-0242ac130006.000002",
		"c65525fa-c7a3-11e8-a878-0242ac130005.000003",
	}

	err = os.WriteFile(f.Name(), []byte(strings.Join(uuids, "\n")), 0o644)
	c.Assert(err, IsNil)

	obtainedUUIDs, err := ParseUUIDIndex(f.Name())
	c.Assert(err, IsNil)
	c.Assert(obtainedUUIDs, DeepEquals, uuids)

	// test GetUUIDBySuffix
	uuid := uuids[1]
	uuidWS := GetUUIDBySuffix(uuids, uuid[len(uuid)-6:])
	c.Assert(uuidWS, Equals, uuid)

	uuidWS = GetUUIDBySuffix(uuids, "100000")
	c.Assert(uuidWS, Equals, "")
}

func (t *testUtilsSuite) TestSuffixForUUID(c *C) {
	cases := []struct {
		uuid           string
		ID             int
		uuidWithSuffix string
	}{
		{"c65525fa-c7a3-11e8-a878-0242ac130005", 1, "c65525fa-c7a3-11e8-a878-0242ac130005.000001"},
		{"c6ae5afe-c7a3-11e8-a19d-0242ac130006", 2, "c6ae5afe-c7a3-11e8-a19d-0242ac130006.000002"},
	}

	for _, cs := range cases {
		uuidWS := AddSuffixForUUID(cs.uuid, cs.ID)
		c.Assert(uuidWS, Equals, cs.uuidWithSuffix)

		uuidWOS, id, err := ParseRelaySubDir(cs.uuidWithSuffix)
		c.Assert(err, IsNil)
		c.Assert(uuidWOS, Equals, cs.uuid)
		c.Assert(id, Equals, cs.ID)

		suffix := SuffixIntToStr(cs.ID)
		hasSuffix := strings.HasSuffix(cs.uuidWithSuffix, suffix)
		c.Assert(hasSuffix, Equals, true)
	}

	_, _, err := ParseRelaySubDir("uuid-with-out-suffix")
	c.Assert(err, NotNil)

	_, _, err = ParseRelaySubDir("uuid-invalid-suffix-len.01")
	c.Assert(err, NotNil)

	_, _, err = ParseRelaySubDir("uuid-invalid-suffix-fmt.abc")
	c.Assert(err, NotNil)

	_, _, err = ParseRelaySubDir("uuid-invalid-fmt.abc.000001")
	c.Assert(err, NotNil)
}

func (t *testUtilsSuite) TestGenFakeRotateEvent(c *C) {
	var (
		nextLogName = "mysql-bin.000123"
		logPos      = uint64(456)
		serverID    = uint32(101)
	)

	ev, err := GenFakeRotateEvent(nextLogName, logPos, serverID)
	c.Assert(err, IsNil)
	c.Assert(ev.Header.ServerID, Equals, serverID)
	c.Assert(ev.Header.Timestamp, Equals, uint32(0))
	c.Assert(ev.Header.LogPos, Equals, uint32(0))
	c.Assert(ev.Header.EventType, Equals, replication.ROTATE_EVENT)

	evR, ok := ev.Event.(*replication.RotateEvent)
	c.Assert(ok, IsTrue)
	c.Assert(string(evR.NextLogName), Equals, nextLogName)
	c.Assert(evR.Position, Equals, logPos)
}
