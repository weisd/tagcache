package cache

import (
	"github.com/weisd/tagcache"
	_ "github.com/weisd/tagcache/redis"
	"testing"
)

func Test_TagCache(t *testing.T) {

	c, err := tagcache.New(tagcache.Options{Adapter: "redis", AdapterConfig: `{"Addr":"127.0.0.1:6379"}`, Section: "dada"})
	if err != nil {
		t.Fatal(err)
	}

	// base use
	err = c.Set("da", "weisd", 300)
	if err != nil {
		t.Fatal(err)
	}

	res := c.Get("da")

	if res != "weisd" {
		t.Fatal("base Set faield")
	}

	t.Log("ok")

	// use tags/namespace
	err = c.Tags([]string{"dd"}).Set("da", "weisd", 300)
	if err != nil {
		t.Fatal(err)
	}
	res = c.Tags([]string{"dd"}).Get("da")

	if res != "weisd" {
		t.Fatal("tags Set faield")
	}

	t.Log("ok")

	err = c.Tags([]string{"aa"}).Set("aa", "aaa", 300)
	if err != nil {
		t.Fatal(err)
	}

	res = c.Tags([]string{"aa"}).Get("aa")

	if res != "aaa" {
		t.Fatal("not aaa")
	}

	t.Log("ok")

	err = c.Tags([]string{"aa", "cc"}).Set("cc", "dada", 300)
	if err != nil {
		t.Fatal(err)
	}

	res = c.Tags([]string{"aa", "cc"}).Get("cc")

	if res != "dada" {
		t.Fatal("not aaa")
	}

	t.Log("ok")

	// flush namespace
	err = c.Tags([]string{"aa"}).Flush()
	if err != nil {
		t.Fatal(err)
	}

	res = c.Tags([]string{"aa"}).Get("aa")
	if res != "" {
		t.Fatal("flush faield")
	}

	res = c.Tags([]string{"aa", "cc"}).Get("cc")
	if res != "" {
		t.Fatal("flush faield")
	}

	res = c.Tags([]string{"aa"}).Get("bb")
	if res != "" {
		t.Fatal("flush faield")
	}

	// still store in
	res = c.Tags([]string{"dd"}).Get("da")
	if res != "weisd" {
		t.Fatal("where ")
	}

	t.Log("ok")

	// c.Flush()

}
