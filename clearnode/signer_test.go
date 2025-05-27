package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEIPSignature(t *testing.T) {
	ok, err := VerifyEip712Data("0x21F7D1F35979B125f6F7918fC16Cb9101e5882d7", "69a3bbcc-c806-48e5-aa16-2f61a3dd741f", "0xAc18738D607D82E6b09B08a5C41Bc8ad89cf91F7", "Yellow App Store", "0x3d09ee2f03e51577e5530fb9ac30400852409583de8a91571ca59945c77d7e9c7a7f1cb8723549e561d1c3ca14ee564f1f2a0e959fa08265d0490f6da52908701b")
	assert.True(t, ok)
	assert.NoError(t, err)
}
