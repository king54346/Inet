package main

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	c := NewClient("1", "127.0.0.1:8000")
	c.Run()
}
