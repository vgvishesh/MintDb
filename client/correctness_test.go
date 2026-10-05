package client

import "strings"

type op struct {
	del   bool
	key   string
	value string
}

func setOp(key, value string) op { return op{key: key, value: value} }
func delOp(key string) op         { return op{del: true, key: key} }

type persistenceCase struct {
	name string
	ops  []op
	want map[string]string // keys that must exist with exactly this value
	gone []string          // keys that must not exist
}

// Every case is valid client input, so the in-memory state is correct. The
// question is whether replaying mint.aof after a crash rebuilds the same state.
var persistenceCases = []persistenceCase{
	{
		// Control case: plain keys and values must survive a restart.
		name: "plain keys and values",
		ops:  []op{setOp("k1", "v1"), setOp("k2", "v2"), delOp("k1")},
		want: map[string]string{"k2": "v2"},
		gone: []string{"k1"},
	},
	{
		// Records are "key,value" and replay keeps only Split(record, ",")[1].
		name: "comma in value",
		ops:  []op{setOp("tk", "this is a data structure, its a hash map with key value pairs,")},
		want: map[string]string{"tk": "this is a data structure, its a hash map with key value pairs,"},
	},
	{
		// "user,42,alice" replays as key "user" with value "42".
		name: "comma in key",
		ops:  []op{setOp("user,42", "alice")},
		want: map[string]string{"user,42": "alice"},
		gone: []string{"user"},
	},
	{
		// Replay treats any record containing "Delete:" as a tombstone, so
		// "note,Delete:victim" deletes victim and never restores note.
		name: "value that looks like a tombstone",
		ops:  []op{setOp("victim", "alive"), setOp("note", "Delete:victim")},
		want: map[string]string{"victim": "alive", "note": "Delete:victim"},
	},
	{
		// "Delete:x,1" is replayed as a tombstone for key "x,1".
		name: "key that looks like a tombstone",
		ops:  []op{setOp("Delete:x", "1")},
		want: map[string]string{"Delete:x": "1"},
	},
	{
		// The tombstone "Delete:a:b" is split on ":" and deletes key "a"
		// instead, so "a:b" comes back to life and "a" is lost.
		name: "colon in deleted key",
		ops:  []op{setOp("a", "keep"), setOp("a:b", "temp"), delOp("a:b")},
		want: map[string]string{"a": "keep"},
		gone: []string{"a:b"},
	},
	{
		// "line2" becomes its own record with no comma, and replay indexes
		// data[1] out of range: the server panics on every startup.
		name: "newline in value",
		ops:  []op{setOp("multi", "line1\nline2")},
		want: map[string]string{"multi": "line1\nline2"},
	},
	{
		// bufio.ScanLines strips a trailing \r from each line.
		name: "trailing carriage return in value",
		ops:  []op{setOp("crlf", "value\r")},
		want: map[string]string{"crlf": "value\r"},
	},
	{
		// bufio.Scanner stops at lines over 64KB and Init never checks
		// scanner.Err(), so this record and every one after it are dropped.
		name: "record larger than 64KB",
		ops:  []op{setOp("big", strings.Repeat("x", 70*1024)), setOp("after", "still here")},
		want: map[string]string{"big": strings.Repeat("x", 70*1024), "after": "still here"},
	},
}

func (s *MintDbSuite) TestPersistenceMatchesMemory() {
	for _, tc := range persistenceCases {
		s.Run(tc.name, func() {
			for _, o := range tc.ops {
				if o.del {
					s.del(o.key)
				} else {
					s.set(o.key, o.value)
				}
			}
			s.checkState(tc, "before restart (in-memory)")

			s.crash()
			s.checkState(tc, "after restart (replayed from disk)")
		})
	}
}

func (s *MintDbSuite) checkState(tc persistenceCase, phase string) {
	for key, want := range tc.want {
		got, found := s.get(key)
		switch {
		case !found:
			s.T().Errorf("%s: key %s missing, want %s", phase, abbrev(key), abbrev(want))
		case got != want:
			s.T().Errorf("%s: key %s = %s, want %s", phase, abbrev(key), abbrev(got), abbrev(want))
		}
	}
	for _, key := range tc.gone {
		if got, found := s.get(key); found {
			s.T().Errorf("%s: key %s = %s, want it absent", phase, abbrev(key), abbrev(got))
		}
	}
}
