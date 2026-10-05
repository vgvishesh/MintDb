package client

func (s *MintDbSuite) TestSetThenGet() {
	s.set("todayKey", "querty")

	value, found := s.get("todayKey")
	s.True(found)
	s.Equal("querty", value)
}

func (s *MintDbSuite) TestGetMissingKey() {
	_, found := s.get("never-set")
	s.False(found)
}

func (s *MintDbSuite) TestSetOverwrites() {
	s.set("k", "v1")
	s.set("k", "v2")

	value, _ := s.get("k")
	s.Equal("v2", value)
}

func (s *MintDbSuite) TestDeleteRemovesKey() {
	s.set("k", "v")
	s.del("k")

	_, found := s.get("k")
	s.False(found)
}
