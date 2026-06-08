package controller

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/gemini"
	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

//@ types

// type that stores gemini ai model which -> stores gemini api callers method in it
type GeminiController struct {
	GeminiAIServiceModel *gemini.GeminiAIService
	pns                  *services.PushNotificationService
}

// func that returns instance of handler type which -> which stores mehtod for calling gemini api for responses
func NewGeminiController(geminiAIServiceModel *gemini.GeminiAIService, pns *services.PushNotificationService) *GeminiController {
	return &GeminiController{
		GeminiAIServiceModel: geminiAIServiceModel,
		pns:                  pns,
	}
}

func (g *GeminiController) ServeGEMINIAIPromptRequest(c *gin.Context) {

	// auth middleware userID check -> recieves token -> attaches user_id from it
	userID := c.GetUint("user_id")

	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.GeminiErrResponse{
			Status:         "access denied;login expired;invalid token",
			Ok:             false,
			GeminiResponse: "early return;prompt did not reach ai",
		})
		return
	}

	var payload models.GeminiPromptRequestPayload
	// atp payload holds the promtp
	err := c.ShouldBindJSON(&payload) //* should comes with this type of binded payload 'body'
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.GeminiErrResponse{
			Status:         "Invalid request payload",
			Ok:             false,
			GeminiResponse: "early return;prompt did not reach ai",
		})
		return
	}

	// call gemini caller for fetching res from recieved prompt
	AiResponse, err := g.GeminiAIServiceModel.PromptGemini(payload.Prompt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.GeminiErrResponse{
			Status:         "failed to generate response",
			Ok:             false,
			GeminiResponse: "invalid prompt",
		})
		return
	}

	airespayload := models.GeminiResponse{
		SenderID:   userID,
		GeminiID:   99999, //* just vessel id
		AIResponse: AiResponse,
	}
	// handler invokes method to send payload to the pns retry which redirect to the pns's chan
	g.pns.RetryDenverAIPromptRequestWithTimeout(g.pns, 2*time.Second, &airespayload) // sends to retryWithTimeout -> pns.geminiaires'chan -> publisher-> consumer-> hub-> h.GeminiReschan -> ws -> client's writer -> client
	slog.Info("successfully sent payload to the pns service✅")

	// todos - add db postgres layer + caching layer for caching same responses to avoid excess credit usage for same prompts
	slog.Info("ai response is served successfully✅", "GeminiResponse", AiResponse)

	// now; it is just optional to send concrete response we are sending response aired via websocket connection
	c.JSON(http.StatusOK, utils.GeminiSuccessResponse{
		Status:         "successfully fetched response from the given user prompt🎉",
		Ok:             true,
		GeminiResponse: AiResponse,
	})
}
