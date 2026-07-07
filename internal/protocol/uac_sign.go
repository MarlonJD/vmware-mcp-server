package protocol

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const UACSignatureAlgorithm = "hmac-sha256-v1"

func SignUAC(request *Request, secret string) error {
	if request == nil || request.UAC == nil {
		return errors.New("uac request is required")
	}
	if strings.TrimSpace(secret) == "" {
		return errors.New("secret is required")
	}
	signature := uacSignature(*request, secret)
	request.UAC.SignatureAlgo = UACSignatureAlgorithm
	request.UAC.Signature = signature
	return nil
}

func VerifyUAC(request Request, secret string) bool {
	if request.UAC == nil || strings.TrimSpace(secret) == "" {
		return false
	}
	if request.UAC.SignatureAlgo != UACSignatureAlgorithm || request.UAC.Signature == "" {
		return false
	}
	expected := uacSignature(request, secret)
	return hmac.Equal([]byte(expected), []byte(request.UAC.Signature))
}

func uacSignature(request Request, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	fields := []string{
		request.ID,
		string(request.Kind),
		request.ExpiresAt.UTC().Format(timeFormatRFC3339Nano()),
		request.UAC.ExpectedApp,
		request.UAC.ExpectedPublisher,
		request.UAC.Reason,
		request.UAC.CommandSummary,
	}
	mac.Write([]byte(strings.Join(fields, "\n")))
	return hex.EncodeToString(mac.Sum(nil))
}

func timeFormatRFC3339Nano() string {
	return "2006-01-02T15:04:05.999999999Z07:00"
}
