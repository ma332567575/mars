package mars

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ma332567575/mars/cluster"
	"github.com/ma332567575/mars/module"
)

func Run(modules []module.Module) {
	for i := 0; i < len(modules); i++ {
		m := modules[i]
		module.Register(m)
	}

	cluster.Init(module.OnMsg)
	module.Init()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill) // trigger close
	sig := <-c
	fmt.Printf("%d", sig)

	cluster.Destroy()
	module.Destroy()
}
