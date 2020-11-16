package nodelist

import (
	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_nodeList_Configure(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		err := (&nodeList{}).Configure("abc", "foo:8080")
		assert.NoError(t, err)
	})
	t.Run("error - nodeID empty", func(t *testing.T) {
		err := (&nodeList{}).Configure("", "foo:8080")
		assert.EqualError(t, err, "nodeID is empty")
	})
	t.Run("ok - address empty", func(t *testing.T) {
		err := (&nodeList{}).Configure("abc", "")
		assert.NoError(t, err)
	})
}

func Test_nodeList_register(t *testing.T) {
	t.Run("ok - not yet registered", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		documentLog := documentlog.NewMockDocumentLog(ctrl)
		documentLog.EXPECT().FindByContentsHash(gomock.Any())
		documentLog.EXPECT().AddDocumentWithContents(gomock.Any(), nodeInfoDocumentType, gomock.Any())
		p2pNetwork := p2p.NewMockP2PNetwork(ctrl)

		nodeList := NewNodeList(documentLog, p2pNetwork).(*nodeList)
		nodeList.Configure("foo", "localhost:5555")
		err := nodeList.register()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("ok - already registered", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		documentLog := documentlog.NewMockDocumentLog(ctrl)
		documentLog.EXPECT().FindByContentsHash(gomock.Any()).Return([]model.DocumentDescriptor{{Document: model.Document{}}}, nil)
		p2pNetwork := p2p.NewMockP2PNetwork(ctrl)
		nodeList := NewNodeList(documentLog, p2pNetwork).(*nodeList)
		nodeList.Configure("foo", "localhost:5555")
		err := nodeList.register()
		if !assert.NoError(t, err) {
			return
		}
	})
}