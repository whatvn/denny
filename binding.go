package denny

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var Validator = binding.Validator

func Binding(ctx *gin.Context) binding.Binding {
	if ctx == nil {
		return nil
	}
	switch ctx.ContentType() {
	case binding.MIMEPOSTForm:
		return binding.FormPost
	case binding.MIMEXML:
		return binding.XML
	default:
		return binding.JSON
	}
}
