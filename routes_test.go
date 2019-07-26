package denny

import (
	"github.com/gin-gonic/gin"
)




func main()  {
	handler := func(ctx *gin.Context) {

	}
	routers := NewRouter()
	routers.Add("/", HttpGet, handler)
}
