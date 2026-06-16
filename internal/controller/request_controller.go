package controller

import (
	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/service"
	"sorint-fleet/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RequestController struct {
	svc *service.RequestService
}

func NewRequestController() *RequestController {
	repo := repository.NewRequestRepository()
	svc := service.NewRequestService(repo)
	return &RequestController{svc: svc}
}

func (rc *RequestController) Create(c *gin.Context) {
	var input dto.CreateRequestDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	r, err := rc.svc.Create(input)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Created(c, r)
}

func (rc *RequestController) Search(c *gin.Context) {
	q := c.Query("q")
	size := 10
	if s := c.Query("size"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			size = v
		}
	}
	hits, err := rc.svc.Search(q, size)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, hits)
}
