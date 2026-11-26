package driver

import "testing"

func TestConfigFromArg(t *testing.T) {
	c, err := ConfigFromArg("postgres://u:p@db.example:6000/shop?sslmode=require")
	if err != nil || c.Engine != EnginePostgres || c.Host != "db.example" || c.Port != 6000 || c.Database != "shop" ||
		c.User != "u" || c.Password != "p" || c.SSLMode != "require" || c.EnvTag != EnvStaging {
		t.Errorf("pg cfg = %+v err=%v", c, err)
	}
	if l, _ := ConfigFromArg("postgres://localhost/x"); l.EnvTag != EnvLocal {
		t.Errorf("localhost tag = %q", l.EnvTag)
	}
	s, _ := ConfigFromArg("/tmp/shop.db")
	if s.Engine != EngineSQLite || s.FilePath != "/tmp/shop.db" || s.Name != "shop.db" {
		t.Errorf("sqlite cfg = %+v", s)
	}
	if _, err := ConfigFromDSN("mysql://x"); err == nil {
		t.Error("non-postgres URL should fail")
	}
}
