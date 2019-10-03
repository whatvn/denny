package denny

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func Binding(ctx *gin.Context) binding.Binding  {
	if ctx == nil {
		return nil
	}
	switch ctx.ContentType() {
	case binding.MIMEJSON:
		return binding.JSON
	case binding.MIMEPOSTForm:
		return binding.FormPost
	case binding.MIMEXML:
		return binding.XML
	default:
		return nil
	}
}
