package module

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/ma332567575/mars/conf"
	"github.com/ma332567575/mars/log"
	"github.com/ma332567575/mars/network"
)

type Module interface {
	OnInit()
	OnDestroy()
	OnMsg(agent network.Agent, msgID uint32, seqID uint32, body []byte) bool
	Run(closeSig chan bool)
}

type module struct {
	mi       Module
	closeSig chan bool
	wg       sync.WaitGroup
}

var modules []*module

func Register(mi Module) {
	m := new(module)
	m.mi = mi
	m.closeSig = make(chan bool, 1)
	modules = append(modules, m)
}

func Init() {
	// init all module
	for i := 0; i < len(modules); i++ {
		modules[i].mi.OnInit()
	}

	// run all module
	for i := 0; i < len(modules); i++ {
		m := modules[i]
		m.wg.Add(1) //for all module run stop, then destory
		go run(m)
	}
}

func OnMsg(agent network.Agent, msgID uint32, seqID uint32, body []byte) {
	for i := 0; i < len(modules); i++ {
		m := modules[i]
		// if deal, return true
		if m.mi.OnMsg(agent, msgID, seqID, body) {
			break
		}
	}
	flylog.GetInstance().Warn("unhandle msgid")
}

func run(m *module) {
	// run module
	m.mi.Run(m.closeSig)
	// run over then done. for destory
	m.wg.Done()
}

func Destroy() {
	for i := 0; i < len(modules); i++ {
		m := modules[i]
		m.closeSig <- true // nfy close
		m.wg.Wait()
		destroy(m)
	}
}

func destroy(m *module) {
	defer func() {
		if r := recover(); r != nil {

			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				//log.Error("%v: %s", r, buf[:l])
				fmt.Printf("%v: %s", r, buf[:l])
			} else {
				//log.Error("%v", r)
				fmt.Printf("%v", r)
			}

		}
	}()

	m.mi.OnDestroy()
}
