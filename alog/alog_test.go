package alog

import (
	"testing"
)

func Test_generateName(t *testing.T) {
	t.Log(generate())
}

func Test_logger(t *testing.T) {
	Print("hello")

	Printf("hello Test_logger")
	Printf("hello Test_logger2")

	Fatalln("heeettt", 1, 2, 3)

	// alog.Error("Errorf")
	// alog.Errorf("Errorf %d\n", 1)

	// alog.Debug("Debug")
	// alog.Debugf("Debugf %d\n", 1)

	// alog.Info("Info")
	// alog.Infof("Infof %d\n", 1)

	// alog.Warm("Warm")
	// alog.Warmf("Warmf %d\n", 1)

}
