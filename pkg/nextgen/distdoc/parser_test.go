package distdoc

import (
	"crypto/x509"
	"encoding/base64"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseDocument(t *testing.T) {
	key, certificate := generateKeyAndCertificate()
	payload, _ := model.ParseHash("f33b5cae968cb88f157999b3551ab0863d2a8f0a")
	payloadAsBytes := []byte(payload.String())
	t.Run("ok", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.NotNil(t, document)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, PayloadHash(payload), document.Payload())
		assert.Equal(t, certificate.Raw, document.SigningCertificate().Raw)
		assert.Equal(t, 1, int(document.Version()))
		assert.Equal(t, "foo/bar", document.PayloadType())
		assert.Equal(t, headers.PrivateParams()[previousHeader].([]string)[0], document.Previous()[0].String())
		assert.Equal(t, headers.PrivateParams()[timelineIDHeader].(string), document.TimelineID().String())
		assert.Equal(t, headers.PrivateParams()[timelineVersionHeader].(int), document.TimelineVersion())
		assert.NotNil(t, document.Data())
		assert.False(t, document.Ref().Empty())
	})
	t.Run("error - input not a JWS (compact serialization format)", func(t *testing.T) {
		document, err := ParseDocument([]byte("not a JWS"))
		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "unable to parse document: failed to parse jws message: invalid compact serialization format")
	})
	t.Run("error - input not a JWS (JSON serialization format)", func(t *testing.T) {
		document, err := ParseDocument([]byte("{}"))
		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: JWS does not contain any signature")
	})
	t.Run("error - sigt header is missing", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		delete(headers.PrivateParams(), signingTimeHeader)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: missing sigt header")
	})
	t.Run("error - invalid sigt header", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(signingTimeHeader, "not a date")
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: invalid sigt header")
	})
	t.Run("error - vers header is missing", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		delete(headers.PrivateParams(), versionHeader)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: missing ver header")
	})
	t.Run("error - prevs header is missing", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		delete(headers.PrivateParams(), previousHeader)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: missing prevs header")
	})
	t.Run("error - invalid prevs (not an array)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(previousHeader, 2)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: invalid prevs header")
	})
	t.Run("error - invalid prevs (invalid entry)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(previousHeader, []string{"not a hash"})
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: invalid prevs header")
	})
	t.Run("error - invalid prevs (invalid entry, not a string)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(previousHeader, []int{5})
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: invalid prevs header")
	})
	t.Run("error - cty header is invalid", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(jws.ContentTypeKey, "")
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: payload type must be formatted as MIME type")
	})
	t.Run("error - invalid version", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(versionHeader, "foobar")
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: invalid ver header")
	})
	t.Run("error - unsupported version", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(versionHeader, 2)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: unsupported version: 2")
	})
	t.Run("error - invalid algorithm", func(t *testing.T) {
		key, certificate := generateRSAKeyAndCertificate()
		headers := makeJWSHeaders(certificate)
		headers.Set(jws.AlgorithmKey, jwa.RS256)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: signing algorithm not allowed: RS256")
	})
	t.Run("error - invalid x5c header (too many entries)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(jws.X509CertChainKey, []string{base64.StdEncoding.EncodeToString(certificate.Raw), base64.StdEncoding.EncodeToString(certificate.Raw)})
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: expected 1 certificate in x5c header")
	})
	t.Run("error - invalid x5c header (not a certificate)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(jws.X509CertChainKey, []string{base64.StdEncoding.EncodeToString([]byte{1, 2, 3})})
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid x5c header")
	})
	t.Run("error - no certificates in x5c header", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(jws.X509CertChainKey, []string{})
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.EqualError(t, err, "document validation failed: expected 1 certificate in x5c header")
	})
	t.Run("error - invalid tid header", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(timelineIDHeader, "not a valid hash")
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid tid header")
	})
	t.Run("error - invalid tid header (not a string)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(timelineIDHeader, 5)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid tid header")
	})
	t.Run("error - invalid tiv header (not a numeric value)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(timelineVersionHeader, "not a numeric value")
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid tiv header")
	})
	t.Run("error - invalid tiv header (not a numeric value)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(timelineVersionHeader, "not a numeric value")
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid tiv header")
	})
	t.Run("error - invalid tiv header (value smaller than 0)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(timelineVersionHeader, -1)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid tiv header")
	})
	t.Run("error - invalid tiv header (value non-integer)", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		headers.Set(timelineVersionHeader, 1.5)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid tiv header")
	})
	t.Run("error - tiv header without tid", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		delete(headers.PrivateParams(), timelineIDHeader)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: tiv specified without tid header")
	})
	t.Run("error - tiv header without tid", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		delete(headers.PrivateParams(), timelineIDHeader)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: tiv specified without tid header")
	})
	t.Run("error - invalid payload", func(t *testing.T) {
		headers := makeJWSHeaders(certificate)
		signature, _ := jws.Sign([]byte("not a valid hash"), headers.Algorithm(), key, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid payload")
	})
	t.Run("error - certificate does not match key", func(t *testing.T) {
		t.SkipNow()
		otherKey, _ := generateKeyAndCertificate()
		headers := makeJWSHeaders(certificate)
		signature, _ := jws.Sign(payloadAsBytes, headers.Algorithm(), otherKey, jws.WithHeaders(headers))

		document, err := ParseDocument(signature)

		assert.Nil(t, document)
		assert.Contains(t, err.Error(), "document validation failed: invalid payload")
	})
}

func makeJWSHeaders(certificate *x509.Certificate) jws.Headers {
	prev, _ := model.ParseHash("4d7a93b838bb1f2176d9daa9574c887313e36620")
	timelineID, _ := model.ParseHash("0e5f927353a9b0b79c27fec27fb32c07c98f411c")
	headerMap := map[string]interface{}{
		jws.AlgorithmKey:      jwa.ES256,
		jws.ContentTypeKey:    "foo/bar",
		jws.CriticalKey:       []string{signingTimeHeader, versionHeader, previousHeader},
		jws.X509CertChainKey:  cert.MarshalX509CertChain([]*x509.Certificate{certificate}),
		signingTimeHeader:     time.Now().UTC().Unix(),
		versionHeader:         1,
		previousHeader:        []string{prev.String()},
		timelineIDHeader:      timelineID.String(),
		timelineVersionHeader: 1,
	}
	headers := jws.NewHeaders()
	for key, value := range headerMap {
		if err := headers.Set(key, value); err != nil {
			logrus.Fatalf("Unable to set header %s: %v", key, err)
		}
	}
	return headers
}
