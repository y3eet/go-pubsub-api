package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AuthCallbackHandler(c *gin.Context) {
	// Print Headers for debugging
	for key, values := range c.Request.Header {
		for _, value := range values {
			fmt.Println(key + ": " + value + "\n")
		}
	}

	// Print Cookies for debugging
	for _, cookie := range c.Request.Cookies() {
		fmt.Println("Cookie: " + cookie.Name + "=" + cookie.Value + "\n")
	}

	// Print request body for debugging
	bodyBytes, err := c.GetRawData()
	if err != nil {
		fmt.Println("Error reading request body: " + err.Error())
	} else {
		fmt.Println("Request Body: " + string(bodyBytes) + "\n")
	}
	c.JSON(http.StatusOK, gin.H{"message": "Auth callback received"})
}
