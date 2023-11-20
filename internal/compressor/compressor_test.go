package compressor

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCompressor_GinGzipMiddleware(t *testing.T) {
	tests := []struct {
		name string
		comp *Compressor
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &Compressor{}
			if got := comp.GinGzipMiddleware(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compressor.GinGzipMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}
