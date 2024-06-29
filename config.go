package main

import (
	xds "github.com/cncf/xds/go/xds/type/v3"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"github.com/envoyproxy/envoy/contrib/golang/filters/http/source/go/pkg/http"
	"google.golang.org/protobuf/types/known/anypb"
)

func init() {
	http.RegisterHttpFilterFactoryAndConfigParser("websocket-example", filterFactory, &parser{})
}

type parser struct {
}

type config struct {
	action string
}

func (p *parser) Parse(conf *anypb.Any, callbacks api.ConfigCallbackHandler) (interface{}, error) {
	configStruct := &xds.TypedStruct{}
	if err := conf.UnmarshalTo(configStruct); err != nil {
		return nil, err
	}

	m := configStruct.Value.AsMap()
	c := &config{}
	if action, ok := m["action"].(string); ok {
		c.action = action
	}

	return c, nil
}

func (p *parser) Merge(parent interface{}, child interface{}) interface{} {
	return child
}

type filter struct {
	api.PassThroughStreamFilter
	config      *config
	callback    api.FilterCallbackHandler
	reqBuffer   []byte
	rspBuffer   []byte
	isWebsocket bool
}

func filterFactory(c interface{}, callbacks api.FilterCallbackHandler) api.StreamFilter {
	return &filter{
		config:    c.(*config),
		callback:  callbacks,
		reqBuffer: make([]byte, 0, 1024),
		rspBuffer: make([]byte, 0, 1024),
	}
}

func main() {
}
