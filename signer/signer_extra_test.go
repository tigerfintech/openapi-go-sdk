package signer

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"testing"
)

// TestLoadPublicKey_Valid verifies loading a valid base64-encoded PKIX RSA public key.
func TestLoadPublicKey_Valid(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成 RSA 密钥对失败: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("序列化公钥失败: %v", err)
	}
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)

	pub, err := LoadPublicKey(pubB64)
	if err != nil {
		t.Fatalf("LoadPublicKey 失败: %v", err)
	}
	if pub == nil {
		t.Fatal("返回的公钥不应为 nil")
	}
	if pub.N.Cmp(privKey.PublicKey.N) != 0 {
		t.Error("公钥 N 模数不匹配")
	}
}

// TestLoadPublicKey_Empty verifies empty string returns an error.
func TestLoadPublicKey_Empty(t *testing.T) {
	_, err := LoadPublicKey("")
	if err == nil {
		t.Fatal("空字符串应返回错误")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Errorf("错误信息应提示为空，实际: %v", err)
	}
}

// TestLoadPublicKey_InvalidBase64 verifies invalid base64 input returns an error.
func TestLoadPublicKey_InvalidBase64(t *testing.T) {
	_, err := LoadPublicKey("!!!not-valid-base64!!!")
	if err == nil {
		t.Fatal("非法 base64 应返回错误")
	}
	if !strings.Contains(err.Error(), "base64-decode") {
		t.Errorf("错误信息应提示 base64 解码失败，实际: %v", err)
	}
}

// TestLoadPublicKey_InvalidDER verifies valid base64 but invalid DER returns an error.
func TestLoadPublicKey_InvalidDER(t *testing.T) {
	// Valid base64 but not a valid public key
	invalidDER := base64.StdEncoding.EncodeToString([]byte("not a real key"))
	_, err := LoadPublicKey(invalidDER)
	if err == nil {
		t.Fatal("无效 DER 数据应返回错误")
	}
	if !strings.Contains(err.Error(), "failed to parse public key") {
		t.Errorf("错误信息应提示解析失败，实际: %v", err)
	}
}

// TestLoadPublicKey_NonRSAKey verifies that a non-RSA public key (ECDSA) returns an error.
func TestLoadPublicKey_NonRSAKey(t *testing.T) {
	ecPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成 ECDSA 密钥对失败: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&ecPrivKey.PublicKey)
	if err != nil {
		t.Fatalf("序列化 ECDSA 公钥失败: %v", err)
	}
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)

	_, err = LoadPublicKey(pubB64)
	if err == nil {
		t.Fatal("非 RSA 公钥应返回错误")
	}
	if !strings.Contains(err.Error(), "not RSA type") {
		t.Errorf("错误信息应提示非 RSA 类型，实际: %v", err)
	}
}

// TestVerifyWithRSA_InvalidBase64Signature verifies that an invalid base64 signature returns an error.
func TestVerifyWithRSA_InvalidBase64Signature(t *testing.T) {
	privPEM, pubKey := generateTestKeyPair(t)
	content := "test content"
	// Sign with valid key to get a valid signature
	validSig, err := SignWithRSA(privPEM, content)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}
	_ = validSig // ensure signing works

	// Now verify with invalid base64 signature
	err = VerifyWithRSA(pubKey, content, "!!!invalid-base64-sig!!!")
	if err == nil {
		t.Fatal("非法 base64 签名应返回错误")
	}
	if !strings.Contains(err.Error(), "Base64 解码失败") {
		t.Errorf("错误信息应提示 base64 解码失败，实际: %v", err)
	}
}

// TestVerifyWithRSA_TamperedContent verifies that verifying a signature against different content fails.
func TestVerifyWithRSA_TamperedContent(t *testing.T) {
	privPEM, pubKey := generateTestKeyPair(t)
	originalContent := "original content"
	tamperedContent := "tampered content"

	signature, err := SignWithRSA(privPEM, originalContent)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	// Verify against tampered content should fail
	err = VerifyWithRSA(pubKey, tamperedContent, signature)
	if err == nil {
		t.Fatal("篡改内容后验签应失败")
	}
}

