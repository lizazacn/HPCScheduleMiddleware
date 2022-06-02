package API

import "github.com/gin-gonic/gin"

func Routes() *gin.Engine{
	r := gin.Default()
	schedule := r.Group("/schedule")
	{
		v1 := schedule.Group("/v1")
		{
			v1.POST("/init", Init)
			v1.POST("/exec", Exec)
		}
	}
	return r
}
