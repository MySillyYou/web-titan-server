package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())

}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_")

func RandStringRunes(n int) string {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}

func CustomerIDOneTesting(append string) error {
	channel := RandStringRunes(14)
	id := RandStringRunes(13)
	idType := RandStringRunes(12)

	cID := GetCustomerID(channel, id, idType, append)
	channel_p, id_p, idType_p, append_p, err := ParseCustomerID(cID)
	if err != nil {
		return err
	}

	if channel_p != channel || id_p != id || idType_p != idType || append != append_p {
		return errors.New(fmt.Sprint("error case", channel, id, idType, append_p))
	}

	return nil
}

func TestGetParseCustomerID(t *testing.T) {
	if err := CustomerIDOneTesting(RandStringRunes(15)); err != nil {
		t.Error(err)
	}

	if err := CustomerIDOneTesting(""); err != nil {
		t.Error(err)
	}
}

func BenchmarkGetParseCustomerID(b *testing.B) {
	for ii := 0; ii < b.N; ii++ {
		CustomerIDOneTesting(RandStringRunes(15))
	}

}
