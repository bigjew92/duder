package main

import "log"

type storage struct {
}

func init() {
	Duder.Storage = storage{}
}

func (s *storage) Test() {
	log.Print("lkajsdlfkjasldfkjalsdkfjlaksdjfaskldjf;laksjd;flkajs;dlfkjas;dfjasd!!!")
}
