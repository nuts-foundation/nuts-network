package distdoc

import (
	"crypto"
	"crypto/sha1"
	"crypto/x509"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jws/sign"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// Version defines a type for distributed document format version.
type Version int

const currentVersion = 1
const signingTimeHeader = "sigt"
const versionHeader = "ver"
const previousHeader = "prevs"
const timelineIDHeader = "tid"
const timelineVersionHeader = "tiv"

var allowedAlgos = []jwa.SignatureAlgorithm{jwa.ES256, jwa.ES384, jwa.ES512, jwa.PS256, jwa.PS384, jwa.PS512}

var errInvalidPayloadType = errors.New("payload type must be formatted as MIME type")
var errInvalidPrevs = errors.New("prevs contains an empty hash")
var errSigningError = errors.New("error while signing document")
var unableToParseDocumentErrFmt = "unable to parse document: %w"
var documentNotValidErrFmt = "document validation failed: %w"
var missingHeaderErrFmt = "missing %s header"
var invalidHeaderErrFmt = "invalid %s header"

// PayloadHash is a Hash of the payload of a Document
type PayloadHash model.Hash

// UnsignedDocument holds the base properties of a document which can be signed to create a Document.
type UnsignedDocument interface {
	// PayloadType returns the MIME-formatted type of the payload. It must contain the context and specific type of the
	// payload, e.g. 'registry/endpoint'.
	PayloadType() string
	// Payload returns the hash of the payload of the document.
	Payload() PayloadHash
	// Previous returns the references of the previous documents this document points to.
	Previous() []model.Hash
	// Version returns the version number of the distributed document format.
	Version() Version
	// TimelineID returns the timeline ID of the document.
	TimelineID() model.Hash
	// TimelineVersion returns the timeline version of the document. If the returned version is < 1 the timeline version
	// is not set.
	TimelineVersion() int
	// Sign signs the document using the given key and certificate. The certificate must correspond with the given
	// signing key and must be meant (key usage) for signing. The moment parameter is included in the signature
	// as time of signing.
	Sign(key crypto.Signer, certificate *x509.Certificate, moment time.Time) (Document, error)
}

// Document defines a signed distributed document as described by RFC004 - Distributed Document Format.
type Document interface {
	UnsignedDocument
	// SigningCertificate returns the certificate that was used for signing this document.
	SigningCertificate() *x509.Certificate
	// SigningTime returns the time that the document was signed.
	SigningTime() time.Time
	// Ref returns the reference to this document.
	Ref() model.Hash
	// Data returns the byte representation of this document which can be used for transport.
	Data() []byte
	// VerifySignature verifies that the signature are signing certificate are correct. If an error occurs or verification
	// fails an error is returned.
	VerifySignature(trustStore *x509.CertPool) error
}

// NewDocument creates a new unsigned document. Parameters payload and payloadType can't be empty, but prevs is optional.
// Prevs must not contain empty or invalid hashes.
func NewDocument(payload PayloadHash, payloadType string, prevs []model.Hash) (UnsignedDocument, error) {
	if !ValidatePayloadType(payloadType) {
		return nil, errInvalidPayloadType
	}
	for _, prev := range prevs {
		if prev.Empty() {
			return nil, errInvalidPrevs
		}
	}
	return &document{
		prevs:       append(prevs),
		payload:     model.Hash(payload),
		payloadType: payloadType,
		version:     currentVersion,
	}, nil
}

func ValidatePayloadType(payloadType string) bool {
	return strings.Contains(payloadType, "/")
}

type document struct {
	prevs           []model.Hash
	payload         model.Hash
	payloadType     string
	certificate     *x509.Certificate
	signingTime     time.Time
	version         Version
	timelineID      model.Hash
	timelineVersion int

	data []byte
	ref  model.Hash
}

func (d document) Data() []byte {
	return d.data
}

func (d document) SigningCertificate() *x509.Certificate {
	return d.certificate
}

func (d document) SigningTime() time.Time {
	return d.signingTime
}

func (d document) PayloadType() string {
	return d.payloadType
}

func (d document) Payload() PayloadHash {
	return PayloadHash(d.payload)
}

func (d document) Previous() []model.Hash {
	return append(d.prevs)
}

func (d document) Ref() model.Hash {
	return d.ref
}

func (d document) Version() Version {
	return d.version
}

func (d document) TimelineID() model.Hash {
	return d.timelineID
}

func (d document) TimelineVersion() int {
	return d.timelineVersion
}

func (d document) VerifySignature(_ *x509.CertPool) error {
	return errors.New("VerifySignature() not implemented yet!")
}

func (d *document) Sign(key crypto.Signer, certificate *x509.Certificate, moment time.Time) (Document, error) {
	prevsAsString := make([]string, len(d.prevs))
	for i, prev := range d.prevs {
		prevsAsString[i] = prev.String()
	}
	// TODO: Make algorithm a parameter
	algo := jwa.ES256
	headerMap := map[string]interface{}{
		jws.AlgorithmKey:     algo,
		jws.ContentTypeKey:   d.payloadType,
		jws.CriticalKey:      []string{signingTimeHeader, versionHeader, previousHeader},
		jws.X509CertChainKey: cert.MarshalX509CertChain([]*x509.Certificate{certificate}),
		signingTimeHeader:    moment.UTC().Unix(),
		previousHeader:       prevsAsString,
		versionHeader:        d.Version(),
	}
	if !d.timelineID.Empty() {
		headerMap[timelineIDHeader] = d.timelineID
		if d.timelineVersion > 0 {
			headerMap[timelineVersionHeader] = d.timelineVersion
		}
	}
	headers := jws.NewHeaders()
	for key, value := range headerMap {
		if err := headers.Set(key, value); err != nil {
			return nil, errors.Wrapf(err, "unable to set header %s", key)
		}
	}
	if signer, err := sign.New(algo); err != nil {
		return nil, err
	} else {
		if data, err := jws.SignMulti([]byte(d.payload.String()), jws.WithSigner(signer, key, nil, headers)); err != nil {
			return nil, errors.Wrap(err, errSigningError.Error())
		} else {
			d.setData(data)
			d.certificate = certificate
			return d, nil
		}
	}
}

func (d *document) setData(data []byte) {
	d.data = data
	d.ref = sha1.Sum(d.data)
}
