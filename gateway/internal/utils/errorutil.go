package utils

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleGRPCError(c *gin.Context, err error) {
	if status, ok := status.FromError(err); ok {
		switch status.Code() {
		case codes.NotFound:
			c.JSON(404, gin.H{"error": status.Message()})
		case codes.InvalidArgument:
			c.JSON(400, gin.H{"error": status.Message()})
		case codes.Unauthenticated:
			c.JSON(401, gin.H{"error": status.Message()})
		case codes.AlreadyExists:
			c.JSON(409, gin.H{"error": status.Message()})
		
		default:
			c.JSON(500, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(500, gin.H{"error": "internal server error"})
}
