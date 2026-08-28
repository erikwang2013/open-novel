package pkg

import "testing"

func TestCryptoRoundTrip(t *testing.T) {
	c, err := NewCrypto("dev-encrypt-key-change-me")
	if err != nil {
		t.Fatal(err)
	}
	plain := `{"coin":"usdttrc20","plans":{"monthly":300,"quarterly":800,"yearly":3000}}`
	enc, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if enc == plain {
		t.Fatal("encrypt returned plaintext")
	}
	dec, err := c.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != plain {
		t.Fatalf("round trip mismatch: got %q want %q", dec, plain)
	}
	// 两次加密随机 nonce 应不同
	enc2, _ := c.Encrypt(plain)
	if enc2 == enc {
		t.Fatal("nonce not randomized")
	}
}

func TestCryptoWrongKey(t *testing.T) {
	c, _ := NewCrypto("key-a")
	enc, _ := c.Encrypt("secret")
	c2, _ := NewCrypto("key-b")
	if _, err := c2.Decrypt(enc); err == nil {
		t.Fatal("decrypt with wrong key should fail")
	}
}

func TestCryptoTampered(t *testing.T) {
	c, _ := NewCrypto("key-a")
	enc, _ := c.Encrypt("secret")
	if _, err := c.Decrypt(enc[:len(enc)-2] + "xx"); err == nil {
		t.Fatal("tampered ciphertext should fail")
	}
}
