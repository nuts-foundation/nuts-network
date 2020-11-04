package distdoc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/nuts-foundation/nuts-crypto/test"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewDocument(t *testing.T) {
	payloadHash, _ := model.ParseHash("f33b5cae968cb88f157999b3551ab0863d2a8f0a")
	payload := PayloadHash(payloadHash)
	t.Run("ok", func(t *testing.T) {
		hash, _ := model.ParseHash("31a3d460bb3c7d98845187c716a30db81c44b615")

		document, err := NewDocument(payload, "some/type", []model.Hash{hash})

		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "some/type", document.PayloadType())
		assert.Equal(t, document.Payload(), payload)
		assert.Equal(t, []model.Hash{hash}, document.Previous())
		assert.Equal(t, Version(1), document.Version())
	})
	t.Run("error - type empty", func(t *testing.T) {
		document, err := NewDocument(payload, "", nil)
		assert.EqualError(t, err, errInvalidPayloadType.Error())
		assert.Nil(t, document)
	})
	t.Run("error - type not a MIME type", func(t *testing.T) {
		document, err := NewDocument(payload, "foo", nil)
		assert.EqualError(t, err, errInvalidPayloadType.Error())
		assert.Nil(t, document)
	})
	t.Run("error - invalid prev", func(t *testing.T) {
		document, err := NewDocument(payload, "foo/bar", []model.Hash{model.EmptyHash()})
		assert.EqualError(t, err, errInvalidPrevs.Error())
		assert.Nil(t, document)
	})
}

func Test_document_Getters(t *testing.T) {
	payload, _ := model.ParseHash("f33b5cae968cb88f157999b3551ab0863d2a8f0a")
	timelineID, _ := model.ParseHash("f33b5cae968cb88f157999b3551ab0863d2a8f0b")
	_, certificate := generateKeyAndCertificate()
	prev1, _ := model.ParseHash("f33b5cae968cb88f157999b3551ab0863d2a8f0a")
	prev2, _ := model.ParseHash("b13098cdd24a364a0a6d16057473ad4d19d7dfb8")
	doc := document{
		prevs:           []model.Hash{prev1, prev2},
		payload:         payload,
		payloadType:     "foo/bar",
		certificate:     certificate,
		signingTime:     time.Unix(1023323333, 0),
		version:         10,
		timelineID:      timelineID,
		timelineVersion: 10,
	}
	doc.setData([]byte{1, 2, 3})

	assert.Equal(t, doc.prevs, doc.Previous())
	assert.Equal(t, PayloadHash(doc.payload), doc.Payload())
	assert.Equal(t, doc.payloadType, doc.PayloadType())
	assert.Equal(t, doc.certificate, doc.SigningCertificate())
	assert.Equal(t, doc.signingTime, doc.SigningTime())
	assert.Equal(t, doc.version, doc.Version())
	assert.Equal(t, doc.timelineID, doc.TimelineID())
	assert.Equal(t, doc.timelineVersion, doc.TimelineVersion())
	assert.Equal(t, doc.data, doc.Data())
	assert.False(t, doc.Ref().Empty())
}

func Test_document_Sign(t *testing.T) {
	payloadHash, _ := model.ParseHash("f33b5cae968cb88f157999b3551ab0863d2a8f0a")
	payload := PayloadHash(payloadHash)
	key, certificate := generateKeyAndCertificate()
	prev1, _ := model.ParseHash("f33b5cae968cb88f157999b3551ab0863d2a8f0a")
	prev2, _ := model.ParseHash("b13098cdd24a364a0a6d16057473ad4d19d7dfb8")
	expectedPrevs := []model.Hash{prev1, prev2}
	contentType := "foo/bar"
	t.Run("ok", func(t *testing.T) {
		doc, _ := NewDocument(payload, contentType, expectedPrevs)

		signedDoc, err := doc.Sign(key, certificate, time.Date(2020, 10, 23, 13, 0, 0, 0, time.UTC))

		if !assert.NoError(t, err) {
			return
		}
		t.Run("verify object properties", func(t *testing.T) {
			assert.Equal(t, certificate.Raw, signedDoc.SigningCertificate().Raw)
			assert.False(t, signedDoc.Ref().Empty())
		})
		t.Run("verify JWS", func(t *testing.T) {
			message, _ := jws.ParseString(string(signedDoc.Data()))
			assert.Len(t, message.Signatures(), 1, "expected 1 signature")
			headers := message.Signatures()[0].ProtectedHeaders()
			// JWS headers
			assert.Equal(t, contentType, headers.ContentType())
			assert.Equal(t, jwa.ES256, headers.Algorithm())
			assert.Len(t, headers.X509CertChain(), 1)
			// Custom headers
			assert.Equal(t, int64(1603458000), int64(headers.PrivateParams()[signingTimeHeader].(float64)))
			assert.Equal(t, 1, int(headers.PrivateParams()[versionHeader].(float64)))
			prevs := headers.PrivateParams()[previousHeader]
			assert.Len(t, prevs, 2, "expected 2 prevs")
			assert.Equal(t, prev1.String(), prevs.([]interface{})[0].(string))
			assert.Equal(t, prev2.String(), prevs.([]interface{})[1].(string))
		})
	})
}

func Test_document_VerifySignature(t *testing.T) {
	t.Run("not implemented yet", func(t *testing.T) {
		err := document{}.VerifySignature(nil)
		assert.EqualError(t, err, "VerifySignature() not implemented yet!")
	})
}

func generateKeyAndCertificate() (*ecdsa.PrivateKey, *x509.Certificate) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	certAsBytes := test.GenerateCertificateEx(time.Now(), key, 1, true, x509.KeyUsageContentCommitment)
	certificate, _ := x509.ParseCertificate(certAsBytes)
	return key, certificate
}

func generateRSAKeyAndCertificate() (*rsa.PrivateKey, *x509.Certificate) {
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	certAsBytes := test.GenerateCertificateEx(time.Now(), key, 1, true, x509.KeyUsageContentCommitment)
	certificate, _ := x509.ParseCertificate(certAsBytes)
	return key, certificate
}
