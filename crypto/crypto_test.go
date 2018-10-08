package crypto

import (
	"strings"
	"testing"
)

func TestEncrypt(t *testing.T) {
	var out string
	TAG := "!seal"
	secret := Secret{}
	word := "pappaeciccia"
	password := DEFAULT_PASSWORD

	err := secret.Encrypt(&word, &out, &password)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if !strings.Contains(out, TAG) {
		t.Logf("Error: expected TAG %s but not found", TAG)
		t.Fail()
	}
}

func TestDecrypt(t *testing.T) {
	var out string

	secret := Secret{}
	word := "AXHToSbAvVZNRwCyaXvtSPIWVSgcL8u6SzED4Q=="
	password := DEFAULT_PASSWORD
	expected := "segretissimo"

	err := secret.Decrypt(&word, &out, &password)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if out != expected {
		t.Errorf("Error: expected %s, got %s", expected, out)
		t.Fail()
	}

}