// TestVerifyWithRSA_TamperedSignature verifies that a modified signature fails verification.
func TestVerifyWithRSA_TamperedSignature(t *testing.T) {
	privPEM, pubKey := generateTestKeyPair(t)
	content := "test content"

	signature, err := SignWithRSA(privPEM, content)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	// Tamper with the signature: decode, modify, re-encode
	sigBytes, _ := base64.StdEncoding.DecodeString(signature)
	if len(sigBytes) > 0 {
		sigBytes[0] ^= 0xFF // flip bits
	}
	tamperedSig := base64.StdEncoding.EncodeToString(sigBytes)

	err = VerifyWithRSA(pubKey, content, tamperedSig)
	if err == nil {
		t.Fatal("篡改签名后验签应失败")
	}
}

// TestParseDERPrivateKey_NonRSA verifies that a non-RSA private key (ECDSA) in PKCS#8 DER
// returns an error when type-asserted to *rsa.PrivateKey.
func TestParseDERPrivateKey_NonRSA(t *testing.T) {
	ecPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成 ECDSA 密钥对失败: %v", err)
	}
	derBytes, err := x509.MarshalPKCS8PrivateKey(ecPrivKey)
	if err != nil {
		t.Fatalf("序列化 ECDSA 私钥失败: %v", err)
	}

	_, err = parseDERPrivateKey(derBytes)
	if err == nil {
		t.Fatal("非 RSA 私钥应返回错误")
	}
	if !strings.Contains(err.Error(), "不是 RSA 类型") {
		t.Errorf("错误信息应提示非 RSA 类型，实际: %v", err)
	}
}

// TestSignWithRSA_EmptyContent verifies signing empty content works (edge case).
func TestSignWithRSA_EmptyContent(t *testing.T) {
	privPEM, pubKey := generateTestKeyPair(t)
	content := ""

	signature, err := SignWithRSA(privPEM, content)
	if err != nil {
		t.Fatalf("签名空内容失败: %v", err)
	}
	if signature == "" {
		t.Fatal("签名结果不应为空")
	}

	// Verify the signature of empty content
	if err := VerifyWithRSA(pubKey, content, signature); err != nil {
		t.Fatalf("验签空内容失败: %v", err)
	}
}

// TestSignWithRSA_LongContent verifies signing a very long string works (boundary value).
func TestSignWithRSA_LongContent(t *testing.T) {
	privPEM, pubKey := generateTestKeyPair(t)
	// Generate a long string (> 10KB)
	content := strings.Repeat("a", 10240)

	signature, err := SignWithRSA(privPEM, content)
	if err != nil {
		t.Fatalf("签名长内容失败: %v", err)
	}

	if err := VerifyWithRSA(pubKey, content, signature); err != nil {
		t.Fatalf("验签长内容失败: %v", err)
	}
}

// TestLoadPrivateKey_NotPEMOrBase64 verifies that invalid data (neither PEM nor base64) returns an error.
func TestLoadPrivateKey_NotPEMOrBase64(t *testing.T) {
	_, err := LoadPrivateKey("this is definitely \x00 not valid pem or base64!!!")
	if err == nil {
		t.Fatal("无效数据应返回错误")
	}
}

// TestLoadPrivateKey_InvalidPEMBlock verifies that PEM with invalid DER bytes returns an error.
func TestLoadPrivateKey_InvalidPEMBlock(t *testing.T) {
	// Construct a PEM block with invalid DER content
	invalidPEM := "-----BEGIN " + "RSA PRIVATE KEY-----\n" +
		base64.StdEncoding.EncodeToString([]byte("invalid der data here")) + "\n" +
		"-----END RSA PRIVATE KEY-----\n"
	_, err := LoadPrivateKey(invalidPEM)
	if err == nil {
		t.Fatal("无效 PEM 块应返回错误")
	}
}

// TestVerifyWithRSA_ValidSignature verifies the full sign-verify round trip with explicit hash.
func TestVerifyWithRSA_ValidSignature(t *testing.T) {
	privPEM, pubKey := generateTestKeyPair(t)
	content := "round-trip test"

	signature, err := SignWithRSA(privPEM, content)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	// Manually verify using crypto/rsa to ensure consistency
	sigBytes, _ := base64.StdEncoding.DecodeString(signature)
	hashed := sha1.Sum([]byte(content))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA1, hashed[:], sigBytes); err != nil {
		t.Fatalf("手动验签失败: %v", err)
	}

	// Also verify via VerifyWithRSA
	if err := VerifyWithRSA(pubKey, content, signature); err != nil {
		t.Fatalf("VerifyWithRSA 验签失败: %v", err)
	}
}
