package engine

import (
	"bytes"
	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog/store"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCmd_GetDocument(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		client := pkg.NewMockNetworkClient(mockCtrl)
		hash, _ := model.ParseHash("bdeab373c65ef320514763bc94eaa827bee14915")
		client.EXPECT().GetDocument(hash).Return(&model.DocumentDescriptor{Document: model.Document{Hash: hash}}, nil)
		client.EXPECT().GetDocumentContents(hash).Return(store.NoopCloser{Reader: bytes.NewReader([]byte("foobar"))}, nil)
		command := cmdWithClient(func() pkg.NetworkClient {
			return client
		})
		command.SetArgs([]string{"get", hash.String()})
		err := command.Execute()
		assert.NoError(t, err)
	})
	t.Run("ok - not found", func(t *testing.T) {
		client := pkg.NewMockNetworkClient(mockCtrl)
		hash, _ := model.ParseHash("bdeab373c65ef320514763bc94eaa827bee14915")
		client.EXPECT().GetDocument(hash)
		command := cmdWithClient(func() pkg.NetworkClient {
			return client
		})
		command.SetArgs([]string{"get", hash.String()})
		err := command.Execute()
		assert.NoError(t, err)
	})
	t.Run("ok - found, no contents", func(t *testing.T) {
		client := pkg.NewMockNetworkClient(mockCtrl)
		hash, _ := model.ParseHash("bdeab373c65ef320514763bc94eaa827bee14915")
		client.EXPECT().GetDocument(hash).Return(&model.DocumentDescriptor{Document: model.Document{Hash: hash}}, nil)
		client.EXPECT().GetDocumentContents(hash).Return(nil, documentlog.ErrMissingDocumentContents)
		command := cmdWithClient(func() pkg.NetworkClient {
			return client
		})
		command.SetArgs([]string{"get", hash.String()})
		err := command.Execute()
		assert.NoError(t, err)
	})
}
